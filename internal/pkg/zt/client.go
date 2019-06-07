/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package zt

import (
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/dhttp"
	"github.com/nalej/network-manager/internal/pkg/entities"
	"github.com/rs/zerolog/log"
)

// Constants
const (
	controllerPath        = "/controller"
	networkAddPath        = "/controller/network/%s______"
	networkDelPath        = "/controller/network/%s"
	networkPath           = controllerPath + "/network"
	networkDetailPath     = networkPath + "/%s"
	networkAuthMemberPath = networkPath + "/%s" + "/member" + "/%s"
	PeerAddressLength     = 10
)

type ZTClient struct {
	client dhttp.Client
}

func NewZTClient(url string, accessToken string) (*ZTClient, derrors.Error) {

	conf, err := dhttp.NewRestURLConfig(url)

	if err != nil {
		log.Error().Msgf("%s", err.Error())
		log.Error().Err(err).Msg("error creating new ZTClient")
		return nil, err
	}

	conf.Headers = map[string]string{
		"X-ZT1-Auth": accessToken,
	}

	client := dhttp.NewClientSling(conf)

	return &ZTClient{client: client}, nil
}

// Add a ZeroTier network to the controller
//   params:
//     entity The Network to be created
//   returns:
//     The added network.
//     Error, if there is an internal error.
// The entries marked [rw] can be set during creation. From those,
// only "name" is required.
func (ztc *ZTClient) Add(networkName string, organizationId string) (*ZTNetwork, derrors.Error) {

	// Get Controller ZT address, as that's needed to create the proper
	status, err := ztc.GetStatus()

	if err != nil {
		log.Error().Err(err).Msg("error getting status when adding network")
		return nil, err
	}

	// Check if we have an address
	if len(status.Address) != PeerAddressLength {
		return nil, derrors.NewInvalidArgumentError("Invalid address in peer status").WithParams(status)
	}

	path := fmt.Sprintf(networkAddPath, status.Address)

	// Send create network request to controller

	network := &ZTNetwork{}

	entity := &ZTNetwork{
		Name: networkName,
		IpAssignmentPools: []IpAssignmentPool{
            {
                IpRangeStart: "192.168.0.1",
                IpRangeEnd: "192.168.15.254",
            },
            /*
		    {
		        IpRangeStart: "fd00:feed:feed:beef:0000:0000:0000:0000",
		        IpRangeEnd: "fd00:feed:feed:beef:ffff:ffff:ffff:ffff",
	        },
            */
        },
		V4AssignMode: &V4AssignMode{
			Zt:       true,
		},
        /*
		V6AssignMode: &V6AssignMode{
			Zt:       true,
			Rfc4193:  true,
			SixPlane: true,
		},
        */
		Routes: []Route{
		    {Target: "192.168.0.0/20"},
        },
	}

	response := ztc.client.Post(path, entity, network)
	if response.Error != nil {
		return nil, derrors.NewInternalError("Error creating new network", response.Error)
	}
	return network, nil
}

// Delete a ZeroTier network from the controller
//   params:
//     entity The Network to be deleted
//   returns: .
//     Error, if there is an internal error.
//func (ztc *ZTClient) Delete(entity *ZTNetwork) derrors.Error {
func (ztc *ZTClient) Delete(networkId string, organizationId string) derrors.Error {
	// Get Controller ZT address, as that's needed to create the proper
	status, err := ztc.GetStatus()

	if err != nil {
		log.Error().Err(err).Msg("error getting status when deleting network")
		return derrors.NewNotFoundError("Error deleting network", err)
	}

	// Check if we have an address
	if len(status.Address) != PeerAddressLength {
		return derrors.NewInvalidArgumentError("Invalid address in peer status").WithParams(status)
	}

	entity := &entities.Network{
		NetworkId:      networkId,
		OrganizationId: organizationId,
	}

	path := fmt.Sprintf(networkDelPath, entity.NetworkId)
	log.Debug().Msg(path)

	// Send delete network request to controller
	response := ztc.client.Delete(path, entity)
	if response.Error != nil {
		return derrors.NewInternalError("Error deleting network", response.Error)
	}

	return nil
}

// Get ZeroTier network information from the controller
//   params:
//     networkID The ZeroTier network ID to get detailed information for
//   returns:
//     The network.
//     Error, if there is an internal error.
func (ztc *ZTClient) Get(networkID string) (*ZTNetwork, derrors.Error) {
	// Get endpoint
	path := fmt.Sprintf(networkDetailPath, networkID)

	// Send get network request to controller
	network := &ZTNetwork{}
	response := ztc.client.Get(path, network)
	if response.Error != nil {
		return nil, derrors.NewNotFoundError("Error retrieving network", response.Error).WithParams(networkID)
	}

	return response.Result.(*ZTNetwork), nil
}

// Retrieves a list of ZeroTier networks from an existing organization
//   params:
//     organizationID The ZeroTier organization ID to get detailed information for
//   returns:
//     The list of networks.
//     Error, if there is an internal error.
func (ztc *ZTClient) List(organizationID string) ([]ZTNetwork, derrors.Error) {
	// Send get network request to controller
	networkList := make([]string, 0)
	response := ztc.client.Get(networkPath, &networkList)
	if response.Error != nil {
		return nil, derrors.NewNotFoundError("Error retrieving networks", response.Error).WithParams(organizationID)
	}
	//var networks []ZTNetwork
	networks := make([]ZTNetwork, len(networkList))
	for i, n := range networkList {
		//for i := 0; i < len(networkList); i++ {
		converted, err := ztc.Get(n)
		if err != nil {
			log.Error().Msgf("Impossible to get network %s", n)
			return nil, err
		}
		networks[i] = *converted
	}
	return networks, nil
}

// Authorize a member to join a network
//	params:
//		Network ID
//		Member ID
//	returns:
//		Error, if there's one
func (ztc *ZTClient) Authorize(networkId string, memberId string) derrors.Error {
	// Create new authorized member
	member := &ZTMember{
		ID:         memberId,
		Nwid:       networkId,
		Authorized: True(),
	}

	// Form path of the request
	path := fmt.Sprintf(networkAuthMemberPath, networkId, memberId)

	// Create empty member
	output := &ZTMember{}

	// Send request to the controller
	request := ztc.client.Post(path, member, output)
	if request.Error != nil {
		return derrors.NewNotFoundError("Error authorizing member", request.Error).WithParams(networkId, memberId)
	}

	return nil
}


// Unauthorize a member to join a network
//	params:
//		Network ID
//		Member ID
//	returns:
//		Error, if there's one
func (ztc *ZTClient) Unauthorize(networkId string, memberId string) derrors.Error {
	// Create new unauthorized member
	member := &ZTMember{
		ID:         memberId,
		Nwid:       networkId,
		Authorized: False(),
	}

	// Form path of the request
	path := fmt.Sprintf(networkAuthMemberPath, networkId, memberId)

	// Create empty member
	output := &ZTMember{}

	// Send request to the controller
	request := ztc.client.Post(path, member, output)
	if request.Error != nil {
		return derrors.NewNotFoundError("Error unauthorizing member", request.Error).WithParams(networkId, memberId)
	}

	return nil
}