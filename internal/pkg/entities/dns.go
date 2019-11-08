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
	"github.com/hashicorp/consul/api"
	"github.com/nalej/grpc-network-go"
)

type DNSEntry struct {
	OrganizationId string
	NetworkId      string
	Fqdn           string
	Ip             string
	ServiceName    string
	Tags           []string
}

func DNSEntryFromGRPC(entry *grpc_network_go.AddDNSEntryRequest) DNSEntry {
	return DNSEntry{
		OrganizationId: entry.OrganizationId,
		Fqdn:           entry.Fqdn,
		Ip:             entry.Ip,
		ServiceName:    entry.ServiceName,
		Tags:           entry.Tags,
	}
}

func (e *DNSEntry) ToGRPC() *grpc_network_go.DNSEntry {
	return &grpc_network_go.DNSEntry{
		OrganizationId: e.OrganizationId,
		NetworkId:      e.NetworkId,
		Fqdn:           e.Fqdn,
		Ip:             e.Ip,
		Tags:           e.Tags,
	}
}

func (e *DNSEntry) ToConsulAPI() *api.AgentServiceRegistration {
	return &api.AgentServiceRegistration{
		Kind:    api.ServiceKind(e.OrganizationId),
		Name:    e.Fqdn,
		Address: e.Ip,
		Tags:    e.Tags,
	}
}

func AddDNSRequestToEntry(e *grpc_network_go.AddDNSEntryRequest) *grpc_network_go.DNSEntry {
	return &grpc_network_go.DNSEntry{
		Fqdn:           e.Fqdn,
		Tags:           e.Tags,
		OrganizationId: e.OrganizationId,
		Ip:             e.Ip,
		//NetworkId:
	}
}
