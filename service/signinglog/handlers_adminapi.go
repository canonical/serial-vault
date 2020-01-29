// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
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

package signinglog

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/request"
	"github.com/CanonicalLtd/serial-vault/service/response"
)

// APIList is the API method to fetch the log records from signing
func APIList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	params := GetSigningLogParams(r)

	// Call the API with the user
	listHandler(w, user, true, params)
}

// APISyncLog is the API method to sync a factory log to the cloud
func APISyncLog(w http.ResponseWriter, r *http.Request) {
	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	request := datastore.SigningLog{}
	err = json.NewDecoder(r.Body).Decode(&request)
	switch {
	// Check we have some data
	case err == io.EOF:
		response.FormatStandardResponse(false, "error-signinglog-data", "", "No signing-log data supplied", w)
		return
		// Check for parsing errors
	case err != nil:
		response.FormatStandardResponse(false, "error-signinglog-json", "", err.Error(), w)
		return
	}

	// Call the API with the user
	syncLogHandler(w, user, true, request)
}

// GetSigningLogParams ...
func GetSigningLogParams(r *http.Request) *datastore.SigningLogParams {
	params := &datastore.SigningLogParams{}

	if offsetParam, ok := r.URL.Query()["offset"]; ok {
		offset, err := strconv.Atoi(offsetParam[0])
		if err == nil {
			params.Offset = offset
		}
	}

	if serialnumber, ok := r.URL.Query()["serialnumber"]; ok {
		params.Serialnumber = serialnumber[0]
	}

	filter, ok := r.URL.Query()["filter"]
	if ok && len(filter) > 0 && filter[0] != "" {
		params.Filter = strings.Split(filter[0], ",")
	}

	return params
}
