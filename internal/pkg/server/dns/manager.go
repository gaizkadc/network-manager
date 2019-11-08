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

package dns

import (
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/network-manager/internal/pkg/consul"
	"github.com/nalej/network-manager/internal/pkg/entities"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	client consul.ConsulClient
}

func NewManager(consulClient *consul.ConsulClient) (*Manager, derrors.Error) {
	return &Manager{
		client: *consulClient,
	}, nil
}

// AddDNSEntry
func (m *Manager) AddDNSEntry(entry *grpc_network_go.AddDNSEntryRequest) derrors.Error {
	log.Debug().Interface("request", entry).Msg("added DNS entry")

	err := m.client.Add(entry.ServiceName, entry.Fqdn, entry.Ip, entry.Tags)

	if err != nil {
		log.Error().Msg("Unable to add DNS entry to the system")
		return derrors.NewGenericError(err.Error())
	}

	return nil
}

// DeleteDNSEntry
func (m *Manager) DeleteDNSEntry(entry *grpc_network_go.DeleteDNSEntryRequest) derrors.Error {
	log.Debug().Interface("request", entry).Msg("delete DNS entry")
	err := m.client.Delete(entry.ServiceName, entry.Tags)

	if err != nil {
		log.Error().Msg("Unable to delete DNS entry from the system")
		return derrors.NewGenericError(err.Error())
	}

	return nil
}

// ListDNSEntries
func (m *Manager) ListDNSEntries(organizationId *grpc_organization_go.OrganizationId) ([]entities.DNSEntry, derrors.Error) {
	serviceList, err := m.client.List(organizationId.OrganizationId)

	if err != nil {
		log.Error().Msg("Unable to retrieve DNS list from the system")
		return nil, derrors.NewGenericError(err.Error())
	}

	entryList := make([]entities.DNSEntry, len(serviceList))
	for i, n := range serviceList {
		entryList[i].OrganizationId = organizationId.OrganizationId
		entryList[i].Fqdn = n.Service
		entryList[i].Ip = n.Address
	}

	return entryList, nil
}
