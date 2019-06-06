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
	NetworkId      string
	Fqdn           string
	Ip             string
	ServiceName    string
	Tags		   []string
}

func DNSEntryFromGRPC(entry *grpc_network_go.AddDNSEntryRequest) DNSEntry {
	return DNSEntry{
		OrganizationId: entry.OrganizationId,
		Fqdn:           entry.Fqdn,
		Ip:             entry.Ip,
		ServiceName:	entry.ServiceName,
		Tags:			entry.Tags,
	}
}

func (e *DNSEntry) ToGRPC() *grpc_network_go.DNSEntry {
	return &grpc_network_go.DNSEntry{
		OrganizationId: e.OrganizationId,
		NetworkId:      e.NetworkId,
		Fqdn:           e.Fqdn,
		Ip:             e.Ip,
		Tags:			e.Tags,
	}
}

func (e *DNSEntry) ToConsulAPI() *api.AgentServiceRegistration {
	return &api.AgentServiceRegistration{
		Kind:    api.ServiceKind(e.OrganizationId),
		Name:    e.Fqdn,
		Address: e.Ip,
		Tags: e.Tags,

	}
}

func AddDNSRequestToEntry(e *grpc_network_go.AddDNSEntryRequest) *grpc_network_go.DNSEntry {
	return &grpc_network_go.DNSEntry{
		Fqdn: e.Fqdn,
		Tags: e.Tags,
		OrganizationId: e.OrganizationId,
		Ip: e.Ip,
		//NetworkId:
	}
}
