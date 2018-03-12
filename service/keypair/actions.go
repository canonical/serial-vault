// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018-2019 Canonical Ltd
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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/auth"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/snapcore/snapd/asserts"
)

// ListResponse is the JSON response from the API Keypairs method
type ListResponse struct {
	Success      bool                `json:"success"`
	ErrorCode    string              `json:"error_code"`
	ErrorSubcode string              `json:"error_subcode"`
	ErrorMessage string              `json:"message"`
	Keypairs     []datastore.Keypair `json:"keypairs"`
}

// ProgressResponse is the JSON response from the API status of keypair generation
type ProgressResponse struct {
	Success      bool                      `json:"success"`
	ErrorCode    string                    `json:"error_code"`
	ErrorSubcode string                    `json:"error_subcode"`
	ErrorMessage string                    `json:"message"`
	Status       []datastore.KeypairStatus `json:"status"`
}

// listHandler is the API method to fetch the signing keys
func listHandler(w http.ResponseWriter, user datastore.User, apiCall bool) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	keypairs, err := datastore.Environ.DB.ListAllowedKeypairs(user)
	if err != nil {
		response.FormatStandardResponse(false, "error-fetch-keypairs", "", err.Error(), w)
		return
	}

	// Return successful JSON response with the list of models
	w.WriteHeader(http.StatusOK)
	formatListResponse(true, "", "", "", keypairs, w)
}

// createHandler is the API method to create a signing key
func createHandler(w http.ResponseWriter, user datastore.User, apiCall bool, keypairWithKey WithPrivateKey) {
	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	// Store the signing-key in the keypair store using the asserts module
	privateKey, sealedPrivateKey, err := datastore.Environ.KeypairDB.ImportSigningKey(keypairWithKey.AuthorityID, keypairWithKey.PrivateKey)
	if err != nil {
		response.FormatStandardResponse(false, "error-keypair-store", "", err.Error(), w)
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
		response.FormatStandardResponse(false, errorCode, "", err.Error(), w)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", "", w)
}

// generateHandler is the API method to generate a signing key
func generateHandler(w http.ResponseWriter, user datastore.User, apiCall bool, keypairWithKey WithPrivateKey) {
	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	go datastore.GenerateKeypair(keypairWithKey.AuthorityID, "", keypairWithKey.KeyName)

	// Return the URL to watch for the response
	statusURL := fmt.Sprintf("/v1/keypairs/status/%s/%s", keypairWithKey.AuthorityID, keypairWithKey.KeyName)
	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Location", statusURL)
	response.FormatStandardResponse(true, "", "", statusURL, w)
}

// enableDisableHandler is the API method to enable/disable a signing key
func enableDisableHandler(w http.ResponseWriter, user datastore.User, apiCall bool, enabled bool, keypairID int) {
	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	// Update the keypair in the local database
	err = datastore.Environ.DB.UpdateAllowedKeypairActive(keypairID, enabled, user)
	if err != nil {
		response.FormatStandardResponse(false, "error-keypair-update", "", err.Error(), w)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", "", w)
}

// assertionHandler is the API method to update a key assertion
func assertionHandler(w http.ResponseWriter, user datastore.User, apiCall bool, assertionRequest AssertionRequest) {
	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	// Check that a keypair ID has been provided
	if assertionRequest.ID == 0 {
		response.FormatStandardResponse(false, "invalid-keypair", "", "ID of keypair not provided", w)
		return
	}

	// Decode the file
	decodedAssertion, err := base64.StdEncoding.DecodeString(assertionRequest.Assertion)
	if err != nil {
		response.FormatStandardResponse(false, "decode-assertion", "", err.Error(), w)
		return
	}

	// Validate the assertion in the request
	assertion, err := asserts.Decode(decodedAssertion)
	if err != nil {
		response.FormatStandardResponse(false, "decode-assertion", "", err.Error(), w)
		return
	}

	// Check that we have an account key assertion
	if assertion.Type().Name != asserts.AccountKeyType.Name {
		response.FormatStandardResponse(false, "invalid-assertion", "", fmt.Sprintf("An assertion of type '%s' is required", asserts.AccountKeyType.Name), w)
		return
	}

	keypair := datastore.Keypair{
		ID:          assertionRequest.ID,
		AuthorityID: assertion.HeaderString("account-id"),
		KeyID:       assertion.HeaderString("public-key-sha3-384"),
		Assertion:   string(decodedAssertion),
	}

	errorCode, err := datastore.Environ.DB.UpdateKeypairAssertion(keypair, user)
	if err != nil {
		response.FormatStandardResponse(false, errorCode, "", err.Error(), w)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", "", w)
}

// statusHandler is the API method to fetch the status of a signing key
func statusHandler(w http.ResponseWriter, user datastore.User, apiCall bool, authorityID, keyName string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	// Check that the user has permissions to this authority-id
	if !datastore.Environ.DB.CheckUserInAccount(user.Username, authorityID) {
		response.FormatStandardResponse(false, "error-auth", "", "Your user does not have permissions for the Signing Authority", w)
		return
	}

	ks, err := datastore.Environ.DB.GetKeypairStatus(authorityID, keyName)
	if err != nil {
		response.FormatStandardResponse(false, "error-keypair-json", "", "Cannot find the status of the keypair", w)
		return
	}

	// Return successful JSON response with the list of models
	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", ks.Status, w)
}

// progressHandler is the API method to fetch the progress of signing key generation
func progressHandler(w http.ResponseWriter, user datastore.User, apiCall bool) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	ks, err := datastore.Environ.DB.ListAllowedKeypairStatus(user)
	if err != nil {
		response.FormatStandardResponse(false, "error-keypair-json", "", "Cannot find the status of the keypairs", w)
		return
	}

	formatProgressResponse(ks, w)
}

func formatListResponse(success bool, errorCode, errorSubcode, message string, keypairs []datastore.Keypair, w http.ResponseWriter) error {
	response := ListResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, Keypairs: keypairs}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the keypairs response.")
		return err
	}
	return nil
}

func formatProgressResponse(status []datastore.KeypairStatus, w http.ResponseWriter) error {
	response := ProgressResponse{Success: true, Status: status}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the keypair status response.")
		return err
	}
	return nil
}
