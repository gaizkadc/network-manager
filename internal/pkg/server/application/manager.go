/*
 * Copyright (C)  2019 Nalej - All Rights Reserved
 */

package application

import (
    "context"
    "fmt"
    "github.com/nalej/derrors"
    "github.com/nalej/grpc-app-cluster-api-go"
    "github.com/nalej/grpc-application-go"
    "github.com/nalej/grpc-application-network-go"
    "github.com/nalej/grpc-deployment-manager-go"
    "github.com/nalej/grpc-infrastructure-go"
    "github.com/nalej/grpc-network-go"
    "github.com/nalej/network-manager/internal/pkg/utils"
    "github.com/rs/zerolog/log"
    "google.golang.org/grpc"
    "time"
)

const (
    ApplicationManagerTimeout = time.Second * 3
    // Number of retries to be done when updating routes
    ApplicationManagerUpdateRetries = 5
)

type Manager struct {
    applicationClient grpc_application_go.ApplicationsClient
    // cluster infrastructure client
    clusterInfrastructure grpc_infrastructure_go.ClustersClient
    // Connection helper to maintain connections with multiple deployment managers
    connHelper *utils.ConnectionsHelper
    appNetClient grpc_application_network_go.ApplicationNetworkClient
}

func NewManager(conn *grpc.ClientConn, connHelper *utils.ConnectionsHelper) *Manager {
    clusterInfrastructure   := grpc_infrastructure_go.NewClustersClient(conn)
    applicationClient       := grpc_application_go.NewApplicationsClient(conn)
    appNetClient            := grpc_application_network_go.NewApplicationNetworkClient(conn)
    return &Manager{
        applicationClient: applicationClient,
        clusterInfrastructure: clusterInfrastructure,
        connHelper: connHelper,
        appNetClient: appNetClient}
}

func (m *Manager) RegisterInboundServiceProxy(request *grpc_network_go.InboundServiceProxy) derrors.Error {
    // Declare a new service proxy in the system model
    ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancel()
    newProxy := grpc_application_go.ServiceProxy{
        OrganizationId: request.OrganizationId,
        AppInstanceId: request.AppInstanceId,
        ServiceGroupId: request.ServiceGroupId,
        ServiceId: request.ServiceId,
        ClusterId: request.ClusterId,
        Ip: request.Ip,
        ServiceInstanceId: request.ServiceInstanceId,
        Fqdn: request.Fqdn,
        ServiceGroupInstanceId: request.ServiceGroupInstanceId,
    }
    _, err := m.applicationClient.AddZtNetworkProxy(ctx,&newProxy)
    if err != nil {
        return derrors.NewInternalError("impossible to add network proxy",err)
    }

    // Inform pods about new available entities
    var updateErr error = nil
    for i:=0; i < ApplicationManagerUpdateRetries; i++ {
        err = m.updateRoutesApplication(request.OrganizationId, request.AppInstanceId, request.Fqdn, request.Ip,
            request.ServiceGroupId, request.ServiceId, false)
        updateErr = err
        if err != nil {
            log.Error().Err(err).Msgf("attempt %d updating routes failed", i)
            time.Sleep(ApplicationManagerTimeout)
        } else {
            break
        }
    }

    if updateErr != nil {
        log.Error().Err(err).Msg("there was an error setting a new route after registering inbound")
        return derrors.NewInternalError("there was an error setting a new route after registering inbound",err)
    }
    return nil
}



