package agent

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestAgent_Start(t *testing.T) {
	// Use intervals longer than the runDuration so that no ticker tick occurs.
	tests := []struct {
		name           string
		pollInterval   time.Duration
		reportInterval time.Duration
		maxSendRate    int
		runDuration    time.Duration // Duration before context cancellation.
	}{
		{
			name:           "short run",
			pollInterval:   1 * time.Second,
			reportInterval: 1 * time.Second,
			maxSendRate:    1,
			runDuration:    300 * time.Millisecond,
		},
		{
			name:           "medium run",
			pollInterval:   1 * time.Second,
			reportInterval: 1 * time.Second,
			maxSendRate:    1,
			runDuration:    500 * time.Millisecond,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := zap.NewNop().Sugar()
			agentInstance := NewAgent(
				tc.pollInterval,
				tc.reportInterval,
				logger,
				tc.maxSendRate,
				"http://localhost:8080",
				"dummyKey",
				"",
			)

			// Start a goroutine to continuously drain the sendQueue.
			drainDone := make(chan struct{})
			go func() {
				for range agentInstance.sendQueue {
					// Drain messages until the channel is closed.
				}
				close(drainDone)
			}()

			ctx, cancel := context.WithTimeout(context.Background(), tc.runDuration)
			defer cancel()

			done := make(chan struct{})
			go func() {
				agentInstance.Start(ctx)
				close(done)
			}()

			// Wait for Agent.Start to return.
			select {
			case <-done:
				// Agent exited as expected.
			case <-time.After(tc.runDuration + 300*time.Millisecond):
				t.Errorf("Agent.Start did not return in expected time for %q", tc.name)
			}

			// Wait for the drain goroutine to complete (i.e. sendQueue is closed).
			select {
			case <-drainDone:
				// Success: sendQueue was closed.
			case <-time.After(100 * time.Millisecond):
				t.Error("expected sendQueue to be closed after Agent.Start returns")
			}
		})
	}
}
