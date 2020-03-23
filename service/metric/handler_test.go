package metric_test

import (
	"net/http/httptest"
	"testing"

	"gopkg.in/check.v1"

	"github.com/CanonicalLtd/serial-vault/service/metric"
)

type MetricSuite struct{}

func TestMetricSuite(t *testing.T) { check.TestingT(t) }

var _ = check.Suite(&MetricSuite{})

func (s *MetricSuite) TestMetricHandler(c *check.C) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	metric.NewServer().ServeHTTP(w, r)
	c.Assert(w.Code, check.Equals, 200)
	c.Assert(w.Header().Get("Content-Type"), check.Matches, "text/plain;.*")
}