func (m *Manager) RegisterOutboundProxy(request *grpc_network_go.OutboundService) derrors.Error {

    ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancel()
    appInstance, err := m.applicationClient.GetAppInstance(ctx, &grpc_application_go.AppInstanceId{
        OrganizationId: request.OrganizationId, AppInstanceId: request.AppInstanceId})
    if err != nil {
        return derrors.NewInternalError("impossible to retrieve application instance", err)
    }

    // find service_group and service ids
    var targetService *grpc_application_go.ServiceInstance = nil
    for _, g := range appInstance.Groups {
        if g.ServiceGroupInstanceId == request.ServiceGroupInstanceId {
            for _, serv := range g.ServiceInstances {
                if serv.ServiceInstanceId == request.ServiceInstanceId {
                    targetService = serv
                    break
                }
            }
        }
    }

    if targetService == nil {
        return derrors.NewInternalError("service instance not found in application instance descriptor")

    }

    // find what services the service can access through the net
    allowedServices := m.getServicesIAccess(appInstance, targetService.Name)

    if len(allowedServices) == 0 {
        // no allowed network connections. stop
        return nil
    }

    // Get available VSA
    ctx2, cancel2 := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancel2()
    net, err := m.applicationClient.GetAppZtNetwork(ctx2, &grpc_application_go.GetAppZtNetworkRequest{
        OrganizationId: request.OrganizationId, AppInstanceId: request.AppInstanceId})
    if err != nil {
        return derrors.NewInternalError("impossible to retrieve network data", err)
    }

    // Store the routing table here
    routing := make(map[string]string,0)
    for _, allowed := range allowedServices {
        vsa := utils.GetVSAName(allowed, appInstance.OrganizationId, appInstance.AppInstanceId)
        virtualIP, found := net.VsaList[vsa]
        if !found {
            log.Error().Interface("virtualIP",net.VsaList).Msgf("unknown virtual ip for VSA %s", vsa)
            continue
        }

        proxiesPerCluster, found := net.AvailableProxies[vsa]
        if !found {
            // we have no available proxies yet
            log.Warn().Interface("knownProxies", net.AvailableProxies).
                Str("vsa",vsa).Msg("no available proxies yet for this VSA")
            continue
        }
        // first we pick the local proxy when possible
        proxiesInCluster, found := proxiesPerCluster.ProxiesPerCluster[targetService.DeployedOnClusterId]
        //var targetProxy proxie
        var targetProxy *grpc_application_go.ServiceProxy = nil
        if !found {
            // not found, simply pick the first we find
            for _, proxies := range proxiesPerCluster.ProxiesPerCluster {
                // Now simply select the first entry
                targetProxy = proxies.List[0]
                break
            }
        } else {
            // take the firs available proxy from our cluster
            targetProxy = proxiesInCluster.List[0]
        }


        if targetProxy == nil {
            // no proxies are available yet for this service
            log.Error().Interface("proxies", net).Str("vsa",vsa).Msg("not found proxies for the service")
            continue
        }

        routing[virtualIP] = targetProxy.Ip
    }

    // Update the routing table
    for virtualIP, ip := range routing {
        route := grpc_deployment_manager_go.ServiceRoute {
            OrganizationId: request.OrganizationId,
            AppInstanceId:  request.AppInstanceId,
            ServiceGroupId: targetService.ServiceGroupId,
            ServiceId:      targetService.ServiceId,
            Vsa:            virtualIP,
            RedirectToVpn:  ip,
            Drop:           false,
        }
        targetCluster, found := m.connHelper.ClusterReference[request.ClusterId]
        if !found {
            return derrors.NewNotFoundError(fmt.Sprintf("impossible to find connection to cluster %s", request.ClusterId))
        }
        clusterAddress := fmt.Sprintf("%s:%d", targetCluster.Hostname, utils.APP_CLUSTER_API_PORT)
        conn, err := m.connHelper.GetAppClusterClients().GetConnection(clusterAddress)
        if err != nil {
            return derrors.NewInternalError("impossible to get cluster connection", err)
        }


        for i:=0; i < ApplicationManagerUpdateRetries; i++ {

            client := grpc_app_cluster_api_go.NewDeploymentManagerClient(conn)
            ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
            defer cancel()
            log.Debug().Str("clusterId", request.ClusterId).Interface("request", route).Msg("set route update")
            _, err = client.SetServiceRoute(ctx, &route)
            if err != nil {
                log.Error().Err(err).Msg("there was an error setting a new route after registering outbound")
                time.Sleep(ApplicationManagerTimeout)
                // return derrors.NewInternalError("there was an error setting a new route after registering outbound",err)
            } else {
                break
            }
        }


    }
    return nil

}


