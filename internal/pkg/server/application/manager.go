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
    "github.com/nalej/grpc-conductor-go"
    "github.com/nalej/grpc-deployment-manager-go"
    "github.com/nalej/grpc-infrastructure-go"
    "github.com/nalej/grpc-network-go"
    "github.com/nalej/grpc-organization-go"
    "github.com/nalej/grpc-utils/pkg/conversions"
    "github.com/nalej/network-manager/internal/pkg/utils"
    "github.com/nalej/network-manager/internal/pkg/zt"
    "github.com/rs/zerolog/log"
    "google.golang.org/grpc"
    "strconv"
    "strings"
    "time"
)

const (
    ApplicationManagerTimeout = time.Second * 3
    // Number of retries to be done when updating routes
    ApplicationManagerUpdateRetries = 5
    ApplicationManagerJoinTimeout = time.Second * 20
    ztInitialRange = 3
    ztFinalRange = 255
)

type Manager struct {
    applicationClient grpc_application_go.ApplicationsClient
    // cluster infrastructure client
    clusterInfrastructure grpc_infrastructure_go.ClustersClient
    // Connection helper to maintain connections with multiple deployment managers
    connHelper *utils.ConnectionsHelper
    appNetClient grpc_application_network_go.ApplicationNetworkClient
    ZTClient           	*zt.ZTClient
}

