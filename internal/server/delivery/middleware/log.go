package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Log creates an Echo middleware that logs HTTP requests and responses.
// It logs the request method, URI, headers, and the processing time of each request.
//
// Parameters:
//   - logger: A sugared logger instance from zap for structured logging.
//
// Returns:
//   - echo.MiddlewareFunc: The middleware function that performs logging.
func Log(logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			start := time.Now()

			logger.Infof("HTTP request_id=%s | request: method=%s uri=%s headers=%v",
				c.Response().Header().Get(echo.HeaderXRequestID),
				c.Request().Method,
				c.Request().RequestURI,
				c.Request().Header,
			)

			if err = next(c); err != nil {
				c.Error(err)
			}

			processingTime := time.Since(start)
			logger.Infof("HTTP request_id=%s | response (processingTime: %s): status=%d size=%d headers=%v",
				c.Response().Header().Get(echo.HeaderXRequestID),
				processingTime,
				c.Response().Status,
				c.Response().Size,
				c.Response().Header(),
			)

			return
		}
	}
}
