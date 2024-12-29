package common

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

type ShutdownFunc func()

type ShutdownManager struct {
	actions []ShutdownFunc
	logger  *zap.SugaredLogger
}

func SetupGracefulShutdown(sm *ShutdownManager) {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stopChan
		sm.ExecuteAll()
		os.Exit(0)
	}()
}

func NewShutdownManager() *ShutdownManager {
	return &ShutdownManager{}
}

func (s *ShutdownManager) Add(action ShutdownFunc) {
	s.actions = append(s.actions, action)
}

func (s *ShutdownManager) AttachLogger(logger *zap.SugaredLogger) {
	s.logger = logger
}

func (s *ShutdownManager) ExecuteAll() {
	withLog := s.logger != nil

	for _, fn := range s.actions {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Логирование или обработка ошибки
					fmt.Printf("Recovered from panic: %v\n", r)
				}
			}()
			if withLog {
				// Какое-то логгирование
			}
			fn()
		}()
	}
}
