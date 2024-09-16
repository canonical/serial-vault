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

package sign

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/CanonicalLtd/serial-vault/service/log"

	"github.com/CanonicalLtd/serial-vault/datastore"
	svlog "github.com/CanonicalLtd/serial-vault/service/log"
	"github.com/CanonicalLtd/serial-vault/service/request"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/snapcore/snapd/asserts"
	"gopkg.in/yaml.v2"
)

// RequestIDResponse is the JSON response from the API Version method
type RequestIDResponse struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"message"`
	RequestID    string `json:"request-id"`
}

// RequestID is the API method to generate a nonce
func RequestID(w http.ResponseWriter, r *http.Request) response.ErrorResponse {
	w.Header().Set("Content-Type", response.JSONHeader)
	// Check that we have an authorised API key header
	_, err := request.CheckModelAPI(r)
	if err != nil {
		svlog.Message("REQUESTID", response.ErrorInvalidAPIKey.Code, response.ErrorInvalidAPIKey.Message)
		return response.ErrorInvalidAPIKey
	}

	err = datastore.Environ.DB.DeleteExpiredDeviceNonces()
	if err != nil {
		svlog.Message("REQUESTID", "delete-expired-nonces", err.Error())
		return response.ErrorGenerateNonce
	}

	nonce, err := datastore.Environ.DB.CreateDeviceNonce()
	if err != nil {
		svlog.Message("REQUESTID", "generate-request-id", err.Error())
		return response.ErrorGenerateNonce
	}

	// Return successful JSON response with the nonce
	formatRequestIDResponse(nonce, w)
	return response.ErrorResponse{Success: true}
}

func parseAssertionStream(r *http.Request) (map[string]asserts.Assertion, response.ErrorResponse) {
	defer r.Body.Close()
	assertions := make(map[string]asserts.Assertion)

	// Use snapd assertion module to decode the assertions in the request stream
	dec := asserts.NewDecoder(r.Body)
	serialRequestAssertion, err := dec.Decode()
	if err == io.EOF {
		svlog.Message("SIGN", "invalid-assertion", response.ErrorEmptyData.Message)
		return nil, response.ErrorEmptyData
	}
	if err != nil {
		svlog.Message("SIGN", "invalid-assertion", err.Error())
		return nil, response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: err.Error(), StatusCode: http.StatusBadRequest}
	}
	// Check that we have a serial-request assertion (the details will have been validated by Decode call)
	if serialRequestAssertion.Type() != asserts.SerialRequestType {
		svlog.Message("SIGN", response.ErrorInvalidType.Code, "The assertion type must be 'serial-request'")
		return nil, response.ErrorInvalidType
	}
	assertions["serial-request"] = serialRequestAssertion

	// Decode the optional model
	modelAssert, err := dec.Decode()
	if err != nil && err != io.EOF {
		svlog.Message("SIGN", "invalid-assertion", err.Error())
		return nil, response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: err.Error(), StatusCode: http.StatusBadRequest}
	}
	if modelAssert != nil {
		if modelAssert.Type() != asserts.ModelType {
			svlog.Message("SIGN", response.ErrorInvalidSecondType.Code, response.ErrorInvalidSecondType.Message)
			return nil, response.ErrorInvalidSecondType
		}
		assertions["model"] = modelAssert
	}

	// Decode the optional serial for remodeling
	serialAssertion, err := dec.Decode()
	if err != nil && err != io.EOF {
		svlog.Message("SIGN", "invalid-assertion", err.Error())
		return nil, response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: err.Error(), StatusCode: http.StatusBadRequest}
	}
	if serialAssertion != nil {
		if serialAssertion.Type() != asserts.SerialType {
			svlog.Message("SIGN", response.ErrorInvalidSecondType.Code, response.ErrorInvalidSecondType.Message)
			return nil, response.ErrorInvalidSecondType
		}
		assertions["serial"] = serialAssertion
	}

	// Stream must be ended now
	_, err = dec.Decode()
	if err != io.EOF {
		if err == nil {
			err = fmt.Errorf("unexpected assertion in the request stream")
		}
		svlog.Message("SIGN", response.ErrorInvalidAssertion.Code, err.Error())
		return nil, response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: err.Error(), StatusCode: http.StatusBadRequest}
	}

	return assertions, response.ErrorResponse{Success: true}
}

func isRemodelingSerialRequest(assertion *asserts.SerialRequest) bool {
	// check original-model, original-brand-id, original-serial
	if assertion.HeaderString("original-model") == "" {
		return false
	}
	if assertion.HeaderString("original-brand-id") == "" {
		return false
	}
	if assertion.HeaderString("original-serial") == "" {
		return false
	}
	return true
}

// Serial is the API method to sign serial assertions from the device
func Serial(w http.ResponseWriter, r *http.Request) response.ErrorResponse {
	// Check that we have an authorised API key header
	apiKey, err := request.CheckModelAPI(r)
	if err != nil {
		svlog.Message("SIGN", response.ErrorInvalidAPIKey.Code, response.ErrorInvalidAPIKey.Message)
		return response.ErrorInvalidAPIKey
	}

	assertions, errResponse := parseAssertionStream(r)
	if !errResponse.Success {
		return errResponse
	}

	serialReq, ok := assertions["serial-request"].(*asserts.SerialRequest)
	if !ok {
		msg := fmt.Sprintf("expected serial-request, got type %q", serialReq.Type().Name)
		svlog.Message("SIGN", response.ErrorInvalidAssertion.Code, msg)
		return response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: msg, StatusCode: http.StatusBadRequest}
	}

	err = asserts.SignatureCheck(serialReq, serialReq.DeviceKey())
	if err != nil {
		msg := fmt.Sprintf("could not validate serial-request self-signature (%s)", err)
		svlog.Message("SIGN", response.ErrorInvalidAssertion.Code, msg)
		return response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: msg, StatusCode: http.StatusBadRequest}
	}
	// Double check the model assertion if present
	modelAssert, ok := assertions["model"]
	if ok {
		if modelAssert.HeaderString("brand-id") != serialReq.HeaderString("brand-id") || modelAssert.HeaderString("model") != serialReq.HeaderString("model") {
			const msg = "Model and serial-request assertion do not match"
			svlog.Message("SIGN", "mismatched-model", msg)
			return response.ErrorResponse{Success: false, Code: "mismatched-model", Message: msg, StatusCode: http.StatusBadRequest}
		}

		// TODO: ideally check the signature of model, need access
		// to the brand public key(s) for models
	}

	if isRemodelingSerialRequest(serialReq) {
		serialAssert := assertions["serial"]
		errResponse := checkRemodelingRequest(serialReq, modelAssert, serialAssert, apiKey)
		if !errResponse.Success {
			return errResponse
		}
	} else {
		// Check the serial assertion
		if _, ok := assertions["serial"]; ok {
			const msg = "unexpected assertion in the request stream"
			svlog.Message("SIGN", response.ErrorInvalidAssertion.Code, msg)
			return response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: msg, StatusCode: http.StatusBadRequest}
		}
	}

	// Verify that the nonce is valid and has not expired
	err = datastore.Environ.DB.ValidateDeviceNonce(serialReq.HeaderString("request-id"))
	if err != nil {
		svlog.Message("SIGN", response.ErrorInvalidNonce.Code, response.ErrorInvalidNonce.Message)
		return response.ErrorInvalidNonce
	}

	// Validate the model by checking that it exists on the database
	model, errResponse := findModel(serialReq.HeaderString("brand-id"), serialReq.HeaderString("model"), serialReq.HeaderString("serial"), apiKey)
	if !errResponse.Success {
		return errResponse
	}

	// Check that the model has an active keypair
	if !model.KeyActive {
		svlog.Message("SIGN", response.ErrorInactiveModel.Code, response.ErrorInactiveModel.Message)
		return response.ErrorInactiveModel
	}

	// Create a basic signing log entry (without the serial number)
	signingLog := datastore.SigningLog{Make: serialReq.HeaderString("brand-id"), Model: serialReq.HeaderString("model"), Fingerprint: serialReq.SignKeyID()}

	// Convert the serial-request headers into a serial assertion
	serialAssertion, err := serialRequestToSerial(serialReq, &signingLog)
	if err != nil {
		svlog.Message("SIGN", response.ErrorCreateAssertion.Code, err.Error())
		return response.ErrorCreateAssertion
	}

	// Sign the assertion with the snapd assertions module
	signedAssertion, err := datastore.Environ.KeypairDB.SignAssertion(asserts.SerialType, serialAssertion.Headers(), serialAssertion.Body(), model.AuthorityID, model.KeyID, model.SealedKey)
	if err != nil {
		svlog.Message("SIGN", "signing-assertion", err.Error())
		return response.ErrorResponse{Success: false, Code: "signing-assertion", Message: err.Error(), StatusCode: http.StatusBadRequest}
	}

	// Store the serial number and device-key fingerprint in the database
	err = datastore.Environ.DB.CreateSigningLog(signingLog)
	if err != nil {
		svlog.Message("SIGN", "logging-assertion", err.Error())
		return response.ErrorResponse{Success: false, Code: "logging-assertion", Message: err.Error(), StatusCode: http.StatusBadRequest}
	}

	// Return successful JSON response with the signed text
	formatSignResponse(signedAssertion, w)
	return response.ErrorResponse{Success: true}
}

func checkRemodelingRequest(serialReq *asserts.SerialRequest, modelAssert, serialAssert asserts.Assertion, apiKey string) response.ErrorResponse {
	originalBrandID := serialReq.HeaderString("original-brand-id")
	originalModel := serialReq.HeaderString("original-model")
	originalSerial := serialReq.HeaderString("original-serial")

	if modelAssert == nil {
		const msg = "Model assertion can't be empty for a remodeling request"
		svlog.Message("SIGN", "invalid-assertion", msg)
		return response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: msg, StatusCode: http.StatusBadRequest}
	}

	// Double check the serial assertion
	if serialAssert == nil {
		const msg = "The current serial assertion can't be empty for a remodeling request"
		svlog.Message("SIGN", "invalid-assertion", msg)
		return response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: msg, StatusCode: http.StatusBadRequest}
	}

	// Validate the original model by checking that it exists on the database
	originalModelAssert, errResponse := findModel(originalBrandID, originalModel, originalSerial, apiKey)
	if !errResponse.Success {
		svlog.Message("SIGN", "invalid-assertion", "original model is not valid")
		return errResponse
	}

	// Validate the new model: it must be defind in the sub-store of the orignal model
	substore, err := datastore.Environ.DB.GetSubstore(originalModelAssert.ID, originalSerial)
	if err != nil {
		svlog.Message("PIVOT", "invalid-substore", "Cannot find sub-store mapping for the model")
		return response.ErrorInvalidSubstore
	}

	// Check if find model maches requested model
	if serialReq.HeaderString("model") != substore.ModelName {
		const msg = "Requested model is invalid"
		svlog.Message("SIGN", "invalid-assertion", msg)
		return response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: msg, StatusCode: http.StatusBadRequest}
	}

	// Check that original-* fields are matching old serial
	if serialAssert.HeaderString("model") != originalModel {
		const msg = "Original model is invalid"
		svlog.Message("SIGN", "invalid-assertion", msg)
		return response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: msg, StatusCode: http.StatusBadRequest}
	}
	if serialAssert.HeaderString("serial") != originalSerial {
		const msg = "Original serial number is invalid"
		svlog.Message("SIGN", "invalid-assertion", msg)
		return response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: msg, StatusCode: http.StatusBadRequest}
	}
	if serialAssert.HeaderString("brand-id") != originalBrandID {
		const msg = "Original brand-id is invalid"
		svlog.Message("SIGN", "invalid-assertion", msg)
		return response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: msg, StatusCode: http.StatusBadRequest}
	}

	// Check that the device key is the same between serial-request and old serial
	if serialAssert.HeaderString("device-key") != serialReq.HeaderString("device-key") {
		const msg = "Device-key is invalid"
		svlog.Message("SIGN", "invalid-assertion", msg)
		return response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: msg, StatusCode: http.StatusBadRequest}
	}

	keyID := substore.FromModel.KeyID
	if keyID != serialAssert.HeaderString("sign-key-sha3-384") {
		msg := fmt.Sprintf("public key id for the model is invalid")
		svlog.Message("SIGN", response.ErrorInvalidAssertion.Code, msg)
		return response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: msg, StatusCode: http.StatusBadRequest}
	}

	oldModelPublicKey, err := datastore.Environ.KeypairDB.PublicKey(keyID)
	if err != nil {
		msg := fmt.Sprintf("could not find public key for the model (%s)", err)
		svlog.Message("SIGN", response.ErrorInvalidAssertion.Code, msg)
		return response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: msg, StatusCode: http.StatusBadRequest}
	}

	err = asserts.SignatureCheck(serialAssert, oldModelPublicKey)
	if err != nil {
		msg := fmt.Sprintf("could not validate serial-request self-signature (%s)", err)
		svlog.Message("SIGN", response.ErrorInvalidAssertion.Code, msg)
		return response.ErrorResponse{Success: false, Code: response.ErrorInvalidAssertion.Code, Message: msg, StatusCode: http.StatusBadRequest}
	}

	return response.ErrorResponse{Success: true}
}

// findModel finds the model by checking that there is an original or pivoted model
func findModel(brandID, modelName, serialNumer, apiKey string) (datastore.Model, response.ErrorResponse) {
	// Assume this is an original (non-pivoted) serial assertion
	// Validate the model by checking that it exists on the database
	model, err := datastore.Environ.DB.FindModel(brandID, modelName, apiKey)
	if err != nil {
		svlog.Message("SIGN", response.ErrorInvalidModel.Code, response.ErrorInvalidModel.Message)
	} else {
		// Found the model, so return it
		return model, response.ErrorResponse{Success: true}
	}

	// Assume that this is a pivoted serial assertion
	// Check for a sub-store model for the pivot
	substore, err := datastore.Environ.DB.GetSubstoreModel(brandID, modelName, serialNumer)
	if err != nil {
		log.Println(err)
		svlog.Message("CHECK", response.ErrorInvalidModelSubstore.Code, response.ErrorInvalidModelSubstore.Message)
		return model, response.ErrorInvalidModelSubstore
	}

	if substore.FromModel.APIKey != apiKey {
		return substore.FromModel, response.ErrorInvalidModelSubstore
	}

	return substore.FromModel, response.ErrorResponse{Success: true}
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
		svlog.Message("SIGN", "create-assertion", response.ErrorEmptySerial.Message)
		return nil, errors.New(response.ErrorEmptySerial.Message)
	}

	// Check that we have not already signed this device, and get the max. revision number for the serial number
	signingLog.SerialNumber = headers["serial"].(string)
	duplicateExists, maxRevision, err := datastore.Environ.DB.CheckForDuplicate(signingLog)
	if err != nil {
		svlog.Message("SIGN", "duplicate-assertion", err.Error())
		return nil, errors.New(response.ErrorDuplicateAssertion.Message)
	}
	if duplicateExists {
		svlog.Message("SIGN", "duplicate-assertion", "The serial number and/or device-key have already been used to sign a device")
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

func formatSignResponse(assertion asserts.Assertion, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", asserts.MediaType)
	w.WriteHeader(http.StatusOK)
	encoder := asserts.NewEncoder(w)
	err := encoder.Encode(assertion)
	if err != nil {
		// Not much we can do if we're here - apart from panic!
		svlog.Message("SIGN", "error-encode-assertion", "Error encoding the assertion.")
		return err
	}

	return nil
}

func formatRequestIDResponse(nonce datastore.DeviceNonce, w http.ResponseWriter) error {
	response := RequestIDResponse{Success: true, RequestID: nonce.Nonce}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		svlog.Message("REQUESTID", "error-form-requestid", err.Error())
		return err
	}
	return nil
}
