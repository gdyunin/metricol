package metrics

type Metric interface {
	Name() string
	Value() string
	Type() MetricType
	ParseFromURLString(string) error
	SetName(string)
	SetValue(string) error
}
