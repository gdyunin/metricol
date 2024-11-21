package metrics

type Metric interface {
	Name() string
	Value() string
	Type() MetricType
	ParseFromURLString(string) error
}
