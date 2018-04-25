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
	"net/http"

	"github.com/snapcore/snapd/asserts"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/auth"
	svlog "github.com/CanonicalLtd/serial-vault/service/log"
	"github.com/CanonicalLtd/serial-vault/service/response"
)

const (
	responseValidModel    = "valid-model"
	responseValidSubstore = "valid-substore"
)

// validateAssertionAction is called by the API method to check a serial assertion
func validateAssertionAction(w http.ResponseWriter, authUser datastore.User, apiCall bool, assertion asserts.Assertion) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(authUser, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, response.ErrorAuth.Code, "", response.ErrorAuth.Message, w)
		return
	}

	// Check that the account is accessbible by the user
	if _, err = datastore.Environ.DB.GetAllowedAccount(assertion.HeaderString("brand-id"), authUser); err != nil {
		response.FormatStandardResponse(false, response.ErrorInvalidAccount.Code, "", response.ErrorInvalidAccount.Message, w)
		return
	}

	// Assume this is an original (non-pivoted) serial assertion
	// Validate the model by checking that it exists on the database
	if modelFound := datastore.Environ.DB.CheckModelExists(assertion.HeaderString("brand-id"), assertion.HeaderString("model")); modelFound {
		response.FormatStandardResponse(true, responseValidModel, "", "", w)
		return
	}

	// Assume that this is a pivoted serial assertion
	// Check for a sub-store model for the pivot
	if _, err = datastore.Environ.DB.GetSubstoreModel(assertion.HeaderString("brand-id"), assertion.HeaderString("model"), assertion.HeaderString("serial")); err != nil {
		svlog.Message("CHECK", "invalid-substore", "Cannot find sub-store model")
		response.FormatStandardResponse(false, response.ErrorInvalidSubstore.Code, "", response.ErrorInvalidSubstore.Message, w)
		return
	}

	// Found the sub-store record
	response.FormatStandardResponse(true, responseValidSubstore, "", "", w)
}
