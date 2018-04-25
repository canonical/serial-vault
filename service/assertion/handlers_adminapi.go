// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
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

package assertion

import (
	"encoding/json"
	"io"
	"net/http"

	svlog "github.com/CanonicalLtd/serial-vault/service/log"
	"github.com/CanonicalLtd/serial-vault/service/request"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/snapcore/snapd/asserts"
)

// APISystemUser is the API method to generate a signed system-user assertion for a device
func APISystemUser(w http.ResponseWriter, r *http.Request) {

	// Validate the user and API key
	authUser, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, response.ErrorAuth.Code, "", err.Error(), w)
		return
	}

	// Decode the body
	user := SystemUserRequest{}
	err = json.NewDecoder(r.Body).Decode(&user)
	switch {
	// Check we have some data
	case err == io.EOF:
		response.FormatStandardResponse(false, response.ErrorEmptyData.Code, "", response.ErrorEmptyData.Message, w)
		return
		// Check for parsing errors
	case err != nil:
		response.FormatStandardResponse(false, response.ErrorDecodeJSON.Code, "", err.Error(), w)
		return
	}

	systemUserAssertionAction(w, authUser, true, user)
}

// APIValidateSerial is the API method to validate a serial assertion for a device
func APIValidateSerial(w http.ResponseWriter, r *http.Request) {
	// Validate the user and API key
	authUser, err := request.CheckUserAPI(r)
	if err != nil {
		response.FormatStandardResponse(false, response.ErrorAuth.Code, "", err.Error(), w)
		return
	}

	assertion, errResponse := parseSerialAssertion(r)
	if !errResponse.Success {
		response.FormatStandardResponse(false, errResponse.Code, "", errResponse.Message, w)
		return
	}

	validateAssertionAction(w, authUser, true, assertion)
}

func parseSerialAssertion(r *http.Request) (asserts.Assertion, response.ErrorResponse) {
	defer r.Body.Close()

	// Get the serial assertion from the body
	dec := asserts.NewDecoder(r.Body)
	assertion, err := dec.Decode()
	if err == io.EOF {
		svlog.Message("CHECK", response.ErrorInvalidAssertion.Code, response.ErrorEmptyData.Message)
		return nil, response.ErrorEmptyData
	}
	if err != nil {
		svlog.Message("CHECK", response.ErrorInvalidAssertion.Code, err.Error())
		return nil, response.ErrorResponse{Success: false, Code: "decode-assertion", Message: err.Error(), StatusCode: http.StatusBadRequest}
	}

	// Check that we have a serial assertion (the details will have been validated by Decode call)
	if assertion.Type() != asserts.SerialType {
		svlog.Message("CHECK", response.ErrorInvalidType.Code, response.ErrorInvalidType.Message)
		return nil, response.ErrorInvalidType
	}
	return assertion, response.ErrorResponse{Success: true}
}
