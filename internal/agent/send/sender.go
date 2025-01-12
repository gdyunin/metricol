package send

import (
	"NewNewMetricol/internal/agent/internal/entity"
	"NewNewMetricol/internal/agent/send/model"
	"NewNewMetricol/pkg/retry"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	// UpdateSingleEndpoint defines the API endpoint for updating a single metric.
	UpdateSingleEndpoint = "/update"
	// UpdateBatchEndpoint defines the API endpoint for updating a batch of metric.
	UpdateBatchEndpoint = "/updates"
)

// MetricsSender provides functionality for sending metrics to a remote server.
type MetricsSender struct {
	httpClient     *resty.Client
	requestBuilder *RequestBuilder
	logger         *zap.SugaredLogger
}

// NewMetricsSender creates and initializes a new MetricsSender instance.
//
// Parameters:
//   - serverAddress: The base URL of the server to which metrics will be sent.
//
// Returns:
//   - *MetricsSender: A new instance of MetricsSender.
func NewMetricsSender(serverAddress string, logger *zap.SugaredLogger) *MetricsSender {
	if !strings.HasPrefix(serverAddress, "http://") && !strings.HasPrefix(serverAddress, "https://") {
		serverAddress = "http://" + strings.TrimPrefix(serverAddress, "/")
	}

	httpClient := resty.New().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBaseURL(serverAddress).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return err != nil || r.StatusCode() >= 500 && r.StatusCode() <= 599
		}).
		SetRetryCount(3).
		SetRetryAfter(func(client *resty.Client, response *resty.Response) (time.Duration, error) {
			currentAttempt := response.Request.Attempt
			if currentAttempt > client.RetryCount {
				return 0, nil
			}

			// 1=>1s; 2=>3s; 3=>5s; ... -- линейная зависимость, которую можно выразить как y = 2x - 1
			return retry.CalcByLinear(currentAttempt, 2, -1), nil
		}).
		SetLogger(logger.Named("http_client"))

	requestBuilder := NewRequestBuilder(httpClient)

	logger.Infof("Init sender with server address: %s", serverAddress)
	return &MetricsSender{
		httpClient:     httpClient,
		requestBuilder: requestBuilder,
		logger:         logger,
	}
}

// SendSingle sends a single metric to the server using gzip compression.
//
// Parameters:
//   - ctx: The context for the HTTP request.
//   - metric: A pointer to the Metric entity to be sent.
//
// Returns:
//   - error: An error if the operation fails, or nil if successful.
func (s *MetricsSender) SendSingle(ctx context.Context, metric *entity.Metric) error {
	modelMetric, err := model.NewFromEntityMetric(metric)
	if err != nil {
		return fmt.Errorf("failed to conver metric to model: %w", err)
	}

	if err = s.prepareAndSend(ctx, modelMetric, UpdateSingleEndpoint); err != nil {
		return fmt.Errorf("failed while preparing or sending request: %w", err)
	}

	s.logger.Infof("Metric %v sended successful", *metric)

	return nil
}

func (s *MetricsSender) SendBatch(ctx context.Context, metrics *entity.Metrics) error {
	modelsMetric, err := model.NewFromEntityMetrics(metrics)
	if err != nil {
		return fmt.Errorf("failed to conver metrics to models: %w", err)
	}

	if err = s.prepareAndSend(ctx, modelsMetric, UpdateBatchEndpoint); err != nil {
		return fmt.Errorf("failed while preparing or sending request: %w", err)
	}

	s.logger.Infof("Metrics sended successful: [%s]", metrics.ToString())

	return nil
}

func (s *MetricsSender) prepareAndSend(ctx context.Context, v any, endpoint string) error {
	req, err := s.prepareRequest(v, endpoint)
	if err != nil {
		return fmt.Errorf("failed to prepare request: %w", err)
	}
	req.SetContext(ctx)

	if _, err = s.doRequest(req); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	return nil
}

func (s *MetricsSender) prepareRequest(v any, endpoint string) (*resty.Request, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metrics: %w", err)
	}

	req, err := s.requestBuilder.BuildWithGzip(http.MethodPost, endpoint, data)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	return req, nil
}

// doRequest executes the given HTTP request and checks for successful response status codes.
//
// Parameters:
//   - r: A pointer to the resty.Request to be executed.
//
// Returns:
//   - *resty.Response: The HTTP response from the server.
//   - error: An error if the request fails or the response status code is not successful.
func (s *MetricsSender) doRequest(r *resty.Request) (resp *resty.Response, err error) {
	resp, err = r.Send()

	if err == nil && (resp.StatusCode() < 200 || resp.StatusCode() > 299) {
		err = fmt.Errorf("server responded with unsuccessful status code: %s", resp.Status())
	}

	return
}
