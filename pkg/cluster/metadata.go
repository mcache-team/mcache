package cluster

import "sync"

type metadataStore struct {
	mu          sync.RWMutex
	addressBook map[string]string
}

func newMetadataStore(initial map[string]string) *metadataStore {
	addressBook := make(map[string]string, len(initial))
	for key, value := range initial {
		addressBook[key] = value
	}
	return &metadataStore{addressBook: addressBook}
}

func (m *metadataStore) Snapshot() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	addressBook := make(map[string]string, len(m.addressBook))
	for key, value := range m.addressBook {
		addressBook[key] = value
	}
	return addressBook
}

func (m *metadataStore) Restore(snapshot map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.addressBook = make(map[string]string, len(snapshot))
	for key, value := range snapshot {
		m.addressBook[key] = value
	}
}

func (m *metadataStore) Lookup(nodeID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.addressBook[nodeID]
	return value, ok
}

func (m *metadataStore) Upsert(nodeID, advertiseAddr string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.addressBook[nodeID] = advertiseAddr
}

func (m *metadataStore) Delete(nodeID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.addressBook, nodeID)
}
