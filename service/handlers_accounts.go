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

	"github.com/CanonicalLtd/serial-vault/account"
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

// AccountResponse is the JSON response from the API Account method
type AccountResponse struct {
	Success      bool              `json:"success"`
	ErrorCode    string            `json:"error_code"`
	ErrorSubcode string            `json:"error_subcode"`
	ErrorMessage string            `json:"message"`
	Account      datastore.Account `json:"account"`
}

// AssertionRequest is the JSON version of a account assertion
type AssertionRequest struct {
	ID        int    `json:"id"`
	Assertion string `json:"assertion"`
}

// // AccountsHandler is the API method to list the account assertions
// func AccountsHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

// 	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
// 	if err != nil {
// 		formatAccountsResponse(false, "error-auth", "", "", nil, w)
// 		return
// 	}

// 	accounts, err := datastore.Environ.DB.ListAllowedAccounts(authUser)
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		formatAccountsResponse(false, "error-accounts-json", "", err.Error(), nil, w)
// 		return
// 	}

// 	// Format the model for output and return JSON response
// 	w.WriteHeader(http.StatusOK)
// 	formatAccountsResponse(true, "", "", "", accounts, w)
// }

// // AccountGetHandler is the API method to fetch an account
// func AccountGetHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

// 	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
// 	if err != nil {
// 		formatBooleanResponse(false, "error-auth", "", "", w)
// 		return
// 	}

// 	vars := mux.Vars(r)
// 	accountID, err := strconv.Atoi(vars["id"])
// 	if err != nil {
// 		w.WriteHeader(http.StatusNotFound)
// 		errorMessage := fmt.Sprintf("%v", vars)
// 		formatBooleanResponse(false, "error-invalid-account", "", errorMessage, w)
// 		return
// 	}

// 	account, err := datastore.Environ.DB.GetAccountByID(accountID, authUser)
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		formatBooleanResponse(false, "error-account", "", err.Error(), w)
// 		return
// 	}

// 	formatAccountResponse(true, "", "", "", account, w)
// }

// // AccountUpdateHandler is the API method to update an account
// func AccountUpdateHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

// 	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
// 	if err != nil {
// 		formatBooleanResponse(false, "error-auth", "", "", w)
// 		return
// 	}

// 	// Decode the JSON body
// 	acct := datastore.Account{}
// 	err = json.NewDecoder(r.Body).Decode(&acct)
// 	switch {
// 	// Check we have some data
// 	case err == io.EOF:
// 		w.WriteHeader(http.StatusBadRequest)
// 		formatBooleanResponse(false, "error-account-data", "", "No account data supplied", w)
// 		return
// 		// Check for parsing errors
// 	case err != nil:
// 		w.WriteHeader(http.StatusBadRequest)
// 		formatBooleanResponse(false, "error-decode-json", "", err.Error(), w)
// 		return
// 	}

// 	err = datastore.Environ.DB.UpdateAccount(acct, authUser)
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		formatBooleanResponse(false, "error-account", "", err.Error(), w)
// 		return
// 	}

// 	formatBooleanResponse(true, "", "", "", w)
// }

// AccountCreateHandler is the API method to create an account
func AccountCreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	_, err := checkIsAdminAndGetUserFromJWT(w, r)
	if err != nil {
		formatBooleanResponse(false, "error-auth", "", "", w)
		return
	}

	// Decode the JSON body
	acct := datastore.Account{}
	err = json.NewDecoder(r.Body).Decode(&acct)
	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-account-data", "", "No account data supplied", w)
		return
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-decode-json", "", err.Error(), w)
		return
	}

	// Fetch the account assertion from the store
	assertion, err := account.FetchAssertionFromStore(asserts.AccountType, []string{acct.AuthorityID})
	if err != nil {
		formatBooleanResponse(false, "error-account", "", "Error fetching the assertion from the store", w)
		return
	}
	acct.Assertion = string(asserts.Encode(assertion))

	// Store the account details
	err = datastore.Environ.DB.CreateAccount(acct)
	if err != nil {
		formatBooleanResponse(false, "error-creating-account", "", "Error creating the account in the database", w)
		return
	}

	formatBooleanResponse(true, "", "", "", w)
}

// AccountsUploadHandler creates or updates an account assertion by uploading a file
func AccountsUploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get the user from the JWT
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

	account := datastore.Account{
		AuthorityID: assertion.HeaderString("account-id"),
		Assertion:   string(decodedAssertion),
	}

	errorCode, err := datastore.Environ.DB.PutAccount(account, authUser)
	if err != nil {
		logMessage("ACCOUNT", "invalid-assertion", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, errorCode, "", err.Error(), w)
		return
	}

	// Return the success response
	formatBooleanResponse(true, "", "", "", w)
}
