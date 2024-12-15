package agent

import (
	"math/rand"
	"testing"
	"time"

	"github.com/gdyunin/metricol.git/internal/config/agent"
	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/stretchr/testify/require"
)

const (
	testIntervalPolling   = 1
	testIntervalReporting = 2
)

func TestAgent_Polling(t *testing.T) {
	tests := []struct {
		name    string
		agent   *Agent
		wantErr bool
	}{
		{
			name: "Polling correctly",
			agent: func() *Agent {
				f := NewMetricsFetcher()
				f.AddMetrics(metrics.NewGauge("RandomValue", 0).SetFetcherAndReturn(rand.Float64))
				return &Agent{
					fetcher: f,
				}
			}(),
			wantErr: false,
		},
		{
			name: "Polling failure",
			agent: func() *Agent {
				f := NewMetricsFetcher()
				f.AddMetrics(metrics.NewGauge("RandomValue", 0))
				return &Agent{
					fetcher: f,
				}
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inProgress := false
			go func() {
				inProgress = true
				tt.agent.Polling(testIntervalPolling)
				inProgress = false
			}()
			time.Sleep(testIntervalPolling * resetErrorCountersIntervals * time.Second)

			if tt.wantErr {
				require.False(t, inProgress)
			} else {
				require.True(t, inProgress)
			}
		})
	}
}

func TestAgent_Reporting(t *testing.T) {
	tests := []struct {
		name    string
		agent   *Agent
		wantErr bool
	}{
		{
			name: "Reporting correctly",
			agent: func() *Agent {
				f := NewMetricsFetcher()
				f.AddMetrics(metrics.NewGauge("RandomValue", 0).SetFetcherAndReturn(rand.Float64))
				s := NewMetricsSender(f, "localhost:8080")
				return &Agent{
					fetcher: f,
					sender:  s,
				}
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inProgress := false
			go func() {
				inProgress = true
				tt.agent.Reporting(testIntervalReporting)
				inProgress = false
			}()
			time.Sleep(testIntervalPolling * resetErrorCountersIntervals * time.Second)

			if tt.wantErr {
				require.False(t, inProgress)
			} else {
				require.True(t, inProgress)
			}
		})
	}
}

func TestDefaultAgent(t *testing.T) {
	tests := []struct {
		name string
		cfg  *agent.Config
	}{
		{
			name: "Create new Agent from config and setup default fetcher",
			cfg: &agent.Config{
				ServerAddress:  "localhost:8080",
				PollInterval:   2,
				ReportInterval: 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := DefaultAgent(tt.cfg)
			require.NotNil(t, a)
			require.NotNil(t, a.fetcher)
			require.NotNil(t, a.sender)
			require.Len(t, a.fetcher.Metrics(), 29)
		})
	}
}

func TestNewAgent(t *testing.T) {
	tests := []struct {
		name string
		cfg  *agent.Config
		want *Agent
	}{
		{
			name: "Create new Agent from config",
			cfg: &agent.Config{
				ServerAddress:  "localhost:8080",
				PollInterval:   2,
				ReportInterval: 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAgent(tt.cfg)
			require.NotNil(t, a)
			require.NotNil(t, a.fetcher)
			require.NotNil(t, a.sender)
		})
	}
}

func Test_setDefaultMetrics(t *testing.T) {
	fetcher := NewMetricsFetcher()
	sender := NewMetricsSender(fetcher, "localhost:8080")
	tests := []struct {
		name  string
		agent *Agent
	}{
		{
			"Successful setup default fetcher",
			&Agent{fetcher: fetcher, sender: sender},
		},
	}
	for _, tt := range tests {
		fetcher := tt.agent.fetcher
		withDefaultMetrics()(tt.agent)
		require.Len(t, fetcher.Metrics(), 29)
	}
}