func NewManager(conn *grpc.ClientConn, connHelper *utils.ConnectionsHelper, ztClient *zt.ZTClient) (*Manager, error) {
    clusterInfrastructure   := grpc_infrastructure_go.NewClustersClient(conn)
    applicationClient       := grpc_application_go.NewApplicationsClient(conn)
    appNetClient            := grpc_application_network_go.NewApplicationNetworkClient(conn)

    return &Manager{
        applicationClient: applicationClient,
        clusterInfrastructure: clusterInfrastructure,
        connHelper: connHelper,
        appNetClient: appNetClient,
        ZTClient:ztClient,
    }, nil
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

// returns the IP range that is going to be used in the new ZT network (rangeMin, rangeMax)
func (m *Manager) getRangeIp (organizationID string, sourceId string, targetId string) (string, string, derrors.Error) {
    log.Debug().Str("organizationID", organizationID).Str("sourceId", sourceId).Str("targetId", targetId).Msg("getRangeIp")
    // get the ipRange for the new ztNetwork
    ctxList, cancelList := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancelList()
    lis, err := m.appNetClient.ListConnections(ctxList, &grpc_organization_go.OrganizationId{
        OrganizationId: organizationID,
    })
    if err != nil {
        log.Error().Err(err).Str("trace", conversions.ToDerror(err).DebugReport()).Msg("error getting connections")
        return "", "", conversions.ToDerror(err)
    }

    ips := make ([]bool, 256)
    // range retrieved x.x.x.x x.x.x.x (192.168.x.1 192.168.x.254)
    for _, conn := range lis.Connections{
        if conn.SourceInstanceId == sourceId || conn.SourceInstanceId == targetId ||
            conn.TargetInstanceId == sourceId || conn.TargetInstanceId == targetId{
                if conn.IpRange != "" {
                    tokens := strings.Split(conn.IpRange, ".") // [x, x, X, x x, x, x, x]
                    if len(tokens) != 7 {
                        log.Error().Str("range", conn.IpRange).Msg("incorrect IP range format")
                        return "", "", derrors.NewInternalError("incorrect IP range format").WithParams(conn.IpRange)
                    }
                    value, convErr := strconv.Atoi(tokens[2])
                    if convErr != nil {
                        log.Error().Str("range", conn.IpRange).Msg("error converting ip range to int")
                        return "", "", derrors.NewInternalError("incorrect IP range format").WithParams(conn.IpRange)
                    }else{
                        ips[value] = true
                    }
                }

        }
    }

    for i:= ztInitialRange; i<ztFinalRange; i++{
        if ! ips[i]  {
            rangeMin := fmt.Sprintf("192.168.%d.1", i)
            rangeMax := fmt.Sprintf("192.168.%d.254", i)
            return rangeMin, rangeMax, nil
        }
    }

    return "", "", derrors.NewInternalError("Free IP Range not found").WithParams(sourceId, targetId)
}

// deployedOnInfo is a struct to keep the service identifier and the cluster where it is deployed on
type deployedOnInfo struct {
    ServiceId string
    ClusterId string
}
// getServiceId returns the service identifier and the cluster identifier where it is deployed on for the service named serviceName
func (m *Manager) getServiceId (instance *grpc_application_go.AppInstance, serviceName string) ([]deployedOnInfo, derrors.Error)   {
    servicesList := make([]deployedOnInfo, 0)

    for _, group := range instance.Groups {
        for _, service := range group.ServiceInstances {
            if service.Name == serviceName {
                servicesList = append(servicesList, deployedOnInfo{service.ServiceId, service.DeployedOnClusterId})
            }
        }
    }
    if len(servicesList) == 0 {
        return servicesList, derrors.NewNotFoundError("service not found in the instance").WithParams(serviceName, instance.AppInstanceId)
    }

    return servicesList, nil
}

// getServiceIdForInbound returns the service identifier that contains the inbound and the cluster identifier where it is deployed on
func (m *Manager) getServiceIdForInbound(instance *grpc_application_go.AppInstance, inbound string) ([]deployedOnInfo, derrors.Error){
    for _, rule := range instance.Rules {
        if rule.Access == grpc_application_go.PortAccess_INBOUND_APPNET &&
            rule.InboundNetInterface ==inbound {
                return m.getServiceId(instance, rule.TargetServiceName)
        }
    }
    return nil, derrors.NewNotFoundError("inbound not found in the instance").WithParams(inbound, instance.AppInstanceId)
}
// getServiceIdForOutbound returns the service identifier that contains the inbound and the cluster identifier where it is deployed on
func (m *Manager) getServiceIdForOutbound(instance *grpc_application_go.AppInstance, outbound string) ([]deployedOnInfo, derrors.Error){
    for _, rule := range instance.Rules {
        if rule.Access == grpc_application_go.PortAccess_OUTBOUND_APPNET &&
            rule.OutboundNetInterface ==outbound {
            return m.getServiceId(instance, rule.TargetServiceName)
        }
    }
    return nil, derrors.NewNotFoundError("outbound not found in the instance").WithParams(outbound, instance.AppInstanceId)
}

func (m *Manager) sendJoin(clusterID string, organizationID string, instanceID string, serviceID string, networkID string, isInbound bool) derrors.Error {
    // update conn helper
    m.connHelper.UpdateClusterConnections(organizationID, m.clusterInfrastructure)

    clusterHostname, exists := m.connHelper.ClusterReference[clusterID]
    if !exists {
        log.Warn().Str("clusterID", clusterID).Msg("impossible to get cluster address")
        return derrors.NewInternalError("impossible to get cluster address").WithParams(clusterID)
    }

    clusterAddress := fmt.Sprintf("%s:%d", clusterHostname.Hostname, utils.APP_CLUSTER_API_PORT)
    log.Debug().Str("clusterAddress", clusterAddress).Msg("getting clusterAddress")
    connTarget, err := m.connHelper.GetAppClusterClients().GetConnection(clusterAddress)
    if err != nil {
        log.Error().Err(err).Msg("impossible to get cluster connection")
        return derrors.NewInternalError("impossible to get cluster connection", err)
    }

    ztRequest := &grpc_deployment_manager_go.JoinZTNetworkRequest{
        OrganizationId: organizationID,
        AppInstanceId: instanceID,
        ServiceId: serviceID,
        IsInbound: isInbound,
        NetworkId: networkID,
    }

    sent := false
    for i := 0; i < ApplicationManagerUpdateRetries && !sent; i++ {
        client := grpc_app_cluster_api_go.NewDeploymentManagerClient(connTarget)
        ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerJoinTimeout)
        defer cancel()
        log.Debug().Str("clusterId", clusterID).Msg("send join ztNetwork")
        _, err = client.JoinZTNetwork(ctx, ztRequest)
        if err != nil {
            log.Error().Err(err).Interface("ztRequest", ztRequest).Str("clusterId", clusterID).Str("clusterAddress", clusterAddress).
                Msg("there was an error sending join message")
            time.Sleep(ApplicationManagerTimeout)
        } else {
            sent = true
        }
    }
    // if we can not send the message in ApplicationManagerUpdate retries -> an error must be sent
    if !sent {
        log.Error().Interface("ztRequest", ztRequest).Msg("max retries sending join message")
        return derrors.NewInternalError("unable to send the join message")
    }

    return nil
}

