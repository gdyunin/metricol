package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/config"
	"github.com/gdyunin/metricol.git/internal/server/delivery"
	"github.com/gdyunin/metricol.git/internal/server/repository"
	"github.com/gdyunin/metricol.git/pkg/convert"
	"github.com/gdyunin/metricol.git/pkg/logging"

	"go.uber.org/zap"
)

const (
	// LoggerNameDelivery is the logger name for the delivery layer.
	loggerNameDelivery = "delivery"
	// LoggerNameRepository is the logger name for the repository layer.
	loggerNameRepository = "repository"
	// LoggerNameGracefulShutdown is the logger name for the graceful shutdown events.
	loggerNameGracefulShutdown = "graceful_shutdown"
	// GracefulShutdownTimeout is the time to wait for ongoing tasks to complete during shutdown.
	gracefulShutdownTimeout = 5 * time.Second
	// DefaultBackupFileName is the name of the default backup file.
	defaultBackupFileName = "backup.txt"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func printAppInfo() {
	fmt.Println("+------------------+----------------------+")
	fmt.Printf("| %-16s | %-20s |\n", "Build version", buildVersion)
	fmt.Printf("| %-16s | %-20s |\n", "Build date", buildDate)
	fmt.Printf("| %-16s | %-20s |\n", "Build commit", buildCommit)
	fmt.Println("+------------------+----------------------+")
}

// mainContext initializes the main application context with a cancel function.
//
// Returns:
//   - context.Context: The main context for application lifecycle management.
//   - context.CancelFunc: A cancel function to signal application termination.
func mainContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// baseLogger initializes a structured logger for the application.
//
// Returns:
//   - *zap.SugaredLogger: A configured logger instance.
//   - func(): A function to flush any buffered logs.
func baseLogger() *zap.SugaredLogger {
	return logging.Logger(logging.LevelINFO)
}

// loadConfig parses the application's configuration file.
//
// Returns:
//   - *config.Config: The parsed configuration.
//   - error: An error if parsing fails.
func loadConfig() (*config.Config, error) {
	cfg, err := config.ParseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

// deliveryWithShutdown holds the Echo server and shutdown actions.
//
// Fields:
//   - server: The Echo server instance.
//   - shutdownActions: A list of functions to execute during shutdown.
type deliveryWithShutdown struct {
	server          *delivery.EchoServer
	shutdownActions []func()
}

// initComponentsWithShutdownActs initializes application components and generates shutdown actions.
//
// Parameters:
//   - cfg: The application configuration.
//   - logger: The structured logger instance.
//
// Returns:
//   - *deliveryWithShutdown: A wrapper for Echo server and its shutdown actions.
//   - error: An error if initialization fails.
func initComponentsWithShutdownActs(
	cfg *config.Config,
	logger *zap.SugaredLogger,
) (*deliveryWithShutdown, error) {
	shutdownActions := make([]func(), 0)

	repoWithShutdownFunc, err := initRepo(cfg, logger.Named(loggerNameRepository))
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}
	shutdownActions = append(shutdownActions, repoWithShutdownFunc.shutdown)

	echoDelivery := delivery.NewEchoServer(
		cfg.ServerAddress,
		cfg.SigningKey,
		repoWithShutdownFunc.repository,
		logger.Named(loggerNameDelivery),
	)

	return &deliveryWithShutdown{
		server:          echoDelivery,
		shutdownActions: shutdownActions,
	}, nil
}

// repoWithShutdown holds the repository instance and its shutdown function.
//
// Fields:
//   - repository: The repository instance.
//   - shutdown: The function to cleanly shut down the repository.
type repoWithShutdown struct {
	repository repository.Repository
	shutdown   func()
}

// initRepo initializes the repository component and its shutdown function.
//
// Parameters:
//   - cfg: The application configuration.
//   - logger: The structured logger instance for the repository.
//
// Returns:
//   - *repoWithShutdown: A wrapper for repository and its shutdown function.
//   - error: An error if initialization fails.
func initRepo(cfg *config.Config, logger *zap.SugaredLogger) (*repoWithShutdown, error) {
	var doNothing = func() {}

	if cfg.DatabaseDSN != "" {
		r, err := repository.NewPostgreSQL(logger, cfg.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize PostgreSQL repository: %w", err)
		}
		return &repoWithShutdown{repository: r, shutdown: r.Shutdown}, nil
	}

	if cfg.FileStoragePath != "" {
		r := repository.NewInFileRepository(
			logger,
			cfg.FileStoragePath,
			defaultBackupFileName,
			convert.IntegerToSeconds(cfg.StoreInterval),
			cfg.Restore,
		)
		return &repoWithShutdown{repository: r, shutdown: r.Shutdown}, nil
	}

	return &repoWithShutdown{
		repository: repository.NewInMemoryRepository(logger),
		shutdown:   doNothing,
	}, nil
}

// setupGracefulShutdown configures the graceful shutdown mechanism for the application.
//
// Parameters:
//   - ctxCancel: The cancel function to terminate the application context.
//   - logger: The structured logger instance for shutdown events.
//   - shutdownActions: A variadic list of functions to execute during shutdown.
func setupGracefulShutdown(
	ctxCancel context.CancelFunc,
	logger *zap.SugaredLogger,
	shutdownActions ...func(),
) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		logger.Info("Received termination signal (SIGTERM or SIGINT). Initiating graceful shutdown...")
		ctxCancel() // Cancel the application context.

		for _, act := range shutdownActions {
			go func(fn func()) {
				fn()
			}(act)
		}

		logger.Infof(
			"Context canceled. Allowing %d seconds for cleanup operations before forced application exit...",
			gracefulShutdownTimeout/time.Second,
		)
		time.Sleep(gracefulShutdownTimeout) // Wait for a graceful shutdown.
		logger.Warn("Timeout reached. Forcing application to exit.")
		os.Exit(0) // Exit the application.
	}()
}

func startProf(ctx context.Context, addr string) error {
	srv := &http.Server{Addr: addr, Handler: nil}
	errCh := make(chan error)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			errCh <- fmt.Errorf("error profiling server run: %w", err)
			close(errCh)
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown profiling server error: %w", err)
		}
	case listenErr := <-errCh:
		return fmt.Errorf("running profiling server error: %w", listenErr)
	}

	return nil
}
