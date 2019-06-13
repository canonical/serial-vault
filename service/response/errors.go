// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2018 Canonical Ltd
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

package response

import (
	"fmt"
	"net/http"

	svlog "github.com/CanonicalLtd/serial-vault/service/log"
)

// ErrorResponse is a generic JSON error response structure from an API method
type ErrorResponse struct {
	Success    bool   `json:"success"`
	Code       string `json:"error_code"`
	SubCode    string `json:"error_subcode"`
	Message    string `json:"message"`
	StatusCode int
}

// Standard error messages
var (
	ErrorAuth                      = ErrorResponse{false, "error-auth", "", "Your user does not have permissions for the Signing Authority", http.StatusBadRequest}
	ErrorAuthDisabled              = ErrorResponse{false, "error-auth", "", "This feature is not enabled for this account", http.StatusBadRequest}
	ErrorInvalidID                 = ErrorResponse{false, "invalid-record", "", "Invalid record ID", http.StatusBadRequest}
	ErrorInvalidAPIKey             = ErrorResponse{false, "invalid-api-key", "", "Invalid API key used", http.StatusBadRequest}
	ErrorNilData                   = ErrorResponse{false, "nil-data", "", "Uninitialized POST data", http.StatusBadRequest}
	ErrorInvalidData               = ErrorResponse{false, "invalid-data", "", "Invalid data supplied", http.StatusBadRequest}
	ErrorEmptyData                 = ErrorResponse{false, "empty-data", "", "No data supplied for signing", http.StatusBadRequest}
	ErrorDecodeJSON                = ErrorResponse{false, "error-decode-json", "", "Error decoding JSON", http.StatusBadRequest}
	ErrorInvalidType               = ErrorResponse{false, "invalid-type", "", "The assertion type must be 'serial'", http.StatusBadRequest}
	ErrorInvalidSecondType         = ErrorResponse{false, "invalid-second-type", "", "The 2nd assertion type must be 'model'", http.StatusBadRequest}
	ErrorInvalidNonce              = ErrorResponse{false, "invalid-nonce", "", "Nonce is invalid or expired", http.StatusBadRequest}
	ErrorInvalidSubstore           = ErrorResponse{false, "invalid-substore", "", "Cannot find sub-store mapping for the model", http.StatusBadRequest}
	ErrorInactiveModel             = ErrorResponse{false, "invalid-model", "", "The model is linked with an inactive signing-key", http.StatusBadRequest}
	ErrorInvalidAccount            = ErrorResponse{false, "invalid-account", "", "The account cannot be found", http.StatusBadRequest}
	ErrorInvalidAssertion          = ErrorResponse{false, "invalid-assertion", "", "The assertion is invalid", http.StatusBadRequest}
	ErrorInvalidKeypair            = ErrorResponse{false, "invalid-keypair", "", "The keypair is invalid", http.StatusBadRequest}
	ErrorFetchKeypairs             = ErrorResponse{false, "fetch-keypairs", "", "Error fetching the signing-keys", http.StatusBadRequest}
	ErrorFetchKeypair              = ErrorResponse{false, "fetch-keypair", "", "Error fetching the signing-key", http.StatusBadRequest}
	ErrorStoreKeypair              = ErrorResponse{false, "store-keypair", "", "Error string the signing-key", http.StatusBadRequest}
	ErrorEmptySerial               = ErrorResponse{false, "create-assertion", "", "The serial number is missing from both the header and body", http.StatusBadRequest}
	ErrorCreateAssertion           = ErrorResponse{false, "create-assertion", "", "Error converting the serial-request to a serial assertion", http.StatusBadRequest}
	ErrorDecodeAssertion           = ErrorResponse{false, "decode-assertion", "", "Error decoding the assertion", http.StatusBadRequest}
	ErrorCheckAssertion            = ErrorResponse{false, "duplicate-assertion", "", "Error checking the serial-request. Please try again later", http.StatusBadRequest}
	ErrorCreateModelAssertion      = ErrorResponse{false, "create-assertion", "", "Error with the model assertion headers", http.StatusBadRequest}
	ErrorCreateSystemUserAssertion = ErrorResponse{false, "create-assertion", "", "Error with the system-user assertion", http.StatusBadRequest}
	ErrorDuplicateAssertion        = ErrorResponse{false, "duplicate-assertion", "", "The serial number and/or device-key have already been used to sign a device", http.StatusBadRequest}
	ErrorAccountAssertion          = ErrorResponse{false, "account-assertion", "", "Error retrieving the account assertion from the database", http.StatusBadRequest}
	ErrorSignAssertion             = ErrorResponse{false, "signing-assertion", "", "Error signing the assertion", http.StatusBadRequest}
	ErrorGenerateNonce             = ErrorResponse{false, "generate-nonce", "", "Error generating a nonce. Please try again later", http.StatusBadRequest}
)

// ErrorInvalidModel returns error message about wrong model
func ErrorInvalidModel(from, modelName, brand, apiKey string) ErrorResponse {
	msg := fmt.Sprintf("Cannot find model %s with the matching brand %s and apiKey %s", modelName, brand, apiKey)
	svlog.Message(from, "invalid-model", msg)
	return ErrorResponse{false, "invalid-model", "", msg, http.StatusBadRequest}
}

// ErrorInvalidModelID returns InvalidModelID error message
func ErrorInvalidModelID(from string, ModelID int) ErrorResponse {
	msg := fmt.Sprintf("Cannot find model with the selected ID %d", ModelID)
	svlog.Message(from, "invalid-model", msg)

	return ErrorResponse{false, "invalid-model", "", msg, http.StatusBadRequest}
}

// ErrorInvalidModelSubstore returns invalid model substore error
func ErrorInvalidModelSubstore(from, brandID, modelName, apiKey, serialNumer string) ErrorResponse {
	msg := fmt.Sprintf("Cannot find a matching model or sub-store model %s with the matching brand %s apiKey %s and serialNumer %s",
		modelName, brandID, apiKey, serialNumer)
	svlog.Message(from, "invalid-model", msg)
	return ErrorResponse{false, "invalid-model", "", msg, http.StatusBadRequest}
}
