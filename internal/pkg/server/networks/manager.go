/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package networks

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
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/network-manager/internal/pkg/entities"
	"github.com/nalej/network-manager/internal/pkg/utils"
	"github.com/nalej/network-manager/internal/pkg/zt"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"strings"
	"time"
)

const (
	// timeout for queries about network values
	NetworkQueryTimeout = time.Second * 10
	ApplicationManagerTimeout = time.Second * 3
	ZTRangeMin = "192.168.0.1"
	ZTRangeMax = "192.168.0.254"
	ApplicationManagerUpdateRetries = 5
	ApplicationManagerJoinTimeout = time.Second * 10
)

// Manager structure with the remote clients required to manage networks.
type Manager struct {
	//NetworkProvider network.Provider
	OrganizationClient	grpc_organization_go.OrganizationsClient
	ApplicationClient 	grpc_application_go.ApplicationsClient
	AppNetClient      	grpc_application_network_go.ApplicationNetworkClient
	ZTClient           	*zt.ZTClient
	// Connection helper to maintain connections with multiple deployment managers
	connHelper 			*utils.ConnectionsHelper
	// cluster infrastructure client
	clusterInfrastructure grpc_infrastructure_go.ClustersClient
}

// NewManager creates a new manager.
func NewManager(organizationConn *grpc.ClientConn, ztClient *zt.ZTClient, helper *utils.ConnectionsHelper) (*Manager, error) {
	orgClient := grpc_organization_go.NewOrganizationsClient(organizationConn)
	appClient := grpc_application_go.NewApplicationsClient(organizationConn)
	appnetClient := grpc_application_network_go.NewApplicationNetworkClient(organizationConn)
	clusterClient   := grpc_infrastructure_go.NewClustersClient(organizationConn)

	return &Manager{
		OrganizationClient: orgClient,
		ApplicationClient: 	appClient,
		AppNetClient: 		appnetClient,
		ZTClient:           ztClient,
		connHelper: 		helper,
		clusterInfrastructure: clusterClient,
	}, nil
}

// AddNetwork adds a new network to the system.
func (m *Manager) AddNetwork(addNetworkRequest *grpc_network_go.AddNetworkRequest) (*entities.Network, derrors.Error) {

	// Check if organization exists
	_, err := m.OrganizationClient.GetOrganization(context.Background(),
		&grpc_organization_go.OrganizationId{OrganizationId: addNetworkRequest.OrganizationId})
	if err != nil {
		return nil, derrors.NewNotFoundError("invalid organizationID", err)
	}

	// Check if application exists
	_, err = m.ApplicationClient.GetAppInstance(context.Background(),&grpc_application_go.AppInstanceId{
		OrganizationId: addNetworkRequest.OrganizationId, AppInstanceId: addNetworkRequest.AppInstanceId})
	if err != nil {
		return nil, derrors.NewNotFoundError("not found application instance")
	}

	// use zt client to add network
	ztNetwork, err := m.ZTClient.Add(addNetworkRequest.Name, addNetworkRequest.OrganizationId, ZTRangeMin, ZTRangeMax)

	if err != nil {
		return nil, derrors.NewGenericError("Cannot add ZeroTier network", err)
	}

	toAdd := ztNetwork.ToNetwork(addNetworkRequest.OrganizationId)

	// the network generation was correct, add the entry to the system model
	netReq := grpc_application_go.AddAppZtNetworkRequest{
		OrganizationId: addNetworkRequest.OrganizationId,
		AppInstanceId: addNetworkRequest.AppInstanceId,
		NetworkId: toAdd.NetworkId,
		VsaList: addNetworkRequest.Vsa,
	}
	_, err = m.ApplicationClient.AddAppZtNetwork(context.Background(), &netReq)
	if err != nil {
		return nil, derrors.NewUnavailableError("impossible to add zt network to system model", err)
	}

	return &toAdd, nil
}

