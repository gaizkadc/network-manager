/*
 * Copyright (C)  2018 Nalej - All Rights Reserved
 */

package network

import (
	"github.com/nalej/derrors"
	"github.com/nalej/network-manager/internal/pkg/entities"
	"sync"
)

type MockupNetworkProvider struct {
	sync.Mutex
	// Networks indexed by network identifier.
	networks map[string]entities.Network
}

func NewMockupNetworkProvider() *MockupNetworkProvider {
	return &MockupNetworkProvider{
		networks: make(map[string]entities.Network, 0),
	}
}

func (m *MockupNetworkProvider) unsafeExists(networkID string) bool {
	_, exists := m.networks[networkID]
	return exists
}

// Add a new network to the system.
func (m *MockupNetworkProvider) Add(network entities.Network) derrors.Error {
	m.Lock()
	defer m.Unlock()
	if !m.unsafeExists(network.NetworkId){
		m.networks[network.NetworkId] = network
		return nil
	}
	return derrors.NewAlreadyExistsError(network.NetworkId)
}

// Exists checks if a network exists on the system.
func (m * MockupNetworkProvider) Exists(networkID string) bool {
	m.Lock()
	defer m.Unlock()
	return m.unsafeExists(networkID)
}

// Get a network.
func (m * MockupNetworkProvider) Get(networkID string) (*entities.Network, derrors.Error) {
	m.Lock()
	defer m.Unlock()
	network, exists := m.networks[networkID]
	if exists {
		return &network, nil
	}
	return nil, derrors.NewNotFoundError(networkID)
}

// Delete a network.
func (m * MockupNetworkProvider) Remove(networkID string) derrors.Error {
	m.Lock()
	defer m.Unlock()
	if !m.unsafeExists(networkID){
		return derrors.NewNotFoundError(networkID)
	}
	delete(m.networks, networkID)
	return nil
}