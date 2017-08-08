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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/gorilla/mux"
	"github.com/snapcore/snapd/asserts"
)

// KeypairWithPrivateKey is the JSON version of a keypair, including the base64 armored, signing-key
type KeypairWithPrivateKey struct {
	ID          int    `json:"id"`
	AuthorityID string `json:"authority-id"`
	PrivateKey  string `json:"private-key"`
}

// KeypairListHandler fetches the available keypairs for display from the database.
// Only viewable reference data is stored in the database, not the restricted private key.
func KeypairListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
	if err != nil {
		formatKeypairsResponse(false, "error-auth", "", "", nil, w)
		return
	}

	keypairs, err := listAllowedKeypairs(authUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		formatKeypairsResponse(false, "error-fetch-keypairs", "", err.Error(), nil, w)
		return
	}

	// Return successful JSON response with the list of models
	w.WriteHeader(http.StatusOK)
	formatKeypairsResponse(true, "", "", "", keypairs, w)
}

func listAllowedKeypairs(authUser datastore.User) ([]datastore.Keypair, error) {
	switch authUser.Role {
	case 0:
		fallthrough
	case datastore.Superuser:
		return listAllKeypairs()
	case datastore.Admin:
		return listKeypairsFilteredByUser(authUser.Username)
	default:
		return []datastore.Keypair{}, nil
	}
}

func listAllKeypairs() ([]datastore.Keypair, error) {
	return datastore.Environ.DB.ListKeypairs("")
}

func listKeypairsFilteredByUser(username string) ([]datastore.Keypair, error) {
	return datastore.Environ.DB.ListKeypairs(username)
}

// KeypairCreateHandler is the API method to create a new keypair that can be used
// for signing serial assertions. The keypairs are stored in the signing database
// and the authority-id/key-id is stored in the models database. Models can then be
// linked to one of the existing signing-keys.
func KeypairCreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
	if err != nil {
		formatBooleanResponse(false, "error-auth", "", "", w)
		return
	}

	// Check that we have a message body
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-nil-data", "", "Uninitialized POST data", w)
		return
	}
	defer r.Body.Close()

	// Decode the JSON body
	keypairWithKey := KeypairWithPrivateKey{}
	err = json.NewDecoder(r.Body).Decode(&keypairWithKey)
	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-keypair-data", "", "No keypair data supplied", w)
		return
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-keypair-json", "", err.Error(), w)
		return
	}

	// Validate the authority-id
	keypairWithKey.AuthorityID = strings.TrimSpace(keypairWithKey.AuthorityID)
	if len(keypairWithKey.AuthorityID) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-keypair-json", "", "The authority-id is mandatory", w)
		return
	}

	// Check that the user has permissions to this authority-id
	if !datastore.Environ.DB.CheckUserInAccount(authUser.Username, keypairWithKey.AuthorityID) {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-auth", "", "Your user does not have permissions for the Signing Authority", w)
		return
	}

	// Store the signing-key in the keypair store using the asserts module
	privateKey, sealedPrivateKey, err := datastore.Environ.KeypairDB.ImportSigningKey(keypairWithKey.AuthorityID, keypairWithKey.PrivateKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-keypair-store", "", err.Error(), w)
		return
	}

	// Store the signing-key in the database
	keypair := datastore.Keypair{
		AuthorityID: keypairWithKey.AuthorityID,
		KeyID:       privateKey.PublicKey().ID(),
		SealedKey:   sealedPrivateKey,
	}
	errorCode, err := datastore.Environ.DB.PutKeypair(keypair)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, errorCode, "", err.Error(), w)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	formatBooleanResponse(true, "", "", "", w)
}

// KeypairDisableHandler disables an existing keypair, which will mean that any
// linked Models will not be able to be signed. The asserts module does not allow
// a keypair to be deleted, so the keypair will just be disabled in the local database.
func KeypairDisableHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
	if err != nil {
		formatBooleanResponse(false, "error-auth", "", "", w)
		return
	}

	// Get the keypair primary key
	vars := mux.Vars(r)
	keypairID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("%v", vars["id"])
		formatBooleanResponse(false, "error-invalid-keypair", "", errorMessage, w)
		return
	}

	// Update the keypair in the local database
	err = updateAllowedKeypairActive(keypairID, false, authUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-keypair-update", "", err.Error(), w)
		return
	}

	formatBooleanResponse(true, "", "", "", w)
}

