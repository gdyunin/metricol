package general

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Ping() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Respond with HTTP 200 status and "pong".
		return c.String(http.StatusOK, "pong")
	}
}