func (m *Manager) sendLeave(organizationID string, clusterID string, instanceID string, serviceID string, isInbound bool, ztNetworkId string) derrors.Error {
    // update conn helper
    m.connHelper.UpdateClusterConnections(organizationID, m.clusterInfrastructure)

    clusterHostname, exists := m.connHelper.ClusterReference[clusterID]
    if !exists {
        log.Warn().Str("clusterID", clusterID).Msg("impossible to get cluster address")
        return derrors.NewInternalError("impossible to get cluster address").WithParams(clusterID)
    }

    clusterAddress := fmt.Sprintf("%s:%d", clusterHostname.Hostname, utils.APP_CLUSTER_API_PORT)
    connTarget, err := m.connHelper.GetAppClusterClients().GetConnection(clusterAddress)
    if err != nil {
        log.Error().Err(err).Msg("impossible to get cluster connection")
        return derrors.NewInternalError("impossible to get cluster connection", err)
    }

    ztRequest := &grpc_deployment_manager_go.LeaveZTNetworkRequest{
        OrganizationId: organizationID,
        AppInstanceId: instanceID,
        ServiceId: serviceID,
        IsInbound: isInbound,
        NetworkId: ztNetworkId,
    }

    sent := false
    for i := 0; i < ApplicationManagerUpdateRetries && !sent; i++ {
        client := grpc_app_cluster_api_go.NewDeploymentManagerClient(connTarget)
        ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerJoinTimeout)
        defer cancel()
        log.Debug().Str("clusterId", clusterID).Msg("send leave ztNetwork")
        _, err = client.LeaveZTNetwork(ctx, ztRequest)
        if err != nil {
            log.Error().Err(err).Interface("ztRequest", ztRequest).Msg("there was an error sending leave message")
            time.Sleep(ApplicationManagerTimeout)
        } else {
            sent = true
        }
    }
    // if we can not send the message in ApplicationManagerUpdate retries -> an error must be sent
    if !sent {
        log.Error().Interface("ztRequest", ztRequest).Msg("max retries sending leave message")
        return derrors.NewInternalError("unable to send the leave message")
    }

    return nil
}

