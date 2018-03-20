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

package assertion

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/service/log"
	"github.com/CanonicalLtd/serial-vault/service/request"
	"github.com/CanonicalLtd/serial-vault/service/response"
)

// ModelAssertionRequest is the JSON version of a model assertion request
type ModelAssertionRequest struct {
	BrandID string `json:"brand-id"`
	Name    string `json:"model"`
}

// ModelAssertion is the API method to generate a model assertion
func ModelAssertion(w http.ResponseWriter, r *http.Request) response.ErrorResponse {
	// Validate the model API key
	apiKey, err := request.CheckModelAPI(r)
	if err != nil {
		log.Message("MODEL", "invalid-api-key", "Invalid API key used")
		return response.ErrorInvalidAPIKey
	}

	defer r.Body.Close()

	// Decode the JSON body
	request := ModelAssertionRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	switch {
	// Check we have some data
	case err == io.EOF:
		return response.ErrorEmptyData
		// Check for parsing errors
	case err != nil:
		return response.ErrorResponse{Success: false, Code: "error-decode-json", SubCode: "", Message: err.Error(), StatusCode: http.StatusBadRequest}
	}

	return modelAssertionHandler(w, apiKey, request)
}
