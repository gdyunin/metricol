package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func WithLogger(logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			start := time.Now()
			err = next(c)
			duration := time.Since(start)

			logger.Infof(
				"HTTP request: method=%s uri=%s headers=%v | "+
					"Process: duration=%s | "+
					"Response: status=%d response_size=%d headers=%v",
				c.Request().Method, c.Request().RequestURI, c.Request().Header,
				duration,
				c.Response().Status, c.Response().Size, c.Response().Header(),
			)
			return
		}
	}
}
