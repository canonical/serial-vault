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

package core

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/auth"
	"github.com/CanonicalLtd/serial-vault/service/log"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/gorilla/csrf"
)

// VersionResponse is the JSON response from the API Version method
type VersionResponse struct {
	Version string `json:"version"`
}

// HealthResponse is the JSON response from the health check method
type HealthResponse struct {
	Database string `json:"database"`
}

// TokenResponse is the JSON response from the API Version method
type TokenResponse struct {
	EnableUserAuth bool `json:"enableUserAuth"`
}

// Version is the API method to return the version of the service
func Version(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", response.JSONHeader)
	w.WriteHeader(http.StatusOK)

	response := VersionResponse{Version: datastore.Environ.Config.Version}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		message := fmt.Sprintf("Error encoding the version response: %v", err)
		log.Message("VERSION", "get-version", message)
	}
}

// Health is the API method to return if the app is up and db.Ping() doesn't return an error
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", response.JSONHeader)
	err := datastore.Environ.DB.HealthCheck()
	var database string

	if err != nil {
		database = err.Error()
		w.WriteHeader(http.StatusBadRequest)
	} else {
		database = "healthy"
	}
	response := HealthResponse{Database: database}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		message := fmt.Sprintf("Error ecoding the health response: %v", err)
		log.Message("HEALTH", "health", message)
	}
}

// Token returns CSRF protection new token in a X-CSRF-Token response header
// This method is also used by the /authtoken endpoint to return the JWT. The method
// indicates to the UI whether OpenID user auth is enabled
func Token(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", response.JSONHeader)
	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	// Check the JWT and return it in the authorization header, if valid
	auth.JWTCheck(w, r)

	response := TokenResponse{EnableUserAuth: datastore.Environ.Config.EnableUserAuth}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		message := fmt.Sprintf("Error encoding the token response: %v", err)
		log.Message("TOKEN", "get-token", message)
	}
}
