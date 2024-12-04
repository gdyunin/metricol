package agent

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/stretchr/testify/require"
)

// TestMetricsSender_Send tests the Send method of MetricsSender.
// It verifies that metrics are sent correctly to the server.
func TestMetricsSender_Send(t *testing.T) {
	tests := []struct {
		name           string
		metricsFetcher Fetcher
	}{
		{
			"Simple send test with Gauge metric",
			func() *MetricsFetcher {
				f := NewMetricsFetcher()
				f.AddMetrics(metrics.NewGauge("RandomValue", 0).SetFetcherAndReturn(rand.Float64))
				return f
			}(),
		},
		{
			"Simple send test with Counter metric",
			func() *MetricsFetcher {
				f := NewMetricsFetcher()
				f.AddMetrics(metrics.NewCounter("PollCount", 0).SetFetcherAndReturn(func() int64 {
					return 1
				}))
				return f
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Ensure the request method is POST
				require.Equal(t, http.MethodPost, r.Method)
				// Ensure the request path contains "/update/"
				require.Contains(t, r.URL.Path, "/update/")
				// Ensure the request path contains a valid metric type
				require.Regexp(t, fmt.Sprintf("/%s|%s/", metrics.MetricTypeGauge, metrics.MetricTypeCounter), r.URL.Path)
				// Ensure the request path ends with a metric value
				require.Regexp(t, "/\\d+$", r.URL.Path)
				w.WriteHeader(http.StatusOK)
			}))
			defer testServer.Close()

			sender := NewMetricsSender(tt.metricsFetcher, strings.TrimPrefix(testServer.URL, "http://"))
			err := sender.Send()
			require.NoError(t, err)
		})
	}
}

// TestNewMetricsSender tests the creation of a new MetricsSender instance.
// It verifies that the server address and dependencies are initialized correctly.
func TestNewMetricsSender(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		wantAddress string
	}{
		{
			"Simple MetricSender init with address",
			"http://localhost:8080",
			"http://localhost:8080",
		},
		{
			"MetricSender init with another address",
			"http://anotherhost:1234",
			"http://anotherhost:1234",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := NewMetricsSender(NewMetricsFetcher(), tt.address)
			require.Equal(t, tt.wantAddress, sender.serverAddress)
			require.NotNil(t, sender.client)
			require.NotNil(t, sender.metricsFetcher)
		})
	}
}
