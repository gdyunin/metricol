package collect

import (
	"context"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"
	"go.uber.org/zap"
)

type Strategy interface {
	Collect() (*entity.Metrics, error)
}

type StreamCollector struct {
	collectStrategy Strategy
	streamTo        chan *entity.Metrics
	logger          *zap.SugaredLogger
	interval        time.Duration
}

func NewStreamCollector(
	streamTo chan *entity.Metrics,
	interval time.Duration,
	collectStrategy Strategy,
	logger *zap.SugaredLogger,
) *StreamCollector {
	return &StreamCollector{
		streamTo:        streamTo,
		interval:        interval,
		collectStrategy: collectStrategy,
		logger:          logger,
	}
}

func (sc *StreamCollector) StartStreaming(ctx context.Context) {
	ticker := time.NewTicker(sc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			sc.logger.Info("Context canceled: stopping stream")
			if _, opened := <-sc.streamTo; opened {
				close(sc.streamTo)
			}
			return
		case <-ticker.C:
			collected, err := sc.collectStrategy.Collect()

			if err != nil {
				sc.logger.Errorf("Collect failed with %T and error: %v", sc.collectStrategy, err)
				continue
			}

			if collected == nil || collected.Length() == 0 {
				sc.logger.Errorf("Recieved empty batch from %T and skip", sc.collectStrategy)
				continue
			}

			sc.streamTo <- collected
		}
	}
}
