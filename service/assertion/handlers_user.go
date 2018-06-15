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

package assertion

import (
	"encoding/json"
	"io"
	"net/http"

	"time"

	"github.com/CanonicalLtd/serial-vault/service/auth"
	"github.com/CanonicalLtd/serial-vault/service/response"
)

const oneYearDuration = time.Duration(24*365) * time.Hour
const userAssertionRevision = "1"

// SystemUserRequest is the JSON version of the request to create a system-user assertion
type SystemUserRequest struct {
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	ModelID  int      `json:"model"`
	Since    string   `json:"since"`
	Until    string   `json:"until"`
	SSHKeys  []string `json:"sshKeys"`
}

// PivotSystemUserRequest is the JSON version of the request to create a system-user assertion
type PivotSystemUserRequest struct {
	SystemUserRequest
	Brand        string `json:"brand-id"`
	ModelName    string `json:"model-name"`
	SerialNumber string `json:"serial"`
}

// SystemUserResponse is the response from a system-user creation
type SystemUserResponse struct {
	Success      bool   `json:"success"`
	ErrorCode    string `json:"error_code"`
	ErrorSubcode string `json:"error_subcode"`
	ErrorMessage string `json:"message"`
	Assertion    string `json:"assertion"`
}

// SystemUserAssertion is the API method to generate a signed system-user assertion for a device
func SystemUserAssertion(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
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

	if user.Password == "" && len(user.SSHKeys) == 0 {
		//return error
	}

	systemUserAssertionAction(w, authUser, false, user)
}
