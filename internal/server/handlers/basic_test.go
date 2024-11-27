package handlers

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestBadRequest(t *testing.T) {
	handler := http.HandlerFunc(BadRequest)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	tests := []struct {
		method       string
		expectedCode int
		expectedBody string
	}{
		{method: http.MethodGet, expectedCode: http.StatusBadRequest, expectedBody: ""},
		{method: http.MethodPut, expectedCode: http.StatusBadRequest, expectedBody: ""},
		{method: http.MethodDelete, expectedCode: http.StatusBadRequest, expectedBody: ""},
		{method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: ""},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			// Send request
			req := resty.New().R()
			req.Method = tt.method
			req.URL = srv.URL
			resp, err := req.Send()

			// Want no errors
			require.NoError(t, err, "error making request")

			// Want status code
			expectedCode := tt.expectedCode
			actualCode := resp.StatusCode()
			require.Equalf(t, expectedCode, actualCode, "expected response code %q, but got %q", tt.expectedCode, actualCode)
		})
	}
}

func TestNotFound(t *testing.T) {
	handler := http.HandlerFunc(NotFound)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	tests := []struct {
		method       string
		expectedCode int
		expectedBody string
	}{
		{method: http.MethodGet, expectedCode: http.StatusNotFound, expectedBody: ""},
		{method: http.MethodPut, expectedCode: http.StatusNotFound, expectedBody: ""},
		{method: http.MethodDelete, expectedCode: http.StatusNotFound, expectedBody: ""},
		{method: http.MethodPost, expectedCode: http.StatusNotFound, expectedBody: ""},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			// Send request
			req := resty.New().R()
			req.Method = tt.method
			req.URL = srv.URL
			resp, err := req.Send()

			// Want no errors
			require.NoErrorf(t, err, "error making request")

			// Want status code
			expectedCode := tt.expectedCode
			actualCode := resp.StatusCode()
			require.Equalf(t, expectedCode, actualCode, "expected response code %s, but got %s",
				strconv.Itoa(expectedCode), strconv.Itoa(actualCode))
		})
	}
}
