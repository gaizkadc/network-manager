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

package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/nalej/derrors"
	"github.com/rs/zerolog/log"
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

func (a *ConsulClient) Add(serviceName string, fqdn string, ip string, tags []string) derrors.Error {

	// Register an external agent so we do not need a local agent running to control that service
	entry := &api.CatalogRegistration{
		Node:       fqdn,
		Datacenter: "dc1",
		Address:    ip,
		Service: &api.AgentService{
			Service: fqdn,
			ID:      fqdn,
			Address: ip,
			Tags:    tags,
		},
		NodeMeta: map[string]string{
			"external-node": "true",
		},
	}
	_, err := a.client.Catalog().Register(entry, &api.WriteOptions{Datacenter: "dc1"})
	if err != nil {
		log.Error().Msg("impossible to register service in catalog")
		return derrors.NewGenericError(err.Error())
	}

	return nil
}

// Delete uses the fqdn as the consul service id and removes
// the entry with that id. If tags are indicated, all the service entries
// with those tags are removed.
func (a *ConsulClient) Delete(fqdn string, tags []string) derrors.Error {

	// delete the associated node

	if tags == nil || len(tags) == 0 {
		// remove using the FQDN
		return a.deleteEntryById(fqdn)
	}

	// delete entries using tags
	return a.deleteEntryByTags(tags)
}

func (a *ConsulClient) deleteEntryById(id string) derrors.Error {

	// Remove the associated consul node to get rid of any everything.
	dereg := api.CatalogDeregistration{
		Datacenter: "dc1",
		Node:       id,
	}
	_, err := a.client.Catalog().Deregister(&dereg, &api.WriteOptions{Datacenter: "dc1"})
	if err != nil {
		log.Error().Err(err).Str("serviceId", id).Msg("service not found")
		return derrors.NewInternalError("service not found", err)
	}
	return nil
}

func (a *ConsulClient) deleteEntryByTags(tags []string) derrors.Error {

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

	toDelete := make([]string, 0)
	for serviceId, serviceTags := range services {
		if serviceTags == nil || len(serviceTags) == 0 {
			continue
		}
		// build a map to determine if the tags are available
		available := make(map[string]bool, len(serviceTags))
		for _, availableTag := range serviceTags {
			available[availableTag] = true
		}

		var allTagsFound bool = true
		for _, toFind := range tags {
			if _, found := available[toFind]; !found {
				allTagsFound = false
			}
		}

		if allTagsFound {
			toDelete = append(toDelete, serviceId)
		}
	}

	log.Debug().Msgf("%d service entries to be deleted", len(toDelete))
	for _, serviceId := range toDelete {
		// find in what node is the service registered
		serv, _, err := a.client.Catalog().Service(serviceId, "", &q)
		if err != nil {
			log.Error().Err(err).Msgf("impossible to retrieve information for service %s", serviceId)
			return derrors.NewGenericError("impossible to retrieve service information", err)
		}
		for _, servEntry := range serv {
			// build the client for the specific node
			config := api.DefaultConfig()
			config.Address = fmt.Sprintf("%s:%d", servEntry.Address, ConsulDNSPort)
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

func (a *ConsulClient) getGenericID(organizationID string, fqdn string) string {
	return fmt.Sprintf("%s-%s", organizationID, fqdn)
}

func (a *ConsulClient) AddGenericEntry(organizationID string, fqdn string, ip string, tags ...string) derrors.Error {
	entryTags := []string{organizationID}
	entryTags = append(entryTags, tags...)
	entry := &api.AgentServiceRegistration{
		ID:      a.getGenericID(organizationID, fqdn),
		Name:    fqdn,
		Address: ip,
		Tags:    entryTags,
	}

	err := a.client.Agent().ServiceRegister(entry)

	if err != nil {
		log.Error().Str("err", err.Error()).Msg("Cannot add generic entry")
		return derrors.NewGenericError(err.Error())
	}

	return nil
}

func (a *ConsulClient) DeleteGenericEntry(organizationID string, fqdn string) derrors.Error {
	serviceID := a.getGenericID(organizationID, fqdn)
	q := api.QueryOptions{}

	// find in what node is the service registered
	serv, _, err := a.client.Catalog().Service(serviceID, "", &q)
	if err != nil {
		log.Error().Err(err).Msgf("impossible to retrieve information for service %s", serviceID)
		return derrors.NewGenericError("impossible to retrieve service information", err)
	}
	for _, servEntry := range serv {
		// build the client for the specific node
		config := api.DefaultConfig()
		config.Address = fmt.Sprintf("%s:%d", servEntry.Address, ConsulDNSPort)
		auxCli, err := api.NewClient(config)
		log.Debug().Str("address", servEntry.Address).Str("serviceID", servEntry.ServiceID).Msgf("delete service")
		err = auxCli.Agent().ServiceDeregister(servEntry.ServiceID)

		if err != nil {
			log.Error().Err(err).Msgf("impossible to retrieve information for service %s", serviceID)
		}
	}
	return nil
}
