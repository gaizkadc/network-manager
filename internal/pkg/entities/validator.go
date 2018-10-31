/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package entities

import (
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-organization-go"
)

const emptyOrganizationId = "organization_id cannot be empty"
const emptyNetworkId = "network_id cannot be empty"
const emptyNetworkName = "network_name cannot be empty"
const emptyFQDN = "FQDN cannot be empty"

func ValidAddNetworkRequest (networkRequest *grpc_network_go.AddNetworkRequest) derrors.Error {
	if networkRequest.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	} else if networkRequest.Name == "" {
		return derrors.NewInvalidArgumentError(emptyNetworkName)
	}
	return nil
}

func ValidOrganizationId(organizationID *grpc_organization_go.OrganizationId) derrors.Error {
	if organizationID.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	return nil
}

func ValidNetworkId (networkId *grpc_network_go.NetworkId) derrors.Error {
	if networkId.NetworkId == "" {
		return derrors.NewInvalidArgumentError(emptyNetworkId)
	}
	return nil
}

func ValidFQDN (fqdn *grpc_network_go.DNSEntry) derrors.Error {
	if fqdn.Fqdn == "" {
		return derrors.NewInvalidArgumentError(emptyFQDN)
	}
	return nil
}

func ValidDeleteNetworkRequest (networkId *grpc_network_go.NetworkId) derrors.Error {
	if networkId.NetworkId == "" {
		return derrors.NewInvalidArgumentError(emptyNetworkId)
	}
	return nil
}