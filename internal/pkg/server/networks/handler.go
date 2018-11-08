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
	log.Debug().Str("organizationID", addNetworkRequest.OrganizationId).
		Str("networkName", addNetworkRequest.Name).Msg("add network")
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
	log.Debug().Str("organizationID", networkID.OrganizationId).
		Str("networkID", networkID.NetworkId).Msg("get network")
	err := entities.ValidNetworkId(networkID)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}

	network, err := h.Manager.GetNetwork(networkID)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	log.Debug().Str("networkID", network.NetworkId).Msg("get network")

	return network.ToGRPC(), nil
}

// DeleteNetwork deletes a network from the system.
func (h * Handler) DeleteNetwork (ctx context.Context, deleteNetworkRequest *grpc_network_go.DeleteNetworkRequest) (*grpc_common_go.Success, error) {
	log.Debug().Str("organizationID", deleteNetworkRequest.OrganizationId).
		Str("networkID", deleteNetworkRequest.NetworkId).Msg("delete network")
	err := entities.ValidDeleteNetworkRequest(deleteNetworkRequest)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}

	err = h.Manager.DeleteNetwork(deleteNetworkRequest)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	log.Debug().Str("networkID", deleteNetworkRequest.NetworkId).Msg("network deleted")

	return &grpc_common_go.Success{}, nil
}

func (h *Handler) JoinNetwork(ctx context.Context, in *grpc_network_go.NetworkId) (*grpc_common_go.Success, error) {
	panic("join network not implemented yet")
	return nil, nil
}

func (h *Handler) ListNetworks(ctx context.Context, organizationID *grpc_organization_go.OrganizationId) (*grpc_network_go.NetworkList, error) {
	log.Debug().Str("organizationID", organizationID.OrganizationId).Msg("list networks")
	err := entities.ValidOrganizationId(organizationID)
	if err != nil {
		log.Error().Msg("Unable to retrieve network list from the system")
		return nil, conversions.ToGRPCError(err)
	}

	networkList, err := h.Manager.ListNetworks(organizationID)

	foundNetworks := make([]*grpc_network_go.Network,len(networkList))
	for i, n := range networkList {
		foundNetworks[i] = n.ToGRPC()
	}

	grpcNetworkList := grpc_network_go.NetworkList{Networks: foundNetworks}

	return &grpcNetworkList, nil
}

func (h * Handler) AuthorizeMember (ctx context.Context, authorizeMemberRequest *grpc_network_go.AuthorizeMemberRequest) (*grpc_common_go.Success, error) {
	log.Debug().Str("organizationID", authorizeMemberRequest.OrganizationId).
		Str("networkID", authorizeMemberRequest.NetworkId).
		Str("memberID", authorizeMemberRequest.MemberId).Msg("authorize member")

	// Validation
	err := entities.ValidAuthorizeMemberRequest(authorizeMemberRequest)
	if err != nil {
		log.Error().Msg("Unable to validate member authorization request")
		return nil, conversions.ToGRPCError(err)
	}

	// Request
	err = h.Manager.AuthorizeMember(authorizeMemberRequest)
	if err != nil {
		log.Error().Msgf("Unable to authorize member %s", authorizeMemberRequest.MemberId)
		return nil, conversions.ToGRPCError(err)
	}

	return &grpc_common_go.Success{}, nil
}