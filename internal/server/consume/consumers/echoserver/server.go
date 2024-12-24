package echoserver

import (
	"fmt"
	"html/template"
	"io"

	"github.com/gdyunin/metricol.git/internal/server/adapter"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/handle/general"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/handle/update"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/handle/value"
	middleware2 "github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/middleware"
	"github.com/gdyunin/metricol.git/internal/server/entity"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type EchoServer struct {
	server        *echo.Echo
	adp           *adapter.EchoAdapter
	log           *zap.SugaredLogger
	serverAddress string
}

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
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
	s.setupMiddlewares()
	s.setupRenderer()
	s.setupRouters()
}

func (s *EchoServer) setupMiddlewares() {
	s.server.Pre(middleware.RemoveTrailingSlash())
	s.server.Use(
		middleware2.WithLogger(s.log.Named("request")),
		//middleware2.WithGzip(),
	)
}

func (s *EchoServer) setupRenderer() {
	s.server.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("web/templates/*.html")),
	}
}

func (s *EchoServer) setupRouters() {
	updateGroup := s.server.Group("/update", middleware2.WithGzip())
	updateGroup.POST("", update.FromJSON(s.adp))
	updateGroup.POST("/:type/:id/:value", update.FromURI(s.adp))

	valueGroup := s.server.Group("/value")
	valueGroup.POST("", value.FromJSON(s.adp))
	valueGroup.GET("/:type/:id", value.FromURI(s.adp))

	s.server.GET("/", general.MainPage(s.adp), middleware2.WithGzip())
	s.server.GET("/ping", general.Ping())
}
