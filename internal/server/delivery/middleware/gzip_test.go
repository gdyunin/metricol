package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func emptyGzipFooter() string {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_ = w.Close()
	return buf.String()
}

func TestGzipMiddleware(t *testing.T) {
	logger := zap.NewNop().Sugar()
	cases := []struct {
		name               string
		acceptEncoding     string
		contentType        string
		responseBody       string
		expectGzip         bool
		expectedStatusCode int
	}{
		{
			name:               "No Accept-Encoding gzip",
			acceptEncoding:     "deflate",
			contentType:        "application/json",
			responseBody:       "plain response",
			expectGzip:         false,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Gzip accepted but content type not matched",
			acceptEncoding:     "gzip",
			contentType:        "application/xml",
			responseBody:       "plain xml response",
			expectGzip:         false,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Gzip accepted with application/json content type",
			acceptEncoding:     "gzip",
			contentType:        "application/json",
			responseBody:       "json response",
			expectGzip:         true,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Gzip accepted with text/html content type",
			acceptEncoding:     "gzip",
			contentType:        "text/html; charset=utf-8",
			responseBody:       "html response",
			expectGzip:         true,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Gzip accepted but no content type set",
			acceptEncoding:     "gzip",
			contentType:        "",
			responseBody:       "no content type response",
			expectGzip:         false,
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			req.Header.Set(echo.HeaderAcceptEncoding, tc.acceptEncoding)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			nextHandler := func(c echo.Context) error {
				c.Response().Header().Set(echo.HeaderContentType, tc.contentType)
				_, err := c.Response().Write([]byte(tc.responseBody))
				return err
			}
			middleware := Gzip(logger)
			handler := middleware(nextHandler)
			err := handler(c)
			if err != nil {
				t.Fatalf("handler returned error: %v", err)
			}
			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			if tc.expectGzip {
				ce := rec.Header().Get(echo.HeaderContentEncoding)
				if ce != gzipHeaderValue {
					t.Errorf("expected Content-Encoding header %q, got %q", gzipHeaderValue, ce)
				}
				gr, err := gzip.NewReader(rec.Body)
				if err != nil {
					t.Fatalf("failed to create gzip reader: %v", err)
				}
				decompressed, err := io.ReadAll(gr)
				if err != nil {
					t.Fatalf("failed to read decompressed body: %v", err)
				}
				_ = gr.Close()
				assert.Equal(t, tc.responseBody, string(decompressed))
			} else {
				ce := rec.Header().Get(echo.HeaderContentEncoding)
				if ce != "" {
					t.Errorf("expected no Content-Encoding header, got %q", ce)
				}
				var expected string
				if strings.Contains(tc.acceptEncoding, "gzip") {
					expected = tc.responseBody + emptyGzipFooter()
				} else {
					expected = tc.responseBody
				}
				assert.Equal(t, expected, rec.Body.String())
			}
		})
	}
}

type errorResponseWriter struct{}

func (e *errorResponseWriter) Header() http.Header         { return http.Header{} }
func (e *errorResponseWriter) Write(_ []byte) (int, error) { return 0, fmt.Errorf("write error") }
func (e *errorResponseWriter) WriteHeader(_ int)           {}

type errorWriter struct{}

func (ew errorWriter) Write(_ []byte) (int, error) { return 0, fmt.Errorf("error from errorWriter") }

func TestGzipWriterWrite(t *testing.T) {
	data := []byte("test data")

	{
		rr := httptest.NewRecorder()
		gw := gzip.NewWriter(rr)
		writer := &gzipWriter{
			ResponseWriter: rr,
			gzipWriter:     gw,
			withGzip:       true,
		}
		n, err := writer.Write(data)
		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		err = writer.gzipWriter.Close()
		assert.NoError(t, err)
		gr, err := gzip.NewReader(rr.Body)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		decompressed, err := io.ReadAll(gr)
		if err != nil {
			t.Fatalf("failed to read decompressed data: %v", err)
		}
		_ = gr.Close()
		assert.Equal(t, string(data), string(decompressed))
	}

	{
		rr := httptest.NewRecorder()
		writer := &gzipWriter{
			ResponseWriter: rr,
			withGzip:       false,
		}
		n, err := writer.Write(data)
		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Equal(t, string(data), rr.Body.String())
	}

	{
		writer := &gzipWriter{
			ResponseWriter: &errorResponseWriter{},
			withGzip:       false,
		}
		_, err := writer.Write(data)
		if err == nil || !strings.Contains(err.Error(), "write error") {
			t.Errorf("expected error from underlying writer, got %v", err)
		}
	}

	{
		rr := httptest.NewRecorder()
		gw := gzip.NewWriter(errorWriter{})
		writer := &gzipWriter{
			ResponseWriter: rr,
			gzipWriter:     gw,
			withGzip:       true,
		}
		dataLarge := bytes.Repeat([]byte("A"), 300)
		_, err := writer.Write(dataLarge)
		if err == nil || !strings.Contains(err.Error(), "error from errorWriter") {
			t.Errorf("expected error from gzip writer, got %v", err)
		}
	}
}
