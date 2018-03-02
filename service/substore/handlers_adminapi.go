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

package substore

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/request"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/gorilla/mux"
)

// APIList is the API method to fetch the sub-store models
func APIList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	vars := mux.Vars(r)
	accountID, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.FormatStandardResponse(false, "error-invalid-account", "", err.Error(), w)
		return
	}

	// Call the API with the user
	listHandler(w, user, true, accountID)
}

// APIUpdate is the API method to update a sub-store model
func APIUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	vars := mux.Vars(r)
	storeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.FormatStandardResponse(false, "error-invalid-store", "", err.Error(), w)
		return
	}

	// Decode the JSON body
	store := datastore.Substore{}
	err = json.NewDecoder(r.Body).Decode(&store)
	switch {
	// Check we have some data
	case err == io.EOF:
		response.FormatStandardResponse(false, "error-store-data", "", "No sub-store data supplied.", w)
		return
		// Check for parsing errors
	case err != nil:
		response.FormatStandardResponse(false, "error-decode-json", "", err.Error(), w)
		return
	}

	// Call the API with the user
	updateHandler(w, user, true, storeID, store)
}

// APICreate is the API method to create a sub-store model
func APICreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	// Decode the JSON body
	store := datastore.Substore{}
	err = json.NewDecoder(r.Body).Decode(&store)
	switch {
	// Check we have some data
	case err == io.EOF:
		response.FormatStandardResponse(false, "error-store-data", "", "No sub-store data supplied.", w)
		return
		// Check for parsing errors
	case err != nil:
		response.FormatStandardResponse(false, "error-decode-json", "", err.Error(), w)
		return
	}

	// Call the API with the user
	createHandler(w, user, true, store)
}

// APIDelete is the API method to delete a sub-store model
func APIDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	vars := mux.Vars(r)
	storeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.FormatStandardResponse(false, "error-invalid-store", "", err.Error(), w)
		return
	}

	// Call the API with the user
	deleteHandler(w, user, true, storeID)
}
