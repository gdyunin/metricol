package send

import (
	"fmt"
	"github.com/gdyunin/metricol.git/internal/agent/fetch"
	"github.com/go-resty/resty/v2"
	"net/http"
	"net/url"
	"path"
)

type MetricsSender struct {
	metricsFetcher fetch.Fetcher
	serverAddress  string
	client         *resty.Client
}

func NewMetricsSender(fetcher fetch.Fetcher, address string) *MetricsSender {
	return &MetricsSender{
		metricsFetcher: fetcher,
		serverAddress:  address,
		client:         resty.New(),
	}
}

func (m *MetricsSender) Send() {
	for _, mm := range m.metricsFetcher.Metrics() {
		u := url.URL{
			Scheme: "http",
			Path:   path.Join(m.serverAddress, "/update/", mm.Type(), mm.Name(), mm.StringValue()),
		}
		req := m.client.R()
		req.Method = http.MethodPost
		req.Header.Set("Content-Type", "text/plain; charset=utf-8")
		req.URL = u.String()

		if _, err := req.Send(); err != nil {
			// A logger could be added in the future
			fmt.Println(err.Error())
		}

	}
}
