package nonce

import (
	"sync"
)

// Manager handles sequential nonce assignment for multiple concurrent bots
// using the same wallet. Thread-safe.
type Manager struct {
	mu      sync.Mutex
	current uint64
}

// NewManager creates a nonce manager starting from the given nonce
func NewManager(startNonce uint64) *Manager {
	return &Manager{
		current: startNonce,
	}
}

// GetNonce returns the next available nonce and increments the counter
// This is atomic and thread-safe
func (m *Manager) GetNonce() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	nonce := m.current
	m.current++
	return nonce
}

// Reset sets the nonce counter to a new value
// Useful if a transaction fails and you need to retry with the same nonce
func (m *Manager) Reset(nonce uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.current = nonce
}

// Current returns the current nonce value without incrementing
func (m *Manager) Current() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.current
}

