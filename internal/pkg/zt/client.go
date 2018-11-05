package zt

import (
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/dhttp"
	"github.com/rs/zerolog/log"
)

type ClientZT struct {
	client dhttp.Client
}


func NewZTClient(url string, accessToken string) (*ClientZT, error) {
	log.Debug().Msgf("connecting to %s", url)

	conf, err := dhttp.NewRestURLConfig(url)

	if err != nil {
		log.Error().Msgf("%s",err.Error())
		log.Error().Err(err).Msg("error creating new ZTClient")
		return nil, err
	}

	conf.Headers = map[string]string{
		"X-ZT1-Auth": accessToken,
	}

	client := dhttp.NewClientSling(conf)

	return  &ClientZT{client: client,}, nil
}

// Constants
const (
	controllerPath = "/controller"
	networkAddPath = "/controller/network/%s______"
	networkDelPath = "/controller/network/%s"
	networkPath = controllerPath + "/network"
	networkDetailPath = networkPath + "/%s"
	PeerAddressLength = 10
)

func (ztc *ClientZT) GetStatus() (*PeerStatus, derrors.Error) {
	result := PeerStatus{}
	response := ztc.client.Get("/status", &result)

	log.Debug().Msgf("show the thing %d",response.Status)
	if response.Error != nil {
		log.Error().Err(response.Error).Msg("error getting status")
		return nil, response.Error
	}

	return &result, nil
}

// Add a ZeroTier network to the controller
//   params:
//     entity The Network to be created
//   returns:
//     The added network.
//     Error, if there is an internal error.
// The entries marked [rw] can be set during creation. From those,
// only "name" is required.
func (ztc *ClientZT) Add(entity *ZTNetwork) (*ZTNetwork, derrors.Error) {
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
	response := ztc.client.Post(path, entity, network)
	if response.Error != nil {
		return nil, derrors.NewInternalError("Error creating new network", response.Error)
	}
	return network, nil
}

// Get ZeroTier network information from the controller
//   params:
//     networkID The ZeroTier network ID to get detailed information for
//   returns:
//     The network.
//     Error, if there is an internal error.
func (ztc *ClientZT) Get(networkID string) (*ZTNetwork, derrors.Error) {
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

// Delete a ZeroTier network from the controller
//   params:
//     entity The Network to be deleted
//   returns: .
//     Error, if there is an internal error.
func (ztc *ClientZT) Delete(entity *ZTNetwork) derrors.Error {
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

	path := fmt.Sprintf(networkDelPath, entity.Nwid)
	log.Debug().Msg(path)

	// Send delete network request to controller
	response := ztc.client.Delete(path, entity)
	if response.Error != nil {
		return derrors.NewInternalError("Error deleting network", response.Error)
	}

	return nil
}

// Retrieves a list of ZeroTier networks from an existing organization
//   params:
//     organizationID The ZeroTier organization ID to get detailed information for
//   returns:
//     The list of networks.
//     Error, if there is an internal error.
func (ztc *ClientZT) List(organizationID string) ([]ZTNetwork, derrors.Error) {

	// Send get network request to controller
	networkList := make ([]string, 0)
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