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
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/CanonicalLtd/serial-vault/crypt"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/random"
	"github.com/CanonicalLtd/serial-vault/service/auth"
	svlog "github.com/CanonicalLtd/serial-vault/service/log"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/snapcore/snapd/asserts"
	"github.com/snapcore/snapd/release"
)

// systemUserAssertionAction is called by the API method to generate a system-user assertion
func systemUserAssertionAction(w http.ResponseWriter, authUser datastore.User, apiCall bool, user SystemUserRequest) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(authUser, datastore.Standard, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, response.ErrorAuth.Code, "", "", w)
		return
	}

	// Get the model:
	model, err := datastore.Environ.DB.GetAllowedModel(user.ModelID, datastore.User{})
	if err != nil {
		log.Println(err)
		svlog.Message("USER", response.ErrorInvalidModelID.Code, response.ErrorInvalidModelID.Message)
		response.FormatStandardResponse(false, response.ErrorInvalidModelID.Code, "", response.ErrorInvalidModelID.Message, w)
		return
	}

	// Generate the system-user assertion and return the response
	resp := GenerateSystemUserAssertion(user, model)
	if !resp.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		svlog.Message("USER", response.ErrorCreateSystemUserAssertion.Code, err.Error())
		response.FormatStandardResponse(false, response.ErrorCreateSystemUserAssertion.Code, "", err.Error(), w)
	}

}

// GenerateSystemUserAssertion creates a system-user assertion from the model and user details
func GenerateSystemUserAssertion(user SystemUserRequest, model datastore.Model) SystemUserResponse {
	// Check that the model has an active system-user keypair
	if !model.KeyActiveUser {
		svlog.Message("USER", response.ErrorInactiveModel.Code, response.ErrorInactiveModel.Message)
		return SystemUserResponse{ErrorCode: response.ErrorInactiveModel.Code, ErrorMessage: response.ErrorInactiveModel.Message}
	}

	// Fetch the account assertion from the database
	account, err := datastore.Environ.DB.GetAccount(model.AuthorityIDUser)
	if err != nil {
		svlog.Message("USER", response.ErrorAccountAssertion.Code, err.Error())
		return SystemUserResponse{ErrorCode: response.ErrorAccountAssertion.Code, ErrorMessage: response.ErrorAccountAssertion.Message}
	}

	// Create the system-user assertion headers from the request
	assertionHeaders := userRequestToAssertion(user, model)

	// Sign the system-user assertion using the system-user key
	signedAssertion, err := datastore.Environ.KeypairDB.SignAssertion(asserts.SystemUserType, assertionHeaders, nil, model.AuthorityIDUser, model.KeyIDUser, model.SealedKeyUser)
	if err != nil {
		svlog.Message("USER", response.ErrorSignAssertion.Code, err.Error())
		return SystemUserResponse{ErrorCode: response.ErrorSignAssertion.Code, ErrorMessage: err.Error()}
	}

	// Get the signed assertion
	serializedAssertion := asserts.Encode(signedAssertion)

	// Format the composite assertion
	composite := fmt.Sprintf("%s\n%s\n%s", account.Assertion, model.AssertionUser, serializedAssertion)

	return SystemUserResponse{Success: true, Assertion: composite}
}

func userRequestToAssertion(user SystemUserRequest, model datastore.Model) map[string]interface{} {
	// Create the salt from a random string
	reg, _ := regexp.Compile("[^A-Za-z0-9]+")
	randomText, err := random.GenerateRandomString(32)
	if err != nil {
		svlog.Message("USER", response.ErrorSignAssertion.Code, err.Error())
		return map[string]interface{}{}
	}
	baseSalt := reg.ReplaceAllString(randomText, "")

	// Encrypt the password
	salt := fmt.Sprintf("$6$%s$", baseSalt)
	password := crypt.CLibCryptUser(user.Password, salt)

	// Set the since and end date/times
	since, err := time.Parse(time.RFC3339, user.Since)
	if err != nil {
		since = time.Now().UTC()
	}
	until, err := time.Parse(time.RFC3339, user.Until)
	if err != nil {
		until = since.Add(oneYearDuration)
	}
	if since.After(until) {
		until = since.Add(oneYearDuration)
	}

	// Create the serial assertion header from the serial-request headers
	headers := map[string]interface{}{
		"type":              asserts.SystemUserType.Name,
		"revision":          userAssertionRevision,
		"authority-id":      model.AuthorityIDUser,
		"brand-id":          model.AuthorityIDUser,
		"email":             user.Email,
		"name":              user.Name,
		"username":          user.Username,
		"models":            []interface{}{model.Name},
		"series":            []interface{}{release.Series},
		"since":             since.Format(time.RFC3339),
		"until":             until.Format(time.RFC3339),
		"sign-key-sha3-384": model.KeyIDUser,
	}

	if len(user.SSHKeys) > 0 {
		// Convert the keys to an interface slice
		keys := make([]interface{}, len(user.SSHKeys))
		for i, k := range user.SSHKeys {
			keys[i] = k
		}
		headers["ssh-keys"] = keys
	} else {
		headers["password"] = password
	}

	// Create a new serial assertion
	return headers
}
