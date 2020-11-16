// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package service

import (
	"context"
	"encoding/json"
	"runtime/debug"

	"net/http"
	"os"
	"time"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/log"
	"github.com/CanonicalLtd/serial-vault/service/response"
	servicesentry "github.com/CanonicalLtd/serial-vault/service/sentry"
	"github.com/getsentry/sentry-go"
	"github.com/gorilla/csrf"
)

// Logger Handle logging for the web service
func Logger(start time.Time, r *http.Request) {
	log.Infof("%s %s %s", r.Method, r.URL.Path, time.Since(start))
}

// ErrorHandler is a standard error handler middleware that generates the error response
// and sentry reports
func ErrorHandler(f func(http.ResponseWriter, *http.Request) response.ErrorResponse) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// Call the handler and it will return a custom error
		e := f(w, r)
		if !e.Success {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(e.StatusCode)

			servicesentry.Report(ctx, e)

			// Encode the response as JSON
			if err := json.NewEncoder(w).Encode(e); err != nil {
				log.Printf("Error forming the signing response: %v\n", err)
			}
		}
	}
}

// Middleware to pre-process web service requests
func Middleware(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()
		hub := sentry.CurrentHub().Clone()
		// add the current request to the sentry scope
		// it will be automatically added to the report
		hub.Scope().SetRequest(r)
		ctx = sentry.SetHubOnContext(ctx, hub)
		defer recoverWithSentry(hub, w, r)
		// Log the request
		Logger(start, r)

		inner.ServeHTTP(w, r.WithContext(ctx))
	})
}

// MiddlewareWithCSRF to pre-process web service requests with CSRF protection
var MiddlewareWithCSRF = func(inner http.Handler) http.Handler {
	// configure request forgery protection
	csrfSecure := true
	csrfSecureEnv := os.Getenv("CSRF_SECURE")
	if csrfSecureEnv == "disable" {
		csrfSecure = false
	}

	CSRF := csrf.Protect(
		[]byte(datastore.Environ.Config.CSRFAuthKey),
		csrf.Secure(csrfSecure),
		csrf.HttpOnly(csrfSecure),
	)

	return CSRF(Middleware(inner))
}

func recoverWithSentry(hub *sentry.Hub, w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		log.Errorf("recover from panic: %v", err)
		debug.PrintStack()

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		e := response.ErrorInternal
		w.WriteHeader(e.StatusCode)

		// Encode the response as JSON
		if err := json.NewEncoder(w).Encode(e); err != nil {
			log.Printf("Error forming the error response after recovering from panic: %v\n", err)
		}

		hub.RecoverWithContext(
			context.WithValue(r.Context(), sentry.RequestContextKey, r),
			err,
		)
	}
}
