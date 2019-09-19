/*
 * Copyright (C)  2019 Nalej - All Rights Reserved
 */

package application

import (
    "context"
    "github.com/nalej/grpc-application-network-go"
    "github.com/nalej/grpc-common-go"
    "github.com/nalej/grpc-network-go"
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
func (h *Handler) AddConnection(ctx context.Context, addRequest *grpc_application_network_go.AddConnectionRequest) (*grpc_common_go.Success, error){
    err := h.Manager.AddConnection(addRequest)
    if err != nil {
        return nil, err
    }
    return &grpc_common_go.Success{}, nil
}
// RemoveConnection removes a connection
func (h *Handler) RemoveConnection(ctx context.Context, removeRequest *grpc_application_network_go.RemoveConnectionRequest) (*grpc_common_go.Success, error){
    err := h.Manager.RemoveConnection(removeRequest)
    if err != nil {
        return nil, err
    }
        return &grpc_common_go.Success{}, nil
}