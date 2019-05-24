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

func (m * Manager) AddEntry(request *grpc_network_go.AddServiceDNSEntryRequest) derrors.Error {
	err := m.client.AddGenericEntry(request.OrganizationId, request.Fqdn, request.Ip, request.Tags...)
	return err
}

func (m * Manager) DeleteEntry(request *grpc_network_go.DeleteServiceDNSEntryRequest) derrors.Error {
	err := m.client.DeleteGenericEntry(request.OrganizationId, request.Fqdn)
	return err
}

func (m * Manager) ListEntries(request *grpc_organization_go.OrganizationId) ([]*grpc_network_go.ServiceDNSEntry, derrors.Error) {
	panic("implement me")
}