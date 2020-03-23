package metric

import (
	"net/http"
	"strconv"

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

func collectAPIStats(api string, inner http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ww := &recordResponse{ResponseWriter: w}
		inner.ServeHTTP(ww, r)
		HTTPIncomingRequestCounterVec.WithLabelValues(r.Method, ww.Code(), api).Inc()
	}
}

type recordResponse struct {
	http.ResponseWriter
	status       int
	bytesWritten int64
	wroteHeader  bool
}

func (r *recordResponse) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status = code
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *recordResponse) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytesWritten += int64(n)
	return n, err
}

func (r *recordResponse) Code() string {
	return strconv.Itoa(r.status)
}
