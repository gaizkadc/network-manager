/*
 * Copyright (C)  2019 Nalej - All Rights Reserved
 */

package application

import (
    "context"
    "github.com/nalej/derrors"
    "github.com/nalej/grpc-application-go"
    "github.com/nalej/grpc-network-go"
    "google.golang.org/grpc"
    "time"
)

const (
    ApplicationManagerTimeout = time.Second * 5
)

type Manager struct {
    applicationClient grpc_application_go.ApplicationsClient
}

func NewManager(conn *grpc.ClientConn) *Manager {
    applicationClient := grpc_application_go.NewApplicationsClient(conn)
    return &Manager{applicationClient: applicationClient}
}

func (m *Manager) RegisterInboundServiceProxy(request *grpc_network_go.InboundServiceProxy) derrors.Error {
    // Declare a new service proxy in the system model
    ctx, cancel := context.WithTimeout(context.Background(), ApplicationManagerTimeout)
    defer cancel()
    newProxy := grpc_application_go.ServiceProxy{
        OrganizationId: request.OrganizationId,
        AppInstanceId: request.AppInstanceId,
        ServiceGroupId: request.ServiceGroupId,
        ServiceId: request.ServiceId,
        ClusterId: request.ClusterId,
        Ip: request.Ip,
        ServiceInstanceId: request.ServiceInstanceId,
        Fqdn: request.Fqdn,
        ServiceGroupInstanceId: request.ServiceGroupInstanceId,
    }
    _, err := m.applicationClient.AddZtNetworkProxy(ctx,&newProxy)
    if err != nil {
        return derrors.NewInternalError("impossible to add network proxy",err)
    }

    // Inform clusters about the new available proxy

    return nil
}

func (m *Manager) RegisterOutboundProxy(request *grpc_network_go.OutboundService) derrors.Error {
    return nil
}