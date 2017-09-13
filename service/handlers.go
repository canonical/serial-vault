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
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/gorilla/csrf"

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
	Success      bool                `json:"success"`
	ErrorCode    string              `json:"error_code"`
	ErrorSubcode string              `json:"error_subcode"`
	ErrorMessage string              `json:"message"`
	Keypairs     []datastore.Keypair `json:"keypairs"`
}

// TokenResponse is the JSON response from the API Version method
type TokenResponse struct {
	EnableUserAuth bool `json:"enableUserAuth"`
}

// VersionHandler is the API method to return the version of the service
func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	response := VersionResponse{Version: datastore.Environ.Config.Version}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		message := fmt.Sprintf("Error encoding the version response: %v", err)
		logMessage("VERSION", "get-version", message)
	}
}

// TokenHandler returns CSRF protection new token in a X-CSRF-Token response header
// This method is also used by the /authtoken endpoint to return the JWT. The method
// indicates to the UI whether OpenID user auth is enabled
func TokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	// Check the JWT and return it in the authorization header, if valid
	JWTCheck(w, r)

	response := TokenResponse{EnableUserAuth: datastore.Environ.Config.EnableUserAuth}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		message := fmt.Sprintf("Error encoding the token response: %v", err)
		logMessage("TOKEN", "get-token", message)
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

	defer r.Body.Close()

	// Use snapd assertion module to decode the assertions in the request stream
	dec := asserts.NewDecoder(r.Body)
	assertion, err := dec.Decode()
	if err == io.EOF {
		logMessage("SIGN", "invalid-assertion", "No data supplied for signing")
		return ErrorEmptyData
	}
	if err != nil {
		logMessage("SIGN", "invalid-assertion", err.Error())
		return ErrorResponse{false, "decode-assertion", "", err.Error(), http.StatusBadRequest}
	}

	// Decode the optional model
	modelAssert, err := dec.Decode()
	if err != nil && err != io.EOF {
		logMessage("SIGN", "invalid-assertion", err.Error())
		return ErrorResponse{false, "decode-assertion", "", err.Error(), http.StatusBadRequest}
	}

	// Stream must be ended now
	_, err = dec.Decode()
	if err != io.EOF {
		if err == nil {
			err = fmt.Errorf("unexpected assertion in the request stream")
		}
		logMessage("SIGN", "invalid-assertion", err.Error())
		return ErrorResponse{false, "decode-assertion", "", err.Error(), http.StatusBadRequest}
	}

	// Check that we have a serial-request assertion (the details will have been validated by Decode call)
	if assertion.Type() != asserts.SerialRequestType {
		logMessage("SIGN", "invalid-type", "The assertion type must be 'serial-request'")
		return ErrorInvalidType
	}

	// Double check the model assertion if present
	if modelAssert != nil {
		if modelAssert.Type() != asserts.ModelType {
			logMessage("SIGN", "invalid-second-type", "The 2nd assertion type must be 'model'")
			return ErrorInvalidSecondType
		}
		if modelAssert.HeaderString("brand-id") != assertion.HeaderString("brand-id") || modelAssert.HeaderString("model") != assertion.HeaderString("model") {
			const msg = "Model and serial-request assertion do not match"
			logMessage("SIGN", "mismatched-model", msg)
			return ErrorResponse{false, "mismatched-model", "", msg, http.StatusBadRequest}
		}

		// TODO: ideally check the signature of model, need access
		// to the brand public key(s) for models
	}

	// Verify that the nonce is valid and has not expired
	err = datastore.Environ.DB.ValidateDeviceNonce(assertion.HeaderString("request-id"))
	if err != nil {
		logMessage("SIGN", "invalid-nonce", "Nonce is invalid or expired")
		return ErrorInvalidNonce
	}

	// Validate the model by checking that it exists on the database
	model, err := datastore.Environ.DB.FindModel(assertion.HeaderString("brand-id"), assertion.HeaderString("model"), r.Header.Get("api-key"))
	if err != nil {
		logMessage("SIGN", "invalid-model", "Cannot find model with the matching brand and model")
		return ErrorInvalidModel
	}

	// Check that the model has an active keypair
	if !model.KeyActive {
		logMessage("SIGN", "invalid-model", "The model is linked with an inactive signing-key")
		return ErrorInactiveModel
	}

	// Create a basic signing log entry (without the serial number)
	signingLog := datastore.SigningLog{Make: assertion.HeaderString("brand-id"), Model: assertion.HeaderString("model"), Fingerprint: assertion.SignKeyID()}

	// Convert the serial-request headers into a serial assertion
	serialAssertion, err := serialRequestToSerial(assertion, &signingLog)
	if err != nil {
		logMessage("SIGN", "create-assertion", err.Error())
		return ErrorCreateAssertion
	}

	// Sign the assertion with the snapd assertions module
	signedAssertion, err := datastore.Environ.KeypairDB.SignAssertion(asserts.SerialType, serialAssertion.Headers(), serialAssertion.Body(), model.AuthorityID, model.KeyID, model.SealedKey)
	if err != nil {
		logMessage("SIGN", "signing-assertion", err.Error())
		return ErrorResponse{false, "signing-assertion", "", err.Error(), http.StatusInternalServerError}
	}

	// Store the serial number and device-key fingerprint in the database
	err = datastore.Environ.DB.CreateSigningLog(signingLog)
	if err != nil {
		logMessage("SIGN", "logging-assertion", err.Error())
		return ErrorResponse{false, "logging-assertion", "", err.Error(), http.StatusInternalServerError}
	}

	// Return successful JSON response with the signed text
	formatSignResponse(true, "", "", "", signedAssertion, w)
	return ErrorResponse{Success: true}
}

