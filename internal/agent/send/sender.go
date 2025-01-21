package send

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"
	"github.com/gdyunin/metricol.git/internal/agent/send/model"
	"github.com/gdyunin/metricol.git/pkg/retry"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	// UpdateBatchEndpoint defines the API endpoint for updating a batch of metrics.
	updateBatchEndpoint = "/updates"
	// AttemptsDefaultCount defines default count of attempts for retry calls.
	attemptsDefaultCount = 4
)

// MetricsSender provides functionality for sending metrics to a remote server.
// It handles requests with retry logic and gzip compression.
type MetricsSender struct {
	httpClient     *resty.Client      // HTTP client configured with retry and logging.
	requestBuilder *RequestBuilder    // Helper for building HTTP requests.
	logger         *zap.SugaredLogger // Logger for structured logging.
	signingKey     string             // Key used for signing requests to the server.
}

// NewMetricsSender creates and initializes a new MetricsSender instance.
//
// Parameters:
//   - serverAddress: The base URL of the server to which metrics will be sent.
//   - logger: A logger for logging messages and errors.
//
// Returns:
//   - *MetricsSender: A new instance of MetricsSender with pre-configured settings.
func NewMetricsSender(serverAddress string, signingKey string, logger *zap.SugaredLogger) *MetricsSender {
	// [ДЛЯ РЕВЬЮ]: Это должно быть в отдельной функции и гораздо сложнее. Но для текущих нужд пока так сойдет)).
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
		SetRetryCount(attemptsDefaultCount).
		SetRetryAfter(func(client *resty.Client, response *resty.Response) (time.Duration, error) {
			currentAttempt := response.Request.Attempt
			if currentAttempt > client.RetryCount {
				return 0, nil
			}
			// Linear retry delay: y = 2x - 1, where x is the attempt number.
			// For example, the delays for the first three attempts are:
			// Attempt 1: 2*1 - 1 = 1 second.
			// Attempt 2: 2*2 - 1 = 3 seconds.
			// Attempt 3: 2*3 - 1 = 5 seconds.
			// This logic is a requirement of the technical specification.
			return retry.CalcByLinear(currentAttempt, retry.DefaultLinearCoefficientScaling, -1), nil
		}).
		SetLogger(logger.Named("http_client"))

	requestBuilder := NewRequestBuilder(httpClient)

	logger.Infof("Initialized MetricsSender with server address: %s", serverAddress)
	return &MetricsSender{
		httpClient:     httpClient,
		requestBuilder: requestBuilder,
		logger:         logger,
		signingKey:     signingKey,
	}
}

// SendBatch sends a batch of metrics to the server using gzip compression.
//
// Parameters:
//   - ctx: The context for the HTTP request.
//   - metrics: A pointer to the Metrics entity containing multiple metrics to be sent.
//
// Returns:
//   - error: An error if the operation fails, or nil if successful.
func (s *MetricsSender) SendBatch(ctx context.Context, metrics *entity.Metrics) error {
	modelsMetric, err := model.NewFromEntityMetrics(metrics)
	if err != nil {
		return fmt.Errorf("conversion of metrics to models failed: %w", err)
	}

	if err = s.prepareAndSend(ctx, modelsMetric, updateBatchEndpoint); err != nil {
		return fmt.Errorf("error during preparation or sending of batch request: %w", err)
	}

	return nil
}

// prepareAndSend prepares the HTTP request and sends it to the specified endpoint.
//
// Parameters:
//   - ctx: The context for the HTTP request.
//   - v: The payload to be sent, serialized as JSON.
//   - endpoint: The API endpoint for the request.
//
// Returns:
//   - error: An error if the operation fails, or nil if successful.
func (s *MetricsSender) prepareAndSend(ctx context.Context, v any, endpoint string) error {
	req, err := s.prepareRequest(v, endpoint)
	if err != nil {
		return fmt.Errorf("request preparation failed: %w", err)
	}
	req.SetContext(ctx)

	if _, err = s.doRequest(req); err != nil {
		return fmt.Errorf("request execution failed: %w", err)
	}

	return nil
}

// prepareRequest builds an HTTP request with gzip compression.
//
// Parameters:
//   - v: The payload to be sent, serialized as JSON.
//   - endpoint: The API endpoint for the request.
//
// Returns:
//   - *resty.Request: The prepared HTTP request.
//   - error: An error if the request could not be prepared.
func (s *MetricsSender) prepareRequest(v any, endpoint string) (*resty.Request, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("serialization of metrics to JSON failed: %w", err)
	}

	req, err := s.requestBuilder.BuildWithGzip(http.MethodPost, endpoint, data, s.signingKey)
	if err != nil {
		return nil, fmt.Errorf("gzip-compressed request build failed: %w", err)
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
		err = fmt.Errorf("unsuccessful response from server: status code %s", resp.Status())
	} else if err != nil {
		s.logger.Errorf("Error during request execution: %v", err)
	}

	return
}
