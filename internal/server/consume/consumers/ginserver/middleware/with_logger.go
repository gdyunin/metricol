package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// WithLogger logs details of incoming HTTP requests and the server's responses.
// The middleware logs the following information:
// - HTTP method and request URI.
// - Processing time for the request.
// - HTTP status code of the response.
// - Size of the response in bytes.
// - Content-Type of the request.
// - Whether the response body is encoded (based on "Accept-Encoding" header).
//
// Parameters:
// - logger: A SugaredLogger instance from the Uber Zap logging library for structured and formatted logging.
//
// Usage:
// - Attach this middleware to your Gin engine to log incoming requests and their responses.
func WithLogger(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record the start time of request processing.
		start := time.Now()

		// Process the request.
		c.Next()

		// Calculate the duration of the request processing.
		duration := time.Since(start)

		// Log the HTTP request and response details.
		logger.Infof(
			"HTTP request: method=%s, uri=%s | "+
				"duration=%s | status=%d | "+
				"response_size=%d bytes | "+
				"content_type=%s | "+
				"body_encoded=%t",
			c.Request.Method,                     // HTTP method (e.g., GET, POST).
			c.Request.RequestURI,                 // Request URI.
			duration,                             // Time taken to process the request.
			c.Writer.Status(),                    // HTTP status code.
			c.Writer.Size(),                      // Size of the response body in bytes.
			c.GetHeader("Content-Type"),          // Content-Type header of the request.
			c.GetHeader("Accept-Encoding") != "", // Whether the response body is encoded.
		)
	}
}
