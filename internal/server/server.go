// Package server provides an HTTP server for handling metrics.
package server

import (
	"fmt"
	"net/http"

	"github.com/gdyunin/metricol.git/internal/config/server"
	"github.com/gdyunin/metricol.git/internal/server/handlers"
	"github.com/gdyunin/metricol.git/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

// Server represents the HTTP server for the metrics application.
type Server struct {
	store         *storage.Store // The storage backend for metrics.
	router        *chi.Mux       // The router for handling HTTP requests.
	serverAddress string         // The address on which the server listens.
}

// NewServer creates a new Server instance with the given configuration.
func NewServer(cfg *server.Config, options ...func(*Server)) *Server {
	s := &Server{
		store:         storage.NewStore(), // Creates a new storage instance.
		router:        chi.NewRouter(),    // Creates a new router instance.
		serverAddress: cfg.ServerAddress,  // Sets the server address from the config.
	}

	for _, o := range options {
		o(s) // Apply each option to the server instance.
	}

	return s
}

// DefaultServer initializes a Server with default routes based on the provided configuration.
func DefaultServer(cfg *server.Config) *Server {
	return NewServer(cfg, withDefaultRoutes())
}

// Start begins listening for HTTP requests on the server's address.
func (s *Server) Start() error {
	return fmt.Errorf("error server run %w", http.ListenAndServe(s.serverAddress, s.router))
}

// withDefaultRoutes sets up default routes for the server.
func withDefaultRoutes() func(*Server) {
	return func(s *Server) {
		// Setup GET methods for retrieving metrics.
		setupDefaultGetRoutes(s)
		// Setup POST methods for updating metrics.
		setupDefaultPostRoutes(s)
	}
}

// setupDefaultGetRoutes configures the GET routes for the server.
func setupDefaultGetRoutes(s *Server) {
	s.router.Get("/", handlers.MainPageHandler(s.store))                                 // Main page handler.
	s.router.Get("/value/{metricType}/{metricName}", handlers.MetricGetHandler(s.store)) // Handler to get metric values.
}

// setupDefaultPostRoutes configures the POST routes for updating metrics.
func setupDefaultPostRoutes(s *Server) {
	s.router.Route("/update/", func(r chi.Router) {
		r.Post("/", handlers.BadRequest) // Handle case where metric type is not passed.
		r.Route("/{metricType}", func(r chi.Router) {
			r.Post("/", handlers.NotFound) // Handle case where metric name is not passed.
			r.Route("/{metricName}", func(r chi.Router) {
				r.Post("/", handlers.BadRequest)                              // Handle case where metric value is not passed.
				r.Post("/{metricValue}", handlers.MetricPostHandler(s.store)) // Handle metric post query.
			})
		})
	})
}