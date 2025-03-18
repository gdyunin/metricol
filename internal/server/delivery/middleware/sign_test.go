package middleware

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gdyunin/metricol.git/pkg/sign"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSignMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		responseBody   string
		expectedHeader string
	}{
		{
			name:           "No key provided",
			key:            "",
			responseBody:   "test body",
			expectedHeader: "",
		},
		{
			name:           "Key provided",
			key:            "secret",
			responseBody:   "test body",
			expectedHeader: hex.EncodeToString(sign.MakeSign([]byte("test body"), "secret")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mw := Sign(tt.key)

			// Handler to be wrapped
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, tt.responseBody)
			}

			// Execute the middleware
			_ = mw(handler)(c)

			// Assertions
			if tt.expectedHeader != "" {
				assert.Equal(
					t,
					tt.expectedHeader,
					rec.Header().Get("HashSHA256"),
					"Expected and actual hash do not match",
				)
			} else {
				assert.Empty(t, rec.Header().Get("HashSHA256"))
			}
			assert.Equal(t, tt.responseBody, rec.Body.String())
		})
	}
}
