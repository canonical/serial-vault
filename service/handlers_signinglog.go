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
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
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

	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
	if err != nil {
		formatSigningLogResponse(false, "error-auth", "", "", nil, w)
		return
	}

	logs, err := listAllowedSigningLog(authUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		formatSigningLogResponse(false, "error-fetch-signinglog", "", err.Error(), nil, w)
		return
	}

	// Return successful JSON response with the list of models
	w.WriteHeader(http.StatusOK)
	formatSigningLogResponse(true, "", "", "", logs, w)
}

func listAllowedSigningLog(authUser datastore.User) ([]datastore.SigningLog, error) {
	switch authUser.Role {
	case 0:
		fallthrough
	case datastore.Superuser:
		return listAllSigningLogs()
	case datastore.Admin:
		return listSigningLogsFilteredByUser(authUser.Username)
	}
	return []datastore.SigningLog{}, nil
}

func listAllSigningLogs() ([]datastore.SigningLog, error) {
	return datastore.Environ.DB.ListSigningLog("")
}

func listSigningLogsFilteredByUser(username string) ([]datastore.SigningLog, error) {
	return datastore.Environ.DB.ListSigningLog(username)
}

// SigningLogFiltersHandler is the API method to fetch the log filter values
func SigningLogFiltersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
	if err != nil {
		formatSigningLogResponse(false, "error-auth", "", "", nil, w)
		return
	}

	filters, err := allowedSigningLogFilterValues(authUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		formatSigningLogFiltersResponse(false, "error-fetch-signinglogfilters", "", err.Error(), filters, w)
		return
	}

	// Encode the response as JSON
	w.WriteHeader(http.StatusOK)
	formatSigningLogFiltersResponse(true, "", "", "", filters, w)
}

func allowedSigningLogFilterValues(authUser datastore.User) (datastore.SigningLogFilters, error) {
	switch authUser.Role {
	case 0:
		fallthrough
	case datastore.Superuser:
		return allAllowedSigningLogFilterValues()
	case datastore.Admin:
		return signingLogFilterValuesFilteredByUser(authUser.Username)
	}
	return datastore.SigningLogFilters{}, nil
}

func allAllowedSigningLogFilterValues() (datastore.SigningLogFilters, error) {
	return datastore.Environ.DB.SigningLogFilterValues("")
}

func signingLogFilterValuesFilteredByUser(username string) (datastore.SigningLogFilters, error) {
	return datastore.Environ.DB.SigningLogFilterValues(username)
}
