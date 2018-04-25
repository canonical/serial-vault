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

package response

import "net/http"

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
	ErrorAuth                      = ErrorResponse{false, "error-auth", "", "Authorization error", http.StatusBadRequest}
	ErrorAuthDisabled              = ErrorResponse{false, "error-auth", "", "This feature is not enabled for this account", http.StatusBadRequest}
	ErrorInvalidAPIKey             = ErrorResponse{false, "invalid-api-key", "", "Invalid API key used", http.StatusBadRequest}
	ErrorNilData                   = ErrorResponse{false, "nil-data", "", "Uninitialized POST data", http.StatusBadRequest}
	ErrorEmptyData                 = ErrorResponse{false, "empty-data", "", "No data supplied for signing", http.StatusBadRequest}
	ErrorDecodeJSON                = ErrorResponse{false, "error-decode-json", "", "Error decoding JSON", http.StatusBadRequest}
	ErrorInvalidType               = ErrorResponse{false, "invalid-type", "", "The assertion type must be 'serial'", http.StatusBadRequest}
	ErrorInvalidSecondType         = ErrorResponse{false, "invalid-second-type", "", "The 2nd assertion type must be 'model'", http.StatusBadRequest}
	ErrorInvalidNonce              = ErrorResponse{false, "invalid-nonce", "", "Nonce is invalid or expired", http.StatusBadRequest}
	ErrorInvalidModel              = ErrorResponse{false, "invalid-model", "", "Cannot find model with the matching brand and model", http.StatusBadRequest}
	ErrorInvalidModelID            = ErrorResponse{false, "invalid-model", "", "Cannot find model with the selected ID", http.StatusBadRequest}
	ErrorInvalidSubstore           = ErrorResponse{false, "invalid-substore", "", "Cannot find sub-store mapping for the model", http.StatusBadRequest}
	ErrorInactiveModel             = ErrorResponse{false, "invalid-model", "", "The model is linked with an inactive signing-key", http.StatusBadRequest}
	ErrorInvalidAccount            = ErrorResponse{false, "invalid-account", "", "The account cannot be found", http.StatusBadRequest}
	ErrorInvalidAssertion          = ErrorResponse{false, "invalid-assertion", "", "The assertion is invalid", http.StatusBadRequest}
	ErrorEmptySerial               = ErrorResponse{false, "create-assertion", "", "The serial number is missing from both the header and body", http.StatusBadRequest}
	ErrorCreateAssertion           = ErrorResponse{false, "create-assertion", "", "Error converting the serial-request to a serial assertion", http.StatusBadRequest}
	ErrorCheckAssertion            = ErrorResponse{false, "duplicate-assertion", "", "Error checking the serial-request. Please try again later", http.StatusBadRequest}
	ErrorCreateModelAssertion      = ErrorResponse{false, "create-assertion", "", "Error with the model assertion headers", http.StatusBadRequest}
	ErrorCreateSystemUserAssertion = ErrorResponse{false, "create-assertion", "", "Error with the system-user assertion", http.StatusBadRequest}
	ErrorDuplicateAssertion        = ErrorResponse{false, "duplicate-assertion", "", "The serial number and/or device-key have already been used to sign a device", http.StatusBadRequest}
	ErrorAccountAssertion          = ErrorResponse{false, "account-assertion", "", "Error retrieving the account assertion from the database", http.StatusBadRequest}
	ErrorSignAssertion             = ErrorResponse{false, "signing-assertion", "", "Error signing the assertion", http.StatusBadRequest}
	ErrorGenerateNonce             = ErrorResponse{false, "generate-nonce", "", "Error generating a nonce. Please try again later", http.StatusBadRequest}
)
