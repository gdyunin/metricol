package echoserver

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/server/adapter"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echoserver/handle"
	"github.com/gdyunin/metricol.git/internal/server/entity"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type EchoServer struct {
	server        *echo.Echo
	adp           *adapter.GinController
	log           *zap.SugaredLogger
	serverAddress string
}

func NewEchoServer(addr string, repo entity.MetricRepository, logger *zap.SugaredLogger) *EchoServer {
	s := EchoServer{
		server:        echo.New(),
		adp:           adapter.NewGinController(repo),
		log:           logger,
		serverAddress: addr,
	}

	s.setupServer()
	return &s
}

func (e *EchoServer) StartConsume() error {
	err := e.server.Start(e.serverAddress)
	if err != nil {
		return fmt.Errorf("emergency stop: failed to start Gin server on address %s: %w", e.serverAddress, err)
	}
	return nil
}

func (e *EchoServer) setupServer() {
	// Set up middlewares.
	e.setupMiddlewares()

	// Define the routes for the server.
	e.setupRouters()
}

func (e *EchoServer) setupMiddlewares() {
	e.server.Use(
		middleware.Logger(), // Adds request logging middleware.
		//middleware.WithGzip(),                         // Adds gzip compression middleware.
	)
}

func (e *EchoServer) setupRouters() {
	// Main page route.
	//e.server.GET("/", handle.MainPageHandler(g.adp))

	// "/update" routes for updating metric values.
	//updateGroup := e.server.Group("update")
	// Update metric values using JSON parameters.
	e.server.POST("/update", handle.UpdateHandlerWithJSONParams(e.adp))
	// Update metric values using URI parameters.
	e.server.POST("/:type/:id/*value", handle.UpdateHandlerWithURIParams(e.adp))

	// "/value" routes for retrieving metric values.
	//valueGroup := e.server.Group("/value")
	// Retrieve metric values using JSON parameters.
	e.server.POST("/value", handle.ValueHandlerWithJSONParams(e.adp))
	// Retrieve metric values using URI parameters.
	e.server.GET("/:type/:id", handle.ValueHandlerWithURIParams(e.adp))

}
