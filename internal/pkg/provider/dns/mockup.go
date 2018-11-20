/*
 * Copyright (C)  2018 Nalej - All Rights Reserved
 */

package dns

import (
	"github.com/nalej/derrors"
	"github.com/nalej/network-manager/internal/pkg/entities"
	"sync"
)

type MockupDNSEntryProvider struct {
	sync.Mutex
	// DNS entries indexed by cluster identifier.
	entries map[string]entities.DNSEntry
}

func NewMockupDNSEntryProvider() *MockupDNSEntryProvider {
	return &MockupDNSEntryProvider{
		entries: make(map[string]entities.DNSEntry, 0),
	}
}

func (m *MockupDNSEntryProvider) unsafeExists(FQDN string) bool {
	_, exists := m.entries[FQDN]
	return exists
}

// Clear cleans the contents of the mockup.
func (m *MockupDNSEntryProvider) Clear() {
	m.Lock()
	m.entries = make(map[string]entities.DNSEntry, 0)
	m.Unlock()
}

// List all DNS entries from a network
func (m *MockupDNSEntryProvider) List(networkID string) derrors.Error {
	m.Lock()
	defer m.Unlock()
	if !m.unsafeExists(networkID) {
		return derrors.NewNotFoundError(networkID)
	}
	return nil
}

// Add a new DNS Entry to the system.
func (m *MockupDNSEntryProvider) Add(entry entities.DNSEntry) derrors.Error {
	m.Lock()
	defer m.Unlock()
	if !m.unsafeExists(entry.Fqdn) {
		m.entries[entry.Fqdn] = entry
		return nil
	}
	return derrors.NewAlreadyExistsError(entry.Fqdn)
}

// Delete a DNS Entry
func (m *MockupDNSEntryProvider) Delete(FQDN string) derrors.Error {
	m.Lock()
	defer m.Unlock()
	if !m.unsafeExists(FQDN) {
		return derrors.NewNotFoundError(FQDN)
	}
	delete(m.entries, FQDN)
	return nil
}
