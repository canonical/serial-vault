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

// NonceResponse is the JSON response from the API Version method
type NonceResponse struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"message"`
	Nonce        string `json:"nonce"`
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
func SignHandler(w http.ResponseWriter, r *http.Request) {

	// Check that we have an authorised API key header
	err := checkAPIKey(r.Header.Get("api-key"))
	if err != nil {
		logMessage("SIGN", "invalid-api-key", "Invalid API key used")
		w.WriteHeader(http.StatusBadRequest)
		formatSignResponse(false, "error-api-key", "", "Invalid API key used", nil, w)
		return
	}

	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		logMessage("SIGN", "invalid-assertion", "Uninitialized POST data")
		formatSignResponse(false, "error-nil-data", "", "Uninitialized POST data", nil, w)
		return
	}

	// Read the full request body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logMessage("SIGN", "invalid-assertion", err.Error())
		formatSignResponse(false, "error-sign-read", "", err.Error(), nil, w)
		return
	}
	if len(data) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		logMessage("SIGN", "invalid-assertion", "No data supplied for signing")
		formatSignResponse(false, "error-sign-empty", "", "No data supplied for signing", nil, w)
		return
	}

	defer r.Body.Close()

	// Use the snapd assertions module to decode the body and validate
	assertion, err := asserts.Decode(data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logMessage("SIGN", "invalid-assertion", err.Error())
		formatSignResponse(false, "error-decode-assertion", "", err.Error(), nil, w)
		return
	}

	// Check that we have a serial-request assertion (the details will have been validated by Decode call)
	if assertion.Type() != asserts.SerialRequestType {
		w.WriteHeader(http.StatusBadRequest)
		logMessage("SIGN", "invalid-assertion", "The assertion type must be 'serial-request'")
		formatSignResponse(false, "error-decode-assertion", "error-invalid-type", "The assertion type must be 'serial'", nil, w)
		return
	}

	// Verify that the nonce is valid and has not expired
	err = Environ.DB.ValidateDeviceNonce(assertion.HeaderString("request-id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logMessage("SIGN", "invalid-nonce", "Nonce is invalid or expired")
		formatSignResponse(false, "error-decode-assertion", "error-invalid-nonce", "Nonce is invalid or expired", nil, w)
		return
	}

	// Validate the model by checking that it exists on the database
	model, err := Environ.DB.FindModel(assertion.HeaderString("brand-id"), assertion.HeaderString("model"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		logMessage("SIGN", "invalid-model", "Cannot find model with the matching brand, model and revision")
		formatSignResponse(false, "error-model-not-found", "", "Cannot find model with the matching brand, model and revision", nil, w)
		return
	}

	// Check that the model has an active keypair
	if !model.KeyActive {
		w.WriteHeader(http.StatusBadRequest)
		logMessage("SIGN", "invalid-model", "The model is linked with an inactive signing-key")
		formatSignResponse(false, "error-model-not-active", "", "The model is linked with an inactive signing-key", nil, w)
		return
	}

	// Convert the serial-request headers into a serial assertion
	serialAssertion, err := serialRequestToSerial(assertion)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logMessage("SIGN", "convert-assertion", err.Error())
		formatSignResponse(false, "error-signing-assertions", "convert-assertion", "Error converting the serial-request to a serial assertion (hint: check the body)", nil, w)
		return
	}

	// Check that we have not already signed this device (the serial number came from the assertion body)
	signingLog := SigningLog{Make: assertion.HeaderString("brand-id"), Model: assertion.HeaderString("model"), SerialNumber: serialAssertion.HeaderString("serial"), Fingerprint: assertion.SignKeyID()}
	duplicateExists, err := Environ.DB.CheckForDuplicate(signingLog)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logMessage("SIGN", "duplicate-assertion", err.Error())
		formatSignResponse(false, "error-signing-assertions", "", err.Error(), nil, w)
		return
	}
	if duplicateExists {
		w.WriteHeader(http.StatusBadRequest)
		logMessage("SIGN", "duplicate-assertion", "The serial number and/or device-key have already been used to sign a device")
		formatSignResponse(false, "error-signing-assertions", "error-duplicate", "The serial number and/or device-key have already been used to sign a device", nil, w)
		return
	}

	// Sign the assertion with the snapd assertions module
	signedAssertion, err := Environ.KeypairDB.SignAssertion(asserts.SerialType, serialAssertion.Headers(), serialAssertion.Body(), model.AuthorityID, model.KeyID, model.SealedKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logMessage("SIGN", "signing-assertion", err.Error())
		formatSignResponse(false, "error-signing-assertions", "", err.Error(), nil, w)
		return
	}

	// Store the serial number and device-key fingerprint in the database
	err = Environ.DB.CreateSigningLog(signingLog)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logMessage("SIGN", "logging-assertion", err.Error())
		formatSignResponse(false, "error-signing-assertions", "", err.Error(), nil, w)
		return
	}

	// Return successful JSON response with the signed text
	formatSignResponse(true, "", "", "", signedAssertion, w)
}

// serialRequestToSerial converts a serial-request to a serial assertion
func serialRequestToSerial(assertion asserts.Assertion) (asserts.Assertion, error) {
	headers := assertion.Headers()
	headers["type"] = asserts.SerialType.Name
	headers["authority-id"] = headers["brand-id"]
	headers["timestamp"] = time.Now().Format(time.RFC3339)
	delete(headers, "request-id")

	// Decode the body which must be YAML, ignore errors
	body := make(map[string]interface{})
	yaml.Unmarshal(assertion.Body(), &body)

	// Get the extra headers from the body
	headers["serial"] = body["serial"]

	// Create a new serial assertion
	content, signature := assertion.Signature()
	return asserts.Assemble(headers, assertion.Body(), content, signature)

}

// NonceHandler is the API method to generate a nonce
func NonceHandler(w http.ResponseWriter, r *http.Request) {
	// Check that we have an authorised API key header
	err := checkAPIKey(r.Header.Get("api-key"))
	if err != nil {
		logMessage("NONCE", "invalid-api-key", "Invalid API key used")
		w.WriteHeader(http.StatusBadRequest)
		formatSignResponse(false, "error-api-key", "", "Invalid API key used", nil, w)
		return
	}

	nonce, err := Environ.DB.CreateDeviceNonce()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logMessage("NONCE", "generate-nonce", err.Error())
		formatNonceResponse(false, err.Error(), DeviceNonce{}, w)
		return
	}

	// Return successful JSON response with the nonce
	formatNonceResponse(true, "", nonce, w)
}
