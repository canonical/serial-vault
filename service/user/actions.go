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

package user

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/auth"
	"github.com/CanonicalLtd/serial-vault/service/response"
)

// ListResponse is the response from a user list request
type ListResponse struct {
	Success      bool             `json:"success"`
	ErrorCode    string           `json:"error_code"`
	ErrorSubcode string           `json:"error_subcode"`
	ErrorMessage string           `json:"message"`
	Users        []datastore.User `json:"users"`
}

// GetResponse is the response from a user creation/update
type GetResponse struct {
	Success      bool           `json:"success"`
	ErrorCode    string         `json:"error_code"`
	ErrorSubcode string         `json:"error_subcode"`
	ErrorMessage string         `json:"message"`
	User         datastore.User `json:"user"`
}

// AccountsResponse is the JSON response from the API Accounts method
type AccountsResponse struct {
	Success      bool                `json:"success"`
	ErrorCode    string              `json:"error_code"`
	ErrorSubcode string              `json:"error_subcode"`
	ErrorMessage string              `json:"message"`
	Accounts     []datastore.Account `json:"accounts"`
}

// listHandler is the API method to fetch the user records
func listHandler(w http.ResponseWriter, user datastore.User, apiCall bool) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Superuser, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	users, err := datastore.Environ.DB.ListUsers()
	if err != nil {
		response.FormatStandardResponse(false, "error-fetch-users", "", err.Error(), w)
		return
	}

	// Return successful JSON response with the list of models
	w.WriteHeader(http.StatusOK)
	formatListResponse(users, w)
}

// getHandler is the API method to fetch the log records from signing
func getHandler(w http.ResponseWriter, user datastore.User, apiCall bool, userID int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Superuser, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	u, err := datastore.Environ.DB.GetUser(userID)
	if err != nil {
		response.FormatStandardResponse(false, "error-fetch-users", "", err.Error(), w)
		return
	}

	// Return successful JSON response with the list of models
	w.WriteHeader(http.StatusOK)
	formatUserResponse(u, w)
}

func createHandler(w http.ResponseWriter, authUser datastore.User, apiCall bool, user datastore.User) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(authUser, datastore.Superuser, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	user.ID, err = datastore.Environ.DB.CreateUser(user)
	if err != nil {
		response.FormatStandardResponse(false, "error-creating-user", "", "", w)
		return
	}

	// Return successful JSON response
	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", "", w)
}

func formatListResponse(users []datastore.User, w http.ResponseWriter) error {
	response := ListResponse{Success: true, Users: users}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the users response.")
		return err
	}
	return nil
}

func formatUserResponse(user datastore.User, w http.ResponseWriter) error {
	response := GetResponse{Success: true, User: user}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the user response.")
		return err
	}
	return nil
}

func updateHandler(w http.ResponseWriter, authUser datastore.User, apiCall bool, user datastore.User) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(authUser, datastore.Superuser, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	err = datastore.Environ.DB.UpdateUser(user)
	if err != nil {
		log.Println("Error updating the store:", err)
		response.FormatStandardResponse(false, "error-stores-substore", "", "Error updating the store", w)
		return
	}

	// Return successful JSON response
	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", "", w)
}

func deleteHandler(w http.ResponseWriter, user datastore.User, apiCall bool, userID int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Superuser, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	err = datastore.Environ.DB.DeleteUser(userID)
	if err != nil {
		response.FormatStandardResponse(false, "error-deleting-user", "", err.Error(), w)
		return
	}

	// Return successful JSON response
	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", "", w)
}

func getOtherAccountsHandler(w http.ResponseWriter, authUser datastore.User, apiCall bool, userID int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(authUser, datastore.Superuser, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth2", "", "", w)
		return
	}

	user, err := datastore.Environ.DB.GetUser(userID)
	if err != nil {
		response.FormatStandardResponse(false, "error-get-user", "", err.Error(), w)
		return
	}

	accounts, err := datastore.Environ.DB.ListNotUserAccounts(user.Username)
	if err != nil {
		response.FormatStandardResponse(false, "error-get-non-user-accounts", "", err.Error(), w)
		return
	}

	// Format the model for output and return JSON response
	w.WriteHeader(http.StatusOK)
	formatAccountsResponse(accounts, w)
}

func formatAccountsResponse(accounts []datastore.Account, w http.ResponseWriter) error {
	response := AccountsResponse{Success: true, Accounts: accounts}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the accounts response.")
		return err
	}
	return nil
}
