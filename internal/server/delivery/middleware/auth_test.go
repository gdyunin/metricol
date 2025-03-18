package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// errorReadCloser simulates a reader that always returns an error.
type errorReadCloser struct{}

func (e errorReadCloser) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}
func (e errorReadCloser) Close() error {
	return nil
}

// computeValidSign is a helper that computes the valid base64-encoded HMAC-SHA256 signature.
func computeValidSign(body []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// TestAuthMiddleware tests the Auth middleware using table-driven tests.
func TestAuthMiddleware(t *testing.T) {
	cases := []struct {
		name               string
		key                string
		requestBody        string
		simulateBodyError  bool
		headerSign         string
		nextHandlerStatus  int // status code returned by the next handler
		expectedStatusCode int // final expected HTTP status code
		expectedBody       string
	}{
		{
			name:               "Empty key, no signature validation",
			key:                "",
			requestBody:        "test body",
			simulateBodyError:  false,
			headerSign:         "", // header is ignored when key is empty
			nextHandlerStatus:  http.StatusOK,
			expectedStatusCode: http.StatusOK,
			expectedBody:       "next",
		},
		{
			name:               "No signature header provided",
			key:                "secret",
			requestBody:        "test body",
			simulateBodyError:  false,
			headerSign:         "",
			nextHandlerStatus:  http.StatusOK,
			expectedStatusCode: http.StatusOK,
			expectedBody:       "next",
		},
		{
			name:               "Valid signature provided",
			key:                "secret",
			requestBody:        "hello",
			simulateBodyError:  false,
			headerSign:         computeValidSign([]byte("hello"), "secret"),
			nextHandlerStatus:  http.StatusOK,
			expectedStatusCode: http.StatusOK,
			expectedBody:       "next",
		},
		{
			name:               "Invalid signature provided",
			key:                "secret",
			requestBody:        "hello",
			simulateBodyError:  false,
			headerSign:         "invalid-sign", // not a valid base64 string, so decoding fails
			nextHandlerStatus:  http.StatusOK,  // next is not called due to signature failure
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       http.StatusText(http.StatusBadRequest),
		},
		{
			name:               "Error reading body",
			key:                "secret",
			requestBody:        "ignored", // body content is ignored because of simulated error
			simulateBodyError:  true,
			headerSign:         "any",
			nextHandlerStatus:  http.StatusOK,
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       http.StatusText(http.StatusInternalServerError),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			var req *http.Request

			if tc.simulateBodyError {
				// simulate an error while reading body
				req = httptest.NewRequest(http.MethodPost, "/", nil)
				req.Body = errorReadCloser{}
			} else {
				req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.requestBody))
			}

			if tc.headerSign != "" {
				req.Header.Set("HashSHA256", tc.headerSign)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Use a flag to verify if the next handler is called.
			nextCalled := false
			nextHandler := func(c echo.Context) error {
				nextCalled = true
				return c.String(tc.nextHandlerStatus, "next")
			}

			// Wrap the next handler with the Auth middleware.
			middleware := Auth(tc.key)
			handler := middleware(nextHandler)
			err := handler(c)
			if err != nil {
				t.Errorf("handler returned an error: %v", err)
			}

			// Verify the response.
			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			assert.Equal(t, tc.expectedBody, rec.Body.String())

			// If the expected status is that of the next handler, ensure it was indeed called.
			if tc.expectedStatusCode == tc.nextHandlerStatus {
				assert.True(t, nextCalled, "expected next handler to be called")
			}
		})
	}
}

// TestCheckSign tests the checkSign helper function.
func TestCheckSign(t *testing.T) {
	cases := []struct {
		name     string
		body     []byte
		key      string
		sign     string
		expected bool
	}{
		{
			name:     "Valid signature",
			body:     []byte("hello"),
			key:      "secret",
			sign:     computeValidSign([]byte("hello"), "secret"),
			expected: true,
		},
		{
			name:     "Invalid signature due to wrong key",
			body:     []byte("hello"),
			key:      "wrong",
			sign:     computeValidSign([]byte("hello"), "secret"),
			expected: false,
		},
		{
			name:     "Malformed base64 signature",
			body:     []byte("hello"),
			key:      "secret",
			sign:     "not_base64!",
			expected: false,
		},
		{
			name:     "Valid signature with different data",
			body:     []byte("test data"),
			key:      "anotherkey",
			sign:     computeValidSign([]byte("test data"), "anotherkey"),
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := checkSign(tc.body, tc.sign, tc.key)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestGetRawBody tests the getRawBody helper function.
func TestGetRawBody(t *testing.T) {
	cases := []struct {
		name           string
		body           io.ReadCloser
		expectedBody   []byte
		expectingError bool
	}{
		{
			name:           "Normal body",
			body:           io.NopCloser(bytes.NewBufferString("normal body")),
			expectedBody:   []byte("normal body"),
			expectingError: false,
		},
		{
			name:           "Error reading body",
			body:           errorReadCloser{},
			expectedBody:   nil,
			expectingError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Body = tc.body

			raw, err := getRawBody(req)
			if tc.expectingError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBody, raw)

				// Also check that the request.Body was reset and can be read again.
				bodyAfter, err := io.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBody, bodyAfter)
			}
		})
	}
}
