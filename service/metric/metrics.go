package metric

import "github.com/prometheus/client_golang/prometheus"

var HTTPIncomingRequestCounterVec = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_in_requests",
		Help: "metric for incoming http requests count",
	},
	[]string{"method", "status", "api"},
)

func init() {
	prometheus.MustRegister(HTTPIncomingRequestCounterVec)
}
