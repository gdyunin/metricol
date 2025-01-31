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

			c.Response().Writer = &signerWriter{
				ResponseWriter: c.Response().Writer,
				key:            key,
				body:           &bytes.Buffer{},
			}

			return next(c)
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
	i, err := w.ResponseWriter.Write(data)
	if err != nil {
		err = fmt.Errorf("error write data in signer writer: %w", err)
	}
	return i, err
}

func (w *signerWriter) WriteHeader(statusCode int) {
	w.Header().Set(
		"HashSHA256",
		hex.EncodeToString(sign.MakeSign(w.body.Bytes(), w.key)),
	)
	w.ResponseWriter.WriteHeader(statusCode)
}
