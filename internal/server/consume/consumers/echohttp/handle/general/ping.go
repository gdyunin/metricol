package general

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Ping creates an HTTP handler function that responds with "pong".
// This function is used as a health check endpoint.
//
// Returns:
//   - An Echo `HandlerFunc` that responds with HTTP 200 status and "pong" as the body.
func Ping() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Respond with HTTP 200 status and "pong".
		return c.String(http.StatusOK, "pong")
	}
}
