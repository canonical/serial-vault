// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
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

package pivot

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/CanonicalLtd/serial-vault/service/log"

	"github.com/CanonicalLtd/serial-vault/account"
	"github.com/CanonicalLtd/serial-vault/datastore"
	svlog "github.com/CanonicalLtd/serial-vault/service/log"
	"github.com/CanonicalLtd/serial-vault/service/request"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/snapcore/snapd/asserts"
)

// Response is the JSON response from the API Sub-Stores method
type Response struct {
	Success      bool               `json:"success"`
	ErrorMessage string             `json:"message"`
	Pivot        datastore.Substore `json:"pivot"`
}

// Model is the API method to determine the pivot details of a model
func Model(w http.ResponseWriter, r *http.Request) response.ErrorResponse {

	assertion, errResponse := parseSerialAssertion(r)
	if !errResponse.Success {
		return errResponse
	}

	substore, errResponse := findModelPivot(assertion.HeaderString("brand-id"), assertion.HeaderString("model"), assertion.HeaderString("serial"), r.Header.Get("api-key"))
	if !errResponse.Success {
		return errResponse
	}

	// Return the model pivot details (store and model name)
	formatPivotResponse(true, "", substore, w)
	return response.ErrorResponse{Success: true}
}

// ModelAssertion is the API method to get the model assertions for a pivoted model
// The serial assertion of the original model is supplied, and the assertions for the pivoted model are returned
func ModelAssertion(w http.ResponseWriter, r *http.Request) response.ErrorResponse {
	assertion, errResponse := parseSerialAssertion(r)
	if !errResponse.Success {
		return errResponse
	}

	// Check that the reseller functionality is enabled for the brand
	acc, err := datastore.Environ.DB.GetAccount(assertion.HeaderString("brand-id"))
	if err != nil {
		return response.ErrorResponse{Success: false, Code: "error-account", Message: err.Error(), StatusCode: http.StatusBadRequest}
	}
	if !acc.ResellerAPI {
		return response.ErrorResponse{Success: false, Code: "error-auth", Message: "This feature is not enabled for this account", StatusCode: http.StatusBadRequest}
	}

	substore, errResponse := findModelPivot(assertion.HeaderString("brand-id"), assertion.HeaderString("model"), assertion.HeaderString("serial"), r.Header.Get("api-key"))
	if !errResponse.Success {
		return errResponse
	}

	assertions := []asserts.Assertion{}

	// Build the model assertion headers for the original model
	assertionHeaders, keypair, err := datastore.ModelAssertionHeadersForModel(substore.FromModel)
	if err != nil {
		svlog.Message("PIVOT", "create-assertion", err.Error())
		return response.ErrorCreateModelAssertion
	}

	// Override the model assertion headers with the sub-store details
	assertionHeaders["model"] = substore.ModelName
	assertionHeaders["store"] = substore.Store

	// Sign the assertion with the snapd assertions module
	signedAssertion, err := datastore.Environ.KeypairDB.SignAssertion(asserts.ModelType, assertionHeaders, []byte(""), substore.FromModel.BrandID, keypair.KeyID, keypair.SealedKey)
	if err != nil {
		svlog.Message("PIVOT", "signing-assertion", err.Error())
		return response.ErrorResponse{Success: false, Code: "signing-assertion", Message: err.Error(), StatusCode: http.StatusBadRequest}
	}

	// Add the account assertion to the assertions list
	fetchAssertionFromStore(&assertions, asserts.AccountType, []string{substore.FromModel.BrandID})

	// Add the account-key assertion to the assertions list
	fetchAssertionFromStore(&assertions, asserts.AccountKeyType, []string{keypair.KeyID})

	// Add the model assertion after the account and account-key assertions
	assertions = append(assertions, signedAssertion)

	// Return successful response with the signed assertions
	formatAssertionResponse(assertions, w)
	return response.ErrorResponse{Success: true}

}

