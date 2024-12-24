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
				"HTTP request: method=%s, uri=%s | "+
					"duration=%s | status=%d | "+
					"response_size=%d bytes | "+
					"response_content_type=%s | "+
					"response_body_encoded=%t",
				c.Request().Method,
				c.Request().RequestURI,
				duration,
				c.Response().Status,
				c.Response().Size,
				c.Response().Header().Get("Content-Type"),
				c.Response().Header().Get("Content-Encoding") != "",
			)
			return
		}
	}
}