// DeleteNetwork deletes a network from the system.
func (m *Manager) DeleteNetwork(deleteNetworkRequest *grpc_network_go.DeleteNetworkRequest) derrors.Error {

	// get the network id from the system model
	ztNetwork, err := m.ApplicationClient.GetAppZtNetwork(context.Background(),
		&grpc_application_go.GetAppZtNetworkRequest{
			OrganizationId: deleteNetworkRequest.OrganizationId,
			AppInstanceId: deleteNetworkRequest.AppInstanceId})
	if err != nil {
		return derrors.NewInternalError("impossible to get network id to delete", err)
	}

	// Use zt client to delete network
	err = m.ZTClient.Delete(ztNetwork.NetworkId, deleteNetworkRequest.OrganizationId)
	if err != nil {
		return derrors.NewGenericError("Cannot delete ZeroTier network", err)
	}

	// get ZTMembers
	list, err := m.ApplicationClient.ListAuthorizedZTNetworkMembers(context.Background(), &grpc_application_go.ListAuthorizedZtNetworkMemberRequest{
		OrganizationId: deleteNetworkRequest.OrganizationId,
		AppInstanceId: 	deleteNetworkRequest.AppInstanceId,
		ZtNetworkId: 	ztNetwork.NetworkId,
	})
	if err != nil {
		log.Error().Err(err).Msg("error getting zero tier members, unable to delete them")
	}
	// List returns a list with all the members, and remove removes all the members in a time call one time per ServiceGroupInstanceId, ServiceApplicationInstanceId and ZtNetworkId
	removed := make (map[string]bool, 0)
	for _, member := range list.Members {
		_, found := removed[fmt.Sprintf("%s#%s", member.ServiceApplicationInstanceId, member.ServiceGroupInstanceId)]
		if ! found {
			_, err = m.ApplicationClient.RemoveAuthorizedZtNetworkMember(context.Background(), &grpc_application_go.RemoveAuthorizedZtNetworkMemberRequest{
				OrganizationId:               member.OrganizationId,
				AppInstanceId:                member.AppInstanceId,
				ServiceGroupInstanceId:       member.ServiceGroupInstanceId,
				ServiceApplicationInstanceId: member.ServiceApplicationInstanceId,
				ZtNetworkId:                  member.NetworkId,
			})
			if err != nil {
				log.Error().Str("organizationId", member.OrganizationId).Str("AppInstanceId", member.AppInstanceId).
					Str("ServiceGroupInstanceId", member.ServiceGroupInstanceId).Str("ServiceApplicationInstanceId", member.ServiceApplicationInstanceId).
					Str("ServiceApplicationInstanceId", member.ServiceApplicationInstanceId).
					Err(err).Msg("error removing zero tier members")
			}else{
				removed[fmt.Sprintf("%s#%s", member.ServiceApplicationInstanceId, member.ServiceGroupInstanceId)] = true
			}
		}
	}

	// Delete the network entry
	_, err = m.ApplicationClient.RemoveAppZtNetwork(context.Background(), &grpc_application_go.RemoveAppZtNetworkRequest{
		OrganizationId: deleteNetworkRequest.OrganizationId,
		AppInstanceId: deleteNetworkRequest.AppInstanceId,
	})
	if err != nil {
		return derrors.NewGenericError("cannot delete zt network entry from the system model")
	}

	return nil
}

// GetNetwork gets an existing network from the system.
func (m *Manager) GetNetwork(networkId *grpc_network_go.NetworkId) (*entities.Network, derrors.Error) {

	// Check if organization exists
	_, err := m.OrganizationClient.GetOrganization(context.Background(),
		&grpc_organization_go.OrganizationId{OrganizationId: networkId.OrganizationId})
	if err != nil {
		return nil, derrors.NewNotFoundError("invalid organizationID", err)
	}

	// use zt client to get network
	ztNetwork, err := m.ZTClient.Get(networkId.NetworkId)

	if err != nil {
		return nil, derrors.NewGenericError("Cannot get ZeroTier network", err)
	}

	toReturn := ztNetwork.ToNetwork(networkId.OrganizationId)

	return &toReturn, nil
}

