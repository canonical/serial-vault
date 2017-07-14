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

package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/snapcore/snapd/asserts"
)

// AccountsResponse is the JSON response from the API Accounts method
type AccountsResponse struct {
	Success      bool                `json:"success"`
	ErrorCode    string              `json:"error_code"`
	ErrorSubcode string              `json:"error_subcode"`
	ErrorMessage string              `json:"message"`
	Accounts     []datastore.Account `json:"accounts"`
}

// AssertionRequest is the JSON version of a account assertion
type AssertionRequest struct {
	ID        int    `json:"id"`
	Assertion string `json:"assertion"`
}

// AccountsHandler is the API method to list the account assertions
func AccountsHandler(w http.ResponseWriter, r *http.Request) {

	// Get the user from the JWT
	username, err := checkUserPermissions(w, r)
	if err != nil {
		formatAccountsResponse(false, "error-auth", "", "", nil, w)
		return
	}

	accounts, err := datastore.Environ.DB.ListAccounts(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		formatAccountsResponse(false, "error-accounts-json", "", err.Error(), nil, w)
		return
	}

	// Format the model for output and return JSON response
	w.WriteHeader(http.StatusOK)
	formatAccountsResponse(true, "", "", "", accounts, w)
}

// AccountsUpsertHandler creates or updates an account assertion
func AccountsUpsertHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get the user from the JWT
	username, err := checkUserPermissions(w, r)
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

	// Decode the file
	decodedAssertion, err := base64.StdEncoding.DecodeString(assertionRequest.Assertion)
	if err != nil {
		logMessage("ACCOUNT", "invalid-assertion", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "decode-assertion", "", err.Error(), w)
		return
	}

	// Validate the assertion in the request
	assertion, err := asserts.Decode(decodedAssertion)
	if err != nil {
		logMessage("ACCOUNT", "invalid-assertion", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "decode-assertion", "", err.Error(), w)
		return
	}

	// Check that we have an account assertion
	if assertion.Type().Name != asserts.AccountType.Name {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "invalid-assertion", "", fmt.Sprintf("An assertion of type '%s' is required", asserts.AccountType.Name), w)
		return
	}

	// Check that the user has permissions for the authority-id
	if !datastore.Environ.DB.CheckUserInAccount(username, assertion.AuthorityID()) {
		formatBooleanResponse(false, "error-auth", "", "You do not have permissions for that authority", w)
		return
	}

	// Store or update the account assertion in the database
	errorCode, err := datastore.Environ.DB.PutAccount(datastore.Account{AuthorityID: assertion.HeaderString("account-id"), Assertion: string(decodedAssertion)})
	if err != nil {
		logMessage("ACCOUNT", "invalid-assertion", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, errorCode, "", err.Error(), w)
		return
	}

	// Return the success response
	formatBooleanResponse(true, "", "", "", w)
}
