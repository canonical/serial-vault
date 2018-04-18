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
	"fmt"
	"io"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/service/assertion"
	svlog "github.com/CanonicalLtd/serial-vault/service/log"
	"github.com/CanonicalLtd/serial-vault/service/request"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/snapcore/snapd/asserts"
)

// SystemUserAssertion is the API method to generate a system-user assertion for a pivoted model
func SystemUserAssertion(w http.ResponseWriter, r *http.Request) response.ErrorResponse {
	// Check that we have an authorised API key header
	_, err := request.CheckModelAPI(r)
	if err != nil {
		svlog.Message("PIVOTUSER", "invalid-api-key", "Invalid API key used")
		return response.ErrorInvalidAPIKey
	}

	// Decode the body
	user := assertion.PivotSystemUserRequest{}
	err = json.NewDecoder(r.Body).Decode(&user)
	switch {
	// Check we have some data
	case err == io.EOF:
		return response.ErrorResponse{Success: false, Code: "error-user-data", Message: "No system-user data supplied", StatusCode: http.StatusBadRequest}
		// Check for parsing errors
	case err != nil:
		return response.ErrorResponse{Success: false, Code: "error-decode-json", Message: err.Error(), StatusCode: http.StatusBadRequest}
	}

	substore, errResponse := findModelPivot(user.Brand, user.ModelName, user.SerialNumber, r.Header.Get("api-key"))
	if !errResponse.Success {
		return errResponse
	}

	// Set up the request details for the pivot model
	model := substore.FromModel
	model.Name = substore.ModelName

	// Generate the system-user assertion for the pivoted model
	resp := assertion.GenerateSystemUserAssertion(user.SystemUserRequest, model)
	if !resp.Success {
		return response.ErrorResponse{Success: false, Code: resp.ErrorCode, Message: resp.ErrorMessage, StatusCode: http.StatusBadRequest}
	}

	w.Header().Set("Content-Type", asserts.MediaType)
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprint(w, resp.Assertion); err != nil {
		svlog.Message("PIVOTUSER", "system-user-assertion", err.Error())
	}

	return response.ErrorResponse{Success: true}
}
