package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		nextHandler echo.HandlerFunc
		expectError bool
	}{
		{
			name: "normal handler",
			nextHandler: func(c echo.Context) error {
				c.Response().Header().Set("Content-Type", "application/json")
				c.Response().Status = 200
				c.Response().Size = 123
				return nil
			},
			expectError: false,
		},
		{
			name: "error handler",
			nextHandler: func(c echo.Context) error {
				return fmt.Errorf("handler error")
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			core, obs := observer.New(zap.InfoLevel)
			logger := zap.New(core).Sugar()
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Response().Header().Set(echo.HeaderXRequestID, "test-req-id")
			middleware := Log(logger)
			handler := middleware(tc.nextHandler)
			err := handler(c)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			logs := obs.All()
			assert.Len(t, logs, 2)
			for _, entry := range logs {
				assert.Contains(t, entry.Message, "HTTP request_id=test-req-id")
			}
		})
	}
}
