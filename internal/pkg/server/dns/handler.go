/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package dns

import (
	"context"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-network-go"
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

func (h * Handler) ListEntries (ctx context.Context, networkID *grpc_network_go.NetworkId) (*grpc_network_go.DNSEntryList, error) {
	panic("list dns entries not implemented yet")
	return nil, nil
}

func (h * Handler) AddDNSEntry (ctx context.Context, entry *grpc_network_go.DNSEntry) (*grpc_common_go.Success, error) {
	log.Debug().Str("networkId", entry.NetworkId).
		Str("fqdn", entry.Fqdn).
		Str("ip", entry.Ip).Msg("add dns entry")
	err := entities.ValidFQDN(entry)
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

func (h * Handler) DeleteDNSEntry (ctx context.Context, entry *grpc_network_go.DNSEntry) (*grpc_common_go.Success, error) {
	log.Debug().Str("networkId", entry.NetworkId).
		Str("fqdn", entry.Fqdn).Msg("delete dns entry")
	err := entities.ValidFQDN(entry)
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