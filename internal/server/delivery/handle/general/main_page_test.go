package general

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTemplate is a minimal implementation of echo.Renderer for testing.
type MockTemplate struct {
	RenderFunc func(w io.Writer, name string, data interface{}, c echo.Context) error
}

func (t *MockTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if t.RenderFunc != nil {
		return t.RenderFunc(w, name, data, c)
	}
	return nil
}

// MockPullerAll implements the PullerAll interface for testing.
type MockPullerAll struct {
	Metrics    *entity.Metrics
	ShouldFail bool
	Delay      time.Duration
}

// PullAll implements the PullerAll interface.
func (m *MockPullerAll) PullAll(ctx context.Context) (*entity.Metrics, error) {
	if m.Delay > 0 {
		select {
		case <-time.After(m.Delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.ShouldFail {
		return nil, errors.New("failed to pull metrics")
	}
	return m.Metrics, nil
}

func TestMainPage(t *testing.T) {
	tests := []struct {
		puller         PullerAll
		name           string
		expectedStatus int
		expectedRows   int
		checkTemplate  bool
	}{
		{
			name: "Success with multiple metrics",
			puller: &MockPullerAll{
				Metrics: &entity.Metrics{
					&entity.Metric{Name: "metric1", Type: entity.MetricTypeCounter, Value: int64(10)},
					&entity.Metric{Name: "metric2", Type: entity.MetricTypeGauge, Value: 20.5},
				},
			},
			expectedStatus: http.StatusOK,
			checkTemplate:  true,
			expectedRows:   2,
		},
		{
			name: "Success with one metric",
			puller: &MockPullerAll{
				Metrics: &entity.Metrics{
					&entity.Metric{Name: "single", Type: entity.MetricTypeCounter, Value: int64(42)},
				},
			},
			expectedStatus: http.StatusOK,
			checkTemplate:  true,
			expectedRows:   1,
		},
		{
			name: "Success with empty metrics",
			puller: &MockPullerAll{
				Metrics: &entity.Metrics{},
			},
			expectedStatus: http.StatusOK,
			checkTemplate:  true,
			expectedRows:   0,
		},
		{
			name: "Error pulling metrics",
			puller: &MockPullerAll{
				ShouldFail: true,
			},
			expectedStatus: http.StatusInternalServerError,
			checkTemplate:  false,
		},
		{
			name:           "Nil metrics",
			puller:         &MockPullerAll{},
			expectedStatus: http.StatusInternalServerError,
			checkTemplate:  false,
		},
		{
			name: "Timeout pulling metrics",
			puller: &MockPullerAll{
				Delay: 6 * time.Second, // Longer than pullAllTimeout (5s).
			},
			expectedStatus: http.StatusInternalServerError,
			checkTemplate:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()

			var templateData interface{}
			var templateCalled bool

			// Set up mock renderer.
			e.Renderer = &MockTemplate{
				RenderFunc: func(_ io.Writer, name string, data interface{}, _ echo.Context) error {
					templateCalled = true
					templateData = data
					assert.Equal(t, "main_page.html", name)
					return nil
				},
			}

			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute handler.
			handler := MainPage(tt.puller)
			err := handler(c)

			// Verify response status.
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Check template rendering.
			if tt.checkTemplate {
				assert.True(t, templateCalled, "Template should have been rendered")
				if templateData != nil {
					tableRows, ok := templateData.([]*tr)
					require.True(t, ok, "Template data should be []*tr")
					assert.Len(t, tableRows, tt.expectedRows)

					// If we have metrics to check, verify they were passed correctly.
					if tt.puller != nil && tt.puller.(*MockPullerAll).Metrics != nil {
						metrics := tt.puller.(*MockPullerAll).Metrics
						for i, metric := range *metrics {
							if i < len(tableRows) {
								assert.Equal(t, metric.Name, tableRows[i].Name)
							}
						}
					}
				}
			} else {
				assert.False(t, templateCalled, "Template should not have been rendered")
				assert.Equal(t, http.StatusText(http.StatusInternalServerError), rec.Body.String())
			}
		})
	}
}

// ExampleMainPage demonstrates how to use the MainPage handler.
// It sets up a dummy puller that returns two metrics, creates an Echo instance with a mock renderer,
// invokes the MainPage handler, and prints the rendered output.
func ExampleMainPage() {
	// Create a dummy puller that returns two metrics.
	puller := &MockPullerAll{
		Metrics: &entity.Metrics{
			&entity.Metric{
				Name:  "metric1",
				Type:  entity.MetricTypeCounter,
				Value: int64(10),
			},
			&entity.Metric{
				Name:  "metric2",
				Type:  entity.MetricTypeGauge,
				Value: 20.5,
			},
		},
	}

	// Create a new Echo instance.
	e := echo.New()

	// Set up a mock renderer that writes to a buffer.
	e.Renderer = &MockTemplate{
		RenderFunc: func(w io.Writer, tmplName string, data interface{}, _ echo.Context) error {
			// Write the template name.
			_, _ = fmt.Fprintf(w, "Template: %s\n", tmplName)
			// Assert that data is a slice of table rows.
			rows, ok := data.([]*tr)
			if ok {
				for _, row := range rows {
					_, _ = fmt.Fprintf(w, "%s: %s\n", row.Name, row.Value)
				}
			}
			return nil
		},
	}

	// Create an HTTP request and recorder.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Invoke the MainPage handler.
	handler := MainPage(puller)
	_ = handler(c)

	// Print the rendered output.
	fmt.Print(rec.Body.String())

	// Output:
	// Template: main_page.html
	// metric1: 10
	// metric2: 20.5
}
