package handlers

import (
	"github.com/gdyunin/metricol.git/internal/metrics"
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

func TestMetricGetHandler(t *testing.T) {
	// Init test repo
	testRepository := storage.NewWarehouse()

	// Add several metrics
	metricsList := []struct {
		name  string
		value string
		mType string
	}{
		{"test_gauge0", "42.0", metrics.MetricTypeGauge},
		{"test_gauge43", "42.431", metrics.MetricTypeGauge},
		{"test_gauge542", "42.43641", metrics.MetricTypeGauge},
		{"test_counter4242", "4242", metrics.MetricTypeCounter},
	}
	for _, mData := range metricsList {
		m, _ := metrics.NewFromStrings(mData.name, mData.value, mData.mType)
		_ = testRepository.PushMetric(m)
	}

	// Run test server
	handler := MetricGetHandler(testRepository)
	router := chi.NewRouter()
	router.Get("/value/{metricType}/{metricName}", handler)
	srv := httptest.NewServer(router)
	defer srv.Close()

	expectedContentType := "text/plain; charset=utf-8"
	tests := []struct {
		name          string
		repository    storage.Repository
		queryType     string
		queryMetric   string
		expectedCode  int
		expectedValue string
	}{
		{
			name:          "get gauge",
			repository:    testRepository,
			queryType:     "gauge",
			queryMetric:   "test_gauge542",
			expectedCode:  http.StatusOK,
			expectedValue: "42.43641",
		},
		{
			name:          "get counter",
			repository:    testRepository,
			queryType:     "counter",
			queryMetric:   "test_counter4242",
			expectedCode:  http.StatusOK,
			expectedValue: "4242",
		},
		{
			name:          "get unknown counter",
			repository:    testRepository,
			queryType:     "counter",
			queryMetric:   "unknown_metric",
			expectedCode:  http.StatusNotFound,
			expectedValue: "",
		},
		{
			name:          "get unknown gauge",
			repository:    testRepository,
			queryType:     "gauge",
			queryMetric:   "unknown_metric",
			expectedCode:  http.StatusNotFound,
			expectedValue: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Send request
			req := resty.New().R()
			req.Method = http.MethodGet
			req.URL, _ = url.JoinPath(srv.URL, "value/", tt.queryType, tt.queryMetric)
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
			if expectedBody := tt.expectedValue; expectedBody != "" {
				actualBody := resp.String()
				defer func() { _ = resp.RawBody().Close() }()
				require.Equalf(t, expectedBody, actualBody, "expected %s but got %s", expectedBody, actualBody)
			}
		})
	}
}
