// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
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
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/CanonicalLtd/serial-vault/account"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/snapcore/snapd/asserts"
)

// ModelAssertionRequest is the JSON version of a model assertion request
type ModelAssertionRequest struct {
	BrandID string `json:"brand-id"`
	Name    string `json:"model"`
}

// ModelAssertionHandler is the API method to generate a model assertion
func ModelAssertionHandler(w http.ResponseWriter, r *http.Request) ErrorResponse {

	// Check that we have an authorised API key header
	err := checkAPIKey(r.Header.Get("api-key"))
	if err != nil {
		logMessage("MODEL", "invalid-api-key", "Invalid API key used")
		return ErrorInvalidAPIKey
	}

	defer r.Body.Close()

	// Decode the JSON body
	request := ModelAssertionRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	switch {
	// Check we have some data
	case err == io.EOF:
		return ErrorEmptyData
		// Check for parsing errors
	case err != nil:
		return ErrorResponse{false, "error-decode-json", "", err.Error(), http.StatusBadRequest}
	}

	// Check that the reseller functionality is enabled for the brand
	acc, err := datastore.Environ.DB.GetAccount(request.BrandID)
	if err != nil {
		return ErrorResponse{false, "error-account", "", err.Error(), http.StatusBadRequest}
	}
	if !acc.ResellerAPI {
		return ErrorResponse{false, "error-auth", "", "This feature is not enabled for this account", http.StatusBadRequest}
	}

	// Validate the model by checking that it exists on the database
	model, err := datastore.Environ.DB.FindModel(request.BrandID, request.Name, r.Header.Get("api-key"))
	if err != nil {
		logMessage("MODEL", "invalid-model", "Cannot find model with the matching brand and model")
		return ErrorInvalidModel
	}

	assertions := []asserts.Assertion{}

	// Build the model assertion headers
	assertionHeaders, keypair, err := createModelAssertionHeaders(model)
	if err != nil {
		logMessage("MODEL", "create-assertion", err.Error())
		return ErrorCreateModelAssertion
	}

	// Sign the assertion with the snapd assertions module
	signedAssertion, err := datastore.Environ.KeypairDB.SignAssertion(asserts.ModelType, assertionHeaders, []byte(""), model.BrandID, keypair.KeyID, keypair.SealedKey)
	if err != nil {
		logMessage("MODEL", "signing-assertion", err.Error())
		return ErrorResponse{false, "signing-assertion", "", err.Error(), http.StatusBadRequest}
	}

	// Add the account assertion to the assertions list
	fetchAssertionFromStore(&assertions, asserts.AccountType, []string{model.BrandID})

	// Add the account-key assertion to the assertions list
	fetchAssertionFromStore(&assertions, asserts.AccountKeyType, []string{keypair.KeyID})

	// Add the model assertion after the account and account-key assertions
	assertions = append(assertions, signedAssertion)

	// Return successful response with the signed assertions
	formatAssertionResponse(true, "", "", "", assertions, w)
	return ErrorResponse{Success: true}
}

func fetchAssertionFromStore(assertions *[]asserts.Assertion, modelType *asserts.AssertionType, headers []string) {
	accountKeyAssertion, err := account.FetchAssertionFromStore(modelType, headers)
	if err != nil {
		logMessage("MODEL", "assertion", err.Error())
	} else {
		*assertions = append(*assertions, accountKeyAssertion)
	}
}

func createModelAssertionHeaders(m datastore.Model) (map[string]interface{}, datastore.Keypair, error) {

	// Get the assertion headers for the model
	assert, err := datastore.Environ.DB.GetModelAssert(m.ID)
	if err != nil {
		return nil, datastore.Keypair{}, err
	}

	// Get the keypair for the model assertion
	keypair, err := datastore.Environ.DB.GetKeypair(assert.KeypairID)
	if err != nil {
		return nil, keypair, err
	}

	// Create the model assertion header
	headers := map[string]interface{}{
		"type":              asserts.ModelType.Name,
		"authority-id":      m.BrandID,
		"brand-id":          m.BrandID,
		"series":            fmt.Sprintf("%d", assert.Series),
		"model":             m.Name,
		"architecture":      assert.Architecture,
		"store":             assert.Store,
		"gadget":            assert.Gadget,
		"kernel":            assert.Kernel,
		"sign-key-sha3-384": keypair.KeyID,
		"timestamp":         time.Now().Format(time.RFC3339),
	}

	// Check if the optional required-snaps field is needed
	if len(assert.RequiredSnaps) == 0 {
		return headers, keypair, nil
	}

	snapList := strings.Split(assert.RequiredSnaps, ",")
	reqdSnaps := []interface{}{}
	for _, s := range snapList {
		reqdSnaps = append(reqdSnaps, strings.TrimSpace(s))
	}
	headers["required-snaps"] = reqdSnaps

	return headers, keypair, nil
}