// ListNetworks gets a list of existing networks from an organization.
func (m *Manager) ListNetworks(organizationId *grpc_organization_go.OrganizationId) ([]entities.Network, derrors.Error) {

	// Check if organization exists
	_, err := m.OrganizationClient.GetOrganization(context.Background(),
		&grpc_organization_go.OrganizationId{OrganizationId: organizationId.OrganizationId})
	if err != nil {
		return nil, derrors.NewNotFoundError("invalid organizationID", err)
	}

	// use zt client to get network
	ztNetworkList, err := m.ZTClient.List(organizationId.OrganizationId)
	if err != nil {
		return nil, derrors.NewGenericError("Cannot get ZeroTier network list", err)
	}

	networkList := make([]entities.Network, len(ztNetworkList))

	for i, n := range ztNetworkList {
		networkList[i] = n.ToNetwork(organizationId.OrganizationId)
	}

	return networkList, nil
}

// Authorize a member to join a network
func (m *Manager) AuthorizeMember(authorizeMemberRequest *grpc_network_go.AuthorizeMemberRequest) derrors.Error {
	// Check if there is a network already defined with this id
	ctx, cancel := context.WithTimeout(context.Background(), NetworkQueryTimeout)
	defer cancel()

	net, err := m.ApplicationClient.GetAppZtNetwork(ctx,&grpc_application_go.GetAppZtNetworkRequest{
		OrganizationId: authorizeMemberRequest.OrganizationId,
		AppInstanceId: authorizeMemberRequest.AppInstanceId,
		})
	if err != nil {
		log.Error().Err(err).Msg("impossible to find the requested network")
		return derrors.NewNotFoundError("impossible to find the requested network", err)
	}
	if net.NetworkId != authorizeMemberRequest.NetworkId {
		return derrors.NewFailedPreconditionError(fmt.Sprintf("application network %s does not match %s",
			net.NetworkId, authorizeMemberRequest.NetworkId))
	}

	err = m.ZTClient.Authorize(authorizeMemberRequest.NetworkId, authorizeMemberRequest.MemberId)
	if err != nil {
		return derrors.NewNotFoundError("Unable to authorize member", err)
	}

	// We can assume the client was successfully authorized
	authRequest := grpc_application_go.AddAuthorizedZtNetworkMemberRequest{
		AppInstanceId: authorizeMemberRequest.AppInstanceId,
		OrganizationId: authorizeMemberRequest.OrganizationId,
		ServiceApplicationInstanceId: authorizeMemberRequest.ServiceApplicationInstanceId,
		ServiceGroupInstanceId: authorizeMemberRequest.ServiceGroupInstanceId,
		NetworkId: authorizeMemberRequest.NetworkId,
		MemberId: authorizeMemberRequest.MemberId,
		IsProxy: authorizeMemberRequest.IsProxy,
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), NetworkQueryTimeout)
	defer cancel2()
	_, err = m.ApplicationClient.AddAuthorizedZtNetworkMember(ctx2,&authRequest)
	if err != nil {
		return derrors.NewNotFoundError("impossible to add authorized zt network entry", err)
	}


	return nil
}

