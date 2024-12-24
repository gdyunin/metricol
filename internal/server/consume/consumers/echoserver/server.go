package echoserver

import (
	"fmt"
	"log"

	"github.com/gdyunin/metricol.git/internal/server/adapter"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/handle/update"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/handle/value"
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

func NewEchoServer(addr string, repo entity.MetricRepository, logger *zap.SugaredLogger) *EchoServer {
	s := EchoServer{
		server:        echo.New(),
		adp:           adapter.NewEchoAdapter(repo),
		log:           logger,
		serverAddress: addr,
	}

	s.setupServer()

	return &s
}

func (s *EchoServer) StartConsume() error {
	s.server.Server.ErrorLog = log.Default()
	err := s.server.Start(s.serverAddress)
	if err != nil {
		return fmt.Errorf("emergency stop: failed to start Gin server on address %s: %w", s.serverAddress, err)
	}
	return nil
}

func (s *EchoServer) setupServer() {
	// Set up middlewares.
	//g.setupMiddlewares()

	// Define the routes for the server.
	s.setupRouters()

	s.server.Use(middleware.Logger(), middleware.RemoveTrailingSlash())
}

func (s *EchoServer) setupRouters() {
	// Main page route.
	//s.server.GET("/", handle.MainPageHandler(s.adp))

	s.server.GET("/ping", func(c echo.Context) error {
		return c.String(200, "pong")
	})

	s.server.POST("/update", update.FromJSON(s.adp))
	// Update metric values using URI parameters.
	s.server.POST("/update/:type/:id/:value", update.FromURI(s.adp))

	// Retrieve metric values using JSON parameters.
	s.server.POST("/value", value.FromJSON(s.adp))
	// Retrieve metric values using URI parameters.
	s.server.GET("value/:type/:id", value.FromURI(s.adp))
}
