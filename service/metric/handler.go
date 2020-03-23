package metric

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ServiceMode represents if the service is running in SigningMode or AdminMode
// this information will be added to the metric
type ServiceMode int

const (
	// SigningMode represents that the service is running in Signing/API Mode
	SigningMode ServiceMode = iota
	// AdminMode represents that the service is running in Admin/UI Mode
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

// ServeHTTP returns a metrics server.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.metrics.ServeHTTP(w, r)
}