// Unauthorize member to join a network
func (m *Manager) UnauthorizeMember(unauthorizeMemberRequest *grpc_network_go.DisauthorizeMemberRequest) derrors.Error {
	// Check if there is already a member
	ctx, cancel := context.WithTimeout(context.Background(), NetworkQueryTimeout)
	defer cancel()

	member, err := m.ApplicationClient.GetAuthorizedZtNetworkMember(ctx, &grpc_application_go.GetAuthorizedZtNetworkMemberRequest{
		OrganizationId: unauthorizeMemberRequest.OrganizationId,
		ServiceGroupInstanceId: unauthorizeMemberRequest.ServiceGroupInstanceId,
		ServiceApplicationInstanceId: unauthorizeMemberRequest.ServiceApplicationInstanceId,
		AppInstanceId: unauthorizeMemberRequest.AppInstanceId,
	})

	if err != nil {
		return derrors.NewNotFoundError("impossible to retrieve zt network member", err)
	}

	log.Debug().Interface("members",member).Msg("members to unauthorize")

	// remove this members
	for _, serviceMember := range member.Members {
		removeRequest := &grpc_application_go.RemoveAuthorizedZtNetworkMemberRequest{
			AppInstanceId: serviceMember.AppInstanceId,
			OrganizationId: serviceMember.OrganizationId,
			ServiceGroupInstanceId: serviceMember.ServiceGroupInstanceId,
			ServiceApplicationInstanceId: serviceMember.ServiceApplicationInstanceId,
			ZtNetworkId: serviceMember.NetworkId,
		}
		log.Debug().Interface("removeAuthorizedZtNetworkMember", removeRequest).Msg("remove authorized network")

		ctx, cancel := context.WithTimeout(context.Background(), NetworkQueryTimeout)
		_, err = m.ApplicationClient.RemoveAuthorizedZtNetworkMember(ctx,removeRequest)
		cancel()
		if err != nil {
			return derrors.NewNotFoundError("impossible to remove authorized zt network member", err)
		}
	}


	// Unauthorize from zt network
	// retrieve me
	for _, serviceMember := range member.Members {
		log.Debug().Str("networkId", serviceMember.NetworkId).Str("memberId",serviceMember.MemberId).
			Msg("unauthorize access to ZT member")
		err = m.ZTClient.Unauthorize(serviceMember.NetworkId, serviceMember.MemberId)
		if err != nil {
			return derrors.NewNotFoundError("impossible to unauthorize member in zt network", err)
		}
	}

	return nil
}

// AuthorizeZTConnection message received from ZT-NALEJ to authorize the member in a ZTNetwork
func (m *Manager) AuthorizeZTConnection(request *grpc_network_go.AuthorizeZTConnectionRequest) derrors.Error{

	ctx, cancel := context.WithTimeout(context.Background(), NetworkQueryTimeout)
	defer cancel()
	// Check if the instance is joined in this zt-network
	list, err := m.AppNetClient.ListZTNetworkConnection(ctx, &grpc_application_network_go.ZTNetworkId{
		OrganizationId:	request.OrganizationId,
		ZtNetworkId: 	request.NetworkId,
	})
	if err != nil {
		return conversions.ToDerror(err)
	}

	found := false
	for _, conn := range list.Connections {
		if conn.AppInstanceId == request.AppInstanceId {
			found = true
		}
	}
	if !found {
		return derrors.NewNotFoundError("instance not found for this zt-network").WithParams(request.AppInstanceId)
	}

	// the instance is allowed for this ZT-network
	// send the authorize request
	err = m.ZTClient.Authorize(request.NetworkId, request.MemberId)
	if err != nil {
		return derrors.NewInternalError("Unable to authorize member", err)
	}

	log.Info().Interface("authorize", request).Msg("Authorization sent to zt-client")


	return nil
}


func (m *Manager) getFQDN(serviceName string, organizationId string, appInstanceId string, outboundName string) string {
	// replace any space
	aux := strings.Replace(strings.ToLower(serviceName), " ", "", -1)

	value := fmt.Sprintf("%s-%s-%s-OUT-%s", aux, organizationId[0:10],	appInstanceId[0:10], outboundName)
	return value
}

// sendUpdateRoute sends a message to update the route to the request pod (if there is an inbound with IP)
func (m *Manager) sendUpdateRoute(request *grpc_network_go.RegisterZTConnectionRequest, inbounds []*grpc_application_network_go.ZTNetworkConnection, allConnected bool) derrors.Error {

	log.Debug().Interface("request", request).Interface("inbounds", inbounds).Msg("sendUpdateRoute")
	// get the best inbound IP (for now,the first one)
	var inboundReg *grpc_application_network_go.ZTNetworkConnection
	for _, inbound := range inbounds {
		if inbound.ZtIp != "" {
			inboundReg = inbound
			break
		}
	}
	if inboundReg == nil {
		log.Info().Interface("request", request).Msg("no ip found for inbound")
		return nil
	}

	return m.sendUpdateRouteToOutbounds(&grpc_network_go.RegisterZTConnectionRequest{
		OrganizationId: inboundReg.OrganizationId,
		AppInstanceId:  inboundReg.AppInstanceId,
		ZtIp:           inboundReg.ZtIp,
		ClusterId:      inboundReg.ClusterId,
		ServiceId:      inboundReg.ServiceId,
		IsInbound:      true,
	}, []*grpc_application_network_go.ZTNetworkConnection{
		{
			OrganizationId: request.OrganizationId,
			ZtNetworkId:    request.NetworkId,
			AppInstanceId:  request.AppInstanceId,
			ServiceId:      request.ServiceId,
			ZtMember:       request.MemberId,
			ZtIp:           request.ZtIp,
			ClusterId:      request.ClusterId,
			Side:           grpc_application_network_go.ConnectionSide_SIDE_OUTBOUND,
		},
	}, allConnected)

}

