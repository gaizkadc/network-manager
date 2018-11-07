/*
 * Copyright (C)  2018 Nalej - All Rights Reserved
 */

package entities

import (
	"github.com/hashicorp/consul/api"
	"github.com/nalej/grpc-network-go"
)

type DNSEntry struct {
	OrganizationId string
	NetworkId string
	Fqdn string
	Ip string
}

func DNSEntryFromGRPC(entry *grpc_network_go.AddDNSEntryRequest) DNSEntry {
	return DNSEntry{
		OrganizationId: entry.OrganizationId,
		NetworkId: entry.NetworkId,
		Fqdn: entry.Fqdn,
		Ip: entry.Ip,
	}
}

func (e *DNSEntry) ToGRPC() *grpc_network_go.DNSEntry{
	return &grpc_network_go.DNSEntry{
		OrganizationId: e.OrganizationId,
		NetworkId: e.NetworkId,
		Fqdn: e.Fqdn,
		Ip: e.Ip,
	}
}

func (e *DNSEntry) ToConsulAPI () *api.AgentServiceRegistration{
	return &api.AgentServiceRegistration{
		Kind: api.ServiceKind(e.OrganizationId),
		Name: e.Fqdn,
		Address: e.Ip,
	}
}

func AddDNSRequestToEntry (e *grpc_network_go.AddDNSEntryRequest) *grpc_network_go.DNSEntry {
	return &grpc_network_go.DNSEntry{
		OrganizationId: e.OrganizationId,
		NetworkId: e.NetworkId,
		Fqdn: e.Fqdn,
		Ip: e.Ip,
	}
}

func DeleteDNSRequestToEntry (e *grpc_network_go.DeleteDNSEntryRequest) *grpc_network_go.DNSEntry {
	return &grpc_network_go.DNSEntry{
		OrganizationId: e.OrganizationId,
		Fqdn: e.Fqdn,
	}
}