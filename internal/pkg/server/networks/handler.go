/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package networks

import (
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/network-manager/internal/pkg/entities"
	"github.com/nalej/grpc-network-go"
	"github.com/rs/zerolog/log"
	"context"
)

type Handler struct {
	Manager Manager
}

func NewHandler(manager Manager) *Handler{
	return &Handler{manager}
}

// AddNetwork adds a network to the system.
func (h *Handler) AddNetwork (ctx context.Context, addNetworkRequest *grpc_network_go.AddNetworkRequest) (*grpc_network_go.Network, error) {
	log.Debug().Str("request_id", addNetworkRequest.RequestId).
		Str("organizationID", addNetworkRequest.OrganizationId).
		Str("network_id", addNetworkRequest.NetworkId).
		Str("network_name", addNetworkRequest.Name).Msg("add network")
	err := entities.ValidAddNetworkRequest(addNetworkRequest)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}

	network, err := h.Manager.AddNetwork(addNetworkRequest)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	log.Debug().Str("networkID", network.NetworkId).Msg("network has been added")

	return network.ToGRPC(), nil
}



// GetNetwork retrieves the network information.
func (h * Handler) GetNetwork (ctx context.Context, networkID *grpc_network_go.NetworkId) (*grpc_network_go.Network, error){
	panic("get network not implemented yet")

	return nil, nil
}

// DeleteNetwork deletes a network from the system.
func (h * Handler) DeleteNetwork (ctx context.Context, deleteNetworkRequest *grpc_network_go.DeleteNetworkRequest) (*grpc_common_go.Success, error) {

	panic("delete network not implemented yet")

	return nil, nil
}

func (h *Handler) JoinNetwork(ctx context.Context, in *grpc_network_go.NetworkId) (*grpc_common_go.Success, error) {
	panic("join network not implemented yet")
	return nil, nil
}

func (h *Handler) ListNetworks(ctx context.Context, in *grpc_organization_go.OrganizationId) (*grpc_network_go.NetworkList, error) {
	panic("list networks not implemented yet")
	return nil, nil
}