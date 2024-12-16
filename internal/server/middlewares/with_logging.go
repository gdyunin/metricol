package middlewares

import (
	"net/http"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/logger"
)

// responseInfo holds information about the HTTP response status and size.
type responseInfo struct {
	status int // HTTP status code of the response.
	size   int // Size of the response in bytes.
}

// loggingResponseWriter wraps a http.ResponseWriter to capture response information.
type loggingResponseWriter struct {
	http.ResponseWriter               // The original ResponseWriter.
	responseInfo        *responseInfo // Pointer to the responseInfo struct.
}

// Write writes the data to the ResponseWriter and updates the response size.
func (rw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.responseInfo.size += size // Update the size of the response.
	return size, err             //nolint:wrapcheck // err must not be wrapped.
}

// WriteHeader sends an HTTP response header and updates the response status.
func (rw *loggingResponseWriter) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.responseInfo.status = statusCode // Update the status code of the response.
}

// WithLogging is a middleware that logs HTTP requests and responses.
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now() // Record the start time of the request.

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseInfo:   &responseInfo{},
		}

		next.ServeHTTP(&lw, r) // Call the next handler in the chain.

		logger.SugarLogger.Infoln(
			"Received HTTP request:",
			r.Method, r.RequestURI, "|",
			"Processed:",
			"time", time.Since(start), // Calculate the processing time.
			"code", lw.responseInfo.status,
			"size", lw.responseInfo.size,
		)
	})
}
