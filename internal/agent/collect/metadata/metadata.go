// Package metadata provides functionality for maintaining and updating metadata related to polling operations.
// It tracks the number of polls conducted and the random seed generated during the last poll. Thread-safety is ensured
// through the use of mutexes.
package metadata

import (
	"math/rand"
	"sync"
)

// Metadata maintains metrics about polling operations,
// including the number of polls conducted and a random seed for the last poll.
type Metadata struct {
	mu           *sync.RWMutex
	pollsCount   int64
	lastPollSeed float64
}

// NewMetadata creates and initializes a new instance of Metadata.
//
// Returns:
//   - *Metadata: A pointer to the initialized Metadata instance.
func NewMetadata() *Metadata {
	return &Metadata{
		pollsCount:   0,
		lastPollSeed: 0,
		mu:           &sync.RWMutex{},
	}
}

// Update increments the poll count and generates a new random seed.
// This method ensures thread-safe updates using a mutex.
func (m *Metadata) Update() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pollsCount++
	m.lastPollSeed = rand.Float64()
}

// Reset clears the poll count and resets the random seed to zero.
// This method is thread-safe and ensures that the metadata values are reset consistently.
func (m *Metadata) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pollsCount = 0
	m.lastPollSeed = 0
}

// PollsCount retrieves the total number of polls conducted in a thread-safe manner.
//
// Returns:
//   - int64: The current poll count.
func (m *Metadata) PollsCount() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pollsCount
}

// LastPollSeed retrieves the random seed value from the last poll in a thread-safe manner.
//
// Returns:
//   - float64: The last poll seed value.
func (m *Metadata) LastPollSeed() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastPollSeed
}
