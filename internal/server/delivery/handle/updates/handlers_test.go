package updates

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gdyunin/metricol.git/internal/server/delivery/model"
	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMetricsUpdater is a mock implementation of the MetricsUpdater interface.
type MockMetricsUpdater struct {
	mock.Mock
}

// PushMetrics implements the MetricsUpdater interface.
func (m *MockMetricsUpdater) PushMetrics(
	ctx context.Context,
	metrics *entity.Metrics,
) (*entity.Metrics, error) {
	args := m.Called(ctx, metrics)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Metrics), args.Error(1)
}

func TestFromJSON(t *testing.T) {
	tests := []struct {
		mockSetup      func(*MockMetricsUpdater)
		name           string
		requestBody    string
		expectedBody   string
		expectedStatus int
		validateJSON   bool
	}{
		{
			name:        "Valid counter metric",
			requestBody: `[{"id":"test_counter","type":"counter","delta":5}]`,
			mockSetup: func(m *MockMetricsUpdater) {
				expectedMetrics := &entity.Metrics{
					{
						Name:  "test_counter",
						Type:  "counter",
						Value: int64(5),
					},
				}
				m.On("PushMetrics", mock.Anything, mock.MatchedBy(func(metrics *entity.Metrics) bool {
					if len(*metrics) != 1 {
						return false
					}
					metric := (*metrics)[0]
					return metric.Name == "test_counter" && metric.Type == "counter" &&
						metric.Value == int64(5)
				})).Return(expectedMetrics, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":"test_counter","type":"counter","delta":5}]`,
			validateJSON:   true,
		},
		{
			name:        "Valid gauge metric",
			requestBody: `[{"id":"test_gauge","type":"gauge","value":3.14}]`,
			mockSetup: func(m *MockMetricsUpdater) {
				expectedMetrics := &entity.Metrics{
					{
						Name:  "test_gauge",
						Type:  "gauge",
						Value: 3.14,
					},
				}
				m.On("PushMetrics", mock.Anything, mock.MatchedBy(func(metrics *entity.Metrics) bool {
					if len(*metrics) != 1 {
						return false
					}
					metric := (*metrics)[0]
					return metric.Name == "test_gauge" && metric.Type == "gauge" && metric.Value == 3.14
				})).Return(expectedMetrics, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":"test_gauge","type":"gauge","value":3.14}]`,
			validateJSON:   true,
		},
		{
			name:        "Multiple metrics",
			requestBody: `[{"id":"counter1","type":"counter","delta":1},{"id":"gauge1","type":"gauge","value":2.5}]`,
			mockSetup: func(m *MockMetricsUpdater) {
				expectedMetrics := &entity.Metrics{
					{
						Name:  "counter1",
						Type:  "counter",
						Value: int64(1),
					},
					{
						Name:  "gauge1",
						Type:  "gauge",
						Value: 2.5,
					},
				}
				m.On("PushMetrics", mock.Anything, mock.Anything).Return(expectedMetrics, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":"counter1","type":"counter","delta":1},{"id":"gauge1","type":"gauge","value":2.5}]`,
			validateJSON:   true,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `[{"id":"test_counter","type":"counter","delta":5`,
			mockSetup:      func(m *MockMetricsUpdater) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid parameters provided in the request.",
			validateJSON:   false,
		},
		{
			name:           "Empty ID",
			requestBody:    `[{"id":"","type":"counter","delta":5}]`,
			mockSetup:      func(m *MockMetricsUpdater) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid parameters provided in the request.",
			validateJSON:   false,
		},
		{
			name:           "Empty type",
			requestBody:    `[{"id":"test_counter","type":"","delta":5}]`,
			mockSetup:      func(m *MockMetricsUpdater) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid parameters provided in the request.",
			validateJSON:   false,
		},
		{
			name:           "Missing value fields",
			requestBody:    `[{"id":"test_counter","type":"counter"}]`,
			mockSetup:      func(m *MockMetricsUpdater) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid parameters provided in the request.",
			validateJSON:   false,
		},
		{
			name:        "PushMetrics returns error",
			requestBody: `[{"id":"test_counter","type":"counter","delta":5}]`,
			mockSetup: func(m *MockMetricsUpdater) {
				m.On("PushMetrics", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError),
			validateJSON:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			mockUpdater := new(MockMetricsUpdater)
			tt.mockSetup(mockUpdater)

			req := httptest.NewRequest(http.MethodPost, "/metrics", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)

			handler := FromJSON(mockUpdater)
			err := handler(c)

			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.validateJSON {
				var expectedMetrics model.Metrics
				var actualMetrics model.Metrics

				err = json.Unmarshal([]byte(tt.expectedBody), &expectedMetrics)
				assert.NoError(t, err)

				err = json.Unmarshal(rec.Body.Bytes(), &actualMetrics)
				assert.NoError(t, err)

				assert.Equal(t, len(expectedMetrics), len(actualMetrics))

				for i, expected := range expectedMetrics {
					actual := actualMetrics[i]
					assert.Equal(t, expected.ID, actual.ID)
					assert.Equal(t, expected.MType, actual.MType)

					if expected.Delta != nil {
						assert.NotNil(t, actual.Delta)
						assert.Equal(t, *expected.Delta, *actual.Delta)
					}

					if expected.Value != nil {
						assert.NotNil(t, actual.Value)
						assert.Equal(t, *expected.Value, *actual.Value)
					}
				}
			} else {
				assert.Equal(t, tt.expectedBody, strings.TrimSpace(rec.Body.String()))
			}

			mockUpdater.AssertExpectations(t)
		})
	}
}

func TestIsValidMetric(t *testing.T) {
	tests := []struct {
		metric   *model.Metric
		name     string
		expected bool
	}{
		{
			name: "Valid counter metric",
			metric: &model.Metric{
				ID:    "test_counter",
				MType: "counter",
				Delta: func() *int64 { val := int64(5); return &val }(),
			},
			expected: true,
		},
		{
			name: "Valid gauge metric",
			metric: &model.Metric{
				ID:    "test_gauge",
				MType: "gauge",
				Value: func() *float64 { val := 3.14; return &val }(),
			},
			expected: true,
		},
		{
			name: "Empty ID",
			metric: &model.Metric{
				ID:    "",
				MType: "counter",
				Delta: func() *int64 { val := int64(5); return &val }(),
			},
			expected: false,
		},
		{
			name: "Empty type",
			metric: &model.Metric{
				ID:    "test_counter",
				MType: "",
				Delta: func() *int64 { val := int64(5); return &val }(),
			},
			expected: false,
		},
		{
			name: "Missing value fields",
			metric: &model.Metric{
				ID:    "test_counter",
				MType: "counter",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidMetric(tt.metric)
			assert.Equal(t, tt.expected, result)
		})
	}
}
