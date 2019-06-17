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
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/gorilla/csrf"
)

// Logger Handle logging for the web service
func Logger(start time.Time, r *http.Request) {
	log.Printf(
		"%s\t%s\t%s",
		r.Method,
		r.URL.Path,
		time.Since(start),
	)
}

// ErrorHandler is a standard error handler middleware that generates the error response
func ErrorHandler(f func(http.ResponseWriter, *http.Request) response.ErrorResponse) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Call the handler and it will return a custom error
		e := f(w, r)
		if !e.Success {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(e.StatusCode)

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

		// Log the request
		Logger(start, r)

		inner.ServeHTTP(w, r)
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
