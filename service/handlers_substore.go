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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/gorilla/mux"
)

// SubstoresResponse is the JSON response from the API Sub-Stores method
type SubstoresResponse struct {
	Success      bool                 `json:"success"`
	ErrorCode    string               `json:"error_code"`
	ErrorSubcode string               `json:"error_subcode"`
	ErrorMessage string               `json:"message"`
	Substores    []datastore.Substore `json:"substores"`
}

// SubstoresHandler is the API method to list the sub-stores
func SubstoresHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-auth", "", "", w)
		return
	}

	vars := mux.Vars(r)
	accountID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("%v", vars)
		formatBooleanResponse(false, "error-invalid-account", "", errorMessage, w)
		return
	}

	stores, err := datastore.Environ.DB.ListSubstores(accountID, authUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-stores-json", "", err.Error(), w)
		return
	}

	// Format the model for output and return JSON response
	w.WriteHeader(http.StatusOK)
	formatSubstoresResponse(true, "", "", "", stores, w)
}

// SubstoreUpdateHandler is the API method to update a sub-store
func SubstoreUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-auth", "", "", w)
		return
	}

	vars := mux.Vars(r)
	_, err = strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-invalid-store", "", err.Error(), w)
		return
	}

	// Decode the JSON body
	store := datastore.Substore{}
	err = json.NewDecoder(r.Body).Decode(&store)
	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-store-data", "", "No sub-store data supplied.", w)
		return
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-decode-json", "", err.Error(), w)
		return
	}

	err = datastore.Environ.DB.UpdateAllowedSubstore(store, authUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-updating-store", "", err.Error(), w)
		return
	}

	w.WriteHeader(http.StatusOK)
	formatBooleanResponse(true, "", "", "", w)
}

// SubstoreCreateHandler is the API method to update a sub-store
func SubstoreCreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-auth", "", "", w)
		return
	}

	// Decode the JSON body
	store := datastore.Substore{}
	err = json.NewDecoder(r.Body).Decode(&store)
	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-store-data", "", "No sub-store data supplied.", w)
		return
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-decode-json", "", err.Error(), w)
		return
	}

	err = datastore.Environ.DB.CreateAllowedSubstore(store, authUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-creating-store", "", err.Error(), w)
		return
	}

	w.WriteHeader(http.StatusOK)
	formatBooleanResponse(true, "", "", "", w)
}

// SubstoreDeleteHandler is the API method to delete a sub-store model.
func SubstoreDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-auth", "", "", w)
		return
	}

	// Get the sub-store primary key
	vars := mux.Vars(r)
	storeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", vars["id"])
		formatBooleanResponse(false, "error-invalid-store", "", errorMessage, w)
		return
	}

	// Update the database
	errorSubcode, err := datastore.Environ.DB.DeleteAllowedSubstore(storeID, authUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-deleting-store", errorSubcode, err.Error(), w)
		return
	}

	w.WriteHeader(http.StatusOK)
	formatBooleanResponse(true, "", "", "", w)
}