// serialRequestToSerial converts a serial-request to a serial assertion
func serialRequestToSerial(assertion asserts.Assertion, signingLog *datastore.SigningLog) (asserts.Assertion, error) {

	// Create the serial assertion header from the serial-request headers
	serialHeaders := assertion.Headers()
	headers := map[string]interface{}{
		"type":                asserts.SerialType.Name,
		"authority-id":        serialHeaders["brand-id"],
		"brand-id":            serialHeaders["brand-id"],
		"serial":              serialHeaders["serial"],
		"device-key":          serialHeaders["device-key"],
		"sign-key-sha3-384":   serialHeaders["sign-key-sha3-384"],
		"device-key-sha3-384": serialHeaders["sign-key-sha3-384"],
		"model":               serialHeaders["model"],
		"timestamp":           time.Now().Format(time.RFC3339),
	}

	// Get the serial-number from the header, but fallback to the body if it is not there
	if headers["serial"] == nil || headers["serial"].(string) == "" {
		// Decode the body which must be YAML, ignore errors
		body := make(map[string]interface{})
		yaml.Unmarshal(assertion.Body(), &body)

		// Get the extra headers from the body
		headers["serial"] = body["serial"]
	}

	// Check that we have a serial
	if headers["serial"] == nil {
		logMessage("SIGN", "create-assertion", ErrorEmptySerial.Message)
		return nil, errors.New(ErrorEmptySerial.Message)
	}

	// Check that we have not already signed this device, and get the max. revision number for the serial number
	signingLog.SerialNumber = headers["serial"].(string)
	duplicateExists, maxRevision, err := datastore.Environ.DB.CheckForDuplicate(signingLog)
	if err != nil {
		logMessage("SIGN", "duplicate-assertion", err.Error())
		return nil, errors.New(ErrorDuplicateAssertion.Message)
	}
	if duplicateExists {
		logMessage("SIGN", "duplicate-assertion", "The serial number and/or device-key have already been used to sign a device")
	}

	// Set the revision number, incrementing the previously used one
	signingLog.Revision = maxRevision + 1
	headers["revision"] = fmt.Sprintf("%d", signingLog.Revision)

	// If we have a body, set the body length
	if len(assertion.Body()) > 0 {
		headers["body-length"] = serialHeaders["body-length"]
	}

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

	err = datastore.Environ.DB.DeleteExpiredDeviceNonces()
	if err != nil {
		logMessage("REQUESTID", "delete-expired-nonces", err.Error())
		return ErrorGenerateNonce
	}

	nonce, err := datastore.Environ.DB.CreateDeviceNonce()
	if err != nil {
		logMessage("REQUESTID", "generate-request-id", err.Error())
		return ErrorGenerateNonce
	}

	// Return successful JSON response with the nonce
	formatRequestIDResponse(true, "", nonce, w)
	return ErrorResponse{Success: true}
}
