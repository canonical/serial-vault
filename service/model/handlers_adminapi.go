// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
 * License granted by Canonical Limited
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

package model

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

// APIList is the API method to fetch the models
func APIList(w http.ResponseWriter, r *http.Request) {
	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	// Call the API with the user
	listHandler(w, user, true)
}

// APIGet is the API method to fetch a model
func APIGet(w http.ResponseWriter, r *http.Request) {
	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.FormatStandardResponse(false, "error-invalid-model", "", err.Error(), w)
		return
	}

	// Call the API with the user
	getHandler(w, user, true, id)
}

// APIUpdate is the API method to update a model
func APIUpdate(w http.ResponseWriter, r *http.Request) {
	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	vars := mux.Vars(r)
	modelID, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.FormatStandardResponse(false, "error-invalid-model", "", err.Error(), w)
		return
	}

	defer r.Body.Close()

	// Decode the JSON body
	mdl := datastore.Model{}
	err = json.NewDecoder(r.Body).Decode(&mdl)
	switch {
	// Check we have some data
	case err == io.EOF:
		response.FormatStandardResponse(false, "error-model-data", "", "No model data supplied.", w)
		return
		// Check for parsing errors
	case err != nil:
		response.FormatStandardResponse(false, "error-decode-json", "", err.Error(), w)
		return
	}

	updateHandler(w, user, true, modelID, mdl)
}

// APIDelete is the API method to delete a model
func APIDelete(w http.ResponseWriter, r *http.Request) {
	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	vars := mux.Vars(r)
	modelID, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.FormatStandardResponse(false, "error-invalid-model", "", err.Error(), w)
		return
	}

	deleteHandler(w, user, true, modelID)
}

// APICreate is the API method to create a model
func APICreate(w http.ResponseWriter, r *http.Request) {
	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	defer r.Body.Close()

	// Decode the JSON body
	mdl := datastore.Model{}
	err = json.NewDecoder(r.Body).Decode(&mdl)
	switch {
	// Check we have some data
	case err == io.EOF:
		response.FormatStandardResponse(false, "error-model-data", "", "No model data supplied.", w)
		return
		// Check for parsing errors
	case err != nil:
		response.FormatStandardResponse(false, "error-decode-json", "", err.Error(), w)
		return
	}

	createHandler(w, user, true, mdl)
}

// APIAssertionHeaders is the API method to upsert the model assertion header details
func APIAssertionHeaders(w http.ResponseWriter, r *http.Request) {
	// Validate the user and API key
	user, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	defer r.Body.Close()

	// Decode the JSON body
	assert := datastore.ModelAssertion{}
	err = json.NewDecoder(r.Body).Decode(&assert)
	switch {
	// Check we have some data
	case err == io.EOF:
		response.FormatStandardResponse(false, "error-model-data", "", "No model data supplied", w)
		return
		// Check for parsing errors
	case err != nil:
		response.FormatStandardResponse(false, "error-decode-json", "", err.Error(), w)
		return
	}

	assertionHeaders(w, user, true, assert)
}
