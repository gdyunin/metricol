package send

import (
	"fmt"
	"github.com/gdyunin/metricol.git/internal/agent/fetch"
	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsSender_Send(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Need post method
		require.Equal(t, http.MethodPost, r.Method)
		// Have path
		require.Contains(t, r.URL.Path, "/update/")
		// Have metric type
		require.Regexp(t, fmt.Sprintf("/%s|%s/", metrics.MetricTypeGauge, metrics.MetricTypeCounter), r.URL.Path)
		// Have metric value
		require.Regexp(t, "/\\d+$", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	tests := []struct {
		name           string
		metricsFetcher fetch.Fetcher
	}{
		{
			"simple send test",
			fetch.NewMetricsFetcher(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := NewMetricsSender(tt.metricsFetcher, testServer.URL)
			sender.Send()
		})
	}
}

func TestNewMetricsSender(t *testing.T) {
	tests := []struct {
		name        string
		wantAddress string
		want        *MetricsSender
	}{
		{
			"simple MetricSender init",
			"localhost:8080",
			&MetricsSender{
				metricsFetcher: fetch.NewMetricsFetcher(),
				serverAddress:  "localhost:8080",
				client:         resty.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewMetricsSender(fetch.NewMetricsFetcher(), "localhost:8080")
			require.Equal(t, tt.want.serverAddress, ms.serverAddress)
		})
	}
}
