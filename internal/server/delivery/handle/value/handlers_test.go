package value

import (
	"context"
	"errors"
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
