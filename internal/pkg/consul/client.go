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

const (
	// Port exposing the ConsulDNS service
	ConsulDNSPort = 8500
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
		//Kind:    api.ServiceKind(appInstanceId),
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

func (a *ConsulClient) Delete(organizationId string, appInstanceId string) derrors.Error {

	// The current consul API does not filter services by tag
	// https://github.com/hashicorp/consul/issues/4811
	// The workaround is to get all the available services from the catalog and
	// remove those matching the tags

	q := api.QueryOptions{}
	services, _, err := a.client.Catalog().Services(&q)

	if err != nil {
		log.Error().Err(err).Msg("impossible to build DNS query")
		return derrors.NewGenericError("impossible to build DNS query", err)
	}

	toDelete := make([]string,0)
	for serviceId, tags := range services {
		// check if organizationId and appInstanceId are in the tags
		foundOrg := false
		foundApp := false
		for _, t := range tags {
			if t == organizationId {
				foundOrg = true
			}
			if t == appInstanceId {
				foundApp = true
			}
		}
		// if found, this entry has to be deleted
		if foundOrg && foundApp {
			toDelete = append(toDelete, serviceId)
		}
	}

	log.Debug().Msgf("%d service entries to be deleted", len(toDelete))
	for _, serviceId := range toDelete {
		// find in what node is the service registered
		serv, _, err := a.client.Catalog().Service(serviceId,"",&q)
		if err != nil {
			log.Error().Err(err).Msgf("impossible to retrieve information for service %s", serviceId)
			return derrors.NewGenericError("impossible to retrieve service information", err)
		}
		for _, servEntry := range serv {
			// build the client for the specific node
			config := api.DefaultConfig()
			config.Address = fmt.Sprintf("%s:%d",servEntry.Address, ConsulDNSPort)
			auxCli, err := api.NewClient(config)
			log.Debug().Msgf("delete service %s", servEntry.ServiceID)
			err = auxCli.Agent().ServiceDeregister(servEntry.ServiceID)

			if err != nil {
				log.Error().Err(err).Msgf("impossible to retrieve information for service %s", serviceId)
			}
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

//
// Generic entries
//

func (a * ConsulClient) getGenericID(organizationID string, fqdn string) string{
	return fmt.Sprintf("%s-%s", organizationID, fqdn)
}

func (a * ConsulClient) AddGenericEntry(organizationID string, fqdn string, ip string, tags ...string) derrors.Error{
	entryTags := []string{organizationID}
	entryTags = append(entryTags, tags...)
	entry := &api.AgentServiceRegistration{
		ID: a.getGenericID(organizationID, fqdn),
		Name:    fqdn,
		Address: ip,
		Tags: entryTags,
	}

	err := a.client.Agent().ServiceRegister(entry)

	if err != nil {
		log.Error().Str("err", err.Error()).Msg("Cannot add generic entry")
		return derrors.NewGenericError(err.Error())
	}

	return nil
}

func (a * ConsulClient) DeleteGenericEntry(organizationID string, fqdn string) derrors.Error{
	serviceID := a.getGenericID(organizationID, fqdn)
	q := api.QueryOptions{}

	// find in what node is the service registered
	serv, _, err := a.client.Catalog().Service(serviceID,"",&q)
	if err != nil {
		log.Error().Err(err).Msgf("impossible to retrieve information for service %s", serviceID)
		return derrors.NewGenericError("impossible to retrieve service information", err)
	}
	for _, servEntry := range serv {
		// build the client for the specific node
		config := api.DefaultConfig()
		config.Address = fmt.Sprintf("%s:%d",servEntry.Address, ConsulDNSPort)
		auxCli, err := api.NewClient(config)
		log.Debug().Str("address", servEntry.Address).Str("serviceID", servEntry.ServiceID).Msgf("delete service")
		err = auxCli.Agent().ServiceDeregister(servEntry.ServiceID)

		if err != nil {
			log.Error().Err(err).Msgf("impossible to retrieve information for service %s", serviceID)
		}
	}
	return nil
}
