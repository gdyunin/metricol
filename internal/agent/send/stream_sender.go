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
	// Const updateBatchEndpoint defines the API endpoint for updating a batch of metrics.
	updateBatchEndpoint = "/updates"
	// Const attemptsDefaultCount defines the default number of attempts for retry calls.
	attemptsDefaultCount = 4
	// Const retryCalcContextKey is the key used to store the retry calculator in the request context.
	retryCalcContextKey contextKey = "retryCalculator"
)

// StreamSender provides functionality for sending batches of metrics to a remote server.
// It retrieves metrics from a channel, converts them to the model format, compresses the payload,
// and sends the HTTP request with built-in retry logic.
type StreamSender struct {
	httpClient     *resty.Client   // httpClient is the client used to send HTTP requests.
	requestBuilder *RequestBuilder // requestBuilder constructs HTTP requests with optional gzip compression.
	logger         *zap.SugaredLogger
	streamFrom     chan *entity.Metrics // streamFrom is the channel from which metrics batches are received.
	signingKey     string               // signingKey is used for signing the request payload.
	cryptoKey      string
	interval       time.Duration // interval defines the period between send attempts.
	maxPoolSize    int           // maxPoolSize limits the number of concurrent sending goroutines.
}

// NewStreamSender creates and initializes a new StreamSender instance.
//
// Parameters:
//   - streamFrom: A channel from which metric batches (entity.Metrics) are received.
//   - interval: The interval for sending metrics batches.
//   - maxPoolSize: The maximum number of concurrent sending operations.
//   - serverAddress: The base URL of the server to which metrics will be sent.
//   - signingKey: A key used for signing requests.
//   - logger: A logger for recording messages and errors.
//
// Returns:
//   - *StreamSender: A pointer to the initialized StreamSender.
func NewStreamSender(
	streamFrom chan *entity.Metrics,
	interval time.Duration,
	maxPoolSize int,
	serverAddress string,
	signingKey string,
	cryptoKey string,
	logger *zap.SugaredLogger,
) *StreamSender {
	// Ensure the server address has the proper HTTP scheme.
	if !strings.HasPrefix(serverAddress, "http://") && !strings.HasPrefix(serverAddress, "https://") {
		serverAddress = "http://" + strings.TrimPrefix(serverAddress, "/")
	}

	httpClient := resty.New().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBaseURL(serverAddress).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return err != nil || (r.StatusCode() >= 500 && r.StatusCode() <= 599)
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
		cryptoKey:      cryptoKey,
		streamFrom:     streamFrom,
		interval:       interval,
		maxPoolSize:    maxPoolSize,
	}
}

// StartStreaming begins the process of periodically sending metrics batches to the server.
// It uses a ticker to trigger send operations and stops when the provided context is canceled.
//
// Parameters:
//   - ctx: The context to control cancellation of the streaming operation.
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

// sendWithPool retrieves metric batches from the streamFrom channel and sends them concurrently.
// It launches up to maxPoolSize goroutines to handle sending in parallel.
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

// SendBatch sends a batch of metrics to the server using gzip compression and retry logic.
// It first converts the metrics from the entity format to the model format, then prepares and sends the request.
//
// Parameters:
//   - ctx: The context for the HTTP request.
//   - metrics: A pointer to an entity.Metrics batch containing the metrics to be sent.
//
// Returns:
//   - error: An error if the sending process fails; otherwise, nil.
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

// prepareAndSend prepares the HTTP request with the provided payload and sends it to the specified endpoint.
// It serializes the payload to JSON, compresses it, and then executes the request.
//
// Parameters:
//   - ctx: The context for the HTTP request.
//   - v: The payload to be sent, typically a converted metrics model.
//   - endpoint: The API endpoint for the request.
//
// Returns:
//   - error: An error if the preparation or execution of the request fails; otherwise, nil.
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

// prepareRequest builds an HTTP request with gzip compression from the provided payload.
// It serializes the payload to JSON, compresses the data, and constructs the request using the RequestBuilder.
// A retry calculator is added to the request context for managing retry intervals.
//
// Parameters:
//   - v: The payload to be sent, which is serialized to JSON.
//   - endpoint: The API endpoint for the request.
//
// Returns:
//   - *resty.Request: The prepared HTTP request.
//   - error: An error if serialization or request construction fails.
func (s *StreamSender) prepareRequest(v any, endpoint string) (*resty.Request, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("serialization of metrics to JSON failed: %w", err)
	}

	req, err := s.requestBuilder.BuildWithParams(http.MethodPost, endpoint, data, s.signingKey, s.cryptoKey)
	if err != nil {
		return nil, fmt.Errorf("request with params build failed: %w", err)
	}

	retryCalculator := retry.NewLinearRetryIterator(retry.DefaultLinearCoefficientScaling, -1)
	req.SetContext(context.WithValue(req.Context(), retryCalcContextKey, retryCalculator))

	return req, nil
}

// doRequest executes the given HTTP request and verifies that the response indicates success.
// It returns the HTTP response or an error if the request fails or if the response status code is not successful.
//
// Parameters:
//   - r: A pointer to the resty.Request to be executed.
//
// Returns:
//   - *resty.Response: The HTTP response from the server.
//   - error: An error if the request fails or if the server returns an unsuccessful status code.
func (s *StreamSender) doRequest(r *resty.Request) (resp *resty.Response, err error) {
	resp, err = r.Send()

	if err == nil && (resp.StatusCode() < 200 || resp.StatusCode() > 299) {
		err = fmt.Errorf("unsuccessful response from server: status code %s", resp.Status())
	} else if err != nil {
		s.logger.Errorf("Error during request execution: %v", err)
	}

	return
}
