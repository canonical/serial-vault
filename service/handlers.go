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
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/snapcore/snapd/asserts"
)

// VersionResponse is the JSON response from the API Version method
type VersionResponse struct {
	Version string `json:"version"`
}

// RequestIDResponse is the JSON response from the API Version method
type RequestIDResponse struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"message"`
	RequestID    string `json:"request-id"`
}

// SignResponse is the JSON response from the API Sign method
type SignResponse struct {
	Success      bool   `json:"success"`
	ErrorCode    string `json:"error_code"`
	ErrorSubcode string `json:"error_subcode"`
	ErrorMessage string `json:"message"`
}

// KeypairsResponse is the JSON response from the API Keypairs method
type KeypairsResponse struct {
	Success      bool      `json:"success"`
	ErrorCode    string    `json:"error_code"`
	ErrorSubcode string    `json:"error_subcode"`
	ErrorMessage string    `json:"message"`
	Keypairs     []Keypair `json:"keypairs"`
}

// VersionHandler is the API method to return the version of the service
func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	response := VersionResponse{Version: Environ.Config.Version}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		message := fmt.Sprintf("Error encoding the version response: %v", err)
		logMessage("VERSION", "get-version", message)
	}
}

// SignHandler is the API method to sign assertions from the device
func SignHandler(w http.ResponseWriter, r *http.Request) ErrorResponse {

	// Check that we have an authorised API key header
	err := checkAPIKey(r.Header.Get("api-key"))
	if err != nil {
		logMessage("SIGN", "invalid-api-key", "Invalid API key used")
		return ErrorInvalidAPIKey
	}

	if r.Body == nil {
		logMessage("SIGN", "invalid-assertion", "Uninitialized POST data")
		return ErrorNilData
	}

	// Read the full request body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logMessage("SIGN", "invalid-assertion", err.Error())
		return ErrorResponse{false, "error-sign-read", "", err.Error(), http.StatusBadRequest}
	}
	if len(data) == 0 {
		logMessage("SIGN", "invalid-assertion", "No data supplied for signing")
		return ErrorEmptyData
	}

	defer r.Body.Close()

	// Use the snapd assertions module to decode the body and validate
	assertion, err := asserts.Decode(data)
	if err != nil {
		logMessage("SIGN", "invalid-assertion", err.Error())
		return ErrorResponse{false, "decode-assertion", "", err.Error(), http.StatusBadRequest}
	}

	// Check that we have a serial-request assertion (the details will have been validated by Decode call)
	if assertion.Type() != asserts.SerialRequestType {
		logMessage("SIGN", "invalid-type", "The assertion type must be 'serial-request'")
		return ErrorInvalidType
	}

	// Verify that the nonce is valid and has not expired
	err = Environ.DB.ValidateDeviceNonce(assertion.HeaderString("request-id"))
	if err != nil {
		logMessage("SIGN", "invalid-nonce", "Nonce is invalid or expired")
		return ErrorInvalidNonce
	}

	// Validate the model by checking that it exists on the database
	model, err := Environ.DB.FindModel(assertion.HeaderString("brand-id"), assertion.HeaderString("model"))
	if err != nil {
		logMessage("SIGN", "invalid-model", "Cannot find model with the matching brand, model and revision")
		return ErrorInvalidModel
	}

	// Check that the model has an active keypair
	if !model.KeyActive {
		logMessage("SIGN", "invalid-model", "The model is linked with an inactive signing-key")
		return ErrorInactiveModel
	}

	// Convert the serial-request headers into a serial assertion
	serialAssertion, err := serialRequestToSerial(assertion)
	if err != nil {
		logMessage("SIGN", "create-assertion", err.Error())
		return ErrorCreateAssertion
	}

	// Check that we have not already signed this device (the serial number came from the assertion body)
	signingLog := SigningLog{Make: assertion.HeaderString("brand-id"), Model: assertion.HeaderString("model"), SerialNumber: serialAssertion.HeaderString("serial"), Fingerprint: assertion.SignKeyID()}
	duplicateExists, err := Environ.DB.CheckForDuplicate(signingLog)
	if err != nil {
		logMessage("SIGN", "duplicate-assertion", err.Error())
		return ErrorDuplicateAssertion
	}
	if duplicateExists {
		logMessage("SIGN", "duplicate-assertion", "The serial number and/or device-key have already been used to sign a device")
		return ErrorDuplicateAssertion
	}

	// Sign the assertion with the snapd assertions module
	signedAssertion, err := Environ.KeypairDB.SignAssertion(asserts.SerialType, serialAssertion.Headers(), serialAssertion.Body(), model.AuthorityID, model.KeyID, model.SealedKey)
	if err != nil {
		logMessage("SIGN", "signing-assertion", err.Error())
		return ErrorResponse{false, "signing-assertion", "", err.Error(), http.StatusInternalServerError}
	}

	// Store the serial number and device-key fingerprint in the database
	err = Environ.DB.CreateSigningLog(signingLog)
	if err != nil {
		logMessage("SIGN", "logging-assertion", err.Error())
		return ErrorResponse{false, "logging-assertion", "", err.Error(), http.StatusInternalServerError}
	}

	// Return successful JSON response with the signed text
	formatSignResponse(true, "", "", "", signedAssertion, w)
	return ErrorResponse{Success: true}
}

// serialRequestToSerial converts a serial-request to a serial assertion
func serialRequestToSerial(assertion asserts.Assertion) (asserts.Assertion, error) {

	// Create the serial assertion header from the serial-request headers
	serialHeaders := assertion.Headers()
	headers := map[string]interface{}{
		"type":                asserts.SerialType.Name,
		"authority-id":        serialHeaders["brand-id"],
		"brand-id":            serialHeaders["brand-id"],
		"device-key":          serialHeaders["device-key"],
		"sign-key-sha3-384":   serialHeaders["sign-key-sha3-384"],
		"device-key-sha3-384": serialHeaders["sign-key-sha3-384"],
		"model":               serialHeaders["model"],
		"timestamp":           time.Now().Format(time.RFC3339),
		"body-length":         serialHeaders["body-length"],
	}

	// Decode the body which must be YAML, ignore errors
	body := make(map[string]interface{})
	yaml.Unmarshal(assertion.Body(), &body)

	// Get the extra headers from the body
	headers["serial"] = body["serial"]

	// Create a new serial assertion
	content, signature := assertion.Signature()
	return asserts.Assemble(headers, assertion.Body(), content, signature)

}

// RequestIDHandler is the API method to generate a nonce
func RequestIDHandler(w http.ResponseWriter, r *http.Request) ErrorResponse {
	// Check that we have an authorised API key header
	err := checkAPIKey(r.Header.Get("api-key"))
	if err != nil {
		logMessage("REQUESTID", "invalid-api-key", "Invalid API key used")
		return ErrorInvalidAPIKey
	}

	nonce, err := Environ.DB.CreateDeviceNonce()
	if err != nil {
		logMessage("REQUESTID", "generate-request-id", err.Error())
		return ErrorGenerateNonce
	}

	// Return successful JSON response with the nonce
	formatRequestIDResponse(true, "", nonce, w)
	return ErrorResponse{Success: true}
}
