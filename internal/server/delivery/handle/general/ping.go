package general

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Ping returns an HTTP handler function that responds with "pong".
//
// This can be used as a health check endpoint to verify the server is running.
//
// Returns:
//   - An echo.HandlerFunc that sends "pong" with a 200 OK status.
func Ping() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	}
}
