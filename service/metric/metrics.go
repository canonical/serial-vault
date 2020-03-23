package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

// HTTPIncomingRequestCounterVec is prometheus metric for incoming http requests count
var HTTPIncomingRequestCounterVec = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_in_requests",
		Help: "metric for incoming HTTP requests count",
	},
	[]string{"method", "status", "view"},
)

// HTTPIncomingLatencyHistogramVec is prometheus metric for incoming requests latency
var HTTPIncomingLatencyHistogramVec = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_in_latency",
		Help:    "metric for incoming requests latency",
		Buckets: []float64{4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192},
	},
	[]string{"method", "status", "view"},
)

// HTTPIncomingErrorsCounterVec is prometheus metric for HTTP errors
var HTTPIncomingErrorsCounterVec = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_in_errors",
		Help: "fetric for HTTP errors",
	},
	[]string{"method", "status", "view"},
)

// HTTPIncomingTimeoutsCounterVec is metric for incoming http timeouts
var HTTPIncomingTimeoutsCounterVec = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_in_timeouts",
		Help: "metric for incoming HTTP timeouts",
	},
	[]string{"method", "view"},
)

func init() {
	InitMetrics()
}

// InitMetrics register all the metrics
func InitMetrics() {
	prometheus.MustRegister(HTTPIncomingRequestCounterVec)
	prometheus.MustRegister(HTTPIncomingLatencyHistogramVec)
	prometheus.MustRegister(HTTPIncomingErrorsCounterVec)
	prometheus.MustRegister(HTTPIncomingTimeoutsCounterVec)
}
