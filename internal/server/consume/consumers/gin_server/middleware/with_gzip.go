package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// gzipWriter is a custom response writer that compresses HTTP responses using gzip.
// It wraps the original Gin ResponseWriter and an underlying gzip.Writer.
type gzipWriter struct {
	gin.ResponseWriter           // Original Gin response writer.
	Writer             io.Writer // Gzip writer for compressing the response data.
}

// Write writes compressed data to the underlying gzip writer.
// Returns the number of bytes written and any errors encountered.
func (w *gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// WithGzip applies gzip compression to HTTP responses if the client supports it.
// It checks the "Accept-Encoding" header to determine if gzip is supported.
// If an error occurs during compression setup, an appropriate HTTP error response is sent.
func WithGzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptEncoding := c.GetHeader("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") {
			gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
			if err != nil {
				// Log and return a 500 Internal Server Error if gzip writer creation fails.
				c.String(http.StatusInternalServerError, "Failed to create gzip writer: %s", err.Error())
				return
			}
			defer func() { _ = gz.Close() }()

			// Replace the response writer with the gzip writer and set the appropriate header.
			c.Writer = &gzipWriter{ResponseWriter: c.Writer, Writer: gz}
			c.Header("Content-Encoding", "gzip")
		}

		contentEncoding := c.GetHeader("Content-Encoding")
		if contentEncoding != "" && !strings.Contains(contentEncoding, "gzip") {
			c.String(http.StatusBadRequest, "Unsupported content encoding: %s", contentEncoding)
			return
		}

		if strings.Contains(contentEncoding, "gzip") {
			gz, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				// Log and return a 500 Internal Server Error if gzip reader creation fails.
				c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				return
			}
			defer func() { _ = gz.Close() }()

			// Replace the request body with the decompressed gzip reader.
			c.Request.Body = gz
		}

		c.Next()
	}
}
