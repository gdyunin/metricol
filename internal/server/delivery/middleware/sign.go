package middleware

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gdyunin/metricol.git/pkg/sign"
	"github.com/labstack/echo/v4"
)

func Sign(key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if key == "" {
				return next(c)
			}

			w := &signerWriter{
				ResponseWriter: c.Response().Writer,
				key:            key,
				body:           &bytes.Buffer{},
			}
			c.Response().Writer = w

			err = next(c)
			if err != nil {
				c.Error(err)
			}

			// Ensure WriteHeader is called after the response body is populated
			if w.body.Len() > 0 {
				w.WriteHeader(http.StatusOK)
			}

			return nil
		}
	}
}

type signerWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
	key  string
}

func (w *signerWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	fmt.Printf("Data being hashed: %s\n", w.body.String()) // Debugging line
	i, err := w.ResponseWriter.Write(data)
	if err != nil {
		err = fmt.Errorf("error write data in signer writer: %w", err)
	}
	return i, err
}

func (w *signerWriter) WriteHeader(statusCode int) {
	fmt.Printf("WriteHeader called with statusCode: %d\n", statusCode) // Debugging line
	w.Header().Set(
		"HashSHA256",
		hex.EncodeToString(sign.MakeSign(w.body.Bytes(), w.key)),
	)
	w.ResponseWriter.WriteHeader(statusCode)
}
