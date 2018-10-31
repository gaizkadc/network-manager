package zt

import (
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/dhttp"
)

type ClientZT struct {
	client dhttp.Client
}


func NewZTClient(url string, accessToken string) (*ClientZT, error) {

	conf, err := dhttp.NewRestURLConfig(url)
	if err != nil {
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
	networkAddPath = "/controller/network/%s______"
	PeerAddressLength = 10
)

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


func (ztc *ClientZT) GetStatus() (*PeerStatus, derrors.Error) {
	result := PeerStatus{}
	response := ztc.client.Get("/status", result)

	if response.Error != nil {
		return nil, response.Error
	}

	return &result, nil
}


