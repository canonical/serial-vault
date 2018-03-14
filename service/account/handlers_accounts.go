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
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/auth"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/gorilla/mux"
)

// AssertionRequest is the JSON version of a account assertion
type AssertionRequest struct {
	ID        int    `json:"id"`
	Assertion string `json:"assertion"`
}

// List is the API method to list the account assertions
func List(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	listHandler(w, authUser, false)
}

// Create is the API method to create an account
func Create(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	defer r.Body.Close()

	// Decode the JSON body
	acct := datastore.Account{}
	err = json.NewDecoder(r.Body).Decode(&acct)
	switch {
	// Check we have some data
	case err == io.EOF:
		response.FormatStandardResponse(false, "error-account-data", "", "No account data supplied.", w)
		return
		// Check for parsing errors
	case err != nil:
		response.FormatStandardResponse(false, "error-decode-json", "", err.Error(), w)
		return
	}

	createHandler(w, authUser, false, acct)
}

// Get is the API method to fetch an account
func Get(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.FormatStandardResponse(false, "error-invalid-acccount", "", err.Error(), w)
		return
	}

	getHandler(w, authUser, false, id)
}

// Update is the API method to update a model
func Update(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	defer r.Body.Close()

	// Decode the JSON body
	acct := datastore.Account{}
	err = json.NewDecoder(r.Body).Decode(&acct)
	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		response.FormatStandardResponse(false, "error-account-data", "", "No account data supplied", w)
		return
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		response.FormatStandardResponse(false, "error-decode-json", "", err.Error(), w)
		return
	}

	updateHandler(w, authUser, false, acct)
}

// Upload is the API method to upload an account assertion
func Upload(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	defer r.Body.Close()

	// Decode the JSON body
	assertionRequest := AssertionRequest{}
	err = json.NewDecoder(r.Body).Decode(&assertionRequest)
	switch {
	// Check we have some data
	case err == io.EOF:
		response.FormatStandardResponse(false, "error-assertion-data", "", "No assertion data supplied.", w)
		return
		// Check for parsing errors
	case err != nil:
		response.FormatStandardResponse(false, "error-decode-json", "", err.Error(), w)
		return
	}

	uploadHandler(w, authUser, false, assertionRequest)
}
