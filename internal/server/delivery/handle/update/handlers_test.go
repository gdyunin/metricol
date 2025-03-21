package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/delivery/model"
	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockMetricsUpdater implements the MetricsUpdater interface for testing.
type MockMetricsUpdater struct {
	ReturnedMetric *entity.Metric
	Delay          time.Duration
	ShouldFail     bool
}

// PushMetric implements the MetricsUpdater interface.
func (m *MockMetricsUpdater) PushMetric(ctx context.Context, metric *entity.Metric) (*entity.Metric, error) {
	if m.Delay > 0 {
		select {
		case <-time.After(m.Delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.ShouldFail {
		return nil, errors.New("failed to push metric")
	}

	if m.ReturnedMetric != nil {
		return m.ReturnedMetric, nil
	}
	return metric, nil
}

func TestFromJSON(t *testing.T) {
	tests := []struct {
		updater        MetricsUpdater
		name           string
		requestBody    string
		expectedBody   string
		expectedStatus int
		checkJSON      bool
	}{
		{
			name: "Valid counter metric",
			updater: &MockMetricsUpdater{
				ReturnedMetric: &entity.Metric{
					Name:  "test_counter",
					Type:  entity.MetricTypeCounter,
					Value: int64(42),
				},
			},
			requestBody:    `{"id":"test_counter","type":"counter","delta":42}`,
			expectedStatus: http.StatusOK,
			checkJSON:      true,
		},
		{
			name: "Valid gauge metric",
			updater: &MockMetricsUpdater{
				ReturnedMetric: &entity.Metric{
					Name:  "test_gauge",
					Type:  entity.MetricTypeGauge,
					Value: 42.5,
				},
			},
			requestBody:    `{"id":"test_gauge","type":"gauge","value":42.5}`,
			expectedStatus: http.StatusOK,
			checkJSON:      true,
		},
		{
			name:           "Invalid JSON payload",
			updater:        &MockMetricsUpdater{},
			requestBody:    `{"id":"test_invalid",type:"gauge","value":42.5}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON payload provided.",
		},
		{
			name: "PushMetric fails",
			updater: &MockMetricsUpdater{
				ShouldFail: true,
			},
			requestBody:    `{"id":"test_failure","type":"counter","delta":42}`,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError),
		},
		{
			name: "Timeout",
			updater: &MockMetricsUpdater{
				Delay: 6 * time.Second, // Longer than metricUpdateTimeout (5s).
			},
			requestBody:    `{"id":"test_timeout","type":"counter","delta":42}`,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := FromJSON(tt.updater)
			err := handler(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.checkJSON {
				var returnedMetric model.Metric
				err := json.Unmarshal(rec.Body.Bytes(), &returnedMetric)
				require.NoError(t, err)

				updater := tt.updater.(*MockMetricsUpdater)
				assert.Equal(t, updater.ReturnedMetric.Name, returnedMetric.ID)
				assert.Equal(t, updater.ReturnedMetric.Type, returnedMetric.MType)

				switch returnedMetric.MType {
				case entity.MetricTypeCounter:
					require.NotNil(t, returnedMetric.Delta)
					intValue, _ := updater.ReturnedMetric.Value.(int64)
					assert.Equal(t, intValue, *returnedMetric.Delta)
				case entity.MetricTypeGauge:
					require.NotNil(t, returnedMetric.Value)
					floatValue, _ := updater.ReturnedMetric.Value.(float64)
					assert.Equal(t, floatValue, *returnedMetric.Value)
				}
			} else if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, strings.TrimSpace(rec.Body.String()))
			}
		})
	}
}

func TestFromURI(t *testing.T) {
	tests := []struct {
		updater        MetricsUpdater
		setupContext   func(c echo.Context)
		name           string
		expectedBody   string
		expectedStatus int
	}{
		{
			name:    "Valid counter metric",
			updater: &MockMetricsUpdater{},
			setupContext: func(c echo.Context) {
				c.SetParamNames("type", "id", "value")
				c.SetParamValues("counter", "test_counter", "42")
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Metric update successful.",
		},
		{
			name:    "Valid gauge metric",
			updater: &MockMetricsUpdater{},
			setupContext: func(c echo.Context) {
				c.SetParamNames("type", "id", "value")
				c.SetParamValues("gauge", "test_gauge", "42.5")
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Metric update successful.",
		},
		{
			name:    "Missing value parameter",
			updater: &MockMetricsUpdater{},
			setupContext: func(c echo.Context) {
				c.SetParamNames("type", "id")
				c.SetParamValues("counter", "test_counter")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Required 'value' parameter is missing.",
		},
		{
			name:    "Invalid counter value",
			updater: &MockMetricsUpdater{},
			setupContext: func(c echo.Context) {
				c.SetParamNames("type", "id", "value")
				c.SetParamValues("counter", "test_counter", "not_a_number")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Provided counter value is invalid.",
		},
		{
			name:    "Invalid gauge value",
			updater: &MockMetricsUpdater{},
			setupContext: func(c echo.Context) {
				c.SetParamNames("type", "id", "value")
				c.SetParamValues("gauge", "test_gauge", "not_a_number")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Provided gauge value is invalid.",
		},
		{
			name:    "Unsupported metric type",
			updater: &MockMetricsUpdater{},
			setupContext: func(c echo.Context) {
				c.SetParamNames("type", "id", "value")
				c.SetParamValues("unsupported", "test_unsupported", "42")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Unsupported metric type.",
		},
		{
			name: "PushMetric fails",
			updater: &MockMetricsUpdater{
				ShouldFail: true,
			},
			setupContext: func(c echo.Context) {
				c.SetParamNames("type", "id", "value")
				c.SetParamValues("counter", "test_counter", "42")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError),
		},
		{
			name: "Timeout",
			updater: &MockMetricsUpdater{
				Delay: 6 * time.Second, // Longer than metricUpdateTimeout (5s).
			},
			setupContext: func(c echo.Context) {
				c.SetParamNames("type", "id", "value")
				c.SetParamValues("counter", "test_counter", "42")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/update/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.setupContext != nil {
				tt.setupContext(c)
			}

			handler := FromURI(tt.updater)
			err := handler(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus >= 400 && tt.expectedStatus < 600 {
				if strings.HasPrefix(rec.Header().Get(echo.HeaderContentType), echo.MIMEApplicationJSON) {
					var errorResponse map[string]string
					err := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
					if err == nil && errorResponse["message"] != "" {
						assert.Equal(t, tt.expectedBody, errorResponse["message"])
					} else {
						responseBody := strings.TrimSpace(rec.Body.String())
						assert.Contains(t, responseBody, tt.expectedBody)
					}
				} else {
					responseBody := strings.TrimSpace(rec.Body.String())
					if strings.HasPrefix(responseBody, "code=") {
						msgParts := strings.Split(responseBody, "message=")
						if len(msgParts) > 1 {
							assert.Equal(t, tt.expectedBody, msgParts[1])
						} else {
							assert.Equal(t, tt.expectedBody, responseBody)
						}
					} else {
						assert.Equal(t, tt.expectedBody, responseBody)
					}
				}
			} else {
				assert.Equal(t, tt.expectedBody, strings.TrimSpace(rec.Body.String()))
			}

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, echo.MIMETextPlain, rec.Header().Get(echo.HeaderContentType))
			}
		})
	}
}

func TestValidateMetricValue(t *testing.T) {
	tests := []struct {
		metric     *model.Metric
		name       string
		valueStr   string
		errorMsg   string
		errorCode  int
		shouldPass bool
	}{
		{
			name: "Valid counter value",
			metric: &model.Metric{
				ID:    "test_counter",
				MType: entity.MetricTypeCounter,
			},
			valueStr:   "42",
			shouldPass: true,
		},
		{
			name: "Valid gauge value",
			metric: &model.Metric{
				ID:    "test_gauge",
				MType: entity.MetricTypeGauge,
			},
			valueStr:   "42.5",
			shouldPass: true,
		},
		{
			name: "Empty value",
			metric: &model.Metric{
				ID:    "test_empty",
				MType: entity.MetricTypeCounter,
			},
			valueStr:   "",
			shouldPass: false,
			errorCode:  http.StatusBadRequest,
			errorMsg:   "Required 'value' parameter is missing.",
		},
		{
			name: "Invalid counter value",
			metric: &model.Metric{
				ID:    "test_invalid_counter",
				MType: entity.MetricTypeCounter,
			},
			valueStr:   "not_a_number",
			shouldPass: false,
			errorCode:  http.StatusBadRequest,
			errorMsg:   "Provided counter value is invalid.",
		},
		{
			name: "Invalid gauge value",
			metric: &model.Metric{
				ID:    "test_invalid_gauge",
				MType: entity.MetricTypeGauge,
			},
			valueStr:   "not_a_number",
			shouldPass: false,
			errorCode:  http.StatusBadRequest,
			errorMsg:   "Provided gauge value is invalid.",
		},
		{
			name: "Unsupported metric type",
			metric: &model.Metric{
				ID:    "test_unsupported",
				MType: "unsupported",
			},
			valueStr:   "42",
			shouldPass: false,
			errorCode:  http.StatusBadRequest,
			errorMsg:   "Unsupported metric type.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMetricValue(tt.metric, tt.valueStr)

			if tt.shouldPass {
				assert.NoError(t, err)

				if tt.metric.MType == entity.MetricTypeCounter {
					require.NotNil(t, tt.metric.Delta)
					expectedDelta, _ := strconv.ParseInt(tt.valueStr, 10, 64)
					assert.Equal(t, expectedDelta, *tt.metric.Delta)
				} else if tt.metric.MType == entity.MetricTypeGauge {
					require.NotNil(t, tt.metric.Value)
					expectedValue, _ := strconv.ParseFloat(tt.valueStr, 64)
					assert.Equal(t, expectedValue, *tt.metric.Value)
				}
			} else {
				require.Error(t, err)
				var httpErr *echo.HTTPError
				ok := errors.As(err, &httpErr)
				require.True(t, ok, "Expected *echo.HTTPError, got %T", err)
				assert.Equal(t, tt.errorCode, httpErr.Code)
				assert.Equal(t, tt.errorMsg, httpErr.Message)
			}
		})
	}
}

// dummyUpdater is a simple implementation of MetricsUpdater that returns the metric unchanged.
type dummyUpdater struct{}

// PushMetric returns the input metric without modification.
func (d *dummyUpdater) PushMetric(_ context.Context, m *entity.Metric) (*entity.Metric, error) {
	return m, nil
}

func ExampleFromJSON() {
	updater := &dummyUpdater{}

	// Create a new Echo instance.
	e := echo.New()

	// Prepare a POST request with a JSON payload representing a counter metric.
	req := httptest.NewRequest(
		http.MethodPost,
		"/update",
		strings.NewReader(`{"id":"test_counter","type":"counter","delta":42}`),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Invoke the FromJSON handler.
	handler := FromJSON(updater)
	if err := handler(c); err != nil {
		panic(err)
	}

	// Print the response body.
	// Expected output: {"delta":42,"id":"test_counter","type":"counter"}
	// Note: JSON key order may vary.
	fmt.Print(rec.Body.String())

	// Output:
	// {"delta":42,"id":"test_counter","type":"counter"}
}

func ExampleFromURI() {
	updater := &dummyUpdater{}

	// Create a new Echo instance.
	e := echo.New()

	// Prepare a POST request for URI-based metric update.
	req := httptest.NewRequest(http.MethodPost, "/update/counter/test_counter/42", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Set URI parameters: type, id, and value.
	c.SetParamNames("type", "id", "value")
	c.SetParamValues("counter", "test_counter", "42")

	// Invoke the FromURI handler.
	handler := FromURI(updater)
	if err := handler(c); err != nil {
		panic(err)
	}

	// Print the response body.
	// Expected output: Metric update successful.
	fmt.Print(rec.Body.String())

	// Output:
	// Metric update successful.
}
