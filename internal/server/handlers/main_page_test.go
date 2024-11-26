package handlers

import (
	"fmt"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"github.com/gdyunin/metricol.git/internal/server/metrics/builder"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestMainPageHandler(t *testing.T) {
	expectedContentType := "text/html; charset=utf-8"
	tests := []struct {
		name         string
		method       string
		repository   storage.Repository
		expectedCode int
	}{
		{
			name:         "empty repo",
			method:       http.MethodGet,
			repository:   storage.NewWarehouse(),
			expectedCode: http.StatusOK,
		},
		{
			name:   "single counter metric in repo",
			method: http.MethodGet,
			repository: func() storage.Repository {
				w := storage.NewWarehouse()

				// Add single metric
				m, _ := builder.NewMetric(metrics.MetricTypeCounter)
				_ = m.SetName("test counter")
				_ = m.SetValue("42")
				_ = w.PushMetric(m)

				return w
			}(),
			expectedCode: http.StatusOK,
		},
		{
			name:   "single gauge metric in repo",
			method: http.MethodGet,
			repository: func() storage.Repository {
				w := storage.NewWarehouse()

				// Add single metric
				m, _ := builder.NewMetric(metrics.MetricTypeGauge)
				_ = m.SetName("test gauge")
				_ = m.SetValue("42.4242")
				_ = w.PushMetric(m)

				return w
			}(),
			expectedCode: http.StatusOK,
		},
		{
			name:   "multi metric in repo",
			method: http.MethodGet,
			repository: func() storage.Repository {
				w := storage.NewWarehouse()

				// Add several metrics
				metricsList := []struct {
					name  string
					value string
					mType metrics.MetricType
				}{
					{"test_gauge0", "42.0", metrics.MetricTypeGauge},
					{"test_gauge43", "42.431", metrics.MetricTypeGauge},
					{"test_gauge542", "42.43641", metrics.MetricTypeGauge},
					{"test_counter4242", "4242", metrics.MetricTypeCounter},
				}

				for _, mData := range metricsList {
					m, _ := builder.NewMetric(mData.mType)
					_ = m.SetName(mData.name)
					_ = m.SetValue(mData.value)
					_ = w.PushMetric(m)
				}

				return w
			}(),
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			// Run test server
			handler := MainPageHandler(tt.repository)
			srv := httptest.NewServer(handler)
			defer srv.Close()

			// Send request
			req := resty.New().R()
			req.Method = tt.method
			req.URL = srv.URL
			resp, err := req.Send()

			// Want no errors
			require.NoError(t, err, "error making request")

			// Want status code
			expectedCode := tt.expectedCode
			actualCode := resp.StatusCode()
			require.Equalf(t, expectedCode, actualCode, "expected response code %s, but got %s",
				strconv.Itoa(expectedCode), strconv.Itoa(actualCode))

			// Want Content-Type
			actualContentType := resp.Header().Get("Content-Type")
			require.Equalf(t, expectedContentType, actualContentType, "expected Content-Type %s, but got %s", expectedContentType, actualContentType)

			// Want body
			actualBody := resp.String()
			defer func() { _ = resp.RawBody().Close() }()
			for _, metricType := range tt.repository.Metrics() {
				for name, value := range metricType {
					require.Containsf(t, actualBody, fmt.Sprintf(rowTemplate, name, value),
						"expected metric %s with value %s don`t exist in got body %s", name, value, actualBody)
				}
			}
		})
	}
}
