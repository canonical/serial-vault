package sentry_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/response"
	report "github.com/CanonicalLtd/serial-vault/service/sentry"
	"github.com/getsentry/sentry-go"
	"github.com/gorilla/mux"
	"gopkg.in/check.v1"
)

const (
	appName          = "testApp"
	version          = "1.2.3"
	testErrorMessage = "test error message"
	testErrorCode    = "test-error-code"
	testErrorSubCode = "test-error-sub-code"
)

type SentrySuite struct{}

var _ = check.Suite(&SentrySuite{})

func TestSentrySuite(t *testing.T) { check.TestingT(t) }

func (s *SentrySuite) TestSendSentryReport400(c *check.C) {
	result := &sentry.Event{}

	// setup fake sentry service
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := json.NewDecoder(req.Body).Decode(&result)
		c.Assert(err, check.IsNil)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	dns := fmt.Sprintf("http://123@%s/123", ts.Listener.Addr())
	err := report.Init(dns, appName, version)
	c.Assert(err, check.IsNil)

	router := mux.NewRouter()
	router.Handle("/error",
		service.Middleware(service.ErrorHandler(testError400))).
		Methods("POST")

	w := httptest.NewRecorder()

	body := bytes.NewBufferString("secret assertion")
	r, err := http.NewRequest("POST", "/error", body)
	c.Assert(err, check.IsNil)

	router.ServeHTTP(w, r)
	c.Assert(w.Code, check.Equals, 400)

	// check sentry report is empty
	c.Assert(result.Message, check.Equals, "")
}

func (s *SentrySuite) TestSendSentryReport501(c *check.C) {
	result := &sentry.Event{}
	var wg sync.WaitGroup
	wg.Add(1)

	// setup fake sentry service
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer wg.Done()
		err := json.NewDecoder(req.Body).Decode(&result)
		c.Assert(err, check.IsNil)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	dns := fmt.Sprintf("http://123@%s/123", ts.Listener.Addr())
	err := report.Init(dns, appName, version)
	c.Assert(err, check.IsNil)

	router := mux.NewRouter()
	router.Handle("/error",
		service.Middleware(service.ErrorHandler(testError501))).
		Methods("POST")

	w := httptest.NewRecorder()

	body := bytes.NewBufferString("secret assertion")
	r, err := http.NewRequest("POST", "/error", body)
	c.Assert(err, check.IsNil)

	r.Header.Add("Foo", "Bar")
	r.Header.Add("api-key", "secret")
	r.Header.Add("Authorization", "Bearer secret")
	r.Header.Add("user", "sync")
	r.Header.Add("cookie", "abc")

	router.ServeHTTP(w, r)
	c.Assert(w.Code, check.Equals, 501)

	// wait until fake sentry service responded
	wg.Wait()

	// check sentry report
	c.Assert(result.Level, check.Equals, sentry.LevelError)
	c.Assert(result.Tags["http_status_code"], check.Equals, "501")

	c.Assert(result.Message, check.Equals, testErrorMessage)
	c.Assert(result.Tags["app_name"], check.Equals, appName)
	c.Assert(result.Tags["error_code"], check.Equals, testErrorCode)
	c.Assert(result.Tags["error_subcode"], check.Equals, testErrorSubCode)

	// check headers
	c.Assert(result.Request.Headers["Foo"], check.Equals, "Bar")
	c.Assert(result.Request.Headers["Api-Key"], check.Equals, "[Filtered]")
	c.Assert(result.Request.Headers["Authorization"], check.Equals, "[Filtered]")
	c.Assert(result.Request.Headers["User"], check.Equals, "[Filtered]")
	c.Assert(result.Request.Headers["Cookie"], check.Equals, "[Filtered]")
}

func (s *SentrySuite) TestSendSentryReportPanic(c *check.C) {
	result := &sentry.Event{}
	var wg sync.WaitGroup
	wg.Add(1)

	// setup fake sentry service
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer wg.Done()
		err := json.NewDecoder(req.Body).Decode(&result)
		c.Assert(err, check.IsNil)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	dns := fmt.Sprintf("http://123@%s/123", ts.Listener.Addr())
	err := report.Init(dns, appName, version)
	c.Assert(err, check.IsNil)

	router := mux.NewRouter()
	router.Handle("/panic",
		service.Middleware(service.ErrorHandler(testPanic))).
		Methods("POST")

	w := httptest.NewRecorder()

	body := bytes.NewBufferString("secret assertion")
	r, err := http.NewRequest("POST", "/panic", body)
	c.Assert(err, check.IsNil)

	r.Header.Add("Foo", "Bar")
	r.Header.Add("api-key", "secret")
	r.Header.Add("Authorization", "Bearer secret")
	r.Header.Add("user", "sync")
	r.Header.Add("cookie", "abc")

	router.ServeHTTP(w, r)
	c.Assert(w.Code, check.Equals, 500)

	// wait until fake sentry service responded
	wg.Wait()

	// check sentry report
	c.Assert(result.Level, check.Equals, sentry.LevelFatal)
	c.Assert(result.Message, check.Equals, testErrorMessage)
	c.Assert(result.Tags["app_name"], check.Equals, appName)

	// check headers
	c.Assert(result.Request.Headers["Foo"], check.Equals, "Bar")
	c.Assert(result.Request.Headers["Api-Key"], check.Equals, "[Filtered]")
	c.Assert(result.Request.Headers["Authorization"], check.Equals, "[Filtered]")
	c.Assert(result.Request.Headers["User"], check.Equals, "[Filtered]")
	c.Assert(result.Request.Headers["Cookie"], check.Equals, "[Filtered]")
}

func (s *SentrySuite) Test_GetFilteredQueryString(c *check.C) {
	filteredQueryString := report.GetFilteredQueryString(`foo=bar&lorem=ipsum&aaa=bbb&foo=abc`)
	c.Assert(filteredQueryString, check.Equals, `aaa=%5BFiltered%5D&foo=%5BFiltered%5D&lorem=%5BFiltered%5D`)
}

func (s *SentrySuite) Test_GetFilteredCookies(c *check.C) {
	filteredQueryString := report.GetFilteredCookies(`cookie1=value1;cookie2=value2`)
	c.Assert(filteredQueryString, check.Equals, `cookie1=[Filtered]; cookie2=[Filtered]`)

	filteredQueryString = report.GetFilteredCookies(`cookie1=value1`)
	c.Assert(filteredQueryString, check.Equals, `cookie1=[Filtered]`)
}

func testError400(w http.ResponseWriter, r *http.Request) response.ErrorResponse {
	return response.ErrorResponse{
		Success:    false,
		Code:       testErrorCode,
		SubCode:    testErrorSubCode,
		StatusCode: 400,
		Message:    testErrorMessage,
	}
}

func testError501(w http.ResponseWriter, r *http.Request) response.ErrorResponse {
	return response.ErrorResponse{
		Success:    false,
		Code:       testErrorCode,
		SubCode:    testErrorSubCode,
		StatusCode: 501,
		Message:    testErrorMessage,
	}
}

func testPanic(w http.ResponseWriter, r *http.Request) response.ErrorResponse {
	panic(testErrorMessage)
}
