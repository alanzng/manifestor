// Package server provides an HTTP proxy server that filters and builds
// HLS and DASH manifests on the fly.
package server

import (
	"net/http"
	"time"
)

// Config holds the configuration for the HTTP server.
type Config struct {
	// Addr is the TCP address to listen on (e.g. ":8080").
	Addr string
	// FetchTimeout is the timeout for fetching upstream manifests.
	// Defaults to 10 seconds.
	FetchTimeout time.Duration
}

// Server is the HTTP proxy server.
type Server struct {
	cfg    Config
	mux    *http.ServeMux
	client *http.Client
}

// New creates a new Server with the given configuration.
func New(cfg Config) *Server {
	if cfg.FetchTimeout == 0 {
		cfg.FetchTimeout = 10 * time.Second
	}
	s := &Server{
		cfg:    cfg,
		mux:    http.NewServeMux(),
		client: &http.Client{Timeout: cfg.FetchTimeout},
	}
	s.routes()
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server. It blocks until the server stops.
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.cfg.Addr, s)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/filter", s.handleFilter)
	s.mux.HandleFunc("/build", s.handleBuild)
}

// handleFilter handles GET /filter requests.
func (s *Server) handleFilter(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// handleBuild handles POST /build requests.
func (s *Server) handleBuild(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
