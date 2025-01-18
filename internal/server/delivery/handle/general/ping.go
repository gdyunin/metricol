package general

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// ConnectionCheckTimeout specifies the duration to wait for a connection check
// before timing out. This constant is used in the Ping handler to ensure that
// health checks complete within a reasonable timeframe.
const connectionCheckTimeout = 3 * time.Second

// ConnectChecker defines an interface for checking connectivity.
//
// The ConnectChecker interface is used to abstract the process of checking
// the connectivity status of a resource, such as a database or external service.
//
// Methods:
//   - CheckConnection: Accepts a context and verifies connectivity. If the
//     connection cannot be established, it returns an error. Otherwise, it
//     completes without returning an error.
type ConnectChecker interface {
	CheckConnection(context.Context) error
}

// Ping returns an HTTP handler function that responds with "pong".
//
// This can be used as a health check endpoint to verify the server is running.
// The handler uses the provided ConnectChecker to perform a connectivity check,
// ensuring the underlying repository or service is operational.
//
// Returns:
//   - An echo.HandlerFunc that sends "pong" with a 200 OK status if the connection check succeeds.
//   - A 500 Internal Server Error status if the connection check fails.
func Ping(checker ConnectChecker) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), connectionCheckTimeout)
		defer cancel()

		if err := checker.CheckConnection(ctx); err != nil {
			return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		return c.String(http.StatusOK, "pong")
	}
}
