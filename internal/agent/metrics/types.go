package metrics

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter MetricType = "counter"
)

type MetricType string

type Metric interface {
	Type() MetricType
	Name() string
	Value() string
	UpdateValue()
}