// KeypairEnableHandler enables an existing keypair, which will mean that any
// linked Models will be able to be signed. The asserts module does not allow
// a keypair to be deleted, so the keypair will just be enabled in the local database.
func KeypairEnableHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
	if err != nil {
		formatBooleanResponse(false, "error-auth", "", "", w)
		return
	}

	// Get the keypair primary key
	vars := mux.Vars(r)
	keypairID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("%v", vars["id"])
		formatBooleanResponse(false, "error-invalid-keypair", "", errorMessage, w)
		return
	}

	// Update the keypair in the local database
	err = updateAllowedKeypairActive(keypairID, true, authUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-keypair-update", "", err.Error(), w)
		return
	}
	formatBooleanResponse(true, "", "", "", w)
}

func updateAllowedKeypairActive(keypairID int, active bool, authUser datastore.User) error {
	switch authUser.Role {
	case 0:
		fallthrough
	case datastore.Superuser:
		return updateKeypairActive(keypairID, active)
	case datastore.Admin:
		return updateKeypairActiveFilteredByUser(keypairID, active, authUser.Username)
	}
	return nil
}

func updateKeypairActive(keypairID int, active bool) error {
	return datastore.Environ.DB.UpdateKeypairActive(keypairID, active, "")
}

func updateKeypairActiveFilteredByUser(keypairID int, active bool, username string) error {
	return datastore.Environ.DB.UpdateKeypairActive(keypairID, active, username)
}

// KeypairAssertionHandler updates the account key assertion on a keypair
func KeypairAssertionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
	if err != nil {
		formatBooleanResponse(false, "error-auth", "", "", w)
		return
	}

	// Check that we have a message body
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-nil-data", "", "Uninitialized POST data", w)
		return
	}
	defer r.Body.Close()

	assertionRequest := AssertionRequest{}
	err = json.NewDecoder(r.Body).Decode(&assertionRequest)
	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-assertion-data", "", "No assertion data supplied", w)
		return
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-assertion-json", "", err.Error(), w)
		return
	}

	// Check that a keypair ID has been provided
	if assertionRequest.ID == 0 {
		logMessage("KEYPAIR", "invalid-keypair", "ID of keypair not provided")
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "invalid-keypair", "", "ID of keypair not provided", w)
		return
	}

	// Decode the file
	decodedAssertion, err := base64.StdEncoding.DecodeString(assertionRequest.Assertion)
	if err != nil {
		logMessage("KEYPAIR", "invalid-assertion", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "decode-assertion", "", err.Error(), w)
		return
	}

	// Validate the assertion in the request
	assertion, err := asserts.Decode(decodedAssertion)
	if err != nil {
		logMessage("KEYPAIR", "invalid-assertion", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "decode-assertion", "", err.Error(), w)
		return
	}

	// Check that we have an account key assertion
	if assertion.Type().Name != asserts.AccountKeyType.Name {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "invalid-assertion", "", fmt.Sprintf("An assertion of type '%s' is required", asserts.AccountKeyType.Name), w)
		return
	}

	// Check that the user has permissions to this authority-id
	if !datastore.Environ.DB.CheckUserInAccount(authUser.Username, assertion.AuthorityID()) {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-auth", "", "Your user does not have permissions for the Authority", w)
		return
	}

	// Store or update the account assertion in the database
	err = datastore.Environ.DB.UpdateKeypairAssertion(assertionRequest.ID, string(decodedAssertion))
	if err != nil {
		logMessage("KEYPAIR", "invalid-assertion", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "invalid-assertion", "", err.Error(), w)
		return
	}

	// Return the success response
	formatBooleanResponse(true, "", "", "", w)
}
