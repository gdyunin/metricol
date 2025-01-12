package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

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

			duration := time.Since(start)
			logger.Infof("HTTP request_id=%s | duration: %s | response: status=%d size=%d headers=%v",
				c.Response().Header().Get(echo.HeaderXRequestID),
				duration,
				c.Response().Status,
				c.Response().Size,
				c.Response().Header(),
			)

			return
		}
	}
}
