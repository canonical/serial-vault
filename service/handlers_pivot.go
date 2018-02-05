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
	"io"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/snapcore/snapd/asserts"
)

// PivotResponse is the JSON response from the API Sub-Stores method
type PivotResponse struct {
	Success      bool               `json:"success"`
	ErrorMessage string             `json:"message"`
	Pivot        datastore.Substore `json:"pivot"`
}

// PivotModelHandler is the API method to determine the pivot details of a model
func PivotModelHandler(w http.ResponseWriter, r *http.Request) ErrorResponse {

	assertion, errResponse := parseSerialAssertion(r)
	if !errResponse.Success {
		return errResponse
	}

	substore, errResponse := findModelPivot(assertion, r.Header.Get("api-key"))
	if !errResponse.Success {
		return errResponse
	}

	// Return the model pivot details (store and model name)
	formatPivotResponse(true, "", substore, w)
	return ErrorResponse{Success: true}
}

// PivotModelAssertionHandler is the API method to get the model assertions for a pivoted model
// The serial assertion of the original model is supplied, and the assertions for the pivoted model are returned
func PivotModelAssertionHandler(w http.ResponseWriter, r *http.Request) ErrorResponse {
	assertion, errResponse := parseSerialAssertion(r)
	if !errResponse.Success {
		return errResponse
	}

	substore, errResponse := findModelPivot(assertion, r.Header.Get("api-key"))
	if !errResponse.Success {
		return errResponse
	}

	assertions := []asserts.Assertion{}

	// Build the model assertion headers for the original model
	assertionHeaders, keypair, err := createModelAssertionHeaders(substore.FromModel)
	if err != nil {
		logMessage("PIVOT", "create-assertion", err.Error())
		return ErrorCreateModelAssertion
	}

	// Override the model assertion headers with the sub-store details
	assertionHeaders["model"] = substore.ToModel.Name
	assertionHeaders["store"] = substore.Store

	// Sign the assertion with the snapd assertions module
	signedAssertion, err := datastore.Environ.KeypairDB.SignAssertion(asserts.ModelType, assertionHeaders, []byte(""), substore.FromModel.BrandID, keypair.KeyID, keypair.SealedKey)
	if err != nil {
		logMessage("MODEL", "signing-assertion", err.Error())
		return ErrorResponse{false, "signing-assertion", "", err.Error(), http.StatusBadRequest}
	}

	// Add the account assertion to the assertions list
	fetchAssertionFromStore(&assertions, asserts.AccountType, []string{substore.FromModel.BrandID})

	// Add the account-key assertion to the assertions list
	fetchAssertionFromStore(&assertions, asserts.AccountKeyType, []string{keypair.KeyID})

	// Add the model assertion after the account and account-key assertions
	assertions = append(assertions, signedAssertion)

	// Return successful response with the signed assertions
	formatAssertionResponse(true, "", "", "", assertions, w)
	return ErrorResponse{Success: true}

}

func parseSerialAssertion(r *http.Request) (asserts.Assertion, ErrorResponse) {
	// Check that we have an authorised API key header
	err := checkAPIKey(r.Header.Get("api-key"))
	if err != nil {
		logMessage("PIVOT", "invalid-api-key", "Invalid API key used")
		return nil, ErrorInvalidAPIKey
	}

	defer r.Body.Close()

	// Get the serial assertion from the body
	dec := asserts.NewDecoder(r.Body)
	assertion, err := dec.Decode()
	if err == io.EOF {
		logMessage("PIVOT", "invalid-assertion", "No data supplied for pivot")
		return nil, ErrorEmptyData
	}
	if err != nil {
		logMessage("PIVOT", "invalid-assertion", err.Error())
		return nil, ErrorResponse{false, "decode-assertion", "", err.Error(), http.StatusBadRequest}
	}

	// Check that we have a serial assertion (the details will have been validated by Decode call)
	if assertion.Type() != asserts.SerialType {
		logMessage("PIVOT", "invalid-type", "The assertion type must be 'serial'")
		return nil, ErrorInvalidType
	}

	return assertion, ErrorResponse{Success: true}
}

func findModelPivot(assertion asserts.Assertion, apiKey string) (datastore.Substore, ErrorResponse) {
	// Validate the model by checking that it exists on the database
	model, err := datastore.Environ.DB.FindModel(assertion.HeaderString("brand-id"), assertion.HeaderString("model"), apiKey)
	if err != nil {
		logMessage("PIVOT", "invalid-model", "Cannot find model with the matching brand and model")
		return datastore.Substore{}, ErrorInvalidModel
	}

	// Check for a sub-store model for the pivot
	substore, err := datastore.Environ.DB.GetSubstore(model.ID, assertion.HeaderString("serial"))
	if err != nil {
		logMessage("PIVOT", "invalid-substore", "Cannot find sub-store mapping for the model")
		return substore, ErrorInvalidSubstore
	}

	return substore, ErrorResponse{Success: true}
}
