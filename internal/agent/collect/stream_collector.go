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
	streamTo          chan *entity.Metrics
	logger            *zap.SugaredLogger
	collectStrategies []Strategy
	interval          time.Duration
}

func NewStreamCollector(
	streamTo chan *entity.Metrics,
	interval time.Duration,
	collectStrategies []Strategy,
	logger *zap.SugaredLogger,
) *StreamCollector {
	return &StreamCollector{
		streamTo:          streamTo,
		interval:          interval,
		collectStrategies: collectStrategies,
		logger:            logger,
	}
}

func (sc *StreamCollector) StartStreaming(ctx context.Context) {
	ticker := time.NewTicker(sc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			sc.logger.Info("Context canceled: stopping stream")
			close(sc.streamTo)
			return
		case <-ticker.C:
			for _, strategy := range sc.collectStrategies {
				// [ДЛЯ РЕВЬЮ]: По-хорошему это надо делать через пул рабочих или семафор,
				// [ДЛЯ РЕВЬЮ]: Но у нас тут не предвидится большого числа стратегий одновременно,
				// [ДЛЯ РЕВЬЮ]: Поэтому ограничение будет избыточно на текущем этапе проекта.
				go func(s Strategy) {
					collected, err := s.Collect()
					if err != nil {
						sc.logger.Errorf("Collect failed with %T and error: %v", sc.collectStrategies, err)
						return
					}

					if collected == nil || collected.Length() == 0 {
						sc.logger.Errorf("Recieved empty batch from %T and skip", sc.collectStrategies)
						return
					}

					sc.streamTo <- collected
				}(strategy)
			}
		}
	}
}