// AddConnection adds a new connection between one outbound and one inbound
func (m *Manager) AddConnection(addRequest *grpc_application_network_go.AddConnectionRequest) error{

    rangeMin, rangeMax, ipErr := m.getRangeIp(addRequest.OrganizationId, addRequest.SourceInstanceId, addRequest.TargetInstanceId)
    if ipErr != nil {
        return conversions.ToGRPCError(ipErr)
    }
    // addRequest needs IpRange
    addRequest.IpRange = fmt.Sprintf("%s-%s", rangeMin, rangeMax)

    // get the serviceId for inbound in targetInstanceId
    ctxTarget, cancelTarget:= context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancelTarget()
    targetInstance, err := m.applicationClient.GetAppInstance(ctxTarget, &grpc_application_go.AppInstanceId{
        OrganizationId: addRequest.OrganizationId,
        AppInstanceId: addRequest.TargetInstanceId,
    } )
    if err != nil {
        return err
    }
    targets, gErr := m.getServiceIdForInbound(targetInstance, addRequest.InboundName)
    if gErr != nil {
        return conversions.ToGRPCError(gErr)
    }

    // get the serviceId for outbound in sourceInstanceId
    ctxSource, cancelSource:= context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancelSource()
    sourceInstance, err := m.applicationClient.GetAppInstance(ctxSource, &grpc_application_go.AppInstanceId{
        OrganizationId: addRequest.OrganizationId,
        AppInstanceId: addRequest.SourceInstanceId,
    } )
    if err != nil {
        return err
    }
    sources, gErr := m.getServiceIdForOutbound(sourceInstance, addRequest.OutboundName)
    if gErr != nil {
        return conversions.ToGRPCError(gErr)
    }

    // Create the connection
    ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancel()
    conn, err := m.appNetClient.AddConnection(ctx, addRequest)
    if err != nil {
        log.Error().Str("trace", conversions.ToDerror(err).DebugReport()).Msg("error adding connection")
        return  err
    }

    // Create ZTNetwork
    ztNetwork, err := m.ZTClient.Add(conn.ConnectionId, addRequest.OrganizationId, rangeMin, rangeMax)
    if err != nil {
        log.Error().Str("trace", conversions.ToDerror(err).DebugReport()).Msg("error creating ZTNetwork")
        return err
    }
    log.Info().Str("networkId", ztNetwork.ID).Str("ZtName", ztNetwork.Name).Msg("ZT network created!")


    // Update the connection with the ztNerworkId
    ctxUpdate, cancelUpdate := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancelUpdate()
    _, err = m.appNetClient.UpdateConnection(ctxUpdate, &grpc_application_network_go.UpdateConnectionRequest{
        OrganizationId:     addRequest.OrganizationId,
        SourceInstanceId:   addRequest.SourceInstanceId,
        TargetInstanceId:   addRequest.TargetInstanceId,
        InboundName:        addRequest.InboundName,
        OutboundName:       addRequest.OutboundName,
        UpdateZtNetworkId:  true,
        ZtNetworkId:        ztNetwork.ID,
        UpdateIpRange:      true,
        IpRange:            addRequest.IpRange,
    })
    if err != nil {
        log.Error().Str("trace", conversions.ToDerror(err).DebugReport()).Msg("error updating  connection instance")
        return  err
    }

    // -------------------------------------------------------------------------
    // send a message to the inbound and the outbound to join into this network
    // -------------------------------------------------------------------------
    // TODO: for now, the messages are sent synchronous, check if an asynchronous call would be necessary

    for _, source := range sources {
        // Source (outbound)
        log.Debug().Str("clusterID", source.ClusterId).Str("SourceInstanceId", addRequest.SourceInstanceId).
            Str("sourceServiceId", source.ServiceId).Str("networkId", ztNetwork.ID).
            Msg("Sending join ZT Network")

        // add a register in ZTConnection table
        // when the pod ask for authorization, the record is searched in this table
        ctxAdd, cancelAdd := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
        defer cancelAdd()

        log.Debug().Str("OrganizationID", addRequest.OrganizationId).Str("ztNetwork.ID", ztNetwork.ID).
            Str("sourceInstanceID", addRequest.SourceInstanceId).Str("ServiceId", source.ServiceId).
            Msg("ADD ztNetworkConnection")

        _, err := m.appNetClient.AddZTNetworkConnection(ctxAdd, &grpc_application_network_go.ZTNetworkConnection{
            OrganizationId: addRequest.OrganizationId,
            ZtNetworkId:    ztNetwork.ID,
            AppInstanceId:  addRequest.SourceInstanceId,
            Side:           grpc_application_network_go.ConnectionSide_SIDE_OUTBOUND,
            ServiceId:      source.ServiceId,
            ClusterId:      source.ClusterId,

        })
        if err != nil {
            log.Error().Str("OrganizationID", addRequest.OrganizationId).Str("ztNetwork.ID", ztNetwork.ID).
                Str("sourceInstanceID", addRequest.SourceInstanceId).Str("ServiceId", source.ServiceId).
                Msg("error adding ztNetworkConnection")
        }

        sErr := m.sendJoin(source.ClusterId, addRequest.OrganizationId, addRequest.SourceInstanceId, source.ServiceId, ztNetwork.ID, false)

        if sErr != nil {
            log.Error().Str("clusterID", source.ClusterId).Str("SourceInstanceId", addRequest.SourceInstanceId).
                Str("sourceServiceId", source.ServiceId).Str("networkId", ztNetwork.ID).
                Msg("error sending JoinZTNetwork")
        }

    }


    // Target (inbound)
    for _, target := range targets {

        // add a register in ZTConnection table
        // when the pod ask for authorization, the record is searched in this table
        ctxAddT, cancelAddT := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
        defer cancelAddT()
        log.Debug().Str("OrganizationID", addRequest.OrganizationId).Str("ztNetwork.ID", ztNetwork.ID).
            Str("sourceInstanceID", addRequest.TargetInstanceId).Str("ServiceId", target.ServiceId).
            Msg("ADD ztNetworkConnection")
        _, err := m.appNetClient.AddZTNetworkConnection(ctxAddT, &grpc_application_network_go.ZTNetworkConnection{
            OrganizationId: addRequest.OrganizationId,
            ZtNetworkId:    ztNetwork.ID,
            AppInstanceId:  addRequest.TargetInstanceId,
            Side:           grpc_application_network_go.ConnectionSide_SIDE_INBOUND,
            ServiceId:      target.ServiceId,
            ClusterId:      target.ClusterId,
        })
        if err != nil {
            log.Error().Str("OrganizationID", addRequest.OrganizationId).Str("ztNetwork.ID", ztNetwork.ID).
                Str("sourceInstanceID", addRequest.TargetInstanceId).Str("ServiceId", target.ServiceId).
                Msg("error adding ztNetworkConnection")
        }

        sErr := m.sendJoin(target.ClusterId, addRequest.OrganizationId, addRequest.TargetInstanceId, target.ServiceId, ztNetwork.ID, true)
        if sErr != nil {
            log.Error().Str("clusterID", target.ClusterId).Str("TargetInstanceId", addRequest.TargetInstanceId).
                Str("targetServiceId", target.ServiceId).Str("networkId", ztNetwork.ID).
                Msg("error sending JoinZTNetwork")
        }

    }



    return nil
}