// getServiceName returns the name of the service
func (m *Manager) getServiceName (organizationId string, appInstanceId string, serviceId string) (string, derrors.Error){
	ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
	defer cancel()
	instance, err := m.ApplicationClient.GetAppInstance(ctx, &grpc_application_go.AppInstanceId{
		OrganizationId: organizationId,
		AppInstanceId:  appInstanceId,
	})
	if err != nil {
		return "", conversions.ToDerror(err)
	}
	// TODO: zt-nalej should send the ServiceGroupID
	serviceName := ""
	for _, group := range instance.Groups{
		for _, service := range group.ServiceInstances {
			if service.ServiceId == serviceId {
				serviceName = service.Name
			}
		}
	}
	if serviceName == "" {
		//log.Warn().Interface("request",request).Msg("ServiceName not found for inbound service")
		return "", derrors.NewNotFoundError("ServiceName not found for inbound service").WithParams(serviceId)
	}
	return serviceName, nil
}

// getServiceGroupId returns the serviceGroupId to which the service belongs
func (m *Manager) getServiceGroupId(organizationID string, applicationId string, serviceID string) (string, derrors.Error){
	ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
	defer cancel()
	instance, err := m.ApplicationClient.GetAppInstance(ctx, &grpc_application_go.AppInstanceId{
		OrganizationId:	organizationID,
		AppInstanceId: 	applicationId,
	})
	if err != nil {
		return "", conversions.ToDerror(err)
	}
	serviceGroupId := ""
	for _, group := range instance.Groups{
		for _, service := range group.ServiceInstances {
			if service.ServiceId == serviceID{
				serviceGroupId = group.ServiceGroupId
			}
		}
	}
	if serviceGroupId == "" {
		log.Warn().Str("appInstanceId",applicationId).Str("serviceId", serviceID).Msg("serviceGroupId not found for inbound service")
		return "", derrors.NewNotFoundError("serviceGroupId not found for inbound service").WithParams(applicationId, serviceID)
	}
	return serviceGroupId, nil

}

