package metadata

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetadata(t *testing.T) {
	tests := []struct {
		actions       func(m *Metadata)
		name          string
		expectedPolls int64
		expectedSeed  float64
	}{
		{
			name:          "Default values",
			actions:       func(m *Metadata) {},
			expectedPolls: 0,
			expectedSeed:  0,
		},
		{
			name: "Update once",
			actions: func(m *Metadata) {
				m.Update()
			},
			expectedPolls: 1,
			expectedSeed:  -1,
		},
		{
			name: "Update multiple times",
			actions: func(m *Metadata) {
				for range 5 {
					m.Update()
				}
			},
			expectedPolls: 5,
			expectedSeed:  -1,
		},
		{
			name: "Reset after updates",
			actions: func(m *Metadata) {
				m.Update()
				m.Reset()
			},
			expectedPolls: 0,
			expectedSeed:  0,
		},
		{
			name: "Reset without update",
			actions: func(m *Metadata) {
				m.Reset()
			},
			expectedPolls: 0,
			expectedSeed:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMetadata()
			tt.actions(m)

			assert.Equal(t, tt.expectedPolls, m.PollsCount())
			if tt.expectedSeed == -1 {
				assert.NotZero(t, m.LastPollSeed())
			} else {
				assert.Equal(t, tt.expectedSeed, m.LastPollSeed())
			}
		})
	}
}

func TestMetadata_ConcurrentAccess(t *testing.T) {
	m := NewMetadata()
	concurrency := 100

	var wg sync.WaitGroup
	for range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Update()
			m.PollsCount()
			m.LastPollSeed()
		}()
	}

	wg.Wait()

	assert.Equal(t, int64(concurrency), m.PollsCount())
	assert.NotZero(t, m.LastPollSeed())
}
