/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package consul

import (
	"github.com/hashicorp/consul/api"
	"github.com/nalej/derrors"
	"github.com/rs/zerolog/log"
	"fmt"
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
	return &ConsulClient{client: *client}, nil
}

func (a *ConsulClient) Add(organizationId string, appInstanceId string, fqdn string, ip string) derrors.Error {

	entry := &api.AgentServiceRegistration{
		//Kind:    api.ServiceKind(organizationId),
		Name:    fqdn,
		Address: ip,
		Tags: []string{organizationId, appInstanceId},
		//ID: fmt.Sprintf("%s-%s",appInstanceId,organizationId),
	}

	err := a.client.Agent().ServiceRegister(entry)

	if err != nil {
		log.Error().Msg("Could not register service")
		return derrors.NewGenericError(err.Error())
	}

	return nil
}

func (a *ConsulClient) Delete(serviceID string, organizationId string, appInstanceId string) derrors.Error {
	/*
	err := a.client.Agent().ServiceDeregister(serviceID)

	if err != nil {
		log.Error().Msg("Could not deregister service")
		return derrors.NewGenericError(err.Error())
	}
	return nil
	*/

	q := api.QueryOptions{}
	cat, _, err := a.client.Catalog().ServiceMultipleTags(serviceID,[]string{organizationId, appInstanceId},&q)

	if err != nil {
		log.Error().Err(err).Msg("impossible to build DNS query")
		return derrors.NewGenericError("impossible to build DNS query", err)
	}

	log.Debug().Msgf("found %d DNS entries to be deleted", len(cat))

	for _,serv := range cat {
		log.Debug().Msgf("deregister service %s from address %s", serv.ServiceID, serv.Address)
		// Create a client to deregister from the original agent that registered the service.
		config := api.DefaultConfig()
		config.Address = fmt.Sprintf("%s:8500",serv.Address)
		auxCli, err := api.NewClient(config)

		if err != nil {
			log.Error().Err(err).Msgf("impossible to create client to connect to %s", serv.Address)
		}

		err = auxCli.Agent().ServiceDeregister(serv.ServiceID)

		dereg := api.CatalogDeregistration{
			ServiceID: serv.ServiceID,
			Datacenter: serv.Datacenter,
			Node: serv.Node,
			Address: serv.Address,
		}
		q := api.WriteOptions{Datacenter:"dc1"}
		_, err = auxCli.Catalog().Deregister(&dereg,&q)

		if err != nil {
			log.Error().Err(err).Msgf("error deleting DNS entry %s", serv.ServiceID)
		}

	}


	return nil
}

func (a *ConsulClient) List(serviceKind string) ([]Service, derrors.Error) {
	services, err := a.client.Agent().Services()

	if err != nil {
		log.Error().Msg("Could not retrieve service list")
		return nil, derrors.NewGenericError(err.Error())
	}

	serviceList := make([]Service, 0)
	for _, v := range services {
		intermediateServ := ServiceFromConsulAPI(v)
		if string(v.Kind) == serviceKind {
			serviceList = append(serviceList, intermediateServ)
		}
	}

	return serviceList, nil
}
