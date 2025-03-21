package general

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// MockConnectChecker implements the ConnectChecker interface for testing.
type MockConnectChecker struct {
	ShouldError bool
	Delay       time.Duration
}

// CheckConnection implements the ConnectChecker interface.
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
		checker        ConnectChecker
		name           string
		expectedBody   string
		expectedStatus int
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
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/ping", http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := Ping(tt.checker)
			_ = handler(c)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.expectedBody, rec.Body.String())
		})
	}
}

// Define a dummy ConnectChecker that always succeeds.
type dummyChecker struct{}

// CheckConnection always returns nil.
func (d *dummyChecker) CheckConnection(_ context.Context) error {
	return nil
}

func ExamplePing() {
	// Instantiate the dummy checker.
	checker := &dummyChecker{}

	// Create a new Echo instance.
	e := echo.New()

	// Create a new HTTP GET request and recorder.
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Call the Ping handler with the dummy checker.
	handler := Ping(checker)
	_ = handler(c)

	// Print the response body.
	fmt.Print(rec.Body.String())

	// Output:
	// pong
}
