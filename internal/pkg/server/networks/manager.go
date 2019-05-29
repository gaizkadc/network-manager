/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package networks

import (
	"context"
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/network-manager/internal/pkg/entities"
	"github.com/nalej/network-manager/internal/pkg/zt"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"time"
)

const (
	// timeout for queries about network values
	NetworkQueryTimeout = time.Second * 10
)

// Manager structure with the remote clients required to manage networks.
type Manager struct {
	//NetworkProvider network.Provider
	OrganizationClient grpc_organization_go.OrganizationsClient
	ApplicationClient grpc_application_go.ApplicationsClient
	ZTClient           *zt.ZTClient
}

// NewManager creates a new manager.
func NewManager(organizationConn *grpc.ClientConn, url string, accessToken string) (*Manager, error) {
	orgClient := grpc_organization_go.NewOrganizationsClient(organizationConn)
	appClient := grpc_application_go.NewApplicationsClient(organizationConn)

	ztClient, err := zt.NewZTClient(url, accessToken)

	if err != nil {
		log.Error().Err(err).Msgf("impossible to create network for url %s", url)
		return nil, err
	}

	return &Manager{
		OrganizationClient: orgClient,
		ApplicationClient: appClient,
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
	ztNetwork, err := m.ZTClient.Add(addNetworkRequest.Name, addNetworkRequest.OrganizationId)

	if err != nil {
		return nil, derrors.NewGenericError("Cannot add ZeroTier network", err)
	}

	toAdd := ztNetwork.ToNetwork(addNetworkRequest.OrganizationId)

	// the network generation was correct, add the entry to the system model
	netReq := grpc_application_go.AddAppZtNetworkRequest{
		OrganizationId: addNetworkRequest.OrganizationId,
		AppInstanceId: addNetworkRequest.AppInstanceId,
		NetworkId: toAdd.NetworkId,
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

	// Delete all the members of the network

	req := grpc_application_go.RemoveCompleteAppZtNetworkMemberNetRequest{
		OrganizationId: deleteNetworkRequest.OrganizationId,
		AppInstanceId: deleteNetworkRequest.AppInstanceId,
		NetworkId: ztNetwork.NetworkId,
	}
	_, err = m.ApplicationClient.RemoveCompleteAppZtNetworkMemberNet(context.Background(), &req)
	if err != nil {
		log.Error().Err(err).Msg("impossible to delete zt network members from system model")
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

	ztNetwork, err := m.ApplicationClient.GetAppZtNetwork(ctx,&grpc_application_go.GetAppZtNetworkRequest{
		OrganizationId: unauthorizeMemberRequest.OrganizationId,
		AppInstanceId:unauthorizeMemberRequest.AppInstanceId})

	if err != nil {
		return derrors.NewNotFoundError("impossible to retrieve zt network member", err)
	}

	removeRequest := &grpc_application_go.RemoveAuthorizedZtNetworkMemberRequest{
		AppInstanceId: unauthorizeMemberRequest.AppInstanceId,
		OrganizationId: unauthorizeMemberRequest.OrganizationId,
		ServiceGroupInstanceId: unauthorizeMemberRequest.ServiceGroupInstanceId,
		ServiceApplicationInstanceId: unauthorizeMemberRequest.ServiceApplicationInstanceId,
		//IsProxy:
	}
	_, err = m.ApplicationClient.RemoveAuthorizedZtNetworkMember(ctx,removeRequest)
	if err != nil {
		return derrors.NewNotFoundError("impossible to remove authorized zt network member", err)
	}

	err = m.ZTClient.Delete(ztNetwork.NetworkId, unauthorizeMemberRequest.OrganizationId)
	if err != nil {
		return derrors.NewNotFoundError("impossible to unauthorize member in zt network", err)
	}

	return nil
}
