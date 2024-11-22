package metrics

type MetricType string

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter MetricType = "counter"
	MetricTypeOther   MetricType = "other"
)
