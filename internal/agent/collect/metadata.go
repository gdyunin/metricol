package collect

import (
	"math/rand"
	"sync"
)

// Metadata holds information about the number of polls and the last poll seed.
// It provides thread-safe methods to update and access this data.
type Metadata struct {
	pollsCount   int64
	lastPollSeed float64
	mu           *sync.RWMutex
}

// NewMetadata initializes and returns a new instance of Metadata.
// The initial values for pollsCount and lastPollSeed are set to 0.
func NewMetadata() *Metadata {
	return &Metadata{
		pollsCount:   0,
		lastPollSeed: 0,
		mu:           &sync.RWMutex{},
	}
}

// Update increments the pollsCount by 1 and generates a new random seed for lastPollSeed.
// This method is thread-safe.
func (m *Metadata) Update() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pollsCount++
	m.lastPollSeed = rand.Float64()
}

// Reset sets the pollsCount and lastPollSeed to 0.
// This method is thread-safe.
func (m *Metadata) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pollsCount = 0
	m.lastPollSeed = 0
}

// PollsCount retrieves the current value of pollsCount in a thread-safe manner.
func (m *Metadata) PollsCount() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pollsCount
}

// LastPollSeed retrieves the current value of lastPollSeed in a thread-safe manner.
func (m *Metadata) LastPollSeed() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastPollSeed
}
