// Package delivery implements the HTTP delivery layer for the server using the Echo framework.
// It sets up routes, middleware, and renderers to handle HTTP requests for metrics operations,
// including updating metrics, retrieving values, and serving general pages such as ping and main page.
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
	"github.com/gdyunin/metricol.git/internal/server/delivery/handle/updates"
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
	// Const defaultTemplatesPath is the default directory path for HTML templates.
	defaultTemplatesPath = "web/templates/"
	// Const gracefulShutdownTimeout is the time duration to wait for ongoing tasks to complete during shutdown.
	gracefulShutdownTimeout = 5 * time.Second
)

// EchoServer defines the HTTP server powered by the Echo framework.
// It encapsulates the Echo instance, logger, metric controller, address,
// template path, and signing key.
type EchoServer struct {
	echo        *echo.Echo                // echo is the Echo instance used to serve HTTP requests.
	logger      *zap.SugaredLogger        // logger is used for structured logging.
	metricsCtrl *controller.MetricService // metricsCtrl handles metric operations.
	addr        string                    // addr is the server address to listen on.
	tmplPath    string                    // tmplPath is the directory path to the HTML templates.
	signingKey  string                    // signingKey is used for request signing and authentication.
	cryptoKey   string
}

// NewEchoServer creates and configures a new EchoServer instance.
// It initializes the Echo server, metric controller, logging, and template path,
// and then applies the build steps to setup middleware, renderers, and routes.
//
// Parameters:
//   - serverAddress: The address on which the server will listen.
//   - signingKey: The key used for signing requests.
//   - repo: The repository instance used for metric storage.
//   - logger: The logger instance for structured logging.
//
// Returns:
//   - *EchoServer: A pointer to the configured EchoServer instance.
func NewEchoServer(
	serverAddress string,
	signingKey string,
	cryptoKey string,
	repo repository.Repository,
	logger *zap.SugaredLogger,
) *EchoServer {
	echoServer := EchoServer{
		echo:        echo.New(),
		logger:      logger,
		addr:        serverAddress,
		signingKey:  signingKey,
		cryptoKey:   cryptoKey,
		tmplPath:    defaultTemplatesPath,
		metricsCtrl: controller.NewMetricService(repo),
	}

	// Hide Echo's startup banner and port output.
	echoServer.echo.HideBanner = true
	echoServer.echo.HidePort = true

	return echoServer.build()
}

// Start runs the Echo server and initiates graceful shutdown when the provided context is canceled.
// It starts a separate goroutine to handle shutdown signals.
//
// Parameters:
//   - ctx: The context used to manage the server lifecycle and signal shutdown.
func (s *EchoServer) Start(ctx context.Context) {
	go s.handleShutdown(ctx)

	s.logger.Infof("Server is starting on %s", s.addr)
	if err := s.echo.Start(s.addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Fatalf("Server start failed: %v", err)
	} else {
		s.logger.Info("Server exited cleanly")
	}
}

// handleShutdown listens for a shutdown signal from the provided context.
// Upon receiving the signal, it initiates a graceful shutdown of the Echo server
// within a predefined timeout period.
//
// Parameters:
//   - ctx: The context to monitor for shutdown signals.
func (s *EchoServer) handleShutdown(ctx context.Context) {
	s.logger.Info("Starting server shutdown listener")
	<-ctx.Done()
	s.logger.Info("Received shutdown signal")

	shutdownCtx, cancel := context.WithTimeout(ctx, gracefulShutdownTimeout)
	defer cancel()

	if err := s.echo.Shutdown(shutdownCtx); err != nil {
		s.logger.Warnf("Failed to shutdown server gracefully: %v", err)
	} else {
		s.logger.Info("Server shutdown gracefully")
	}
}

// build configures the EchoServer by executing a series of setup steps.
// The setup steps include configuring pre-middlewares, general middlewares,
// template renderers, and HTTP routes.
//
// Returns:
//   - *EchoServer: A pointer to the EchoServer after configuration.
func (s *EchoServer) build() *EchoServer {
	// Execute each build step to configure the server.
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

// setupPreMiddlewares configures the pre-middlewares for the Echo server.
// Pre-middlewares are executed before the router is matched.
// This setup includes assigning a unique request ID and removing trailing slashes.
func (s *EchoServer) setupPreMiddlewares() {
	s.logger.Info("Setting up pre-middlewares")
	s.echo.Pre(
		echoMiddleware.RequestID(),
		echoMiddleware.RemoveTrailingSlash(),
	)
}

// setupGeneralMiddlewares configures the general middlewares for the Echo server.
// These middlewares handle logging, decompression, authentication, signing, and gzip compression.
func (s *EchoServer) setupGeneralMiddlewares() {
	s.logger.Info("Setting up general middlewares")
	requestLogger := s.logger.Named("request")

	s.echo.Use(
		custMiddleware.Log(requestLogger),
		echoMiddleware.Decompress(),
		custMiddleware.Auth(s.signingKey),
		custMiddleware.Sign(s.signingKey),
		custMiddleware.Crypto(s.cryptoKey, requestLogger.Named("crypto")),
		custMiddleware.Gzip(requestLogger.Named("gzip_writer")),
	)
}

// setupRenderers sets up the HTML template renderer for the Echo server.
// It parses all HTML templates in the specified template path and assigns the renderer to Echo.
func (s *EchoServer) setupRenderers() {
	s.logger.Info("Setting up template renderers")
	s.echo.Renderer = render.NewRenderer(
		template.Must(template.ParseGlob(path.Join(s.tmplPath, "*.html"))),
	)
}

// setupRouters configures the HTTP routes for the Echo server.
// It defines route groups and associates them with handler functions for updating metrics,
// retrieving metric values, and serving general pages.
func (s *EchoServer) setupRouters() {
	s.logger.Info("Setting up routes")

	// Route group for single metric updates.
	updateGroup := s.echo.Group("/update")
	updateGroup.POST("", update.FromJSON(s.metricsCtrl))
	updateGroup.POST("/:type/:id/:value", update.FromURI(s.metricsCtrl))

	// Route group for batch metric updates.
	updatesGroup := s.echo.Group("/updates")
	updatesGroup.POST("", updates.FromJSON(s.metricsCtrl))

	// Route group for metric value retrieval.
	valueGroup := s.echo.Group("/value")
	valueGroup.POST("", value.FromJSON(s.metricsCtrl))
	valueGroup.GET("/:type/:id", value.FromURI(s.metricsCtrl))

	// Routes for main page and health check.
	s.echo.GET("/", general.MainPage(s.metricsCtrl))
	s.echo.GET("/ping", general.Ping(s.metricsCtrl))
}
