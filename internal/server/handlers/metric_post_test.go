package handlers

import (
	"github.com/gdyunin/metricol.git/internal/server/metrics/library"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricPostHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}
	type request struct {
		uri    string
		method string
		header http.Header
	}
	tests := []struct {
		name       string
		repository storage.Repository
		want       want
		request    request
	}{
		{
			"post new valid gauge",
			storage.NewWarehouse(),
			want{
				"text/plain",
				http.StatusOK,
			},
			request{
				"/update/gauge/mainLifeQuestion/4.2",
				http.MethodPost,
				func() http.Header {
					header := http.Header{}
					header.Set("Content-Type", "text/plain")
					return header
				}(),
			},
		},
		{
			"post new invalid gauge",
			storage.NewWarehouse(),
			want{
				"",
				http.StatusBadRequest,
			},
			request{
				"/update/gauge/mainLifeQuestion/4.2.2",
				http.MethodPost,
				func() http.Header {
					header := http.Header{}
					header.Set("Content-Type", "text/plain")
					return header
				}(),
			},
		},
		{
			"post new valid counter",
			storage.NewWarehouse(),
			want{
				"text/plain",
				http.StatusOK,
			},
			request{
				"/update/counter/mainLifeQuestion/42",
				http.MethodPost,
				func() http.Header {
					header := http.Header{}
					header.Set("Content-Type", "text/plain")
					return header
				}(),
			},
		},
		{
			"post new invalid counter",
			storage.NewWarehouse(),
			want{
				"",
				http.StatusBadRequest,
			},
			request{
				"/update/counter/mainLifeQuestion/4.2",
				http.MethodPost,
				func() http.Header {
					header := http.Header{}
					header.Set("Content-Type", "text/plain")
					return header
				}(),
			},
		},
		{
			"post repeat valid gauge",
			func() storage.Repository {
				s := storage.NewWarehouse()
				m := library.NewGauge()
				m.SetName("mainLifeQuestion")
				_ = m.SetValue("4.2")
				_ = s.PushMetric(m)
				return s
			}(),
			want{
				"text/plain",
				http.StatusOK,
			},
			request{
				"/update/gauge/mainLifeQuestion/4.2",
				http.MethodPost,
				func() http.Header {
					header := http.Header{}
					header.Set("Content-Type", "text/plain")
					return header
				}(),
			},
		},
		{
			"post repeat valid counter",
			func() storage.Repository {
				s := storage.NewWarehouse()
				m := library.NewCounter()
				m.SetName("mainLifeQuestion")
				_ = m.SetValue("42")
				_ = s.PushMetric(m)
				return s
			}(),
			want{
				"text/plain",
				http.StatusOK,
			},
			request{
				"/update/counter/mainLifeQuestion/42",
				http.MethodPost,
				func() http.Header {
					header := http.Header{}
					header.Set("Content-Type", "text/plain")
					return header
				}(),
			},
		},
		{
			"post valid metric with invalid method [GET]",
			storage.NewWarehouse(),
			want{
				"",
				http.StatusMethodNotAllowed,
			},
			request{
				"/update/counter/mainLifeQuestion/42",
				http.MethodGet,
				func() http.Header {
					header := http.Header{}
					header.Set("Content-Type", "text/plain")
					return header
				}(),
			},
		},
		{
			"post valid metric with invalid method [PUT]",
			storage.NewWarehouse(),
			want{
				"",
				http.StatusMethodNotAllowed,
			},
			request{
				"/update/counter/mainLifeQuestion/42",
				http.MethodPut,
				func() http.Header {
					header := http.Header{}
					header.Set("Content-Type", "text/plain")
					return header
				}(),
			},
		},
		{
			"post valid metric with invalid method [DELETE]",
			storage.NewWarehouse(),
			want{
				"",
				http.StatusMethodNotAllowed,
			},
			request{
				"/update/counter/mainLifeQuestion/42",
				http.MethodDelete,
				func() http.Header {
					header := http.Header{}
					header.Set("Content-Type", "text/plain")
					return header
				}(),
			},
		},
		{
			"post valid metric with invalid header",
			storage.NewWarehouse(),
			want{
				"",
				http.StatusBadRequest,
			},
			request{
				"/update/counter/mainLifeQuestion/42",
				http.MethodPost,
				func() http.Header {
					header := http.Header{}
					header.Set("Content-Type", "application/json")
					return header
				}(),
			},
		},
		{
			"post unknown metric",
			storage.NewWarehouse(),
			want{
				"",
				http.StatusBadRequest,
			},
			request{
				"/update/unknown/mainLifeQuestion/42",
				http.MethodPost,
				func() http.Header {
					header := http.Header{}
					header.Set("Content-Type", "text/plain")
					return header
				}(),
			},
		},
		{
			"post metric with empty name",
			storage.NewWarehouse(),
			want{
				"",
				http.StatusNotFound,
			},
			request{
				"/update/gauge//42",
				http.MethodPost,
				func() http.Header {
					header := http.Header{}
					header.Set("Content-Type", "text/plain")
					return header
				}(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.request.method, tt.request.uri, nil)
			r.Header = tt.request.header
			w := httptest.NewRecorder()

			handler := http.StripPrefix("/update/", MetricPostHandler(tt.repository))
			h := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				handler.ServeHTTP(writer, request)
			})

			h(w, r)
			res := w.Result()

			require.Equal(t, tt.want.statusCode, res.StatusCode)
			if res.StatusCode == http.StatusOK {
				require.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