// RemoveConnection removes a connection
func (m *Manager) RemoveConnection(removeRequest *grpc_application_network_go.RemoveConnectionRequest) error{

    // get the connection_instance to get zt-network and remove it
    ctxGet, cancelGet := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancelGet()

    conn, err := m.appNetClient.GetConnection(ctxGet, &grpc_application_network_go.ConnectionInstanceId{
        OrganizationId: removeRequest.OrganizationId,
        SourceInstanceId: removeRequest.SourceInstanceId,
        TargetInstanceId: removeRequest.TargetInstanceId,
        InboundName: removeRequest.InboundName,
        OutboundName: removeRequest.OutboundName,
    })
    if err != nil {
        log.Error().Err(err).Msg("error getting the connection instance")
        return err
    }

    // Remove Zero tier network
    if conn.ZtNetworkId != "" {
        log.Debug().Msg("Remove zero tier network")
        // remove ZeroTier network
        delErr := m.ZTClient.Delete(conn.ZtNetworkId, removeRequest.OrganizationId)
        if delErr != nil {
            log.Error().Err(delErr).Str("organizationId", removeRequest.OrganizationId).Msg("error deleting zero tier network")
            return conversions.ToGRPCError(delErr)
        }

        // send a message to zt-nalej (through deployment-manager) to leave the network
        ctxList, cancelList := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
        defer cancelList()
        ztConn, err := m.appNetClient.ListZTNetworkConnection(ctxList, &grpc_application_network_go.ZTNetworkId{
            OrganizationId: removeRequest.OrganizationId,
            ZtNetworkId:    conn.ZtNetworkId,
        })
        if err != nil {
            log.Error().Err(err).Str("organizationId", removeRequest.OrganizationId).Str("ZtNetworkId",  conn.ZtNetworkId).
                Msg("error getting zero tier connections")
        }
        for _, ztConn := range ztConn.Connections{
            isInbound := true
            if ztConn.Side == grpc_application_network_go.ConnectionSide_SIDE_OUTBOUND {
                isInbound = false
            }
            sendErr := m.sendLeave(removeRequest.OrganizationId, ztConn.ClusterId, ztConn.AppInstanceId, ztConn.ServiceId, isInbound, ztConn.ZtNetworkId)
            if sendErr != nil {
                log.Error().Err(sendErr).Str("ClusterId", ztConn.ClusterId).Str("AppInstanceId", ztConn.AppInstanceId).
                    Str("ServiceId", ztConn.ServiceId).Msg("error sending leave ztConnection")
            }
        }

        // Remove ZT-Connections
        ctxRemove, cancelRemove := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
        defer cancelRemove()
        _, err = m.appNetClient.RemoveZTNetworkConnectionByNetworkId(ctxRemove, &grpc_application_network_go.ZTNetworkId{
            OrganizationId: removeRequest.OrganizationId,
            ZtNetworkId: conn.ZtNetworkId,
        })
        if err != nil {
            log.Error().Err(delErr).Str("organizationId", removeRequest.OrganizationId).Str("ztNetworkId", conn.ZtNetworkId).
                Msg("error deleting zero tier connections")
            //return err
        }
    }

    // Remove Connection
    ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancel()

    _, err = m.appNetClient.RemoveConnection(ctx, removeRequest)

    if err != nil {
        return  err
    }
    return nil
}

