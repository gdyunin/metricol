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
	common2 "github.com/gdyunin/metricol.git/internal/common"

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
	observers   map[common2.Observer]struct{}
	baseURL     string
	interval    time.Duration
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
		observers: make(map[common2.Observer]struct{}),
		mu:        &sync.RWMutex{},
		log:       logger,
		baseURL:   serverAddress,
	}
}

// waitServer checks if the server is available by sending a ping request.
// It retries up to a maximum number of attempts with increasing intervals.
func (r *RestyClient) waitServer() error {
	const (
		maxRetries  = 10               // Maximum number of retry attempts.
		minInterval = 1 * time.Second  // Minimum interval between attempts.
		maxInterval = 10 * time.Second // Maximum interval between attempts.
	)

	for i := range make([]struct{}, maxRetries) {
		r.log.Infof("Checking server availability... Attempt %d/%d", i+1, maxRetries)

		// Send a ping request to check server availability.
		resp, err := r.client.R().Get(fmt.Sprintf("http://%s/ping", r.baseURL))
		if err == nil && resp.StatusCode() == http.StatusOK {
			r.log.Info("Server is available.")
			return nil
		}

		interval := time.Duration(i+1) * minInterval
		if interval > maxInterval {
			interval = maxInterval
		}

		r.log.Warnf("Server not available: %v. Retrying in %s...", err, interval)
		time.Sleep(interval)
	}

	return errors.New("server did not become available within the retry limit")
}

// StartProduce starts the metrics production process.
// It runs in a loop until an error occurs or the interrupter stops it.
func (r *RestyClient) StartProduce() error {
	if err := r.waitServer(); err != nil {
		return fmt.Errorf("server is not available: %w", err)
	}

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
func (r *RestyClient) RegisterObserver(observer common2.Observer) error {
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
func (r *RestyClient) RemoveObserver(observer common2.Observer) error {
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

// sendAll sends all metrics to the server and notifies observers.
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

// send sends a single metric to the server.
func (r *RestyClient) send(metric *model.Metric) error {
	r.log.Infof("Transmitting metric: %v.", metric)
	req := r.makeRequest()

	body, _ := json.Marshal(metric)
	compressedBody, err := compressBody(body)
	if err != nil {
		r.log.Info("Metric compression failed. Sending uncompressed data.")
		req.SetBody(body)
	} else {
		req.SetBody(compressedBody)
		req.Header.Set("Content-Encoding", "gzip")
	}

	resp, err := req.Send()
	if err != nil {
		return fmt.Errorf("failed to send metric %v: %w", metric, err)
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
		Host:   r.baseURL,
		Path:   "/update",
	}

	req := resty.New().R()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Method = http.MethodPost
	req.URL = u.String()

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
