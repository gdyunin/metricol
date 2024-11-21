package memstorage

import "github.com/gdyunin/metricol.git/internal/server/metrics"

type MemStorage interface {
	PushMetric(metric metrics.Metric)
	PullMetric()
}

type BaseMemStorage struct {
}

func NewBaseMemStorage() BaseMemStorage {
	return BaseMemStorage{}
}

func (b *BaseMemStorage) PushMetric(metric metrics.Metric) {
	//TODO implement me
	//panic("implement me")
}

func (b *BaseMemStorage) PullMetric() {
	//TODO implement me
	//panic("implement me")
}
