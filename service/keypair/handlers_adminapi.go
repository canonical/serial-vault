// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
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

package keypair

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/service/log"
	"github.com/CanonicalLtd/serial-vault/service/request"
	"github.com/CanonicalLtd/serial-vault/service/response"
)

// SyncRequest is the request to fetch keypairs
type SyncRequest struct {
	Secret string `json:"secret"`
}

// APIList is the API method to fetch the log records from signing
func APIList(w http.ResponseWriter, r *http.Request) {
	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	// Call the API with the user
	listHandler(w, user, true)
}

// APISyncKeypairs fetches the signing-keys accessible by a user
// A encryption secret is provided and the keypairs are decrypted and re-encrypted
// using the supplied keystore secret
func APISyncKeypairs(w http.ResponseWriter, r *http.Request) {
	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		log.Error("error-auth", err)
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	request := SyncRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	switch {
	// Check we have some data
	case err == io.EOF:
		response.FormatStandardResponse(false, "error-keypair-data", "", "No keypair sync data supplied", w)
		return
		// Check for parsing errors
	case err != nil:
		response.FormatStandardResponse(false, "error-keypair-json", "", err.Error(), w)
		return
	}

	syncHandler(w, user, true, request)
}
