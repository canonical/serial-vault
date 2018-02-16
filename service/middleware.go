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
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/usso"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
)

// Logger Handle logging for the web service
func Logger(start time.Time, r *http.Request) {
	log.Printf(
		"%s\t%s\t%s",
		r.Method,
		r.RequestURI,
		time.Since(start),
	)
}

// ErrorHandler is a standard error handler middleware that generates the error response
func ErrorHandler(f func(http.ResponseWriter, *http.Request) ErrorResponse) http.HandlerFunc {
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
func MiddlewareWithCSRF(inner http.Handler) http.Handler {
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

	return CSRF(inner)
}

// CORSMiddleware handles the header options for cross-origin requests (used in development only)
func CORSMiddleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		headers := handlers.AllowedHeaders([]string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "X-Requested-With", "Origin"})
		origins := handlers.AllowedOrigins([]string{"http://localhost:3000"})
		methods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
		exposed := handlers.ExposedHeaders([]string{"X-CSRF-Token"})
		credentials := handlers.AllowCredentials()

		return handlers.CORS(headers, origins, methods, exposed, credentials)(h)
	}
}

// JWTCheck extracts the JWT from the request, validates it and returns the token
func JWTCheck(w http.ResponseWriter, r *http.Request) (*jwt.Token, error) {

	// Do not validate access if user authentication is off (default)
	if !datastore.Environ.Config.EnableUserAuth {
		return nil, nil
	}

	// Get the JWT from the header or cookie
	jwtToken, err := usso.JWTExtractor(r)
	if err != nil {
		log.Println("Error in JWT extraction:", err.Error())
		return nil, errors.New("Error in retrieving the authentication token")
	}

	// Verify the JWT string
	token, err := usso.VerifyJWT(jwtToken)
	if err != nil {
		log.Printf("JWT fails verification: %v", err.Error())
		return nil, errors.New("The authentication token is invalid")
	}

	if !token.Valid {
		log.Println("Invalid JWT")
		return nil, errors.New("The authentication token is invalid")
	}

	// Set up the bearer token in the header
	w.Header().Set("Authorization", "Bearer "+jwtToken)

	return token, nil
}
