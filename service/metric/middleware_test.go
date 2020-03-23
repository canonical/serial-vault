package metric

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/check.v1"
)

var expectedPrometheusData = map[string]string{
	"http_in_errors":   `label:<name:"method" value:"GET" > label:<name:"status" value:"500" > label:<name:"view" value:"testError" > counter:<value:1 >`,
	"http_in_latency":  `label:<name:"method" value:"GET" > label:<name:"status" value:"200" > label:<name:"view" value:"testOK" > histogram:<sample_count:1`,
	"http_in_requests": `label:<name:"method" value:"GET" > label:<name:"status" value:"200" > label:<name:"view" value:"testOK" > counter:<value:1 >`,
	"http_in_timeouts": `label:<name:"method" value:"GET" > label:<name:"view" value:"testTimeout" > counter:<value:1 > `,
}

type MiddlewareSuite struct{}

func TestMiddlewareSuiteSuite(t *testing.T) { check.TestingT(t) }

var _ = check.Suite(&MiddlewareSuite{})

func (s *MiddlewareSuite) TestCollectAPIStats(c *check.C) {
	// restore the default prometheus registerer when the unit test is complete.
	snapshot := prometheus.DefaultRegisterer
	defer func() {
		prometheus.DefaultRegisterer = snapshot
	}()

	// creates a blank registry
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry

	InitMetrics()

	w := httptest.NewRecorder()
	router := mux.NewRouter()

	router.Handle("/ok", CollectAPIStats("testOK", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))).Methods("GET")

	router.Handle("/error", CollectAPIStats("testError", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))).Methods("GET")

	router.Handle("/timeout", CollectAPIStats("testTimeout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(504)
	}))).Methods("GET")

	r := httptest.NewRequest("GET", "/ok", nil)
	router.ServeHTTP(w, r)
	r = httptest.NewRequest("GET", "/error", nil)
	router.ServeHTTP(w, r)
	r = httptest.NewRequest("GET", "/timeout", nil)
	router.ServeHTTP(w, r)

	metrics, err := registry.Gather()
	if err != nil {
		c.Error(err)
		return
	}

	if len(metrics) != len(expectedPrometheusData) {
		c.Fatalf("expected %d metrics, got %d", len(expectedPrometheusData), len(metrics))
	}

	for _, metric := range metrics {
		expectedPrefix, ok := expectedPrometheusData[metric.GetName()]
		if !ok {
			c.Fatalf("metric %s not found", metric.GetName())
		}

		if !strings.HasPrefix(metric.Metric[0].String(), expectedPrefix) {
			c.Fatalf("\ngot metric: %s\n  expected: %s\n", metric.Metric[0].String(), expectedPrefix)
		}
	}
}
