package metrics

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter MetricType = "counter"
	MetricTypeOther   MetricType = "other"

	ErrorParseMetricValue  = "error parse metric value"
	ErrorUnknownMetricType = "error unknown metric type"
	ErrorEmptyName         = "name was required but pass empty"
	ErrorEmptyValue        = "value was required but pass empty"
)

type MetricType string

type Metric interface {
	Name() string
	Value() string
	Type() MetricType
	SetName(string) error
	SetValue(string) error
}
