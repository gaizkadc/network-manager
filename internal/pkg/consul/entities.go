/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package consul

import (
	"github.com/hashicorp/consul/api"
)

type Service struct {
	ID string
	Service string
	Address string
}

func ServiceFromConsulAPI (s *api.AgentService) Service {
	return Service {
		ID: s.ID,
		Service: s.Service,
		Address: s.Address,
	}
}