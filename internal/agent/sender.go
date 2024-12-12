package agent

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sync"

	"github.com/gdyunin/metricol.git/internal/metrics"
	"github.com/go-resty/resty/v2"
)

type Observer interface {
	OnNotify()
}

type MetricsSender struct {
	mu             *sync.Mutex
	observers      map[Observer]struct{}
	metricsFetcher Fetcher
	client         *resty.Client
	serverAddress  string
}

func NewMetricsSender(fetcher Fetcher, address string) *MetricsSender {
	return &MetricsSender{
		metricsFetcher: fetcher,
		serverAddress:  address,
		client:         resty.New(),
		observers:      make(map[Observer]struct{}),
	}
}

func (m *MetricsSender) Send() error {
	var buf bytes.Buffer
	for _, mm := range m.metricsFetcher.Metrics() {
		metricType, metricName, metricValue, ok := m.recognizeMetric(mm)
		if !ok {
			buf.WriteString(fmt.Sprintf("error sending metric %v: failed conversion Metric to Struct\n", mm))
			continue
		}

		req := m.makeRequest(metricType, metricName, metricValue)
		if _, err := req.Send(); err != nil {
			buf.WriteString(fmt.Sprintf("error sending metric %v: %v\n", mm, err))
		}
	}

	if buf.Len() != 0 {
		return fmt.Errorf("one or more metrics were not sent to the server: %s", buf.String())
	}

	m.Notify()
	return nil
}

func (m *MetricsSender) RegisterObserver(observer Observer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.observers[observer] = struct{}{}
}

func (m *MetricsSender) RemoveObserver(observer Observer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.observers, observer)
}

func (m *MetricsSender) Notify() {
	for o := range m.observers {
		o.OnNotify()
	}
}

func (m *MetricsSender) makeRequest(mType, mName, mValue string) *resty.Request {
	u := url.URL{
		Scheme: "http",
		Path:   path.Join(m.serverAddress, "/update/", mType, mName, mValue),
	}

	req := m.client.R()
	req.Method = http.MethodPost
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	req.URL = u.String()

	return req
}

func (m *MetricsSender) recognizeMetric(mm metrics.Metric) (metricType, name, value string, ok bool) {
	name = mm.StringName()
	value = mm.StringValue()
	ok = true

	switch mm.(type) {
	case *metrics.Counter:
		metricType = metrics.MetricTypeCounter
	case *metrics.Gauge:
		metricType = metrics.MetricTypeGauge
	default:
		return "", "", "", false
	}

	return
}
