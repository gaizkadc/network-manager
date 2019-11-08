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

package servicedns

import (
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/network-manager/internal/pkg/consul"
)

type Manager struct {
	client consul.ConsulClient
}

func NewManager(consulClient *consul.ConsulClient) Manager {
	return Manager{
		client: *consulClient,
	}
}

func (m *Manager) AddEntry(request *grpc_network_go.AddServiceDNSEntryRequest) derrors.Error {
	err := m.client.AddGenericEntry(request.OrganizationId, request.Fqdn, request.Ip, request.Tags...)
	return err
}

func (m *Manager) DeleteEntry(request *grpc_network_go.DeleteServiceDNSEntryRequest) derrors.Error {
	err := m.client.DeleteGenericEntry(request.OrganizationId, request.Fqdn)
	return err
}

func (m *Manager) ListEntries(request *grpc_organization_go.OrganizationId) ([]*grpc_network_go.ServiceDNSEntry, derrors.Error) {
	panic("implement me")
}
