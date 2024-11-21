package metrics

type MetricType string

const (
	MetricTypeGauge   MetricType = "Gauge"
	MetricTypeCounter MetricType = "Counter"
)
