/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package dns

import (
	"context"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/network-manager/internal/pkg/entities"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	Manager Manager
}

func NewHandler(manager Manager) *Handler{
	return &Handler{manager}
}

func (h * Handler) ListEntries (ctx context.Context, organizationID *grpc_organization_go.OrganizationId) (*grpc_network_go.DNSEntryList, error) {
	log.Debug().Str("organizationId", organizationID.OrganizationId).Msg("list dns entries")

	entryList, err := h.Manager.ListDNSEntries(organizationID)

	if err != nil {
		log.Error().Msg("Unable to retrieve DNS list from the system")
		return nil, derrors.NewGenericError(err.Error())
	}

	foundEntries := make ([]*grpc_network_go.DNSEntry, len(entryList))
	for i, n := range entryList {
		foundEntries[i] = n.ToGRPC()
	}

	grpcEntryList := grpc_network_go.DNSEntryList{DnsEntries: foundEntries}

	return &grpcEntryList, nil
}

func (h * Handler) AddDNSEntry (ctx context.Context, entry *grpc_network_go.AddDNSEntryRequest) (*grpc_common_go.Success, error) {
	log.Debug().Str("organizationId", entry.OrganizationId).
		Str("networkId", entry.NetworkId).
		Str("fqdn", entry.Fqdn).
		Str("ip", entry.Ip).Msg("add dns entry")

	aux := entities.AddDNSRequestToEntry(entry)
	err := entities.ValidFQDN(aux)
	if err != nil {
		log.Error().Msgf("Invalid FQDN: %s", err.Error())
		return nil, conversions.ToGRPCError(err)
	}

	err = h.Manager.AddDNSEntry(entry)
	if err != nil {
		log.Error().Msgf("Unable to add DNS entry: %s", err.Error())
		return nil, conversions.ToGRPCError(err)
	}
	return &grpc_common_go.Success{}, nil
}

func (h * Handler) DeleteDNSEntry (ctx context.Context, entry *grpc_network_go.DeleteDNSEntryRequest) (*grpc_common_go.Success, error) {
	log.Debug().Str("organizationId", entry.OrganizationId).
		Str("fqdn", entry.Fqdn).Msg("delete dns entry")

	aux := entities.DeleteDNSRequestToEntry(entry)
	err := entities.ValidFQDN(aux)
	if err != nil {
		log.Error().Msgf("Invalid FQDN: %s", err.Error())
		return nil, conversions.ToGRPCError(err)
	}

	err = h.Manager.DeleteDNSEntry(entry)
	if err != nil {
		log.Error().Msgf("Unable to delete DNS entry: %s", err.Error())
		return nil, conversions.ToGRPCError(err)
	}
	return &grpc_common_go.Success{}, nil
}