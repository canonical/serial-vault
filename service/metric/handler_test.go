package metric_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CanonicalLtd/serial-vault/service/metric"
)

func TestMetricHandler(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	metric.NewServer().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("expected code 200, got %d", w.Code)
	}
	if !strings.HasPrefix(w.Header().Get("Content-Type"), "text/plain") {
		t.Errorf("expected Content-Type: 'text/plain', got %s", w.Header().Get("Content-Type"))
	}
}
