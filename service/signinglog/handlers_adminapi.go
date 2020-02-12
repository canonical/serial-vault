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

	// Call the API with the user
	listHandler(w, user, true)
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

// GetSigningLogParams parse and set defaults for the search parameters from the request
func GetSigningLogParams(r *http.Request) *datastore.SigningLogParams {
	params := &datastore.SigningLogParams{
		Limit: datastore.ListSigningLogDefaultLimit,
	}
	query := r.URL.Query()

	if offset, err := strconv.ParseUint(query.Get("offset"), 10, 64); err == nil {
		params.Offset = offset
	}

	if fetchAll := query.Get("all"); fetchAll == "true" {
		params.Limit = 0 // Means no limit.
		params.Offset = 0
	}

	if filter := query.Get("filter"); filter != "" {
		params.Filter = strings.Split(filter, ",")
	}

	params.Serialnumber = query.Get("serialnumber")

	return params
}
