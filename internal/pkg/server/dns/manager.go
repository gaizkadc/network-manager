/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
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

func NewManager (consulClient *consul.ConsulClient) (* Manager, derrors.Error){
	return &Manager {
		client: *consulClient,
	}, nil
}

// ListDNSEntries
func (m * Manager) ListDNSEntries (organizationId *grpc_organization_go.OrganizationId) ([]entities.DNSEntry, derrors.Error) {
	serviceList, err := m.client.List(organizationId.OrganizationId)

	if err != nil {
		log.Error().Msg("Unable to retrieve DNS list from the system")
		return nil, derrors.NewGenericError(err.Error())
	}

	entryList := make ([]entities.DNSEntry, len(serviceList))
	for i, n := range serviceList {
		entryList[i].OrganizationId = organizationId.OrganizationId
		entryList[i].Fqdn = n.Service
		entryList[i].Ip = n.Address
	}

	return entryList, nil
}

// AddDNSEntry
func (m * Manager) AddDNSEntry (entry *grpc_network_go.AddDNSEntryRequest) derrors.Error {
	aux := entities.DNSEntryFromGRPC(entry)
	asr := aux.ToConsulAPI()
	err := m.client.Add(asr)

	if err != nil {
		log.Error().Msg("Unable to add DNS entry to the system")
		return derrors.NewGenericError(err.Error())
	}

	return nil
}

// DeleteDNSEntry
func (m * Manager) DeleteDNSEntry (entry *grpc_network_go.DeleteDNSEntryRequest) derrors.Error {
	err := m.client.Delete(entry.Fqdn)

	if err != nil {
		log.Error().Msg("Unable to delete DNS entry from the system")
		return derrors.NewGenericError(err.Error())
	}

	return nil
}