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
	"fmt"
	"net/http"
	"strconv"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/gorilla/mux"
)

// SigningLogResponse is the JSON response from the API Signing Log method
type SigningLogResponse struct {
	Success      bool                   `json:"success"`
	ErrorCode    string                 `json:"error_code"`
	ErrorSubcode string                 `json:"error_subcode"`
	ErrorMessage string                 `json:"message"`
	SigningLog   []datastore.SigningLog `json:"logs"`
}

// SigningLogFiltersResponse is the JSON response from the API Signing Log Filters method
type SigningLogFiltersResponse struct {
	Success           bool                        `json:"success"`
	ErrorCode         string                      `json:"error_code"`
	ErrorSubcode      string                      `json:"error_subcode"`
	ErrorMessage      string                      `json:"message"`
	SigningLogFilters datastore.SigningLogFilters `json:"filters"`
}

// SigningLogHandler is the API method to fetch the log records from signing
func SigningLogHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get the user from the JWT
	username, err := checkUserPermissions(w, r)
	if err != nil {
		formatSigningLogResponse(false, "error-auth", "", "", nil, w)
		return
	}

	logs, err := datastore.Environ.DB.ListSigningLog(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		formatSigningLogResponse(false, "error-fetch-signinglog", "", err.Error(), nil, w)
		return
	}

	// Return successful JSON response with the list of models
	w.WriteHeader(http.StatusOK)
	formatSigningLogResponse(true, "", "", "", logs, w)
}

// SigningLogFiltersHandler is the API method to fetch the log filter values
func SigningLogFiltersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get the user from the JWT
	username, err := checkUserPermissions(w, r)
	if err != nil {
		formatSigningLogResponse(false, "error-auth", "", "", nil, w)
		return
	}

	filters, err := datastore.Environ.DB.SigningLogFilterValues(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		formatSigningLogFiltersResponse(false, "error-fetch-signinglogfilters", "", err.Error(), filters, w)
		return
	}

	// Encode the response as JSON
	w.WriteHeader(http.StatusOK)
	formatSigningLogFiltersResponse(true, "", "", "", filters, w)
}

// SigningLogDeleteHandler is the API method to delete a signing log entry
func SigningLogDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get the signinglog primary key
	vars := mux.Vars(r)
	logID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("%v", vars["id"])
		formatSigningLogResponse(false, "error-invalid-signinglog", "", errorMessage, nil, w)
		return
	}

	// Update the database
	signingLog := datastore.SigningLog{ID: logID}
	errorSubcode, err := datastore.Environ.DB.DeleteSigningLog(signingLog)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatSigningLogResponse(false, "error-deleting-signinglog", errorSubcode, errorMessage, nil, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	formatSigningLogResponse(true, "", "", "", nil, w)
}
