/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package networks

import (
	"context"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/network-manager/internal/pkg/entities"
	"github.com/nalej/network-manager/internal/pkg/zt"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// Manager structure with the remote clients required to manage networks.
type Manager struct {
	//NetworkProvider network.Provider
	OrganizationClient grpc_organization_go.OrganizationsClient
	ZTClient *zt.ClientZT
}

// NewManager creates a new manager.
func NewManager (organizationConn *grpc.ClientConn, url string, accessToken string) (*Manager,error){
	orgClient := grpc_organization_go.NewOrganizationsClient(organizationConn)
	ztClient, err := zt.NewZTClient(url, accessToken)

	if err != nil {
		log.Error().Err(err).Msgf("impossible to create network for url %s", url)
		return nil,err
	}

	return &Manager {
		OrganizationClient: orgClient,
		ZTClient: ztClient,
	}, nil
}

// AddNetwork adds a new network to the system.
func (m * Manager) AddNetwork(addNetworkRequest *grpc_network_go.AddNetworkRequest) (*entities.Network, derrors.Error) {

	// Check if organization exists
	_, err := m.OrganizationClient.GetOrganization(context.Background(),
		&grpc_organization_go.OrganizationId{OrganizationId:  addNetworkRequest.OrganizationId,})
	if err != nil {
		return nil, derrors.NewNotFoundError("invalid organizationID", err)
	}

	// use zt client to add network
	ztNetwork, err := m.ZTClient.Add(&zt.ZTNetwork{Name: addNetworkRequest.Name})

	if err != nil {
		return nil, derrors.NewGenericError("Cannot add ZeroTier network", err)
	}

	 toAdd := ztNetwork.ToNetwork(addNetworkRequest.OrganizationId)

	return &toAdd, nil
}

// GetNetwork gets an existing network from the system.
func (m * Manager) GetNetwork(networkId *grpc_network_go.NetworkId) (*entities.Network, derrors.Error) {

	// Check if organization exists
	_, err := m.OrganizationClient.GetOrganization(context.Background(),
		&grpc_organization_go.OrganizationId{OrganizationId:  networkId.OrganizationId,})
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

// DeleteNetwork deletes a network from the system.
func (m * Manager) DeleteNetwork(deleteNetworkRequest *grpc_network_go.DeleteNetworkRequest) derrors.Error {
	// Check if organization exists
	_, err := m.OrganizationClient.GetOrganization(context.Background(),
		&grpc_organization_go.OrganizationId{OrganizationId:  deleteNetworkRequest.OrganizationId,})
	if err != nil {
		return derrors.NewNotFoundError("invalid organizationID", err)
	}

	// retrieve network to be deleted
	toBeDeleted, err := m.ZTClient.Get(deleteNetworkRequest.NetworkId)
	if err != nil {
		return derrors.NewNotFoundError("cannot retrieve network to be deleted", err)
	}

	// Use zt client to delete network
	err = m.ZTClient.Delete(toBeDeleted)
	if err != nil {
		return derrors.NewGenericError("Cannot delete ZeroTier network", err)
	}

	return nil
}