// getConnections returns the inbound and outbound connections of and instance
func (m *Manager) getConnections(organizationId string, appInstanceId string) (*grpc_application_network_go.ConnectionInstanceList, *grpc_application_network_go.ConnectionInstanceList) {

    appInstance := &grpc_application_go.AppInstanceId{
        OrganizationId: organizationId,
        AppInstanceId:  appInstanceId,
    }
    ctxList, cancelList := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancelList()
    // get ZTConnection
    inboundList, err := m.appNetClient.ListInboundConnections(ctxList, appInstance)
    if err != nil {
        log.Error().Err(err).Msg("Error getting inbound connections")
    }

    ctxList2, cancelList2 := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancelList2()
    // get ZTConnection
    outboundList, err2 := m.appNetClient.ListOutboundConnections(ctxList2, appInstance)
    if err2 != nil {
        log.Error().Err(err).Msg("Error getting outbound connections")
    }

    return inboundList, outboundList
}

// manageConnectionsServiceRunning checks if the service should have the leave the connection
func (m *Manager) manageConnectionsServiceTerminating (instance *grpc_application_go.AppInstance, connection *grpc_application_network_go.ConnectionInstance,
    isInbound bool, service *grpc_conductor_go.ServiceUpdate) {


    var ids []deployedOnInfo
    var idErr derrors.Error
    if isInbound {
        // if the service is the owner of the inbound...
        ids, idErr = m.getServiceIdForInbound(instance, connection.InboundName)
        if idErr != nil {
            log.Error().Err(idErr).Msg("error getting service id")
        }
    }else{
        ids, idErr = m.getServiceIdForOutbound(instance, connection.OutboundName)
        if idErr != nil {
            log.Error().Err(idErr).Msg("error getting service id")
        }
    }
    // check if the serviceId is in the list
    found := false
    for i:=0; i< len(ids) && ! found; i++ {
        if ids[i].ServiceId == service.ServiceId {
            found = true
        }
    }
    if found {

        // getZTConnection to get ZTMember and remove (unauthorized)
        ctxGet, cancelGet := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
        defer cancelGet()
        ztConnection, err := m.appNetClient.ListZTNetworkConnection(ctxGet, &grpc_application_network_go.ZTNetworkId{
            OrganizationId:instance.OrganizationId,
            ZtNetworkId: connection.ZtNetworkId,
        })
        if err != nil {
            log.Error().Err(err).Str("organizationId", instance.OrganizationId).
                Str("networkId", connection.ZtNetworkId).
                Msg("error getting zt-connection unauthorized message can not be sent")
        }else {
            ztMember := ""
            for _, conn := range ztConnection.Connections {
                if conn.AppInstanceId == instance.AppInstanceId && conn.ServiceId == service.ServiceId && conn.ClusterId == service.ClusterId {
                    ztMember = conn.ZtMember
                }
            }
            if ztMember != "" {
                unErr := m.ZTClient.Unauthorize(connection.ZtNetworkId, ztMember)
                if unErr != nil {
                  log.Error().Err(unErr).Msg("error sending unauthorized message")
                }
            } else {
                log.Warn().Msg("ztMember not found, unauthorized message can not be sent")
            }
        }
        ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
        defer cancel()
        _, err = m.appNetClient.RemoveZTNetworkConnection(ctx, &grpc_application_network_go.ZTNetworkConnectionId{
            OrganizationId: service.OrganizationId,
            ZtNetworkId:    connection.ZtNetworkId,
            AppInstanceId:  service.ApplicationInstanceId,
            ServiceId:      service.ServiceId,
            ClusterId:      service.ClusterId,
        })
        if err != nil {
            log.Error().Err(err).Msg("error removing ZTConnection")
        }

        // TODO: Update connection Status (wait until SERVICE_TERMINATED is sent)
    } else {
        log.Debug().Msg("no service id found")
    }
}

