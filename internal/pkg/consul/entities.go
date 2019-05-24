/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package consul

import (
	"github.com/hashicorp/consul/api"
)

type Service struct {
	Kind api.ServiceKind
	//ID string
	Service string
	Address string
}

func ServiceFromConsulAPI(s *api.AgentService) Service {
	return Service{
		Kind: s.Kind,
		//ID: s.ID,
		Service: s.Service,
		Address: s.Address,
	}
}

type GenericEntry struct{
	Fqdn string
	IP string
	Tags []string
}