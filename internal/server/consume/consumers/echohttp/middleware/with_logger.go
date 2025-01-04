package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// WithLogger creates an Echo middleware function that logs details about incoming HTTP requests and responses.
//
// Parameters:
//   - logger: A `*zap.SugaredLogger` instance used to log request and response details.
//
// Returns:
//   - An Echo `MiddlewareFunc` that logs HTTP request and response details, including method, URI, headers, duration, status, and response size.
func WithLogger(logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			// Record the start time of the request.
			start := time.Now()

			// Call the next handler in the chain.
			err = next(c)

			// Calculate the duration of the request processing.
			duration := time.Since(start)

			// Log the request and response details.
			logger.Infof(
				"HTTP request: method=%s uri=%s headers=%v | "+
					"Process: duration=%s | "+
					"Response: status=%d response_size=%d headers=%v",
				c.Request().Method,     // HTTP method (e.g., GET, POST).
				c.Request().RequestURI, // Full request URI.
				c.Request().Header,     // Request headers.
				duration,               // Time taken to process the request.
				c.Response().Status,    // HTTP response status code.
				c.Response().Size,      // Size of the response in bytes.
				c.Response().Header(),  // Response headers.
			)

			return
		}
	}
}
