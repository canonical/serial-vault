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
	"io"
	"net/http"
	"strings"

	"github.com/CanonicalLtd/serial-vault/datastore"
)

// KeypairWithPrivateKey is the JSON version of a keypair, including the base64 armored, signing-key
type KeypairWithPrivateKey struct {
	ID          int    `json:"id"`
	AuthorityID string `json:"authority-id"`
	PrivateKey  string `json:"private-key"`
	KeyName     string `json:"key-name"`
}

// KeypairStatusResponse is the JSON response from the API status of keypair generation
type KeypairStatusResponse struct {
	Success      bool                      `json:"success"`
	ErrorCode    string                    `json:"error_code"`
	ErrorSubcode string                    `json:"error_subcode"`
	ErrorMessage string                    `json:"message"`
	Status       []datastore.KeypairStatus `json:"status"`
}

// KeypairGenerateHandler is the API method to generate a new keypair that can be used
// for signing serial (or model) assertions. The keypairs are stored in the signing database
// and the authority-id/key-id is stored in the models database. Models can then be
// linked to one of the existing signing-keys.
func KeypairGenerateHandler(w http.ResponseWriter, r *http.Request) {

	keypair, ok := verifyKeypair(w, r)
	if !ok {
		return
	}

	go datastore.GenerateKeypair(keypair.AuthorityID, "", keypair.KeyName)

	// Return the URL to watch for the response
	statusURL := fmt.Sprintf("/v1/keypairs/status/%s/%s", keypair.AuthorityID, keypair.KeyName)
	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Location", statusURL)
	formatBooleanResponse(true, "", "", statusURL, w)
}

func verifyKeypair(w http.ResponseWriter, r *http.Request) (KeypairWithPrivateKey, bool) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	keypairWithKey := KeypairWithPrivateKey{}

	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
	if err != nil {
		formatBooleanResponse(false, "error-auth", "", "", w)
		return keypairWithKey, false
	}

	// Check that we have a message body
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-nil-data", "", "Uninitialized POST data", w)
		return keypairWithKey, false
	}
	defer r.Body.Close()

	// Decode the JSON body
	err = json.NewDecoder(r.Body).Decode(&keypairWithKey)
	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-keypair-data", "", "No keypair data supplied", w)
		return keypairWithKey, false
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-keypair-json", "", err.Error(), w)
		return keypairWithKey, false
	}

	// Validate the authority-id
	keypairWithKey.AuthorityID = strings.TrimSpace(keypairWithKey.AuthorityID)
	if len(keypairWithKey.AuthorityID) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-keypair-json", "", "The authority-id is mandatory", w)
		return keypairWithKey, false
	}

	// Check that the user has permissions to this authority-id
	if !datastore.Environ.DB.CheckUserInAccount(authUser.Username, keypairWithKey.AuthorityID) {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-auth", "", "Your user does not have permissions for the Signing Authority", w)
		return keypairWithKey, false
	}

	return keypairWithKey, true

}
