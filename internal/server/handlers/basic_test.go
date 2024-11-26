package handlers

import (
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
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
			req := resty.New().R()
			req.Method = tt.method
			req.URL = srv.URL

			resp, err := req.Send()

			require.NoError(t, err, "error making request")
			require.Equal(t, tt.expectedCode, resp.StatusCode(), "response code didn't match expected")
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
			req := resty.New().R()
			req.Method = tt.method
			req.URL = srv.URL

			resp, err := req.Send()

			require.NoError(t, err, "error making request")
			require.Equal(t, tt.expectedCode, resp.StatusCode(), "response code didn't match expected")
		})
	}
}
