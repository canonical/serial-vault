package sentry

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/getsentry/sentry-go"
)

const placeholder = "[Filtered]"

var headerFilter = []string{"Api-Key", "Authorization", "User", "Cookie"}

// Init initializes sentry client
func Init(dsn, serviceName, serviceVersion string) error {
	sentrySyncTransport := sentry.NewHTTPSyncTransport()
	sentrySyncTransport.Timeout = time.Second * 3

	opt := sentry.ClientOptions{
		Dsn:              dsn,
		Release:          serviceVersion,
		Transport:        sentrySyncTransport,
		AttachStacktrace: true,
		BeforeSend:       getFilteredEvent,
	}
	if err := sentry.Init(opt); err != nil {
		return err
	}

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("app_name", serviceName)
	})
	return nil
}

// Report creates sentry report from response.ErrorResponse
func Report(ctx context.Context, resp response.ErrorResponse) {
	if !shouldReport(resp) {
		return
	}

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		go hub.WithScope(func(scope *sentry.Scope) {
			// create a sentry report
			scope.SetTag("http_status_code", fmt.Sprint(resp.StatusCode))
			scope.SetTag("error_code", resp.Code)
			if resp.SubCode != "" {
				scope.SetTag("error_subcode", resp.SubCode)
			}
			scope.SetLevel(sentry.LevelError)
			hub.CaptureMessage(resp.Message)
		})
	}
}

func getFilteredEvent(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	for _, headerName := range headerFilter {
		if _, ok := event.Request.Headers[headerName]; ok {
			event.Request.Headers[headerName] = placeholder
		}
	}

	if event.Request.QueryString != "" {
		event.Request.QueryString = GetFilteredQueryString(event.Request.QueryString)
	}

	if event.Request.Data != "" {
		event.Request.Data = placeholder
	}

	if event.Request.Cookies != "" {
		event.Request.Cookies = GetFilteredCookies(event.Request.Cookies)
	}

	return event
}

func shouldReport(resp response.ErrorResponse) bool {
	return resp.StatusCode >= 500
}

// GetFilteredQueryString replaces all the values in the query string
func GetFilteredQueryString(q string) string {
	values, err := url.ParseQuery(q)
	if err != nil {
		return placeholder
	}

	for name := range values {
		values.Set(name, placeholder)
	}

	return values.Encode()
}

// GetFilteredCookies replaces all the values in the cookie string
func GetFilteredCookies(rawCookies string) string {
	header := http.Header{}
	header.Add("Cookie", rawCookies)
	request := &http.Request{Header: header}

	var newCookies string
	for _, c := range request.Cookies() {
		s := fmt.Sprintf("%s=%s", c.Name, placeholder)
		if newCookies != "" {
			newCookies = newCookies + "; " + s
		} else {
			newCookies = s
		}

	}
	return newCookies
}
