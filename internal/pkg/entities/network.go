/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package entities

import (
	"github.com/nalej/grpc-network-go"
)

type Network struct {
	// OrganizationId with the organization identifier.
	OrganizationId string
	// NetworkId with the ZeroTier network identifier.
	NetworkId string
	// Name assigned to the network.
	NetworkName string
	// Timestamp of the creation of the network.
	CreationTimestamp int64
}

func (n *Network) ToGRPC() *grpc_network_go.Network{
	return &grpc_network_go.Network{
		OrganizationId: n.OrganizationId,
		NetworkId: n.NetworkId,
		Name: n.NetworkName,
		CreationTimestamp: n.CreationTimestamp,
	}
}

