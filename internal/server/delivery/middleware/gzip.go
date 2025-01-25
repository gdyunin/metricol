package middleware

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const gzipHeaderValue = "gzip"

var contentTypesForGzip = []string{
	"application/json",
	"text/html",
}

// Gzip provides a middleware for Echo that compresses responses using gzip if the client supports it.
func Gzip(logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			acceptEncoding := c.Request().Header.Get(echo.HeaderAcceptEncoding)

			originalWriter := c.Response().Writer
			newWriter := gzip.NewWriter(originalWriter)
			defer func() {
				if closeErr := newWriter.Close(); closeErr != nil {
					logger.Warnf("Error closing gzip writer: %v", closeErr)
				}
			}()

			if strings.Contains(acceptEncoding, gzipHeaderValue) {
				writer := &gzipWriter{
					ResponseWriter: originalWriter,
					gzipWriter:     newWriter,
				}

				c.Response().Writer = writer

				c.Response().Before(func() {
					contentType := c.Response().Header().Get(echo.HeaderContentType)
					for _, ct := range contentTypesForGzip {
						if strings.HasPrefix(contentType, ct) {
							c.Response().Header().Set(echo.HeaderContentEncoding, gzipHeaderValue)
							writer.withGzip = true
							break
						}
					}
				})
			}

			return next(c)
		}
	}
}

// gzipWriter is a wrapper around http.ResponseWriter that supports gzip compression.
type gzipWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
	withGzip   bool
}

// Write writes data to the response, compressing it if gzip is enabled.
func (w *gzipWriter) Write(data []byte) (int, error) {
	if w.withGzip {
		n, err := w.gzipWriter.Write(data)
		if err != nil {
			return n, fmt.Errorf("error writing data with gzip: %w", err)
		}
		return n, nil
	}

	n, err := w.ResponseWriter.Write(data)
	if err != nil {
		return n, fmt.Errorf("error writing data without gzip: %w", err)
	}
	return n, nil
}
