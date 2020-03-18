package metric

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ServiceMode int

const (
	SigningMode ServiceMode = iota
	AdminMode
)

// Server is an http Metrics server.
type Server struct {
	mode    ServiceMode
	metrics http.Handler
}

// NewServer returns a new metrics server.
func NewServer(mode ServiceMode) *Server {
	return &Server{
		mode:    mode,
		metrics: promhttp.Handler(),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.metrics.ServeHTTP(w, r)
}
