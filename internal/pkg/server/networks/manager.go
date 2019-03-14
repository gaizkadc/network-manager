/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package networks

import (
	"context"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/network-manager/internal/pkg/entities"
	"github.com/nalej/network-manager/internal/pkg/zt"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
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
	// Check if organization exists
	_, err := m.OrganizationClient.GetOrganization(context.Background(),
		&grpc_organization_go.OrganizationId{OrganizationId: deleteNetworkRequest.OrganizationId})
	if err != nil {
		return derrors.NewNotFoundError("invalid organizationID", err)
	}

	// get the entry from the system model
	ztNetwork, err := m.ApplicationClient.GetAppZtNetwork(context.Background(),
		&grpc_application_go.GetAppZtNetworkRequest{
			OrganizationId: deleteNetworkRequest.OrganizationId,
			AppInstanceId: deleteNetworkRequest.AppInstanceId})
	if err != nil {
		return derrors.NewInternalError("impossible to get network id to delete", err)
	}

	// delete from the system model
	req := grpc_application_go.RemoveAppZtNetworkRequest{
		OrganizationId: deleteNetworkRequest.OrganizationId,
		AppInstanceId: deleteNetworkRequest.AppInstanceId}
	_, err = m.ApplicationClient.RemoveAppZtNetwork(context.Background(), &req)
	if err != nil {
		log.Error().Err(err).Msg("impossible to delete zt network from system model")
	}

	// Use zt client to delete network
	err = m.ZTClient.Delete(ztNetwork.NetworkId, deleteNetworkRequest.OrganizationId)
	if err != nil {
		return derrors.NewGenericError("Cannot delete ZeroTier network", err)
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
	// Check if organization exists
	_, err := m.OrganizationClient.GetOrganization(context.Background(),
		&grpc_organization_go.OrganizationId{OrganizationId: authorizeMemberRequest.OrganizationId})
	if err != nil {
		return derrors.NewNotFoundError("invalid organizationID", err)
	}

	err = m.ZTClient.Authorize(authorizeMemberRequest.NetworkId, authorizeMemberRequest.MemberId)
	if err != nil {
		return derrors.NewNotFoundError("Unable to authorize member", err)
	}

	return nil
}
