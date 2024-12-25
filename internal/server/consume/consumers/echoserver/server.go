package echoserver

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/gdyunin/metricol.git/internal/server/adapter"
	"github.com/gdyunin/metricol.git/internal/server/backup"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/handle/general"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/handle/update"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/handle/value"
	mw "github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/middleware"
	"github.com/gdyunin/metricol.git/internal/server/entity"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

var compressedContentTypes = [2]string{
	"application/json",
	"text/html",
}

type EchoServer struct {
	server        *echo.Echo
	adp           *adapter.EchoAdapter
	log           *zap.SugaredLogger
	backupper     backup.Backupper
	serverAddress string
}

func NewEchoServer(addr string, repo entity.MetricRepository, logger *zap.SugaredLogger) *EchoServer {
	s := EchoServer{
		server:        echo.New(),
		adp:           adapter.NewEchoAdapter(repo),
		log:           logger,
		serverAddress: addr,
	}

	s.server.HideBanner = true
	s.server.HidePort = true

	s.setupServer()

	return &s
}

func (s *EchoServer) StartConsume() error {
	err := s.server.Start(s.serverAddress)
	if err != nil {
		return fmt.Errorf("emergency stop: failed to start Gin server on address %s: %w", s.serverAddress, err)
	}
	return nil
}

func (s *EchoServer) setupServer() {
	s.setupPreMiddlewares()
	s.setupMiddlewares()
	s.setupRenderer()
	s.setupRouters()
}

func (s *EchoServer) setupPreMiddlewares() {
	s.server.Pre(middleware.RemoveTrailingSlash())
}

func (s *EchoServer) setupMiddlewares() {
	s.server.Use(
		mw.WithLogger(s.log.Named("request")),
		middleware.Decompress(),
		middleware.GzipWithConfig(middleware.GzipConfig{
			Skipper: func(c echo.Context) bool {
				contentType := c.Response().Header().Get("Content-Type")
				for _, ict := range compressedContentTypes {
					if strings.HasPrefix(contentType, ict) {
						return true
					}
				}
				return false
			},
		}),
	)
}

func (s *EchoServer) setupRenderer() {
	s.server.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("web/templates/*.html")),
	}
}

func (s *EchoServer) setupRouters() {
	updateGroup := s.server.Group("/update")
	updateGroup.POST("", update.FromJSON(s.adp))
	updateGroup.POST("/:type/:id/:value", update.FromURI(s.adp))

	valueGroup := s.server.Group("/value")
	valueGroup.POST("", value.FromJSON(s.adp))
	valueGroup.GET("/:type/:id", value.FromURI(s.adp))

	s.server.GET("/", general.MainPage(s.adp))
	s.server.GET("/ping", general.Ping())
}
