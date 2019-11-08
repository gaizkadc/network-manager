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

package application

import (
	"context"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-network-go"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
)

type Handler struct {
	Manager Manager
}

func NewHandler(manager Manager) *Handler {
	return &Handler{manager}
}

// RegisterInboundServiceProxy operation to update rules based on new service proxy being created.
func (h *Handler) RegisterInboundServiceProxy(ctx context.Context, request *grpc_network_go.InboundServiceProxy) (*grpc_common_go.Success, error) {
	err := h.Manager.RegisterInboundServiceProxy(request)
	if err != nil {
		return nil, err
	}
	return &grpc_common_go.Success{}, nil
}

// RegisterOutboundProxy operation to retrieve existing networking rules.
func (h *Handler) RegisterOutboundProxy(ctx context.Context, request *grpc_network_go.OutboundService) (*grpc_common_go.Success, error) {
	err := h.Manager.RegisterOutboundProxy(request)
	if err != nil {
		return nil, err
	}
	return &grpc_common_go.Success{}, nil
}

// AddConnection adds a new connection between one outbound and one inbound
func (h *Handler) AddConnection(ctx context.Context, addRequest *grpc_application_network_go.AddConnectionRequest) (*grpc_common_go.Success, error) {
	err := h.Manager.AddConnection(addRequest)
	if err != nil {
		return nil, err
	}
	return &grpc_common_go.Success{}, nil
}

// RemoveConnection removes a connection
func (h *Handler) RemoveConnection(ctx context.Context, removeRequest *grpc_application_network_go.RemoveConnectionRequest) (*grpc_common_go.Success, error) {
	err := h.Manager.RemoveConnection(removeRequest)
	if err != nil {
		return nil, err
	}
	return &grpc_common_go.Success{}, nil
}

// RegisterZTConnection operation to indicate that the inbound or outbound  are within the ztNetwork
func (h *Handler) RegisterZTConnection(ctx context.Context, in *grpc_network_go.RegisterZTConnectionRequest) (*grpc_common_go.Success, error) {
	return nil, conversions.ToGRPCError(derrors.NewUnimplementedError("not implemented yet"))
}
