package server

import (
	"net/http"
	"testing"

	"github.com/gdyunin/metricol.git/internal/config/server"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestDefaultServer(t *testing.T) {
	tests := []struct {
		name string
		cfg  *server.Config
		want *Server
	}{
		{
			name: "Create default server",
			cfg:  &server.Config{ServerAddress: ":8080"},
			want: &Server{
				serverAddress: ":8080",
				store:         storage.NewStore(),
				router:        chi.NewRouter(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := DefaultServer(tt.cfg)
			require.NotNil(t, s)
			require.NotNil(t, s.store)
			require.NotNil(t, s.router)
			require.Equal(t, tt.want.serverAddress, s.serverAddress)

			// Check if the default routes are set up correctly.
			require.True(t, s.router.Match(chi.NewRouteContext(), http.MethodGet, "/"))
			require.True(t, s.router.Match(chi.NewRouteContext(), http.MethodGet, "/value/{metricType}/{metricName}"))
			require.True(t, s.router.Match(chi.NewRouteContext(), http.MethodPost, "/update/"))
			require.True(t, s.router.Match(chi.NewRouteContext(), http.MethodPost, "/update/{metricType}"))
			require.True(t, s.router.Match(chi.NewRouteContext(), http.MethodPost, "/update/{metricType}/{metricName}"))
			require.True(t, s.router.Match(
				chi.NewRouteContext(),
				http.MethodPost,
				"/update/{metricType}/{metricName}/{metricValue}",
			))
		})
	}
}

func TestNewServer(t *testing.T) {
	tests := []struct {
		name string
		cfg  *server.Config
		want *Server
	}{
		{
			name: "Create server from config",
			cfg:  &server.Config{ServerAddress: ":8080"},
			want: &Server{
				serverAddress: ":8080",
				store:         storage.NewStore(),
				router:        chi.NewRouter(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer(tt.cfg)
			require.NotNil(t, s)
			require.Equal(t, tt.want.serverAddress, s.serverAddress)
			require.NotNil(t, s.store)
			require.NotNil(t, s.router)
		})
	}
}

func TestServer_Start(t *testing.T) {
	tests := []struct {
		name          string
		store         *storage.Store
		router        *chi.Mux
		serverAddress string
		wantErr       bool
	}{
		{
			name:          "Start server successfully",
			store:         storage.NewStore(),
			router:        chi.NewRouter(),
			serverAddress: ":8080",
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				store:         tt.store,
				router:        tt.router,
				serverAddress: tt.serverAddress,
			}

			var err error
			go func() { err = s.Start() }()
			require.True(t, (err != nil) == tt.wantErr)
		})
	}
}

func Test_setDefaultRoutes(t *testing.T) {
	tests := []struct {
		name   string
		router chi.Router
		store  storage.Repository
	}{
		{
			name:   "Set default routes",
			router: chi.NewRouter(),
			store:  storage.NewStore(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setDefaultRoutes(tt.router, tt.store)

			// Check if the default routes are set up correctly.
			require.True(t, tt.router.Match(chi.NewRouteContext(), http.MethodGet, "/"))
			require.True(t, tt.router.Match(chi.NewRouteContext(), http.MethodGet, "/value/{metricType}/{metricName}"))
			require.True(t, tt.router.Match(chi.NewRouteContext(), http.MethodPost, "/update/"))
			require.True(t, tt.router.Match(chi.NewRouteContext(), http.MethodPost, "/update/{metricType}"))
			require.True(t, tt.router.Match(chi.NewRouteContext(), http.MethodPost, "/update/{metricType}/{metricName}"))
			require.True(t, tt.router.Match(
				chi.NewRouteContext(),
				http.MethodPost,
				"/update/{metricType}/{metricName}/{metricValue}",
			))
		})
	}
}
