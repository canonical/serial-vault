package metric

import (
	"net/http"
	"strconv"
)

func CollectAPIStats(api string, inner http.HandlerFunc) http.HandlerFunc {
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
