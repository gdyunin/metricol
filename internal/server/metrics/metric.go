package metrics

type Metric interface {
	Name() string
	Value() string
	Type() MetricType
	SetName(string)
	SetValue(string) error
}
