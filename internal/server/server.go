// Package server provides an HTTP server for handling metrics.
package server

import (
	"fmt"
	"net/http"

	"github.com/gdyunin/metricol.git/internal/config/server"
	"github.com/gdyunin/metricol.git/internal/server/handlers"
	"github.com/gdyunin/metricol.git/internal/server/logger"
	"github.com/gdyunin/metricol.git/internal/server/middlewares"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

// Server represents the HTTP server for the metrics application.
type Server struct {
	store         *storage.Store // The storage backend for metrics.
	router        *chi.Mux       // The router for handling HTTP requests.
	serverAddress string         // The address on which the server listens.
}

// NewServer creates a new Server instance with the given configuration and options.
func NewServer(cfg *server.Config, options ...func(*Server)) *Server {
	s := &Server{
		store:         storage.NewStore(),
		router:        chi.NewRouter(),
		serverAddress: cfg.ServerAddress,
	}

	for _, o := range options {
		o(s)
	}

	return s
}

// DefaultServer initializes a Server with default routes and middlewares based on the provided configuration.
func DefaultServer(cfg *server.Config) *Server {
	return NewServer(cfg, withDefaultMiddlewares, withDefaultRoutes)
}

// Start begins listening for HTTP requests on the server's address.
func (s *Server) Start() error {
	return fmt.Errorf("error server run %w", http.ListenAndServe(s.serverAddress, s.router))
}

// withDefaultRoutes configures the default middlewares for the server's router.
func withDefaultMiddlewares(s *Server) {
	_ = logger.InitializeSugarLogger("INFO")
	s.router.Use(middlewares.WithLogging)
}

// withDefaultRoutes configures the default routes for the server's router.
func withDefaultRoutes(s *Server) {
	// Setup GET methods for retrieving metrics.
	s.router.Get("/", handlers.MainPageHandler(s.store))
	s.router.Post("/value/", handlers.MetricGetFromBodyHandler(s.store))
	s.router.Get("/value/{metricType}/{metricName}", handlers.MetricGetFromURIHandler(s.store))

	// Setup POST methods for updating metrics.
	s.router.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.MetricPostFromBodyHandler(s.store)) // Handle case where metric type is not passed.
		r.Route("/{metricType}", func(r chi.Router) {
			r.Post("/", handlers.NotFound) // Handle case where metric name is not passed.
			r.Route("/{metricName}", func(r chi.Router) {
				r.Post("/", handlers.BadRequest)                                     // Handle case where metric value is not passed.
				r.Post("/{metricValue}", handlers.MetricPostFromURIHandler(s.store)) // Handle metric post query.
			})
		})
	})
}
