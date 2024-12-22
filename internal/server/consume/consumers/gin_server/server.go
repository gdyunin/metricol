package gin_server

import (
	"fmt"

	"github.com/gdyunin/metricol.git/internal/server/adapter"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/gin_server/handle"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/gin_server/middleware"
	"github.com/gdyunin/metricol.git/internal/server/entity"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinServer represents the Gin-based HTTP server with its configuration, middlewares, and routes.
type GinServer struct {
	serverAddress string                 // The address where the server listens for incoming requests.
	server        *gin.Engine            // The Gin engine instance used for routing and handling HTTP requests.
	adp           *adapter.GinController // The controller for handling requests and interacting with repositories.
	log           *zap.SugaredLogger     // Logger for capturing server-related logs.
}

// NewServer initializes and returns a new GinServer instance with the specified address and repositories.
// It sets up the Gin server and returns an error if the setup fails.
func NewServer(addr string, repo entity.MetricRepository, logger *zap.SugaredLogger) (*GinServer, error) {
	s := GinServer{
		serverAddress: addr,
		server:        gin.New(),
		adp:           adapter.NewGinController(repo),
		log:           logger,
	}

	// Attempt to set up the server and handle errors.
	if err := s.setupServer(); err != nil {
		return nil, fmt.Errorf("failed to set up Gin server: %w", err)
	}

	// Load HTML templates for rendering.
	s.server.LoadHTMLGlob("web/templates/*")

	return &s, nil
}

// NewServerWithConfigParser initializes and returns a new GinServer instance using a configuration parser function and repositories.
// It sets up the Gin server with the provided configuration and returns an error if setup fails.
func NewServerWithConfigParser(cfgParser func() (*GinServerConfig, error), repo entity.MetricRepository, logger *zap.SugaredLogger) (*GinServer, error) {
	// Parse the configuration using the provided parser function.
	cfg, err := cfgParser()
	if err != nil {
		return nil, fmt.Errorf("failed to parse Gin server configuration: %w", err)
	}

	// Create a new server using the parsed configuration.
	s, err := NewServer(cfg.ServerAddress, repo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to set up Gin server with configuration %+v: %w", cfg, err)
	}
	return s, nil
}

// StartConsume starts the Gin server and begins consuming requests.
// If the server fails to start, it returns an error with additional context.
func (g *GinServer) StartConsume() error {
	err := g.server.Run(g.serverAddress)
	if err != nil {
		return fmt.Errorf("emergency stop: failed to start Gin server on address %s: %w", g.serverAddress, err)
	}
	return nil
}

// setupServer configures and sets up the Gin server by applying middlewares and defining routes.
// It returns an error if any part of the setup fails.
func (g *GinServer) setupServer() error {
	// Set up middlewares and handle errors during setup.
	if err := g.setupMiddlewares(); err != nil {
		return fmt.Errorf("failed to set up middlewares: %w", err)
	}

	// Define the routes for the server.
	g.setupRouters()
	return nil
}

// setupMiddlewares configures and applies middlewares for the Gin server.
// It returns an error if any middleware setup fails.
func (g *GinServer) setupMiddlewares() error {
	g.server.Use(
		gin.Recovery(), // Provides recovery middleware to handle panics gracefully.
		middleware.WithLogger(g.log.Named("request")), // Adds request logging middleware.
		middleware.WithGzip(),                         // Adds gzip compression middleware.
	)
	return nil
}

// setupRouters configures the routes for the Gin server.
// It defines the main page route, value-related routes, and update-related routes.
func (g *GinServer) setupRouters() {
	// Main page route.
	g.server.GET("/", handle.MainPageHandler(g.adp))

	// "/value" routes for retrieving metric values.
	{
		value := g.server.Group("/value")
		value.POST("", handle.ValueHandlerWithJSONParams(g.adp))                       // Retrieve metric values using JSON parameters.
		value.GET("/:metricType/:metricName", handle.ValueHandlerWithURIParams(g.adp)) // Retrieve metric values using URI parameters.
	}

	// "/update" routes for updating metric values.
	{
		update := g.server.Group("/update")
		update.POST("", handle.UpdateHandlerWithJSONParams(g.adp))                 // Update metric values using JSON parameters.
		update.POST("/:type/:id/*value", handle.UpdateHandlerWithURIParams(g.adp)) // Update metric values using URI parameters.
	}
}