// SerialAssertion is the API method to get the serial assertions for a pivoted model
// The serial assertion of the original model is supplied, and the assertions for the pivoted model are returned
func SerialAssertion(w http.ResponseWriter, r *http.Request) response.ErrorResponse {
	assertion, errResponse := parseSerialAssertion(r)
	if !errResponse.Success {
		return errResponse
	}

	// Check that the reseller functionality is enabled for the brand
	acc, err := datastore.Environ.DB.GetAccount(assertion.HeaderString("brand-id"))
	if err != nil {
		return response.ErrorResponse{Success: false, Code: "error-account", Message: err.Error(), StatusCode: http.StatusBadRequest}
	}
	if !acc.ResellerAPI {
		return response.ErrorResponse{Success: false, Code: "error-auth", Message: "This feature is not enabled for this account", StatusCode: http.StatusBadRequest}
	}

	substore, errResponse := findModelPivot(assertion.HeaderString("brand-id"), assertion.HeaderString("model"), assertion.HeaderString("serial"), r.Header.Get("api-key"))
	if !errResponse.Success {
		return errResponse
	}

	assertions := []asserts.Assertion{}

	// Build the serial assertion headers for the original model
	// Override the model assertion headers with the sub-store details
	assertionHeaders := assertion.Headers()
	assertionHeaders["model"] = substore.ModelName
	assertionHeaders["timestamp"] = time.Now().Format(time.RFC3339)

	// Sign the assertion with the snapd assertions module
	signedAssertion, err := datastore.Environ.KeypairDB.SignAssertion(asserts.SerialType, assertionHeaders, assertion.Body(), substore.FromModel.BrandID, substore.FromModel.KeyID, substore.FromModel.SealedKey)
	if err != nil {
		svlog.Message("PIVOT", "signing-assertion", err.Error())
		return response.ErrorResponse{Success: false, Code: "signing-assertion", Message: err.Error(), StatusCode: http.StatusBadRequest}
	}

	// Add the account assertion to the assertions list
	fetchAssertionFromStore(&assertions, asserts.AccountType, []string{substore.FromModel.BrandID})

	// Add the account-key assertion to the assertions list
	fetchAssertionFromStore(&assertions, asserts.AccountKeyType, []string{substore.FromModel.KeyID})

	// Add the model assertion after the account and account-key assertions
	assertions = append(assertions, signedAssertion)

	// Return successful response with the signed assertions
	formatAssertionResponse(assertions, w)
	return response.ErrorResponse{Success: true}

}

func parseSerialAssertion(r *http.Request) (asserts.Assertion, response.ErrorResponse) {
	// Check that we have an authorised API key header
	_, err := request.CheckModelAPI(r)
	if err != nil {
		svlog.Message("PIVOT", "invalid-api-key", "Invalid API key used")
		return nil, response.ErrorInvalidAPIKey
	}

	defer r.Body.Close()

	// Get the serial assertion from the body
	dec := asserts.NewDecoder(r.Body)
	assertion, err := dec.Decode()
	if err == io.EOF {
		svlog.Message("PIVOT", "invalid-assertion", "No data supplied for pivot")
		return nil, response.ErrorEmptyData
	}
	if err != nil {
		svlog.Message("PIVOT", "invalid-assertion", err.Error())
		return nil, response.ErrorResponse{Success: false, Code: "decode-assertion", Message: err.Error(), StatusCode: http.StatusBadRequest}
	}

	// Check that we have a serial assertion (the details will have been validated by Decode call)
	if assertion.Type() != asserts.SerialType {
		svlog.Message("PIVOT", "invalid-type", "The assertion type must be 'serial'")
		return nil, response.ErrorInvalidType
	}

	return assertion, response.ErrorResponse{Success: true}
}

func findModelPivot(brand, modelName, serial, apiKey string) (datastore.Substore, response.ErrorResponse) {
	// Validate the model by checking that it exists on the database
	model, err := datastore.Environ.DB.FindModel(brand, modelName, apiKey)
	if err != nil {
		svlog.Message("PIVOT", "invalid-model", "Cannot find model with the matching brand and model")
		return datastore.Substore{}, response.ErrorInvalidModel
	}

	// Check for a sub-store model for the pivot
	substore, err := datastore.Environ.DB.GetSubstore(model.ID, serial)
	if err != nil {
		log.Println(err)
		svlog.Message("PIVOT", "invalid-substore", "Cannot find sub-store mapping for the model")
		return substore, response.ErrorInvalidSubstore
	}

	return substore, response.ErrorResponse{Success: true}
}

func formatPivotResponse(success bool, message string, store datastore.Substore, w http.ResponseWriter) error {
	response := Response{Success: success, ErrorMessage: message, Pivot: store}
	return jsonEncode(response, w)
}

func jsonEncode(response interface{}, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the JSON response.")
		return err
	}
	return nil
}

func fetchAssertionFromStore(assertions *[]asserts.Assertion, modelType *asserts.AssertionType, headers []string) {
	assertion, err := account.FetchAssertionFromStore(modelType, headers)
	if err != nil {
		svlog.Message("MODEL", "assertion", err.Error())
	} else {
		*assertions = append(*assertions, assertion)
	}
}

func formatAssertionResponse(assertions []asserts.Assertion, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", asserts.MediaType)
	w.WriteHeader(http.StatusOK)
	encoder := asserts.NewEncoder(w)

	for _, assert := range assertions {
		err := encoder.Encode(assert)
		if err != nil {
			// Not much we can do if we're here - apart from panic!
			svlog.Message("MODEL", "assertion", "Error encoding the assertions.")
			return err
		}
	}

	return nil
}
