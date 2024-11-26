package handlers

import (
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"github.com/gdyunin/metricol.git/internal/server/metrics/builder"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestMetricPostHandler(t *testing.T) {
	expectedContentType := "text/plain; charset=utf-8"
	type metric struct {
		metricType  metrics.MetricType
		metricName  string
		metricValue string
	}
	tests := []struct {
		name           string
		repository     storage.Repository
		metric         metric
		expectedCode   int
		wantRepository storage.Repository
	}{
		{
			name:       "add valid gauge",
			repository: storage.NewWarehouse(),
			metric: metric{
				metricType:  metrics.MetricTypeGauge,
				metricName:  "test_gauge",
				metricValue: "4.2",
			},
			expectedCode: http.StatusOK,
			wantRepository: func() storage.Repository {
				w := storage.NewWarehouse()
				m, _ := builder.NewMetric(metrics.MetricTypeGauge)
				_ = m.SetName("test_gauge")
				_ = m.SetValue("4.2")
				_ = w.PushMetric(m)
				return w
			}(),
		},
		{
			name:       "add invalid gauge",
			repository: storage.NewWarehouse(),
			metric: metric{
				metricType:  metrics.MetricTypeGauge,
				metricName:  "test_counter",
				metricValue: "42invalid",
			},
			expectedCode:   http.StatusBadRequest,
			wantRepository: nil,
		},
		{
			name:       "add valid counter",
			repository: storage.NewWarehouse(),
			metric: metric{
				metricType:  metrics.MetricTypeCounter,
				metricName:  "test_counter",
				metricValue: "42",
			},
			expectedCode: http.StatusOK,
			wantRepository: func() storage.Repository {
				w := storage.NewWarehouse()
				m, _ := builder.NewMetric(metrics.MetricTypeCounter)
				_ = m.SetName("test_counter")
				_ = m.SetValue("42")
				_ = w.PushMetric(m)
				return w
			}(),
		},
		{
			name:       "add invalid counter",
			repository: storage.NewWarehouse(),
			metric: metric{
				metricType:  metrics.MetricTypeCounter,
				metricName:  "test_counter",
				metricValue: "42invalid",
			},
			expectedCode:   http.StatusBadRequest,
			wantRepository: nil,
		},
		{
			name:       "add empty metric name metric",
			repository: storage.NewWarehouse(),
			metric: metric{
				metricType:  metrics.MetricTypeCounter,
				metricName:  "",
				metricValue: "42",
			},
			expectedCode:   http.StatusNotFound,
			wantRepository: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run test server
			handler := MetricPostHandler(tt.repository)
			router := chi.NewRouter()
			router.Post("/update/{metricType}/{metricName}/{metricValue}", handler)
			srv := httptest.NewServer(router)
			defer srv.Close()

			// Send request
			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL, _ = url.JoinPath(srv.URL, "update/", string(tt.metric.metricType), tt.metric.metricName, tt.metric.metricValue)
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
		})
	}
}
