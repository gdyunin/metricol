package send

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

func TestStreamSender_SendBatch(t *testing.T) {
	// Table-driven tests.
	tests := []struct {
		metrics       *entity.Metrics
		name          string
		expectErrPart string
		serverStatus  int
		expectErr     bool
	}{
		{
			name:         "successful send with nil metrics",
			metrics:      nil,
			serverStatus: http.StatusOK,
			expectErr:    false,
		},
		{
			name: "successful send with valid gauge metric",
			metrics: &entity.Metrics{
				{
					Name:  "gauge1",
					Type:  entity.MetricTypeGauge,
					Value: 3.14, // correct type for gauge: float64
				},
			},
			serverStatus: http.StatusOK,
			expectErr:    false,
		},
		{
			name: "successful send with valid counter metric",
			metrics: &entity.Metrics{
				{
					Name:  "counter1",
					Type:  entity.MetricTypeCounter,
					Value: int64(100), // correct type for counter: int64
				},
			},
			serverStatus: http.StatusOK,
			expectErr:    false,
		},
		{
			name: "metrics conversion error for gauge",
			metrics: &entity.Metrics{
				{
					Name:  "gauge_bad",
					Type:  entity.MetricTypeGauge,
					Value: 42, // incorrect type: int instead of float64
				},
			},
			serverStatus:  http.StatusOK,
			expectErr:     true,
			expectErrPart: "unexpected value type for gauge metric",
		},
		{
			name: "server error",
			metrics: &entity.Metrics{
				{
					Name:  "gauge1",
					Type:  entity.MetricTypeGauge,
					Value: 3.14,
				},
			},
			serverStatus:  http.StatusInternalServerError,
			expectErr:     true,
			expectErrPart: "unsuccessful response from server",
		},
	}

	// Iterate through the test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up an HTTP test server that returns the desired status code.
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Optionally, you can inspect the request body.
				var data json.RawMessage
				_ = json.NewDecoder(r.Body).Decode(&data)
				w.WriteHeader(tc.serverStatus)
			}))
			defer ts.Close()

			// Create a resty client pointed at the test server.
			client := resty.New().SetBaseURL(ts.URL)

			// Create a RequestBuilder using the original implementation.
			builder := NewRequestBuilder(client)

			// Create a dummy channel for streamFrom (not used directly in SendBatch).
			dummyChan := make(chan *entity.Metrics)
			// Create a no-op logger.
			logger := zap.NewNop().Sugar()

			// Initialize the StreamSender.
			sender := NewStreamSender(dummyChan, time.Second, 1, ts.URL, "dummySigningKey", "", logger)
			// Override the requestBuilder with our instance that uses the original compressor.
			sender.requestBuilder = builder

			// Call SendBatch.
			err := sender.SendBatch(context.Background(), tc.metrics)

			// Validate the result.
			if tc.expectErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				} else if !strings.Contains(err.Error(), tc.expectErrPart) {
					t.Errorf("expected error message to contain %q, got %q", tc.expectErrPart, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}
