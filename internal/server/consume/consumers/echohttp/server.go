package echohttp

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/gdyunin/metricol.git/internal/server/adapters/consumers"
	"github.com/gdyunin/metricol.git/internal/server/consume"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/handle/general"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/handle/update"
	"github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/handle/value"
	mw "github.com/gdyunin/metricol.git/internal/server/consume/consumers/echohttp/middleware"
	"github.com/gdyunin/metricol.git/internal/server/entities"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// EchoServerConsumerFactory creates instances of EchoServer.
type EchoServerConsumerFactory struct {
	addr   string                     // The address for the Echo server.
	repo   entities.MetricsRepository // Metrics repository for managing metrics.
	logger *zap.SugaredLogger         // Logger for logging messages.
}

// NewEchoServerConsumerFactory creates a new EchoServerConsumerFactory.
//
// Parameters:
//   - addr: The address for the Echo server.
//   - repo: Metrics repository for managing metrics.
//   - logger: Logger for logging messages.
//
// Returns:
//   - A pointer to the initialized EchoServerConsumerFactory instance.
func NewEchoServerConsumerFactory(addr string, repo entities.MetricsRepository, logger *zap.SugaredLogger) *EchoServerConsumerFactory {
	return &EchoServerConsumerFactory{
		addr:   addr,
		repo:   repo,
		logger: logger,
	}
}

// CreateConsumer creates and returns a new EchoServer instance.
func (f *EchoServerConsumerFactory) CreateConsumer() consume.Consumer {
	f.logger.Info("Creating a new echo http server consumer.")
	return NewEchoServer(f.addr, f.repo, f.logger)
}

// compressedContentTypes defines the content types to be skipped by the Gzip middleware.
var compressedContentTypes = [2]string{
	"application/json",
	"text/html",
}

// EchoServer represents a consumer that uses the Echo web framework to handle HTTP requests.
type EchoServer struct {
	server        *echo.Echo             // The Echo server instance.
	adp           *consumers.EchoAdapter // Adapter for handling metrics operations.
	log           *zap.SugaredLogger     // Logger for logging server-related messages.
	serverAddress string                 // The address for the Echo server.
}

// NewEchoServer creates a new EchoServer instance.
//
// Parameters:
//   - addr: The address for the Echo server.
//   - repo: Metrics repository for managing metrics.
//   - logger: Logger for logging messages.
//
// Returns:
//   - A pointer to the initialized EchoServer instance.
func NewEchoServer(addr string, repo entities.MetricsRepository, logger *zap.SugaredLogger) *EchoServer {
	s := EchoServer{
		server:        echo.New(),
		adp:           consumers.NewEchoAdapter(repo),
		log:           logger,
		serverAddress: addr,
	}

	// Configure the Echo server to hide banners and port information.
	s.server.HideBanner = true
	s.server.HidePort = true

	// Set up the server components.
	s.setupServer()

	return &s
}

// StartConsume starts the Echo server to consume metrics.
//
// Returns:
//   - An error if the server fails to start.
func (s *EchoServer) StartConsume() error {
	s.log.Info("Starting consume.")
	err := s.server.Start(s.serverAddress)
	if err != nil {
		return fmt.Errorf("failed to start Echo server on address '%s': %w", s.serverAddress, err)
	}
	return nil
}

// setupServer configures the Echo server with middleware, templates, and routes.
func (s *EchoServer) setupServer() {
	s.setupPreMiddlewares()
	s.setupMiddlewares()
	s.setupRenderer()
	s.setupRouters()
}

// setupPreMiddlewares sets up pre-middleware for the Echo server.
func (s *EchoServer) setupPreMiddlewares() {
	s.server.Pre(middleware.RemoveTrailingSlash())
}

// setupMiddlewares configures middleware for the Echo server.
func (s *EchoServer) setupMiddlewares() {
	s.server.Use(
		mw.WithLogger(s.log.Named("request")), // Request logger middleware.
		middleware.Decompress(),               // Decompression middleware.
		middleware.GzipWithConfig(middleware.GzipConfig{
			Skipper: func(c echo.Context) bool {
				contentType := c.Response().Header().Get("Content-Type")
				for _, ct := range compressedContentTypes {
					if strings.HasPrefix(contentType, ct) {
						return true
					}
				}
				return false
			},
		}),
	)
}

// setupRenderer configures the template renderer for the Echo server.
func (s *EchoServer) setupRenderer() {
	s.server.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("web/templates/*.html")),
	}
}

// setupRouters configures routes for the Echo server.
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
