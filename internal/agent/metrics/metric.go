package metrics

type Metric interface {
	Type() MetricType
	Name() string
	Value() string
	UpdateValue()
}
