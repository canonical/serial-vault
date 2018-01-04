// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2017 Canonical Ltd
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
	"io"
	"net/http"
	"time"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/snapcore/snapd/asserts"
)

// ModelAssertion is the JSON version of a model assertion request
type ModelAssertion struct {
	BrandID      string `json:"brand-id"`
	Name         string `json:"model"`
	Series       string `json:"series"`
	Architecture string `json:"architecture"`
	Store        string `json:"store"`
	Gadget       string `json:"gadget"`
	Kernel       string `json:"kernel"`
}

// ModelAssertionHandler is the API method to generate a model assertion
func ModelAssertionHandler(w http.ResponseWriter, r *http.Request) ErrorResponse {

	// Check that we have an authorised API key header
	err := checkAPIKey(r.Header.Get("api-key"))
	if err != nil {
		logMessage("MODEL", "invalid-api-key", "Invalid API key used")
		return ErrorInvalidAPIKey
	}

	defer r.Body.Close()

	// Decode the JSON body
	request := ModelAssertion{}
	err = json.NewDecoder(r.Body).Decode(&request)
	switch {
	// Check we have some data
	case err == io.EOF:
		return ErrorEmptyData
		// Check for parsing errors
	case err != nil:
		return ErrorResponse{false, "error-decode-json", "", err.Error(), http.StatusBadRequest}
	}

	// Validate the model by checking that it exists on the database
	model, err := datastore.Environ.DB.FindModel(request.BrandID, request.Name, r.Header.Get("api-key"))
	if err != nil {
		logMessage("MODEL", "invalid-model", "Cannot find model with the matching brand and model")
		return ErrorInvalidModel
	}

	// Build the model assertion
	assertionHeaders := createModelAssertion(request, model)
	if err != nil {
		logMessage("MODEL", "create-assertion", err.Error())
		return ErrorCreateAssertion
	}

	// Sign the assertion with the snapd assertions module
	signedAssertion, err := datastore.Environ.KeypairDB.SignAssertion(asserts.ModelType, assertionHeaders, []byte(""), model.AuthorityIDModel, model.KeyIDModel, model.SealedKeyModel)
	if err != nil {
		logMessage("MODEL", "signing-assertion", err.Error())
		return ErrorResponse{false, "signing-assertion", "", err.Error(), http.StatusInternalServerError}
	}

	// Return successful JSON response with the signed text
	formatSignResponse(true, "", "", "", signedAssertion, w)
	return ErrorResponse{Success: true}
}

func createModelAssertion(model ModelAssertion, m datastore.Model) map[string]interface{} {

	// Create the model assertion header
	headers := map[string]interface{}{
		"type":              asserts.ModelType.Name,
		"authority-id":      model.BrandID,
		"brand-id":          model.BrandID,
		"series":            model.Series,
		"model":             model.Name,
		"architecture":      model.Architecture,
		"gadget":            model.Gadget,
		"kernel":            model.Kernel,
		"sign-key-sha3-384": m.KeyIDModel,
		"timestamp":         time.Now().Format(time.RFC3339),
	}

	if len(model.Store) == 0 {
		headers["store"] = model.Store
	}

	return headers
}
