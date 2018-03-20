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

package keypair

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/auth"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/gorilla/mux"
)

// WithPrivateKey is the JSON version of a keypair, including the base64 armored, signing-key
type WithPrivateKey struct {
	ID          int    `json:"id"`
	AuthorityID string `json:"authority-id"`
	PrivateKey  string `json:"private-key"`
	KeyName     string `json:"key-name"`
}

// AssertionRequest is the JSON version of a account assertion
type AssertionRequest struct {
	ID        int    `json:"id"`
	Assertion string `json:"assertion"`
}

// List fetches the available keypairs for display from the database.
// Only viewable reference data is stored in the database, not the restricted private key.
func List(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	listHandler(w, authUser, false)
}

// Create is the API method to create a keypair
// Create a new keypair that can be used for signing serial assertions. The
// keypairs are stored in the signing database and the authority-id/key-id is
// stored in the models database. Models can then be linked to one of the
// existing signing-keys.
func Create(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	keypairWithKey, ok := verifyKeypair(w, r, authUser)
	if !ok {
		return
	}

	createHandler(w, authUser, false, keypairWithKey)
}

// Generate is the API method to generate a new keypair that can be used
// for signing serial (or model) assertions. The keypairs are stored in the signing database
// and the authority-id/key-id is stored in the models database. Models can then be
// linked to one of the existing signing-keys.
func Generate(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	keypairWithKey, ok := verifyKeypair(w, r, authUser)
	if !ok {
		return
	}

	generateHandler(w, authUser, false, keypairWithKey)
}

// Disable disables an existing keypair, which will mean that any
// linked Models will not be able to be signed. The asserts module does not allow
// a keypair to be deleted, so the keypair will just be disabled in the local database.
func Disable(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	// Get the keypair primary key
	vars := mux.Vars(r)
	keypairID, err := strconv.Atoi(vars["id"])
	if err != nil {
		errorMessage := fmt.Sprintf("%v", vars["id"])
		response.FormatStandardResponse(false, "error-invalid-key", "", errorMessage, w)
		return
	}

	enableDisableHandler(w, authUser, false, false, keypairID)
}

// Enable enables an existing keypair, which will mean that any
// linked Models will be able to be signed. The asserts module does not allow
// a keypair to be deleted, so the keypair will just be enabled in the local database.
func Enable(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	// Get the keypair primary key
	vars := mux.Vars(r)
	keypairID, err := strconv.Atoi(vars["id"])
	if err != nil {
		errorMessage := fmt.Sprintf("%v", vars["id"])
		response.FormatStandardResponse(false, "error-invalid-key", "", errorMessage, w)
		return
	}

	enableDisableHandler(w, authUser, false, true, keypairID)
}

// Assertion updates the account key assertion on a keypair
func Assertion(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	defer r.Body.Close()

	assertionRequest := AssertionRequest{}
	err = json.NewDecoder(r.Body).Decode(&assertionRequest)
	switch {
	// Check we have some data
	case err == io.EOF:
		response.FormatStandardResponse(false, "error-assertion-data", "", "No assertion data supplied", w)
		return
		// Check for parsing errors
	case err != nil:
		response.FormatStandardResponse(false, "error-assertion-json", "", err.Error(), w)
		return
	}

	assertionHandler(w, authUser, false, assertionRequest)
}

// Status returns the creation status of a keypair
func Status(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	vars := mux.Vars(r)

	statusHandler(w, authUser, false, vars["authorityID"], vars["keyName"])
}

// Progress returns the status of keypairs that are being generated
func Progress(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	progressHandler(w, authUser, false)
}

func verifyKeypair(w http.ResponseWriter, r *http.Request, authUser datastore.User) (WithPrivateKey, bool) {

	keypairWithKey := WithPrivateKey{}

	defer r.Body.Close()

	// Decode the JSON body
	err := json.NewDecoder(r.Body).Decode(&keypairWithKey)
	switch {
	// Check we have some data
	case err == io.EOF:
		response.FormatStandardResponse(false, "error-keypair-data", "", "No keypair data supplied", w)
		return keypairWithKey, false
		// Check for parsing errors
	case err != nil:
		response.FormatStandardResponse(false, "error-keypair-json", "", err.Error(), w)
		return keypairWithKey, false
	}

	// Validate the authority-id
	keypairWithKey.AuthorityID = strings.TrimSpace(keypairWithKey.AuthorityID)
	if len(keypairWithKey.AuthorityID) == 0 {
		response.FormatStandardResponse(false, "error-keypair-json", "", "The authority-id is mandatory", w)
		return keypairWithKey, false
	}

	// Check that the user has permissions to this authority-id
	if !datastore.Environ.DB.CheckUserInAccount(authUser.Username, keypairWithKey.AuthorityID) {
		response.FormatStandardResponse(false, "error-auth", "", "Your user does not have permissions for the Signing Authority", w)
		return keypairWithKey, false
	}

	return keypairWithKey, true
}
