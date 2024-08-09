package metric

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

var expectedPrometheusData = map[string]string{
	"http_in_errors":   `label:{name:"method" value:"GET"} label:{name:"status" value:"500"} label:{name:"view" value:"testError"} counter:{value:1`,
	"http_in_latency":  `label:{name:"method" value:"GET"} label:{name:"status" value:"200"} label:{name:"view" value:"testOK"} histogram:{sample_count:1`,
	"http_in_requests": `label:{name:"method" value:"GET"} label:{name:"status" value:"200"} label:{name:"view" value:"testOK"} counter:{value:1`,
	"http_in_timeouts": `label:{name:"method" value:"GET"} label:{name:"view" value:"testTimeout"} counter:{value:1`,
}

func TestCollectAPIStats(t *testing.T) {
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
		t.Error(err)
		return
	}

	if len(metrics) != len(expectedPrometheusData) {
		t.Fatalf("expected %d metrics, got %d", len(expectedPrometheusData), len(metrics))
	}

	for _, metric := range metrics {
		expectedPrefix, ok := expectedPrometheusData[metric.GetName()]
		if !ok {
			t.Fatalf("metric %s not found", metric.GetName())
		}

		// convert any amount of spaces to 1
		actual_metric := strings.Join(strings.Fields(metric.Metric[0].String()), " ")
		if !strings.HasPrefix(actual_metric, expectedPrefix) {
			t.Fatalf("\ngot metric: %s\n  expected: %s\n", actual_metric, expectedPrefix)
		}
	}
}
