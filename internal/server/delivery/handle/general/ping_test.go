package general

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// MockConnectChecker implements the ConnectChecker interface for testing
type MockConnectChecker struct {
	ShouldError bool
	Delay       time.Duration
}

// CheckConnection implements the ConnectChecker interface
func (m *MockConnectChecker) CheckConnection(ctx context.Context) error {
	if m.Delay > 0 {
		select {
		case <-time.After(m.Delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if m.ShouldError {
		return errors.New("connection check failed")
	}
	return nil
}

func TestPing(t *testing.T) {
	tests := []struct {
		name           string
		checker        ConnectChecker
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			checker:        &MockConnectChecker{ShouldError: false},
			expectedStatus: http.StatusOK,
			expectedBody:   "pong",
		},
		{
			name:           "Connection Failure",
			checker:        &MockConnectChecker{ShouldError: true},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError),
		},
		{
			name:           "Timeout",
			checker:        &MockConnectChecker{Delay: 4 * time.Second},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError),
		},
		{
			name:           "Near Timeout",
			checker:        &MockConnectChecker{Delay: 2 * time.Second},
			expectedStatus: http.StatusOK,
			expectedBody:   "pong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute the handler
			handler := Ping(tt.checker)
			_ = handler(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.expectedBody, rec.Body.String())
		})
	}
}
