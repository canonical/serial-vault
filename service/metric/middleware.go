package metric

import (
	"net/http"
	"strconv"
	"time"
)

// CollectAPIStats middleware collects different statistics from the HTTP request
// use this middleware like this:
// router := mux.NewRouter()
// router.Handle("/v1/models", metric.CollectAPIStats("modelList", http.HandlerFunc(model.List))).Methods("GET")
func CollectAPIStats(view string, inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := &recordResponse{ResponseWriter: w}
		start := time.Now()

		inner.ServeHTTP(ww, r)

		// count all server side errors
		if ww.status >= 500 {
			HTTPIncomingErrorsCounterVec.WithLabelValues(r.Method, ww.Status(), view).Inc()
		}
		// count 504 Gateway Timeout
		if ww.status == 504 {
			HTTPIncomingTimeoutsCounterVec.WithLabelValues(r.Method, view).Inc()
		}
		latency := float64(time.Since(start).Milliseconds())
		HTTPIncomingLatencyHistogramVec.WithLabelValues(r.Method, ww.Status(), view).Observe(latency)
		HTTPIncomingRequestCounterVec.WithLabelValues(r.Method, ww.Status(), view).Inc()
	})
}

// recordResponse is a proxy around an http.ResponseWriter
type recordResponse struct {
	http.ResponseWriter
	status       int
	bytesWritten int64
	wroteHeader  bool
}

// WriteHeader writes http status code
func (r *recordResponse) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status = code
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(code)
}

// Write records written bytes
func (r *recordResponse) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytesWritten += int64(n)
	return n, err
}

// Status returns the HTTP status of the request as a string
func (r *recordResponse) Status() string {
	return strconv.Itoa(r.status)
}
