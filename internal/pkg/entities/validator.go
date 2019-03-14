/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package entities

import (
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-organization-go"
)

const (
	emptyOrganizationId = "organization_id cannot be empty"
	emptyNetworkId      = "network_id cannot be empty"
	emptyNetworkName    = "network_name cannot be empty"
	emptyFQDN           = "FQDN cannot be empty"
	emptyMemberId       = "Member ID cannot be empty"
	emptyAppId          = "Application instance ID cannot be empty"
)

func ValidAddNetworkRequest(addNetworkRequest *grpc_network_go.AddNetworkRequest) derrors.Error {
	if addNetworkRequest.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	} else if addNetworkRequest.AppInstanceId == "" {
		return derrors.NewInvalidArgumentError(emptyAppId)
	}
	return nil
}

func ValidOrganizationId(organizationID *grpc_organization_go.OrganizationId) derrors.Error {
	if organizationID.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	return nil
}

func ValidNetworkId(networkId *grpc_network_go.NetworkId) derrors.Error {
	if networkId.NetworkId == "" {
		return derrors.NewInvalidArgumentError(emptyNetworkId)
	}
	return nil
}

func ValidFQDN(fqdn *grpc_network_go.DNSEntry) derrors.Error {
	if fqdn.Fqdn == "" {
		return derrors.NewInvalidArgumentError(emptyFQDN)
	}
	return nil
}

func ValidDeleteNetworkRequest(deleteNetworkRequest *grpc_network_go.DeleteNetworkRequest) derrors.Error {
	if deleteNetworkRequest.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if deleteNetworkRequest.AppInstanceId == "" {
		return derrors.NewInvalidArgumentError(emptyNetworkId)
	}
	return nil
}

func ValidAuthorizeMemberRequest(authMemberRequest *grpc_network_go.AuthorizeMemberRequest) derrors.Error {
	if authMemberRequest.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if authMemberRequest.NetworkId == "" {
		return derrors.NewInvalidArgumentError(emptyNetworkId)
	}

	if authMemberRequest.MemberId == "" {
		return derrors.NewInvalidArgumentError(emptyMemberId)
	}

	return nil
}