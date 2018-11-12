/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package consul

import (
	"github.com/hashicorp/consul/api"
	"github.com/nalej/derrors"
	"github.com/rs/zerolog/log"
)

type ConsulClient struct {
	client api.Client
}

func NewConsulClient(address string) (*ConsulClient, derrors.Error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		log.Error().Err(err).Msg("error creating new ConsulClient")
		return nil, derrors.NewGenericError("error creating new ConsulClient", err)
	}
	return &ConsulClient{client: *client,}, nil
}

func (a *ConsulClient) Add (organizationId string, fqdn string, ip string) derrors.Error {
	entry := &api.AgentServiceRegistration{
		Kind: api.ServiceKind(organizationId),
		Name: fqdn,
		Address: ip,
	}
	err := a.client.Agent().ServiceRegister(entry)

	if err != nil {
		log.Error().Msg("Could not register service")
		return derrors.NewGenericError(err.Error())
	}

	return nil
}

func (a *ConsulClient) Delete (serviceID string) derrors.Error {
	err := a.client.Agent().ServiceDeregister(serviceID)

	if err != nil {
		log.Error().Msg("Could not deregister service")
		return derrors.NewGenericError(err.Error())
	}

	return nil
}

func (a *ConsulClient) List (serviceKind string) ([]Service, derrors.Error) {
	services, err := a.client.Agent().Services()

	if err != nil {
		log.Error().Msg("Could not retrieve service list")
		return nil, derrors.NewGenericError(err.Error())
	}

	serviceList := make ([]Service, 0)
	for _, v := range services {
		intermediateServ := ServiceFromConsulAPI(v)
		if string(v.Kind) == serviceKind {
			serviceList = append(serviceList, intermediateServ)
		}
	}

	return serviceList, nil
}