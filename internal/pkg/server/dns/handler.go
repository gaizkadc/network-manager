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
	"context"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/network-manager/internal/pkg/entities"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	Manager Manager
}

func NewHandler(manager Manager) *Handler {
	return &Handler{manager}
}

func (h *Handler) AddDNSEntry(ctx context.Context, entry *grpc_network_go.AddDNSEntryRequest) (*grpc_common_go.Success, error) {
	log.Debug().Interface("request", entry).Msg("add dns entry")

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

func (h *Handler) DeleteDNSEntry(ctx context.Context, entry *grpc_network_go.DeleteDNSEntryRequest) (*grpc_common_go.Success, error) {
	log.Debug().Interface("request", entry).Msg("delete dns entry")

	err := h.Manager.DeleteDNSEntry(entry)
	if err != nil {
		log.Error().Msgf("Unable to delete DNS entry: %s", err.Error())
		return nil, conversions.ToGRPCError(err)
	}
	return &grpc_common_go.Success{}, nil
}

func (h *Handler) ListEntries(ctx context.Context, organizationID *grpc_organization_go.OrganizationId) (*grpc_network_go.DNSEntryList, error) {
	log.Debug().Interface("request", organizationID).Msg("list dns entries")

	err := entities.ValidOrganizationId(organizationID)
	if err != nil {
		log.Error().Msg("Unable to retrieve network list from the system")
		return nil, conversions.ToGRPCError(err)
	}

	entryList, err := h.Manager.ListDNSEntries(organizationID)

	if err != nil {
		log.Error().Msg("Unable to retrieve DNS list from the system")
		return nil, derrors.NewGenericError(err.Error())
	}

	foundEntries := make([]*grpc_network_go.DNSEntry, len(entryList))
	for i, n := range entryList {
		foundEntries[i] = n.ToGRPC()
	}

	grpcEntryList := grpc_network_go.DNSEntryList{DnsEntries: foundEntries}

	return &grpcEntryList, nil
}
