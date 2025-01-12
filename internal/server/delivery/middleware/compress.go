package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Compress() echo.MiddlewareFunc {
	contentTypesToCompress := map[string]struct{}{
		"application/json": {},
		"text/html":        {},
	}

	gzipConfig := middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			contentType := c.Response().Header().Get("Content-Type")
			_, shouldCompress := contentTypesToCompress[contentType]
			return !shouldCompress
		},
	}

	return middleware.GzipWithConfig(gzipConfig)
}
