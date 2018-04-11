// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
 * License granted by Canonical Limited
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

package testlog

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/request"
	"github.com/CanonicalLtd/serial-vault/service/response"
)

// APISyncLog is the API method to sync a factory test log to the cloud
func APISyncLog(w http.ResponseWriter, r *http.Request) {
	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	request := datastore.TestLog{}
	err = json.NewDecoder(r.Body).Decode(&request)
	switch {
	// Check we have some data
	case err == io.EOF:
		response.FormatStandardResponse(false, "error-testlog-data", "", "No testlog data supplied", w)
		return
		// Check for parsing errors
	case err != nil:
		response.FormatStandardResponse(false, "error-testlog-json", "", err.Error(), w)
		return
	}

	// Call the API with the user
	syncLogHandler(w, user, true, request)
}

// APIListLog is the API method to list the unsync-ed factory test logs
func APIListLog(w http.ResponseWriter, r *http.Request) {
	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	// Call the API with the user
	listHandler(w, user, true)
}
