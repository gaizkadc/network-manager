/*
 * Copyright (C)  2018 Nalej - All Rights Reserved
 */

package entities

import (
	"github.com/hashicorp/consul/api"
	"github.com/nalej/grpc-network-go"
)

type DNSEntry struct {
	NetworkId string
	Fqdn string
	Ip string
}

func DNSEntryFromGRPC(entry *grpc_network_go.DNSEntry) DNSEntry {
	return DNSEntry{NetworkId: entry.NetworkId, Fqdn: entry.Fqdn, Ip: entry.Ip}
}

func (e *DNSEntry) ToGRPC() *grpc_network_go.DNSEntry{
	return &grpc_network_go.DNSEntry{
		NetworkId: e.NetworkId,
		Fqdn: e.Fqdn,
		Ip: e.Ip,
	}
}

func (e *DNSEntry) ToConsulAPI () *api.AgentServiceRegistration{
	return &api.AgentServiceRegistration{
		Name: e.Fqdn,
		Address: e.Ip,
	}
}