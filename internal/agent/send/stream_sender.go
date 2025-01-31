package send

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gdyunin/metricol.git/internal/agent/internal/entity"
	"github.com/gdyunin/metricol.git/internal/agent/send/model"
	"github.com/gdyunin/metricol.git/pkg/retry"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type contextKey string

const (
	// UpdateBatchEndpoint defines the API endpoint for updating a batch of metrics.
	updateBatchEndpoint = "/updates"
	// AttemptsDefaultCount defines default count of attempts for retry calls.
	attemptsDefaultCount            = 4
	retryCalcContextKey  contextKey = "retryCalculator"
)

// StreamSender provides functionality for sending metrics to a remote server.
// It handles requests with retry logic and gzip compression.
type StreamSender struct {
	httpClient     *resty.Client
	requestBuilder *RequestBuilder
	logger         *zap.SugaredLogger
	streamFrom     chan *entity.Metrics
	signingKey     string
	interval       time.Duration
	maxPoolSize    int
}

// NewStreamSender creates and initializes a new StreamSender instance.
//
// Parameters:
//   - serverAddress: The base URL of the server to which metrics will be sent.
//   - logger: A logger for logging messages and errors.
//
// Returns:
//   - *StreamSender: A new instance of StreamSender with pre-configured settings.
func NewStreamSender(
	streamFrom chan *entity.Metrics,
	interval time.Duration,
	maxPoolSize int,
	serverAddress string,
	signingKey string,
	logger *zap.SugaredLogger,
) *StreamSender {
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

			retryCalculator, ok := response.Request.Context().Value(retryCalcContextKey).(*retry.LinearRetryIterator)
			if !ok {
				logger.Warn("Retry interval calculator not found in request context, retry cancelled")
				return 0, nil
			}

			switch currentAttempt {
			case 1:
				retryCalculator.SetCurrentAttempt(1)
			case client.RetryCount + 1:
				return 0, nil
			}

			return retryCalculator.Next(), nil
		}).
		SetLogger(logger.Named("http_client"))

	requestBuilder := NewRequestBuilder(httpClient)

	logger.Infof("Initialized StreamSender with server address: %s", serverAddress)
	return &StreamSender{
		httpClient:     httpClient,
		requestBuilder: requestBuilder,
		logger:         logger,
		signingKey:     signingKey,
		streamFrom:     streamFrom,
		interval:       interval,
		maxPoolSize:    maxPoolSize,
	}
}

func (s *StreamSender) StartStreaming(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Context canceled: stopping stream")
			return
		case <-ticker.C:
			s.sendWithPool(ctx)
		}
	}
}

func (s *StreamSender) sendWithPool(ctx context.Context) {
	var wg sync.WaitGroup

	for range s.maxPoolSize {
		select {
		case <-ctx.Done():
			s.logger.Info("Context canceled: cancel send")
			return
		default:
			wg.Add(1)
			go func() {
				defer wg.Done()
				metrics, ok := <-s.streamFrom
				if !ok {
					s.logger.Info("StreamFrom channel was closed, stop sending")
					return
				}
				if metrics == nil || metrics.Length() == 0 {
					return
				}
				s.logger.Infof("Preparing to send %d metrics in batch", metrics.Length())
				err := s.SendBatch(ctx, metrics)
				if err != nil {
					s.logger.Errorf("Failed to send metrics batch: count=%d, error=%v", metrics.Length(), err)
					return
				}
				s.logger.Infof("Successfully sent batch of metrics: count=%d", metrics.Length())
			}()
		}
	}

	wg.Wait()
}

// SendBatch sends a batch of metrics to the server using gzip compression.
//
// Parameters:
//   - ctx: The context for the HTTP request.
//   - metrics: A pointer to the Metrics entity containing multiple metrics to be sent.
//
// Returns:
//   - error: An error if the operation fails, or nil if successful.
func (s *StreamSender) SendBatch(ctx context.Context, metrics *entity.Metrics) error {
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
func (s *StreamSender) prepareAndSend(ctx context.Context, v any, endpoint string) error {
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
func (s *StreamSender) prepareRequest(v any, endpoint string) (*resty.Request, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("serialization of metrics to JSON failed: %w", err)
	}

	req, err := s.requestBuilder.BuildWithGzip(http.MethodPost, endpoint, data, s.signingKey)
	if err != nil {
		return nil, fmt.Errorf("gzip-compressed request build failed: %w", err)
	}

	retryCalculator := retry.NewLinearRetryIterator(retry.DefaultLinearCoefficientScaling, -1)
	req.SetContext(context.WithValue(req.Context(), retryCalcContextKey, retryCalculator))

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
func (s *StreamSender) doRequest(r *resty.Request) (resp *resty.Response, err error) {
	resp, err = r.Send()

	if err == nil && (resp.StatusCode() < 200 || resp.StatusCode() > 299) {
		err = fmt.Errorf("unsuccessful response from server: status code %s", resp.Status())
	} else if err != nil {
		s.logger.Errorf("Error during request execution: %v", err)
	}

	return
}
