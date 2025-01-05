package helpers

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

// ShutdownFunc defines a function type for actions to be executed during shutdown.
type ShutdownFunc func()

// ShutdownManager manages a list of actions to be executed during application shutdown.
// It also supports attaching a logger for better visibility of shutdown processes.
type ShutdownManager struct {
	logger  *zap.SugaredLogger // Logger for recording shutdown-related events.
	actions []ShutdownFunc     // List of shutdown actions.
}

// SetupGracefulShutdown sets up a signal handler for graceful shutdown of the application.
// It listens for system signals such as SIGTERM and SIGINT, executes all registered shutdown actions,
// and then terminates the application.
//
// Parameters:
//   - sm: An instance of ShutdownManager to manage the registered shutdown actions.
func SetupGracefulShutdown(sm *ShutdownManager) {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stopChan
		sm.ExecuteAll()
		os.Exit(0)
	}()
}

// NewShutdownManager creates and returns a new instance of ShutdownManager.
//
// Returns:
//   - A pointer to a newly initialized ShutdownManager.
func NewShutdownManager() *ShutdownManager {
	return &ShutdownManager{}
}

// Add registers a new shutdown action to the ShutdownManager.
//
// Parameters:
//   - action: A function of type ShutdownFunc to be executed during shutdown.
func (s *ShutdownManager) Add(action ShutdownFunc) {
	s.actions = append(s.actions, action)
}

// AttachLogger attaches a logger to the ShutdownManager for logging shutdown events.
//
// Parameters:
//   - logger: A pointer to a SugaredLogger for recording logs.
func (s *ShutdownManager) AttachLogger(logger *zap.SugaredLogger) {
	s.logger = logger
}

// ExecuteAll executes all registered shutdown actions in the order they were added.
// It recovers from any panics during execution and logs them if a logger is attached.
func (s *ShutdownManager) ExecuteAll() {
	withLog := s.logger != nil

	for _, fn := range s.actions {
		func() {
			defer func() {
				if r := recover(); r != nil {
					message := fmt.Sprintf("Recovered from panic during shutdown action: %v", r)
					if withLog {
						s.logger.Errorf(message)
					} else {
						log.Println(message)
					}
				}
			}()
			if withLog {
				s.logger.Info("Executing shutdown action.")
			}
			fn()
		}()
	}
}
