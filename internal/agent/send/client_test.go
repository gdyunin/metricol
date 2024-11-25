package send

import (
	"github.com/gdyunin/metricol.git/internal/agent/fetch"
	"github.com/gdyunin/metricol.git/internal/agent/metrics/library"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestClient_Send(t *testing.T) {
	type args struct {
		s       *fetch.Storage
		address string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			"simple send",
			args{
				s: func() *fetch.Storage {
					s := fetch.NewStorage()
					s.AddMetrics(library.NewCounter("PollCount", func() int64 { return 1 }))
					return s
				}(),
				address: "localhost:8080",
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(tt.args.s, tt.args.address)
			if err := c.Send(); err != nil {
				require.ErrorAs(t, err, &tt.wantErr)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	type args struct {
		s       *fetch.Storage
		address string
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		{
			"generate new client",
			args{
				s:       fetch.NewStorage(),
				address: "localhost:8080",
			},
			&Client{
				storage:  fetch.NewStorage(),
				host:     "localhost:8080",
				basePath: newBasePath(),
				client:   http.Client{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewClient(tt.args.s, tt.args.address)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_newBasePath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			"generation base path",
			"/update/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newBasePath()
			require.Equal(t, tt.want, got)
		})
	}
}
