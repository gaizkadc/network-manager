/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package networks

import (
	"context"
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-network-go"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/network-manager/internal/pkg/entities"
	"github.com/nalej/network-manager/internal/pkg/zt"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"time"
)

const (
	// timeout for queries about network values
	NetworkQueryTimeout = time.Second * 10
	ZTRangeMin = "192.168.0.1"
	ZTRangeMax = "192.168.15.254"
)

// Manager structure with the remote clients required to manage networks.
type Manager struct {
	//NetworkProvider network.Provider
	OrganizationClient	grpc_organization_go.OrganizationsClient
	ApplicationClient 	grpc_application_go.ApplicationsClient
	AppNetClient      	grpc_application_network_go.ApplicationNetworkClient
	ZTClient           	*zt.ZTClient
}

// NewManager creates a new manager.
func NewManager(organizationConn *grpc.ClientConn, ztClient *zt.ZTClient) (*Manager, error) {
	orgClient := grpc_organization_go.NewOrganizationsClient(organizationConn)
	appClient := grpc_application_go.NewApplicationsClient(organizationConn)
	appnetClient := grpc_application_network_go.NewApplicationNetworkClient(organizationConn)

	return &Manager{
		OrganizationClient: orgClient,
		ApplicationClient: 	appClient,
		AppNetClient: 		appnetClient,
		ZTClient:           ztClient,
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

func (m *Manager) AuthorizeZTConnection(request *grpc_network_go.AuthorizeZTConnectionRequest) derrors.Error{

	ctx, cancel := context.WithTimeout(context.Background(), NetworkQueryTimeout)
	defer cancel()
	// Check if the instance is joined in this zt-network
	list, err := m.AppNetClient.ListZTNetworkConnection(ctx, &grpc_application_network_go.ZTNetworkConnectionId{
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
