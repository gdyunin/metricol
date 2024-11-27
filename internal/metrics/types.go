package metrics

const (
	MetricTypeGauge   = "gauge"
	MetricTypeCounter = "counter"

	ErrorParseMetricValue  = "error parse metric value"
	ErrorUnknownMetricType = "error unknown metric type"
	ErrorFetcherNotSet     = "error fetcher not set"
)

type Metric interface {
	Name() string
	StringValue() string
	Type() string
	Update() error
}
