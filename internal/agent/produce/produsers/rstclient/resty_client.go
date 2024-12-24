package rstclient

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/adapter/produce"
	"github.com/gdyunin/metricol.git/internal/agent/common"
	"github.com/gdyunin/metricol.git/internal/agent/entity"
	"github.com/gdyunin/metricol.git/internal/agent/produce/produsers/rstclient/model"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	resetErrorCountersIntervals = 4
	maxErrorsToInterrupt        = 3
)

// RestyClient is a producer that sends metrics to a server using the Resty library.
type RestyClient struct {
	adp         *produce.RestyClientAdapter
	client      *resty.Client
	ticker      *time.Ticker
	interrupter *common.Interrupter
	mu          *sync.RWMutex
	log         *zap.SugaredLogger
	observers   map[common.Observer]struct{}
	interval    time.Duration
	baseUrl     string
}

// NewRestyClient creates a new RestyClient instance with the specified interval,
// server address, repository, and logger.
func NewRestyClient(
	interval time.Duration,
	serverAddress string,
	repo entity.MetricsRepository,
	logger *zap.SugaredLogger,
) *RestyClient {
	rc := resty.New()

	return &RestyClient{
		adp:       produce.NewRestyClientAdapter(repo, logger.Named("resty client adapter")),
		client:    rc,
		interval:  interval,
		observers: make(map[common.Observer]struct{}),
		mu:        &sync.RWMutex{},
		log:       logger,
		baseUrl:   serverAddress,
	}
}

// StartProduce begins the metrics production process.
// It runs in a loop until an error occurs or the interrupter stops it.
func (r *RestyClient) StartProduce() error {
	r.ticker = time.NewTicker(r.interval)
	defer r.ticker.Stop()

	interrupter, err := common.NewInterrupter(r.interval*resetErrorCountersIntervals, maxErrorsToInterrupt)
	if err != nil {
		return fmt.Errorf("failed to initialize interrupter: %w", err)
	}
	r.interrupter = interrupter
	defer r.interrupter.Stop()

	r.log.Info("Metrics production process started.")
	for {
		select {
		case <-r.ticker.C:
			if err := r.sendAll(); err != nil {
				return fmt.Errorf("metrics production process failed: %w", err)
			}
		case <-r.interrupter.C:
			return errors.New("error limit exceeded during metrics production")
		}
	}
}

// RegisterObserver adds a new observer to be notified of events.
func (r *RestyClient) RegisterObserver(observer common.Observer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.observers[observer]; exists {
		return fmt.Errorf("observer %v is already registered", observer)
	}

	r.observers[observer] = struct{}{}
	r.log.Infof("Successfully registered observer: %v.", observer)
	return nil
}

// RemoveObserver removes an existing observer from the notification list.
func (r *RestyClient) RemoveObserver(observer common.Observer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.observers[observer]; exists {
		delete(r.observers, observer)
		r.log.Infof("Successfully removed observer: %v.", observer)
		return nil
	}

	return fmt.Errorf("observer %v is not registered", observer)
}

// NotifyObservers sends notifications to all registered observers.
func (r *RestyClient) NotifyObservers() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for o := range r.observers {
		o.OnNotify()
	}

	r.log.Info("All registered observers have been notified.")
}

// sendAll transmits all metrics to the server and notifies observers.
func (r *RestyClient) sendAll() error {
	r.log.Info("Initiating metrics transmission.")
	metrics := r.adp.Metrics()

	for _, m := range metrics {
		if !r.interrupter.InLimit() {
			return errors.New("metrics transmission halted: error limit exceeded")
		}

		if err := r.send(m); err != nil {
			r.log.Errorf("Failed to send metric '%v': %v", m, err)
			r.interrupter.AddError()
		}
	}

	r.NotifyObservers()
	r.log.Info("Completed metrics transmission.")
	return nil
}

// send transmits a single metric to the server.
func (r *RestyClient) send(metric *model.Metric) error {
	r.log.Infof("Transmitting metric: %v.", metric)
	req := r.makeRequest()

	body, _ := json.Marshal(metric)
	//buf := bytes.NewReader(body)

	//compressedBody, err := compressBody(body)
	//if err != nil {
	//	r.log.Info("Metric compression failed. Sending uncompressed data.")
	//	req.Body = body
	//} else {
	//	req.Body = compressedBody
	//	req.Header.Set("Content-Encoding", "gzip")
	//}
	re, err := resty.New().R().Get("http://" + r.baseUrl + "/ping")
	if err != nil {
		panic(err)
	}
	r.log.Infof("%+v", re)

	req.SetBody(body)
	resp, err := req.Send()
	r.log.Infof("%+v", req)
	r.log.Infof("%+v", req.Header)
	r.log.Infof("%+v", req.URL)
	r.log.Infof("%+v", req.Body)
	r.log.Infof("%+v", req.QueryParam)
	r.log.Infof("%+v", req.Result)

	if err != nil {
		panic(err)
		//return fmt.Errorf("failed to send metric %v: %w", metric, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to send metric %v: server returned status code %d", metric, resp.StatusCode())
	}

	return nil
}

// makeRequest prepares a new HTTP request for transmitting metrics.
func (r *RestyClient) makeRequest() *resty.Request {
	u := url.URL{
		Scheme: "http",
		Host:   r.baseUrl,
		Path:   "/update",
	}

	req := resty.New().R()
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Host", u.Host)
	req.Method = http.MethodPost
	req.URL = u.String()
	//req.Header.Set("Accept-Encoding", "gzip")

	return req
}

// compressBody compresses the given data using GZIP compression.
func compressBody(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GZIP writer: %w", err)
	}

	if _, err = w.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write data to GZIP buffer: %w", err)
	}

	if err = w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close GZIP writer: %w", err)
	}

	return buf.Bytes(), nil
}
