package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type gzipWriter struct {
	http.ResponseWriter           // Original Gin response writer.
	Writer              io.Writer // Gzip writer for compressing the response data.
}

const gzipEncodingHeader = "gzip"

func WithGzip() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			acceptEncoding := c.Request().Header.Get("Accept-Encoding")
			if strings.Contains(acceptEncoding, gzipEncodingHeader) {
				c, err = setCompressor(c)
				if err != nil {
					return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				}
			}

			if c.Request().Body == http.NoBody {
				return next(c)
			}

			contentEncoding := c.Request().Header.Get("Content-Encoding")
			if contentEncoding != "" && !strings.Contains(contentEncoding, gzipEncodingHeader) {
				return c.String(http.StatusBadRequest, fmt.Sprintf("Unsupported content encoding: %s", contentEncoding))
			}
			if strings.Contains(contentEncoding, gzipEncodingHeader) {
				c, err = setDecompressor(c)
				if err != nil {
					return c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				}
			}
			return next(c)
		}
	}
}

func setCompressor(c echo.Context) (echo.Context, error) {
	gz, err := gzip.NewWriterLevel(c.Response().Writer, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	c.Response().Writer = &gzipWriter{
		ResponseWriter: c.Response().Writer,
		Writer:         gz,
	}
	c.Request().Header.Set("Content-Encoding", gzipEncodingHeader)
	return c, nil
}

func setDecompressor(c echo.Context) (echo.Context, error) {
	gz, err := gzip.NewReader(c.Request().Body)
	if err != nil {
		return nil, err
	}

	c.Request().Body = gz
	return c, nil
}
