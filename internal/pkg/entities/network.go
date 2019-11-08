/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
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

func (n *Network) ToGRPC() *grpc_network_go.Network {
	return &grpc_network_go.Network{
		OrganizationId:    n.OrganizationId,
		NetworkId:         n.NetworkId,
		Name:              n.NetworkName,
		CreationTimestamp: n.CreationTimestamp,
	}
}