// sendUpdateRouteToOutbounds send a message to update the route to all outbound pods if the pod is registered (has IP)
func (m *Manager) sendUpdateRouteToOutbounds(request *grpc_network_go.RegisterZTConnectionRequest, outbounds []*grpc_application_network_go.ZTNetworkConnection, allConnected bool) derrors.Error {

	log.Debug().Interface("request", request).Interface("outbounds", outbounds).Msg("sendUpdateRouteToOutbounds")

	if len(outbounds) == 0{
		// no routes to update, nothing to send
		return nil
	}

	// to update the route, we need:
	// 1) vsa
	//
	//serviceName, nErr := m.getServiceName(request)
	serviceName, nErr := m.getServiceName(outbounds[0].OrganizationId,outbounds[0].AppInstanceId, outbounds[0].ServiceId )
	if nErr != nil {
		return nErr
	}

	// Get available VSA
	ctx2, cancel2 := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
	defer cancel2()
	net, err := m.ApplicationClient.GetAppZtNetwork(ctx2, &grpc_application_go.GetAppZtNetworkRequest{
		OrganizationId: outbounds[0].OrganizationId, AppInstanceId: outbounds[0].AppInstanceId})

	if err != nil {
		return derrors.NewInternalError("impossible to retrieve network data", err)
	}

	// Get connection to get the outbound name
	ctx3, cancel3 := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
	defer cancel3()
	conn, err := m.AppNetClient.GetConnectionByZtNetworkId(ctx3, &grpc_application_network_go.ZTNetworkId {
		OrganizationId: outbounds[0].OrganizationId,
		ZtNetworkId:    outbounds[0].ZtNetworkId,
	})

	if err != nil {
		return derrors.NewInternalError("impossible to retrieve connection instance ", err)
	}
	log.Debug().Interface("conn", conn).Msg("connection")

	//fqdn := m.getFQDN(serviceName, request.OrganizationId, request.AppInstanceId, conn.OutboundName)
	fqdn := m.getFQDN(serviceName, outbounds[0].OrganizationId, outbounds[0].AppInstanceId, conn.OutboundName)
	virtualIP, found := net.VsaList[fqdn]
	if !found {
		log.Warn().Interface("vsa",net.VsaList).Str("fqdn", fqdn).Msg("unknown VSA for FQDN ")
	}
	log.Debug().Str("virtualIP", virtualIP).Str("fqdn", fqdn).Msg("getting virtualIP")

	for _, outbound := range outbounds {
		if outbound.ZtIp != "" && outbound.ClusterId != "" {
			// 1) serviceGroupId
			serviceGroupId, nErr  := m.getServiceGroupId(outbound.OrganizationId, outbound.AppInstanceId, outbound.ServiceId)
			if nErr != nil {
				log.Warn().Interface("request",request).Msg("serviceGroupId not found for inbound service")
			}else {

				// get the ip for the VSA
				newRoute := grpc_deployment_manager_go.ServiceRoute{
					Vsa:            virtualIP,
					OrganizationId: outbound.OrganizationId,
					AppInstanceId:  outbound.AppInstanceId,
					ServiceId:      outbound.ServiceId,
					ServiceGroupId: serviceGroupId,
					RedirectToVpn:  request.ZtIp,
					Drop:           false,
				}
				// and the client
				targetCluster, found := m.connHelper.ClusterReference[outbound.ClusterId]
				if !found {
					return derrors.NewNotFoundError(fmt.Sprintf("impossible to find connection to cluster %s", outbound.ClusterId))
				}
				clusterAddress := fmt.Sprintf("%s:%d", targetCluster.Hostname, utils.APP_CLUSTER_API_PORT)
				conn, err := m.connHelper.GetAppClusterClients().GetConnection(clusterAddress)
				if err != nil {
					return derrors.NewInternalError("impossible to get cluster connection", err)
				}

				sent := false
				for i := 0; i < ApplicationManagerUpdateRetries && !sent; i++ {
					client := grpc_app_cluster_api_go.NewDeploymentManagerClient(conn)
					ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerJoinTimeout)
					log.Debug().Str("clusterId", outbound.ClusterId).Interface("request", newRoute).Msg("set route update")
					_, err = client.SetServiceRoute(ctx, &newRoute)
					cancel()
					if err != nil {
						log.Error().Err(err).Str("ClusterId", outbound.ClusterId).Str("AppInstanceId", outbound.AppInstanceId).
							Str("ServiceId", outbound.ServiceId).Str("ztIp", request.ZtIp).
							Msg("there was an error setting a new route-sending the route to the outbounds")
						time.Sleep(ApplicationManagerTimeout)
					} else {
						sent = true
					}

				}
				// if we can not send the message in ApplicationManagerUpdate retries -> an error must be sent
				if !sent {
					log.Error().Err(err).Str("ClusterId", outbound.ClusterId).Str("AppInstanceId", outbound.AppInstanceId).
						Str("ServiceId", outbound.ServiceId).Str("ztIp", request.ZtIp).Msg("max retries sending the route to the outbounds")
					// I can not return an error, sometimes, when a pod is restarting, the message is sent to the
					// terminating pod, and it returns an error.
					// If I return and error -> I'll never updated the connection status
					//return derrors.NewInternalError("there was an error setting a new route",err)
				}
			}
		}
	}

	if allConnected {
		m.updateConnectionStatus(conn, grpc_application_network_go.ConnectionStatus_ESTABLISHED)
	}

	return nil
}

