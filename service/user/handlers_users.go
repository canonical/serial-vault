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

package user

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

// Request is the JSON version of the request to create a user
type Request struct {
	Username string   `json:"username"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	APIKey   string   `json:"api_key"`
	Role     int      `json:"role"`
	Accounts []string `json:"accounts"`
}

// List is the API method to fetch the users
func List(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	listHandler(w, authUser, false)
}

// Get is the API method to fetch a user
func Get(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.FormatStandardResponse(false, "error-invalid-user", "", err.Error(), w)
		return
	}

	getHandler(w, authUser, false, id)
}

// Create is the API method to create a sub-store model
func Create(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	defer r.Body.Close()

	// Decode the JSON body
	userRequest := Request{}
	err = json.NewDecoder(r.Body).Decode(&userRequest)
	switch {
	// Check we have some data
	case err == io.EOF:
		response.FormatStandardResponse(false, "error-user-data", "", "No user data supplied.", w)
		return
		// Check for parsing errors
	case err != nil:
		response.FormatStandardResponse(false, "error-decode-json", "", err.Error(), w)
		return
	}

	// Create a new user
	user := datastore.User{
		Username: userRequest.Username,
		Name:     userRequest.Name,
		Email:    userRequest.Email,
		Role:     userRequest.Role,
		APIKey:   userRequest.APIKey,
		Accounts: datastore.BuildAccountsFromAuthorityIDs(userRequest.Accounts),
	}
	user.ID, err = datastore.Environ.DB.CreateUser(user)

	createHandler(w, authUser, false, user)
}

// Update is the API method to update a user
func Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.FormatStandardResponse(false, "error-invalid-account", "", err.Error(), w)
		return
	}

	defer r.Body.Close()

	// Decode the JSON body
	userRequest := Request{}
	err = json.NewDecoder(r.Body).Decode(&userRequest)
	switch {
	// Check we have some data
	case err == io.EOF:
		response.FormatStandardResponse(false, "error-user-data", "", "No user data supplied.", w)
		return
		// Check for parsing errors
	case err != nil:
		response.FormatStandardResponse(false, "error-decode-json", "", err.Error(), w)
		return
	}

	// Form a database record
	user := datastore.User{
		ID:       userID,
		Username: userRequest.Username,
		Name:     userRequest.Name,
		Email:    userRequest.Email,
		Role:     userRequest.Role,
		APIKey:   userRequest.APIKey,
		Accounts: datastore.BuildAccountsFromAuthorityIDs(userRequest.Accounts),
	}

	updateHandler(w, authUser, false, user)
}

// Delete is the API method to delete a user
func Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.FormatStandardResponse(false, "error-invalid-user", "", err.Error(), w)
		return
	}

	deleteHandler(w, authUser, false, userID)
}

// GetOtherAccounts is the API method to retrieve accounts not belonging to the user
func GetOtherAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		response.FormatStandardResponse(false, "error-invalid-user", "", err.Error(), w)
		return
	}

	getOtherAccountsHandler(w, authUser, false, id)
}
