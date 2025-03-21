package value

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gdyunin/metricol.git/internal/server/internal/controller"
	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMetricsPuller is a mock implementation of the MetricsPuller interface.
type MockMetricsPuller struct {
	mock.Mock
}

// Pull implements the MetricsPuller interface.
func (m *MockMetricsPuller) Pull(
	ctx context.Context,
	metricType string,
	name string,
) (*entity.Metric, error) {
	args := m.Called(ctx, metricType, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Metric), args.Error(1)
}

func normalizeErrorResponse(body string) string {
	if strings.Contains(body, "code=") {
		parts := strings.SplitN(body, ", message=", 2)
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	}
	return strings.TrimSpace(body)
}

func TestFromJSON(t *testing.T) {
	tests := []struct {
		mockSetup      func(*MockMetricsPuller)
		name           string
		requestBody    string
		expectedBody   string
		expectedStatus int
		validateJSON   bool
	}{
		{
			name:        "Metric not found",
			requestBody: `{"id":"non_existent","type":"counter"}`,
			mockSetup: func(m *MockMetricsPuller) {
				m.On("Pull", mock.Anything, "counter", "non_existent").
					Return(nil, controller.ErrNotFoundInRepository)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Metric not found in the repository.",
			validateJSON:   false,
		},
		{
			name:        "Repository error",
			requestBody: `{"id":"error_metric","type":"counter"}`,
			mockSetup: func(m *MockMetricsPuller) {
				m.On("Pull", mock.Anything, "counter", "error_metric").
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError),
			validateJSON:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			mockPuller := new(MockMetricsPuller)
			tt.mockSetup(mockPuller)

			req := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := FromJSON(mockPuller)
			err := handler(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			actualBody := normalizeErrorResponse(rec.Body.String())
			assert.Equal(t, tt.expectedBody, actualBody)

			mockPuller.AssertExpectations(t)
		})
	}
}

func TestFromURI(t *testing.T) {
	tests := []struct {
		paramSetup     func(*echo.Echo, *http.Request) echo.Context
		mockSetup      func(*MockMetricsPuller)
		name           string
		expectedBody   string
		expectedStatus int
	}{
		{
			name: "Metric not found",
			paramSetup: func(e *echo.Echo, req *http.Request) echo.Context {
				c := e.NewContext(req, httptest.NewRecorder())
				c.SetParamNames("type", "id")
				c.SetParamValues("counter", "non_existent")
				return c
			},
			mockSetup: func(m *MockMetricsPuller) {
				m.On("Pull", mock.Anything, "counter", "non_existent").
					Return(nil, controller.ErrNotFoundInRepository)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Metric not found in the repository.",
		},
		{
			name: "Repository error",
			paramSetup: func(e *echo.Echo, req *http.Request) echo.Context {
				c := e.NewContext(req, httptest.NewRecorder())
				c.SetParamNames("type", "id")
				c.SetParamValues("counter", "error_metric")
				return c
			},
			mockSetup: func(m *MockMetricsPuller) {
				m.On("Pull", mock.Anything, "counter", "error_metric").
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			mockPuller := new(MockMetricsPuller)
			tt.mockSetup(mockPuller)

			req := httptest.NewRequest(http.MethodGet, "/value/:type/:id", nil)
			c := tt.paramSetup(e, req)
			rec := c.Response().Writer.(*httptest.ResponseRecorder)

			handler := FromURI(mockPuller)
			err := handler(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			actualBody := normalizeErrorResponse(rec.Body.String())
			assert.Equal(t, tt.expectedBody, actualBody)

			mockPuller.AssertExpectations(t)
		})
	}
}

// dummyPuller implements the MetricsPuller interface.
type dummyPuller struct {
	Metric *entity.Metric
}

// Pull returns a dummy counter metric with a value of 100.
func (d *dummyPuller) Pull(_ context.Context, _, _ string) (*entity.Metric, error) {
	return d.Metric, nil
}

func ExampleFromJSON() {
	puller := &dummyPuller{
		Metric: &entity.Metric{
			Name:  "test_counter",
			Type:  entity.MetricTypeCounter,
			Value: int64(100),
		},
	}

	e := echo.New()
	// Create a POST request with a JSON payload representing a counter metric.
	req := httptest.NewRequest(
		http.MethodPost,
		"/value",
		strings.NewReader(`{"id":"test_counter","type":"counter"}`),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Invoke the FromJSON handler.
	handler := FromJSON(puller)
	if err := handler(c); err != nil {
		panic(err)
	}

	// Print the JSON response.
	// Expected output: {"delta":100,"id":"test_counter","type":"counter"}
	fmt.Print(rec.Body.String())

	// Output:
	// {"delta":100,"id":"test_counter","type":"counter"}
}

func ExampleFromURI() {
	puller := &dummyPuller{
		Metric: &entity.Metric{
			Name:  "test_counter",
			Type:  entity.MetricTypeCounter,
			Value: int64(200),
		},
	}

	e := echo.New()
	// Create a GET request; the body is not needed.
	req := httptest.NewRequest(http.MethodGet, "/value/counter/test_counter", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Set the URI parameters: metric type and id.
	c.SetParamNames("type", "id")
	c.SetParamValues("counter", "test_counter")

	// Invoke the FromURI handler.
	handler := FromURI(puller)
	if err := handler(c); err != nil {
		panic(err)
	}

	// Print the plain text response, which should be "200".
	fmt.Print(rec.Body.String())

	// Output:
	// 200
}
