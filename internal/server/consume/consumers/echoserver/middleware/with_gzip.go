package middleware

import (
	"bytes"
	"compress/gzip"
	"errors"
	"net/http"
	"strings"

	"github.com/gdyunin/metricol.git/pkg/logger"
	"github.com/labstack/echo/v4"
)

const gzipScheme = "gzip"

var encodingContentTypes = []string{
	"application/json",
	"text/html",
}

func WithGzip() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			lg, _ := logger.Logger("INFO")
			lg.Infof("DDDDDDDDDDDDDDDDDDDDDebug!!!!!!! WithGzip start with req headers = %v", c.Request().Header)

			if strings.Contains(c.Request().Header.Get(echo.HeaderContentEncoding), gzipScheme) {
				err := decompressReq(c.Request())
				if err != nil {
					return c.String(500, "")
				}
			}

			if strings.Contains(c.Request().Header.Get(echo.HeaderAcceptEncoding), gzipScheme) {
				err := compressResp(c.Response())
				if err != nil {
					return c.String(500, "")
				}
			}

			return next(c)
		}
	}
}

func decompressReq(r *http.Request) error {
	encHeader := r.Header.Get(echo.HeaderContentEncoding)

	if r.Body == http.NoBody || encHeader == "" {
		return nil
	}

	if !strings.Contains(encHeader, gzipScheme) {
		return errors.New("")
	}

	gz, err := gzip.NewReader(r.Body)
	if err != nil {
		return err
	}
	r.Body = gz

	return nil
}

func compressResp(r *echo.Response) error {
	gz, err := gzip.NewWriterLevel(r.Writer, gzip.BestCompression)
	if err != nil {
		return err
	}
	r.Writer = &gzipResponseWriter{
		rw:         r.Writer,
		gzip:       gz,
		compressed: false,
	}
	return nil
}

type gzipResponseWriter struct {
	rw         http.ResponseWriter
	gzip       *gzip.Writer
	compressed bool
}

func (g *gzipResponseWriter) Header() http.Header {
	return g.rw.Header()
}

func (g *gzipResponseWriter) Write(p []byte) (int, error) {
	lg, _ := logger.Logger("INFO")
	lg.Info(p)
	lg.Info(bytes.NewBuffer(p).String())
	cth := g.Header().Get(echo.HeaderContentType)
	for _, ct := range encodingContentTypes {
		if strings.Contains(cth, ct) {
			g.compressed = true
			g.Header().Set(echo.HeaderContentEncoding, gzipScheme)
			return g.gzip.Write(p)
		}
	}
	return g.rw.Write(p)
}

func (g *gzipResponseWriter) WriteHeader(statusCode int) {
	if g.compressed {
		g.rw.Header().Set(echo.HeaderContentEncoding, gzipScheme)
	} else {
		g.rw.Header().Del(echo.HeaderContentEncoding)
	}
	g.rw.WriteHeader(statusCode)
}
