/*
 * Copyright (C)  2019 Nalej - All Rights Reserved
 */

package application

import (
    "context"
    "fmt"
    "github.com/nalej/grpc-infrastructure-go"
    "github.com/nalej/grpc-app-cluster-api-go"
    "github.com/nalej/network-manager/internal/pkg/utils"
    "github.com/nalej/derrors"
    "github.com/nalej/grpc-application-go"
    "github.com/nalej/grpc-deployment-manager-go"
    "github.com/nalej/grpc-network-go"
    "github.com/rs/zerolog/log"
    "google.golang.org/grpc"
    "time"
)

const (
    ApplicationManagerTimeout = time.Second * 5
)

type Manager struct {
    applicationClient grpc_application_go.ApplicationsClient
    // cluster infrastructure client
    clusterInfrastructure grpc_infrastructure_go.ClustersClient
    // Connection helper to maintain connections with multiple deployment managers
    connHelper *utils.ConnectionsHelper
}

func NewManager(conn *grpc.ClientConn, connHelper *utils.ConnectionsHelper) *Manager {
    clusterInfrastructure := grpc_infrastructure_go.NewClustersClient(conn)
    applicationClient := grpc_application_go.NewApplicationsClient(conn)
    return &Manager{applicationClient: applicationClient, clusterInfrastructure: clusterInfrastructure,
        connHelper: connHelper}
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
    return m.updateRoutesApplication(request.OrganizationId, request.AppInstanceId, request.Fqdn, request.Ip,
        request.ServiceGroupId, request.ServiceId, false)

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
    // proxies for the service
    for fqdn, proxies := range net.AvailableProxies {
        // if a proxy is already available in the same cluster
        targetProxy, found := proxies.ProxiesPerCluster[targetService.DeployedOnClusterId]
        if !found {
           for _, proxy := range proxies.ProxiesPerCluster {
               targetProxy = proxy
               break
           }
        }
        if targetProxy == nil {
            // no proxies are available yet for this service
            log.Error().Interface("proxies", net).Str("fqdn",fqdn).Msg("not found proxies for the service")
            continue
        }
        routing[fqdn] = targetProxy.List[0].Ip
    }

    // Update the routing table
    for vsa, ip := range routing {
        route := grpc_deployment_manager_go.ServiceRoute{
            OrganizationId: request.OrganizationId,
            AppInstanceId: request.AppInstanceId,
            ServiceGroupId: targetService.ServiceGroupId,
            ServiceId: targetService.ServiceGroupId,
            Vsa: vsa,
            RedirectToVpn: ip,
            Drop: false,
        }

        // Send it.
    }






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

    // build an iterable structure to have the list of clusters this application is deployed into
    clusterSet := make(map[string]bool,0)
    for _, group := range appInstance.Groups {
        for _, serv := range group.ServiceInstances {
            clusterSet[serv.DeployedOnClusterId] = true
        }
    }

    for clusterId, _ := range clusterSet {
        newRoute := grpc_deployment_manager_go.ServiceRoute{
            Vsa: fqdn,
            OrganizationId: organizationId,
            AppInstanceId: appInstanceId,
            ServiceId: serviceId,
            ServiceGroupId: serviceGroupId,
            RedirectToVpn: ip,
            Drop: false,
        }

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

    return nil

}

/*
func (m *Manager) updateRoutesApplication(request *grpc_network_go.InboundServiceProxy) derrors.Error {

    // update the status of cluster connections
    m.connHelper.UpdateClusterConnections(request.OrganizationId, m.clusterInfrastructure)

    if m.connHelper.GetAppClusterClients().NumConnections() == 0 {
        return derrors.NewUnavailableError("no available cluster connections")
    }

    // find what cluster we have to inform about changes
    ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancel()
    appInstance, err := m.applicationClient.GetAppInstance(ctx,
        &grpc_application_go.AppInstanceId{AppInstanceId: request.AppInstanceId, OrganizationId: request.OrganizationId})
    if err != nil {
        return derrors.NewUnavailableError("impossible to find application descriptor", err)
    }

    // build an iterable structure to have the list of clusters this application is deployed into
    clusterSet := make(map[string]bool,0)
    for _, group := range appInstance.Groups {
        for _, serv := range group.ServiceInstances {
            clusterSet[serv.DeployedOnClusterId] = true
        }
    }

    for clusterId, _ := range clusterSet {
        newRoute := grpc_deployment_manager_go.ServiceRoute{
            Vsa:request.Fqdn,
            OrganizationId: request.OrganizationId,
            AppInstanceId: request.AppInstanceId,
            ServiceId: request.ServiceId,
            ServiceGroupId: request.ServiceGroupId,
            RedirectToVpn: request.Ip,Drop: false,
        }

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

    return nil

}
*/