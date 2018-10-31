/*
 * Copyright (C)  2018 Nalej - All Rights Reserved
 */

 package network

 import (
	 "github.com/nalej/derrors"
	 "github.com/nalej/network-manager/internal/pkg/entities"
 )

 // Provider for application
 type Provider interface {
	 // Add a new network to the system.
	 Add(network entities.Network) derrors.Error
	 // Exists checks if a network exists on the system.
	 Exists(networkID string) bool
	 // Get a network.
	 Get(networkID string) (* entities.Network, derrors.Error)
	 // Delete a network.
	 Delete(networkID string) derrors.Error
 }