// updateConnectionStatus updates the status of a connection
func (m *Manager) updateConnectionStatus(connectionInstance *grpc_application_network_go.ConnectionInstance, newStatus grpc_application_network_go.ConnectionStatus) {
	updateConnectionRequest := grpc_application_network_go.UpdateConnectionRequest{
		OrganizationId:   connectionInstance.OrganizationId,
		SourceInstanceId: connectionInstance.SourceInstanceId,
		TargetInstanceId: connectionInstance.TargetInstanceId,
		InboundName:      connectionInstance.InboundName,
		OutboundName:     connectionInstance.OutboundName,
		UpdateStatus:     true,
		Status:           newStatus,
	}
	ctxAppnet, cancelAppnet := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
	defer cancelAppnet()
	_, err := m.AppNetClient.UpdateConnection(ctxAppnet, &updateConnectionRequest)
	if err != nil {
		log.Error().Err(err).
			Interface("connectionInstance", connectionInstance).
			Interface("updateConnectionRequest", updateConnectionRequest).
			Msg("error when updating connectionInstance. Unable to update connection status.")
	}
}

// RegisterZTConnection message received from ZT_NALEJ when getting Zero Tier address
func (m *Manager) RegisterZTConnection(request *grpc_network_go.RegisterZTConnectionRequest) derrors.Error {

	// update conn helper
	_ = m.connHelper.UpdateClusterConnections(request.OrganizationId, m.clusterInfrastructure)

	/* OrganizationId, AppInstanceId, ZtIp, NetworkId, MemberId, IsInbound, ClusterID, serviceID*/
	ctxList, cancelList := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
	defer cancelList()


	log.Debug().Interface("request", request).Msg("update zt-networkConnection")
	ctxUpdate, cancelUpdate := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
	defer cancelUpdate()

	_, err := m.AppNetClient.UpdateZTNetworkConnection(ctxUpdate, &grpc_application_network_go.UpdateZTNetworkConnectionRequest{
		OrganizationId: request.OrganizationId,
		ZtNetworkId: 	request.NetworkId,
		AppInstanceId: 	request.AppInstanceId,
		ServiceId: 		request.ServiceId,
		ClusterId: 		request.ClusterId,
		UpdateZtIp: 	true,
		ZtIp: 			request.ZtIp,
		UpdateZtMember: true,
		ZtMember: 		request.MemberId,
	})
	if err != nil {
		log.Error().Err(err).Interface("request", request).Msg("error updating ztIp in the inbound")
	}

	// list contains the inbound and the outbound appInstanceId
	// get all the services involved in this connection
	list, err := m.AppNetClient.ListZTNetworkConnection(ctxList, &grpc_application_network_go.ZTNetworkId{
		OrganizationId: request.OrganizationId,
		ZtNetworkId: request.NetworkId,
	})
	if err != nil {
		log.Error().Err(err).Msg("error getting zt-networkConnection")
		return conversions.ToDerror(err)
	}
	outboundList := make([]*grpc_application_network_go.ZTNetworkConnection, 0)
	inboundList := make([]*grpc_application_network_go.ZTNetworkConnection, 0)

	allConnected := true

	for _, conn := range list.Connections {
		if conn.Side == grpc_application_network_go.ConnectionSide_SIDE_OUTBOUND{
			outboundList = append(outboundList, conn)
		}else{
			inboundList = append(inboundList, conn)
		}
		if conn.ZtIp == "" {
			log.Debug().Interface("connection", conn).Msg("Connection has no IP!!!!!!!!!!!!!!")
			allConnected = false
		}
	}

	if request.IsInbound {
		// send to all the outbound pods a message to add the new route
		return m.sendUpdateRouteToOutbounds(request, outboundList, allConnected)
	}
	// if the inbound IP is stored -> send a route to add this IP itself
	return m.sendUpdateRoute(request, inboundList, allConnected)
}
