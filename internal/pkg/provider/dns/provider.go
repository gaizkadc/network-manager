/*
 * Copyright (C)  2018 Nalej - All Rights Reserved
 */

package dns

import (
	"github.com/nalej/derrors"
	"github.com/nalej/network-manager/internal/pkg/entities"
)

type Provider interface {
	//List DNS entries on the system
	List(dns entities.DNSEntry) derrors.Error
	//Add DNS entry to the system
	Add(dns entities.DNSEntry) derrors.Error
	//Delete DNS entry from the system
	Delete(dns entities.DNSEntry) derrors.Error
}