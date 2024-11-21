package metrics

type Metric interface {
	Name() string
	Type() MetricType
}
