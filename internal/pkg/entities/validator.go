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
	emptyIp             = "Service IP cannot be empty"
	emptyAppInstanceId  = "app_instance_id cannot be empty"
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

func ValidAddServiceDNSEntryRequest(request *grpc_network_go.AddServiceDNSEntryRequest) derrors.Error {
	if request.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if request.Fqdn == "" {
		return derrors.NewInvalidArgumentError(emptyFQDN)
	}
	if request.Ip == "" {
		return derrors.NewInvalidArgumentError(emptyIp)
	}
	return nil
}

func ValidDeleteServiceDNSEntryRequest(request *grpc_network_go.DeleteServiceDNSEntryRequest) derrors.Error {
	if request.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if request.Fqdn == "" {
		return derrors.NewInvalidArgumentError(emptyFQDN)
	}
	return nil
}

func ValidAuthorizeZTConnectionRequest(request *grpc_network_go.AuthorizeZTConnectionRequest) derrors.Error {
	if request.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if request.AppInstanceId == "" {
		return derrors.NewInvalidArgumentError(emptyAppInstanceId)
	}
	if request.NetworkId == "" {
		return derrors.NewInvalidArgumentError(emptyNetworkId)
	}
	if request.MemberId == "" {
		return derrors.NewInvalidArgumentError(emptyMemberId)
	}

	return nil
}
