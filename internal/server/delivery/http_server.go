package delivery

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"path"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/delivery/handle/general"
	"github.com/gdyunin/metricol.git/internal/server/delivery/handle/update"
	"github.com/gdyunin/metricol.git/internal/server/delivery/handle/value"
	custMiddleware "github.com/gdyunin/metricol.git/internal/server/delivery/middleware"
	"github.com/gdyunin/metricol.git/internal/server/delivery/render"
	"github.com/gdyunin/metricol.git/internal/server/internal/controller"
	"github.com/gdyunin/metricol.git/internal/server/repository"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

const (
	DefaultTemplatesPath = "web/templates/"
	// GracefulShutdownTimeout is the time to wait for ongoing tasks to complete during shutdown.
	GracefulShutdownTimeout = 5 * time.Second
)

// EchoServer defines the HTTP server powered by Echo framework.
type EchoServer struct {
	echo        *echo.Echo
	logger      *zap.SugaredLogger
	metricsCtrl *controller.MetricService
	addr        string
	tmplPath    string
}

// NewEchoServer creates and configures a new EchoServer instance.
//
// Parameters:
//   - serverAddress: The address on which the server will listen.
//   - repo: The repository instance for metric storage.
//   - logger: The logger instance for structured logging.
//
// Returns:
//   - A pointer to the configured EchoServer.
func NewEchoServer(serverAddress string, repo repository.Repository, logger *zap.SugaredLogger) *EchoServer {
	echoServer := EchoServer{
		echo:        echo.New(),
		logger:      logger,
		addr:        serverAddress,
		tmplPath:    DefaultTemplatesPath,
		metricsCtrl: controller.NewMetricService(repo),
	}

	echoServer.echo.HideBanner = true
	echoServer.echo.HidePort = true

	return echoServer.build()
}

// Start runs the Echo server and handles graceful shutdown when the context is canceled.
//
// Parameters:
//   - ctx: The context for managing server lifecycle.
func (s *EchoServer) Start(ctx context.Context) {
	go s.handleShutdown(ctx)

	s.logger.Infof("Server is starting on %s", s.addr)
	if err := s.echo.Start(s.addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Fatalf("Server start failed: %v", err)
	} else {
		s.logger.Info("Server exited cleanly")
	}
}

// handleShutdown manages the graceful shutdown process when a signal is received.
//
// Parameters:
//   - ctx: The context to monitor for shutdown signals.
func (s *EchoServer) handleShutdown(ctx context.Context) {
	s.logger.Info("Starting server shutdown listener")
	<-ctx.Done()
	s.logger.Info("Received shutdown signal")
	shutdownCtx, cancel := context.WithTimeout(ctx, GracefulShutdownTimeout)
	defer cancel()
	if err := s.echo.Shutdown(shutdownCtx); err != nil {
		s.logger.Warnf("Failed to shutdown server gracefully: %v", err)
	} else {
		s.logger.Info("Server shutdown gracefully")
	}
}

// build configures the EchoServer by applying setup steps.
//
// Returns:
//   - A pointer to the configured EchoServer.
func (s *EchoServer) build() *EchoServer {
	buildSteps := []func(){
		s.setupPreMiddlewares,
		s.setupGeneralMiddlewares,
		s.setupRenderers,
		s.setupRouters,
	}

	for _, stepFunc := range buildSteps {
		stepFunc()
	}

	s.logger.Info("Server build completed")
	return s
}

// setupPreMiddlewares configures the pre-middleware for the Echo server.
func (s *EchoServer) setupPreMiddlewares() {
	s.logger.Info("Setting up pre-middlewares")
	s.echo.Pre(
		echoMiddleware.RequestID(),
		echoMiddleware.RemoveTrailingSlash(),
	)
}

// setupGeneralMiddlewares configures the general middleware for the Echo server.
func (s *EchoServer) setupGeneralMiddlewares() {
	s.logger.Info("Setting up general middlewares")
	s.echo.Use(
		custMiddleware.Log(s.logger.Named("request")),
		echoMiddleware.Decompress(),
		echoMiddleware.Gzip(),
	)
}

// setupRenderers sets up the HTML template renderer for the Echo server.
func (s *EchoServer) setupRenderers() {
	s.logger.Info("Setting up template renderers")
	s.echo.Renderer = render.NewRenderer(
		template.Must(template.ParseGlob(path.Join(s.tmplPath, "*.html"))),
	)
}

// setupRouters configures the routes for the Echo server.
func (s *EchoServer) setupRouters() {
	s.logger.Info("Setting up routes")
	updateGroup := s.echo.Group("/update")
	updateGroup.POST("", update.FromJSON(s.metricsCtrl))
	updateGroup.POST("/:type/:id/:value", update.FromURI(s.metricsCtrl))

	valueGroup := s.echo.Group("/value")
	valueGroup.POST("", value.FromJSON(s.metricsCtrl))
	valueGroup.GET("/:type/:id", value.FromURI(s.metricsCtrl))

	s.echo.GET("/", general.MainPage(s.metricsCtrl))
	s.echo.GET("/ping", general.Ping(s.metricsCtrl))
}
