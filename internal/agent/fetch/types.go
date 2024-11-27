package fetch

import "github.com/gdyunin/metricol.git/internal/metrics"

type Fetcher interface {
	Fetch()
	Metrics() []metrics.Metric
}
