package general

import "github.com/labstack/echo/v4"

func Ping() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(200, "pong")
	}
}
