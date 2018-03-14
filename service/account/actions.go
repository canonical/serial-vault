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

package account

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

// ListResponse is the JSON response from the API Accounts method
type ListResponse struct {
	Success      bool                `json:"success"`
	ErrorCode    string              `json:"error_code"`
	ErrorSubcode string              `json:"error_subcode"`
	ErrorMessage string              `json:"message"`
	Accounts     []datastore.Account `json:"accounts"`
}

// GetResponse is the JSON response from the API Account method
type GetResponse struct {
	Success      bool              `json:"success"`
	ErrorCode    string            `json:"error_code"`
	ErrorSubcode string            `json:"error_subcode"`
	ErrorMessage string            `json:"message"`
	Account      datastore.Account `json:"account"`
}

// listHandler is the API method to fetch the user records
func listHandler(w http.ResponseWriter, user datastore.User, apiCall bool) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	accounts, err := datastore.Environ.DB.ListAllowedAccounts(user)
	if err != nil {
		response.FormatStandardResponse(false, "error-fetch-models", "", err.Error(), w)
		return
	}

	// Return successful JSON response with the list of models
	w.WriteHeader(http.StatusOK)
	formatListResponse(accounts, w)
}

func createHandler(w http.ResponseWriter, user datastore.User, apiCall bool, acct datastore.Account) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	err = datastore.Environ.DB.CreateAccount(acct)
	if err != nil {
		response.FormatStandardResponse(false, "error-creating-account", "", "Error creating the account in the database", w)
		return
	}

	// Return successful JSON response
	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", "", w)
}

// getHandler is the API method to fetch the accounts
func getHandler(w http.ResponseWriter, user datastore.User, apiCall bool, accountID int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	account, err := datastore.Environ.DB.GetAccountByID(accountID, user)
	if err != nil {
		response.FormatStandardResponse(false, "error-account", "", err.Error(), w)
		return
	}

	// Return successful JSON response with the list of models
	w.WriteHeader(http.StatusOK)
	formatGetResponse(account, w)
}

func updateHandler(w http.ResponseWriter, user datastore.User, apiCall bool, acct datastore.Account) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	err = datastore.Environ.DB.UpdateAccount(acct, user)
	if err != nil {
		log.Println("Error updating the account:", err)
		response.FormatStandardResponse(false, "error-account", "", "Error updating the model", w)
		return
	}

	// Return successful JSON response
	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", "", w)
}

func uploadHandler(w http.ResponseWriter, user datastore.User, apiCall bool, assertionRequest AssertionRequest) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
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

	// Check that we have an account assertion
	if assertion.Type().Name != asserts.AccountType.Name {
		response.FormatStandardResponse(false, "invalid-assertion", "", fmt.Sprintf("An assertion of type '%s' is required", asserts.AccountType.Name), w)
		return
	}

	account := datastore.Account{
		AuthorityID: assertion.HeaderString("account-id"),
		Assertion:   string(decodedAssertion),
	}

	errorCode, err := datastore.Environ.DB.PutAccount(account, user)
	if err != nil {
		response.FormatStandardResponse(false, errorCode, "", err.Error(), w)
		return
	}

	// Return successful JSON response
	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", "", w)
}

func formatListResponse(accounts []datastore.Account, w http.ResponseWriter) error {
	response := ListResponse{Success: true, Accounts: accounts}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the accounts response.")
		return err
	}
	return nil
}

func formatGetResponse(account datastore.Account, w http.ResponseWriter) error {
	response := GetResponse{Success: true, Account: account}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the account response.")
		return err
	}
	return nil
}