func (m *Manager) updateRoutesApplication(organizationId string, appInstanceId string, fqdn string, ip string,
    serviceGroupId string, serviceId string, drop bool) derrors.Error {

    // update the status of cluster connections
    m.connHelper.UpdateClusterConnections(organizationId, m.clusterInfrastructure)

    if m.connHelper.GetAppClusterClients().NumConnections() == 0 {
        return derrors.NewUnavailableError("no available cluster connections")
    }

    // find what cluster we have to inform about changes
    ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancel()
    appInstance, err := m.applicationClient.GetAppInstance(ctx,
        &grpc_application_go.AppInstanceId{AppInstanceId: appInstanceId, OrganizationId: organizationId})
    if err != nil {
        return derrors.NewUnavailableError("impossible to find application descriptor", err)
    }


    // find the service proxy name
    proxyServiceName := ""
    // map to translate service names with service ids
    // service_name -> service_id
    serviceNameDict := make(map[string]string,0)
    // map to get the service group id of  service by its name
    serviceGroupDict := make(map[string]string,0)
    // cluster and list of services to be informed in every cluster
    // service_id -> [cluster_id0,cluster_id1...]
    servicesCluster := make(map[string][]string,0)
    for _, group := range appInstance.Groups {
        for _, service := range group.ServiceInstances {
            if service.ServiceGroupId == serviceGroupId && service.ServiceId == serviceId {
                // fill the name of the proxy service
                proxyServiceName = service.Name
            }
            serviceNameDict[service.Name] = service.ServiceId
            serviceGroupDict[service.Name] = service.ServiceGroupId
            if _, found := servicesCluster[serviceId]; !found {
                servicesCluster[service.ServiceId] = []string{service.DeployedOnClusterId}
            } else {
                servicesCluster[service.ServiceId] = append(servicesCluster[serviceId], service.DeployedOnClusterId)
            }
        }
    }


    // find the list of services that can access this service and their clusters
    // [serviceName0, serviceName1...]
    allowedServices := m.getAllowedServices(appInstance, proxyServiceName)
    if len(allowedServices) == 0 {
        // there are no services allowed to access this service
        log.Debug().Str("targetService", proxyServiceName).Interface("allowedServices", allowedServices).
            Msg("no services are authorized to access this service")
        return nil
    }


    // Get available VSA
    ctx2, cancel2 := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancel2()
    net, err := m.applicationClient.GetAppZtNetwork(ctx2, &grpc_application_go.GetAppZtNetworkRequest{
        OrganizationId: organizationId, AppInstanceId: appInstanceId})
    if err != nil {
        return derrors.NewInternalError("impossible to retrieve network data", err)
    }


    for _, allowedServiceName := range allowedServices {
        allowedServiceId := serviceNameDict[allowedServiceName]
        allowedServiceGroupId := serviceGroupDict[allowedServiceName]

        virtualIP, found := net.VsaList[fqdn]
        if !found {
            log.Warn().Interface("vsa",net.VsaList).Msgf("unknown VSA for FQDN %s", fqdn)
            continue
        }

        clusterIds, _ := servicesCluster[allowedServiceId]

        // get the ip for the VSA
        newRoute := grpc_deployment_manager_go.ServiceRoute{
            Vsa: virtualIP,
            OrganizationId: organizationId,
            AppInstanceId: appInstanceId,
            ServiceId: allowedServiceId,
            ServiceGroupId: allowedServiceGroupId,
            RedirectToVpn: ip,
            Drop: false,
        }

        // for every cluster where this service was deployed
        for _, clusterId := range clusterIds {
            // and the client
            targetCluster, found := m.connHelper.ClusterReference[clusterId]
            if !found {
                return derrors.NewNotFoundError(fmt.Sprintf("impossible to find connection to cluster %s", clusterId))
            }
            clusterAddress := fmt.Sprintf("%s:%d", targetCluster.Hostname, utils.APP_CLUSTER_API_PORT)
            conn, err := m.connHelper.GetAppClusterClients().GetConnection(clusterAddress)
            if err != nil {
                return derrors.NewInternalError("impossible to get cluster connection", err)
            }

            client := grpc_app_cluster_api_go.NewDeploymentManagerClient(conn)
            ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
            defer cancel()
            log.Debug().Str("clusterId", clusterId).Interface("request", newRoute).Msg("set route update")
            _, err = client.SetServiceRoute(ctx, &newRoute)
            if err != nil {
                log.Error().Err(err).Msg("there was an error setting a new route")
                return derrors.NewInternalError("there was an error setting a new route",err)
            }
        }
    }

    return nil
}

