package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func Auth(key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if key == "" {
				return next(c)
			}

			sign := c.Request().Header.Get("HashSHA256")
			if sign == "" {
				return next(c) // https://app.pachca.com/chats?thread_message_id=419958556
			}

			rawBody, err := getRawBody(c.Request())
			if err != nil {
				return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			}

			if !checkSign(rawBody, sign, key) {
				return c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			}

			return next(c)
		}
	}
}

// getRawBody retrieves the raw body of a http.Request.
//
// Parameters:
//   - req: The HTTP request whose body needs to be read.
//
// Returns:
//   - []byte: The raw body of the request as a byte slice.
//   - error: An error if the body retrieval or reading fails; otherwise, nil.
func getRawBody(req *http.Request) ([]byte, error) {
	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	defer func() {
		if err = req.Body.Close(); err != nil {
			log.Errorf("failed to close body: %v", err)
		}
	}()

	req.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	return rawBody, nil
}

// checkSign verifies the HMAC-SHA256 signature of a given message.
//
// Parameters:
//   - body: The original message whose signature needs to be verified.
//   - sign: The provided HMAC-SHA256 signature to verify.
//   - key: The secret key used to generate the HMAC-SHA256 signature.
//
// Returns:
//   - bool: True if the provided signature matches the computed signature; otherwise, false.
func checkSign(body []byte, sign string, key string) bool {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)

	gotSign, err := hex.DecodeString(sign)
	if err != nil {
		return false
	}
	validSign := h.Sum(nil)
	return hmac.Equal(gotSign, validSign)
}
