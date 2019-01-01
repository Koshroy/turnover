package subscriptions

import (
	"net/url"
	"sync"
)

// MemManager manages a list of forwarding targets in-memory
type MemManager struct {
	targets []url.URL

	sync.RWMutex
}

// NewMemManager instantiates a new manager
func NewMemManager() *MemManager {
	return &MemManager{
		targets: make([]url.URL, 0),
	}
}

// Add adds a new forwarding target to the manager
func (m *MemManager) Add(target url.URL) bool {
	m.Lock()
	defer m.Unlock()

	m.targets = append(m.targets, target)
	return true
}

// Remove removes a forwarding target from the manager
func (m *MemManager) Remove(target url.URL) bool {
	m.Lock()
	defer m.Unlock()

	var i = 0
	var url url.URL
	var found = false
	for i, url = range m.targets {
		if url == target {
			found = true
			break
		}
	}

	if found {
		m.targets = append(m.targets[:i], m.targets[i+1:]...)
	}

	return found
}

// List returns a list of forwarding targets
func (m *MemManager) List() []url.URL {
	m.RLock()
	defer m.RUnlock()

	retArr := make([]url.URL, len(m.targets))
	copy(retArr, m.targets)

	return retArr
}
