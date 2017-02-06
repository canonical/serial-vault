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
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// SigningLogResponse is the JSON response from the API Signing Log method
type SigningLogResponse struct {
	Success      bool         `json:"success"`
	ErrorCode    string       `json:"error_code"`
	ErrorSubcode string       `json:"error_subcode"`
	ErrorMessage string       `json:"message"`
	SigningLog   []SigningLog `json:"logs"`
}

type SigningLogRequest struct {
	FromID int      `json:"fromID"`
	Makes  []string `json:"makes"`
	Models []string `json:"models"`
}

// SigningLogHandler is the API method to fetch the log records from signing
func SigningLogHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Check that we have a message body
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		formatSigningLogResponse(false, "error-fetch-signinglog", "", "Uninitialized POST data", nil, w)
		return
	}
	defer r.Body.Close()

	// Decode the JSON body
	request := SigningLogRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatSigningLogResponse(false, "error-fetch-signinglog", "", err.Error(), nil, w)
		return
	}
	log.Println(request.FromID)
	log.Println(request.Makes)
	log.Println(request.Models)
	if request.FromID == 0 {
		request.FromID = MaxFromID
	}

	logs, err := Environ.DB.ListSigningLog(request)
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

	filters, err := Environ.DB.SigningLogFilterValues()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		formatSigningLogResponse(false, "error-fetch-signinglogfilters", "", err.Error(), nil, w)
		return
	}

	// Encode the response as JSON
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(filters); err != nil {
		log.Println("Error forming the signing log filters response.")
	}
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
	signingLog := SigningLog{ID: logID}
	errorSubcode, err := Environ.DB.DeleteSigningLog(signingLog)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatSigningLogResponse(false, "error-deleting-signinglog", errorSubcode, errorMessage, nil, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	formatSigningLogResponse(true, "", "", "", nil, w)
}
