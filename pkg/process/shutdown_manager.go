package process

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type ShutdownHandlerFunc func(ctx context.Context) error

type ShutdownManager struct {
	mu              *sync.Mutex
	handlers        []ShutdownHandlerFunc
	gracefulTimeout time.Duration
	logger          *zap.SugaredLogger
}

func NewShutdownManager(gracefulTimeout time.Duration, logger *zap.SugaredLogger) *ShutdownManager {
	return &ShutdownManager{
		mu:              &sync.Mutex{},
		handlers:        make([]ShutdownHandlerFunc, 0),
		gracefulTimeout: gracefulTimeout,
		logger:          logger,
	}
}

func (sm *ShutdownManager) AddShutdownHandler(h ShutdownHandlerFunc) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.handlers = append(sm.handlers, h)
}

func (sm *ShutdownManager) WaitShutdown() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	stop := <-signalChan

	sm.logger.Infof("Shutdown signal received: %s. Cleaning up...", stop.String())

	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.shutdown()

	sm.logger.Info("Shutdown ended.")
}

func (sm *ShutdownManager) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), sm.gracefulTimeout)
	defer cancel()
	sm.executeAll(ctx)
}

func (sm *ShutdownManager) executeAll(ctx context.Context) {
	var wg sync.WaitGroup

	for _, handler := range sm.handlers {
		wg.Add(1)
		go func(h ShutdownHandlerFunc) {
			defer func() {
				if r := recover(); r != nil {
					sm.logger.Errorf("Panic occurred in shutdown handler: %v", r)
				}
				wg.Done()
			}()
			if err := h(ctx); err != nil {
				sm.logger.Errorf("Error occurred while shutting down: %v", err)
			}
		}(handler)
	}

	wg.Wait()
}