// Local function to get all the services that can access me.
// params:
//  appInstance
//  serviceName service to be checked
// return:
//  list with the name of the services that can be accessed
func (m *Manager) getAllowedServices(appInstance *grpc_application_go.AppInstance, serviceName string) []string {
    //list with the name of all the services
    allServices := make(map[string]bool,0)
    for _, g := range appInstance.Groups {
        for _, s := range g.ServiceInstances {
            allServices[s.Name] = true
        }
    }
    
    allowedServices := make([]string,0)
    for _, rule := range appInstance.Rules {
        // this is the service we are looking for and it is opened
        if rule.TargetServiceName == serviceName {
            switch rule.Access {
            case grpc_application_go.PortAccess_ALL_APP_SERVICES:
                // open access, we have permission to access this
                for name, _ := range allServices {
                    allowedServices = append(allowedServices, name)
                }

            case grpc_application_go.PortAccess_APP_SERVICES:
                // this is only granted for the specified list of services, check if we are in
                if rule.AuthServices != nil {
                    for _, grantedServiceName := range rule.AuthServices {
                        allowedServices = append(allowedServices, grantedServiceName)
                    }
                }
            }
        }
    }
    log.Debug().Interface("grantedServices", allowedServices).Str("serviceName",serviceName).
        Msg("the service can be accessed by...")
    return allowedServices
}

// Local function to obtain the list of services a service can access. This operation is needed
// to inform outbound connections about new VSA
// params:
//  appInstance
//  serviceName service to be checked
// return:
//  list of services that can be access
func (m *Manager) getServicesIAccess(appInstance *grpc_application_go.AppInstance, serviceName string) [] string {
    //list with the name of all the services
    allServices := make(map[string]bool,0)
    for _, g := range appInstance.Groups {
        for _, s := range g.ServiceInstances {
            allServices[s.Name] = true
        }
    }

    allowedServices := make([]string,0)
    for _, rule := range appInstance.Rules {
        // this is the service we are looking for and it is opened
       switch rule.Access {
        case grpc_application_go.PortAccess_ALL_APP_SERVICES:
            // open access, we have permission to access this
             allowedServices = append(allowedServices, rule.TargetServiceName)

        case grpc_application_go.PortAccess_APP_SERVICES:
            // this is only granted if we are in the list
            if rule.AuthServices != nil {
                for _, grantedServiceName := range rule.AuthServices {
                    if grantedServiceName == serviceName {
                        allowedServices = append(allowedServices, rule.TargetServiceName)
                    }
                }
            }
        }
    }
    log.Debug().Interface("grantedServices", allowedServices).Str("serviceName",serviceName).
        Msg("this service can access...")
    return allowedServices
}

// AddConnection adds a new connection between one outbound and one inbound
func (m *Manager) AddConnection(addRequest *grpc_application_network_go.AddConnectionRequest) error{

    // TODO: ZT tasks!!
    ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancel()

    _, err := m.appNetClient.AddConnection(ctx, addRequest)

    if err != nil {
        return  err
    }
    return nil
}

// RemoveConnection removes a connection
func (m *Manager) RemoveConnection(removeRequest *grpc_application_network_go.RemoveConnectionRequest) error{

    // TODO: ZT Tasks!!
    ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancel()

    _, err := m.appNetClient.RemoveConnection(ctx, removeRequest)

    if err != nil {
        return  err
    }
    return nil }