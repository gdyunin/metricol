package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

var (
	emptyRepo        *storage.Store
	singleMetricRepo *storage.Store
	multiMetricRepo  *storage.Store
)

func initTestRepos() {
	if emptyRepo != nil && singleMetricRepo != nil && multiMetricRepo != nil {
		return
	}

	emptyRepo = storage.NewStore()

	singleMetricRepo = storage.NewStore()
	_ = singleMetricRepo.PushMetric(metrics.NewCounter("test_counter", 42))

	multiMetricRepo = storage.NewStore()
	_ = multiMetricRepo.PushMetric(metrics.NewCounter("test_counter", 42))
	_ = multiMetricRepo.PushMetric(metrics.NewCounter("test_counter2", 84))
	_ = multiMetricRepo.PushMetric(metrics.NewGauge("test_gauge", 3.14))
}

func TestMainPageHandler(t *testing.T) {
	cwd, _ := os.Getwd()
	_ = os.Chdir("../../..")
	defer func() { _ = os.Chdir(cwd) }()
	initTestRepos()
	tests := []struct {
		name                string
		repository          *storage.Store
		expectedCode        int
		expectedContentType string
		expectedBody        string
	}{
		{
			name:                "Single metric",
			repository:          singleMetricRepo,
			expectedCode:        http.StatusOK,
			expectedContentType: "text/html",
			expectedBody:        "metric1",
		},
		{
			name:                "Multiple metrics",
			repository:          multiMetricRepo,
			expectedCode:        http.StatusOK,
			expectedContentType: "text/html",
			expectedBody:        "metric1",
		},
		{
			name:                "No metrics",
			repository:          emptyRepo,
			expectedCode:        http.StatusOK,
			expectedContentType: "text/html",
			expectedBody:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run test server.
			handler := MainPageHandler(tt.repository)
			router := chi.NewRouter()
			router.Get("/", handler)
			srv := httptest.NewServer(router)
			defer srv.Close()

			// Send request.
			req := resty.New().R()
			req.Method = http.MethodGet
			req.URL, _ = url.JoinPath(srv.URL, "/")
			resp, err := req.Send()

			// Want no errors
			require.NoError(t, err, "error making request")

			// Want status code
			expectedCode := tt.expectedCode
			actualCode := resp.StatusCode()
			require.Equalf(t, expectedCode, actualCode, "expected response code %s, but got %s",
				strconv.Itoa(expectedCode), strconv.Itoa(actualCode))

			// Want Content-Type
			expectedContentType := tt.expectedContentType
			actualContentType := resp.Header().Get("Content-Type")
			require.Equalf(t, expectedContentType, actualContentType, "expected Content-Type %s, but got %s",
				expectedContentType, actualContentType)

			// Want body
			actualBody := resp.String()
			defer func() { _ = resp.RawBody().Close() }()
			for _, metricType := range tt.repository.Metrics() {
				for name, value := range metricType {
					require.Containsf(
						t,
						actualBody,
						fmt.Sprintf(`%s</td><td>%s`, name, value),
						"expected metric %s with value %s don`t exist in got body %s", name, value, actualBody,
					)
				}
			}
		})
	}
}

func TestFillMetricsTable(t *testing.T) {
	initTestRepos()
	tests := []struct {
		name       string
		repository *storage.Store
		expected   []TableRow
	}{
		{
			name:       "Single metric",
			repository: singleMetricRepo,
			expected: []TableRow{
				{Name: "test_counter", Value: "42"},
			},
		},
		{
			name:       "Multiple metrics",
			repository: multiMetricRepo,
			expected: []TableRow{
				{Name: "test_counter", Value: "42"},
				{Name: "test_counter2", Value: "84"},
				{Name: "test_gauge", Value: "3.14"},
			},
		},
		{
			name:       "No metrics",
			repository: emptyRepo,
			expected:   []TableRow{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fillMetricsTable(tt.repository)

			require.Equal(t, tt.expected, result)
		})
	}
}

func TestParseTemplate(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "Valid template path",
			path: "web/template/main_page.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cwd, _ := os.Getwd()
			_ = os.Chdir("../../..")
			defer func() { _ = os.Chdir(cwd) }()

			template, err := parseTemplate(tt.path)
			require.Nil(t, err)
			require.NotNil(t, template)

			cachedTemplate = nil // Reset cache
		})
	}
}
