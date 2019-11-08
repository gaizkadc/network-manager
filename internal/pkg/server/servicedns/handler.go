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
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/network-manager/internal/pkg/entities"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

type Handler struct {
	Manager Manager
}

func NewHandler(manager Manager) *Handler {
	return &Handler{manager}
}

func (h *Handler) AddEntry(ctx context.Context, request *grpc_network_go.AddServiceDNSEntryRequest) (*grpc_common_go.Success, error) {
	vErr := entities.ValidAddServiceDNSEntryRequest(request)
	if vErr != nil {
		return nil, conversions.ToGRPCError(vErr)
	}
	log.Debug().Str("organization_id", request.OrganizationId).Str("fqdn", request.Fqdn).Str("ip", request.Ip).Msg("Add service DNS entry")
	err := h.Manager.AddEntry(request)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	return &grpc_common_go.Success{}, nil
}

func (h *Handler) DeleteEntry(ctx context.Context, request *grpc_network_go.DeleteServiceDNSEntryRequest) (*grpc_common_go.Success, error) {
	vErr := entities.ValidDeleteServiceDNSEntryRequest(request)
	if vErr != nil {
		return nil, conversions.ToGRPCError(vErr)
	}
	log.Debug().Str("organization_id", request.OrganizationId).Str("fqdn", request.Fqdn).Msg("Remove service DNS entry")
	err := h.Manager.DeleteEntry(request)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	return &grpc_common_go.Success{}, nil
}

func (h *Handler) ListEntries(ctx context.Context, request *grpc_organization_go.OrganizationId) (*grpc_network_go.ServiceDNSEntryList, error) {
	vErr := entities.ValidOrganizationId(request)
	if vErr != nil {
		return nil, conversions.ToGRPCError(vErr)
	}
	list, err := h.Manager.ListEntries(request)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	return &grpc_network_go.ServiceDNSEntryList{
		Entries: list,
	}, nil
}
