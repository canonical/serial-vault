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

	"github.com/CanonicalLtd/serial-vault/service/request"
	"github.com/CanonicalLtd/serial-vault/service/response"
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