// manageConnectionsServiceRunning checks if the service should have the connection and it is, send a join message to the pod
func (m *Manager) manageConnectionsServiceRunning (instance *grpc_application_go.AppInstance, connection *grpc_application_network_go.ConnectionInstance,
    isInbound bool, service *grpc_conductor_go.ServiceUpdate) {

    var ids []deployedOnInfo
    var idErr derrors.Error
    if isInbound {
        // if the service is the owner of the inbound...
        ids, idErr = m.getServiceIdForInbound(instance, connection.InboundName)
        if idErr != nil {
            log.Error().Err(idErr).Msg("error getting service id")
        }
    }else{
        ids, idErr = m.getServiceIdForOutbound(instance, connection.OutboundName)
        if idErr != nil {
            log.Error().Err(idErr).Msg("error getting service id")
        }
    }
    // check if the serviceId is in the list
    found := false
    for i:=0; i< len(ids) && ! found; i++ {
        if ids[i].ServiceId == service.ServiceId {
            found = true
        }
    }
    if found {
        log.Debug().Str("cluster", service.ClusterId).Msg("manageConnectionsServiceRunning - sending join")
        sendErr := m.sendJoin(service.ClusterId, service.OrganizationId, service.ApplicationInstanceId, service.ServiceId, connection.ZtNetworkId, isInbound)
        if sendErr != nil {
            log.Error().Err(sendErr).Msg("error sending Join Message to the pod")
        }else{
            ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
            defer cancel()
            var side grpc_application_network_go.ConnectionSide
            if isInbound {
                side = grpc_application_network_go.ConnectionSide_SIDE_INBOUND
            }else{
                side = grpc_application_network_go.ConnectionSide_SIDE_OUTBOUND
            }
            m.appNetClient.AddZTNetworkConnection(ctx, &grpc_application_network_go.ZTNetworkConnection{
                OrganizationId: service.OrganizationId,
                ZtNetworkId: connection.ZtNetworkId,
                AppInstanceId: service.ApplicationInstanceId,
                ServiceId: service.ServiceId,
                ClusterId: service.ClusterId,
                Side: side,
            })
        }

        // TODO: Update connection Status (wait until SERVICE_TERMINATED is sent)
    }
}

// ManageConnections receives DeploymentServiceUpdateRequest messages from the bus and manage
// connections depending of the service updated (if it has or no connections, if it is added or removed, etc)
func (m *Manager) ManageConnections (request *grpc_conductor_go.DeploymentServiceUpdateRequest) derrors.Error{

    for _, service := range request.List {
        log.Info().Str("serviceId", service.ServiceId).Str("clusterID", service.ClusterId).Str("status", service.Status.String()).Msg("Service updated")
        // get connections to see if the services has or not one of them
        inConn, outConn := m.getConnections(service.OrganizationId, service.ApplicationInstanceId)

        log.Debug().Interface("inbound", inConn).Msg("inbound connections")
        log.Debug().Interface("outbound", outConn).Msg("outbound connections")

        // get the instance to has the relation between services and inbound/outbound interfaces
        ctxGet, cancelGet := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
        instance, err := m.applicationClient.GetAppInstance(ctxGet, &grpc_application_go.AppInstanceId{
            OrganizationId: request.OrganizationId,
            AppInstanceId: service.ApplicationInstanceId,
        })
        if err != nil {
            log.Error().Err(err).Msg("error getting instance")
            cancelGet()
            return conversions.ToDerror(err)
        }
        cancelGet()
        if service.Status == grpc_application_go.ServiceStatus_SERVICE_RUNNING {
            for _, inbound := range inConn.Connections {
                m.manageConnectionsServiceRunning(instance, inbound, true, service)
            }
            for _, outbound := range outConn.Connections {
                m.manageConnectionsServiceRunning(instance, outbound, false, service)
            }
         }else if service.Status == grpc_application_go.ServiceStatus_SERVICE_TERMINATED ||service.Status == grpc_application_go.ServiceStatus_SERVICE_TERMINATING {
             for _, inbound := range inConn.Connections {
                 m.manageConnectionsServiceTerminating(instance, inbound, true, service)
             }
             for _, outbound := range outConn.Connections {
                 m.manageConnectionsServiceTerminating(instance, outbound, false, service)
             }
        }else{
            log.Debug().Str("Status", service.Status.String()).Msg("Nothing to do?")
        }
    }

    return nil
}