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

package service

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/ubuntu-core/snappy/asserts"
)

// ModelSerialize is the JSON version of a model, with the signing key ID
type ModelSerialize struct {
	ID          int    `json:"id"`
	BrandID     string `json:"brand-id"`
	Name        string `json:"model"`
	Type        string `json:"type"`
	KeypairID   int    `json:"keypair-id"`
	Revision    int    `json:"revision"`
	AuthorityID string `json:"authority-id"`
	KeyID       string `json:"key-id"`
}

// VersionResponse is the JSON response from the API Version method
type VersionResponse struct {
	Version string `json:"version"`
}

// SignResponse is the JSON response from the API Sign method
type SignResponse struct {
	Success      bool   `json:"success"`
	ErrorCode    string `json:"error_code"`
	ErrorSubcode string `json:"error_subcode"`
	ErrorMessage string `json:"message"`
	Signature    string `json:"identity"`
}

// ModelsResponse is the JSON response from the API Models method
type ModelsResponse struct {
	Success      bool             `json:"success"`
	ErrorCode    string           `json:"error_code"`
	ErrorSubcode string           `json:"error_subcode"`
	ErrorMessage string           `json:"message"`
	Models       []ModelSerialize `json:"models"`
}

// ModelResponse is the JSON response from the API Get Model method
type ModelResponse struct {
	Success      bool           `json:"success"`
	ErrorCode    string         `json:"error_code"`
	ErrorSubcode string         `json:"error_subcode"`
	ErrorMessage string         `json:"message"`
	Model        ModelSerialize `json:"model"`
}

// KeypairsResponse is the JSON response from the API Keypairs method
type KeypairsResponse struct {
	Success      bool      `json:"success"`
	ErrorCode    string    `json:"error_code"`
	ErrorSubcode string    `json:"error_subcode"`
	ErrorMessage string    `json:"message"`
	Keypairs     []Keypair `json:"keypairs"`
}

// VersionHandler is the API method to return the version of the service
func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	response := VersionResponse{Version: Environ.Config.Version}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding the version response: %v\n", err)
	}
}

// SignHandler is the API method to sign assertions from the device
func SignHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		formatSignResponse(false, "error-nil-data", "", "Uninitialized POST data", nil, w)
		return
	}

	// Read the full request body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatSignResponse(false, "error-sign-read", "", err.Error(), nil, w)
		return
	}
	if len(data) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		formatSignResponse(false, "error-sign-empty", "", "No data supplied for signing", nil, w)
		return
	}

	defer r.Body.Close()

	// Use the ubuntu-core assertions module to decode the body and validate
	assertion, err := asserts.Decode(data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatSignResponse(false, "error-decode-json", "", err.Error(), nil, w)
		return
	}

	// Check that we have a device-serial assertion (the details will have been validated by Decode call)
	if assertion.Type() != asserts.DeviceSerialType {
		w.WriteHeader(http.StatusBadRequest)
		formatSignResponse(false, "error-decode-assertion", "error-invalid-type", "The assertion type must be 'device-serial'", nil, w)
		return
	}

	// Validate the model by checking that it exists on the database
	model, err := Environ.DB.FindModel(assertion.Header("brand-id"), assertion.Header("model"), assertion.Revision())
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		formatSignResponse(false, "error-model-not-found", "", "Cannot find model with the matching brand, model and revision", nil, w)
		return
	}

	// Sign the assertion with the ubuntu-core assertions module
	signedAssertion, err := Environ.KeypairDB.Sign(asserts.DeviceSerialType, assertion.Headers(), assertion.Body(), model.KeyID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		formatSignResponse(false, "error-signing-assertions", "", err.Error(), signedAssertion, w)
		return
	}

	// Return successful JSON response with the signed text
	formatSignResponse(true, "", "", "", signedAssertion, w)
}

func modelForDisplay(model Model) ModelSerialize {
	return ModelSerialize{ID: model.ID, BrandID: model.BrandID, Name: model.Name, Type: ModelType, Revision: model.Revision, KeypairID: model.KeypairID, AuthorityID: model.AuthorityID, KeyID: model.KeyID}
}

// ModelsHandler is the API method to list the models
func ModelsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	models := []ModelSerialize{}

	dbModels, err := Environ.DB.ListModels()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorMessage := fmt.Sprintf("%v", err)
		formatModelsResponse(false, "error-fetch-models", "", errorMessage, nil, w)
		return
	}

	w.WriteHeader(http.StatusOK)

	// Format the database records for output
	for _, model := range dbModels {
		mdl := modelForDisplay(model)
		models = append(models, mdl)
	}

	// Return successful JSON response with the list of models
	formatModelsResponse(true, "", "", "", models, w)
}

// ModelGetHandler is the API method to get a model by ID.
func ModelGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)

	modelID, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("%v", vars)
		formatModelResponse(false, "error-invalid-model", "", errorMessage, ModelSerialize{}, w)
		return
	}

	model, err := Environ.DB.GetModel(modelID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("Model ID: %d.", modelID)
		formatModelResponse(false, "error-get-model", "", errorMessage, ModelSerialize{ID: modelID}, w)
		return
	}

	// Format the model for output and return JSON response
	w.WriteHeader(http.StatusOK)
	mdl := modelForDisplay(model)
	formatModelResponse(true, "", "", "", mdl, w)
}

// ModelUpdateHandler is the API method to update a model.
func ModelUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get the model primary key
	vars := mux.Vars(r)
	modelID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("%v", vars["id"])
		formatModelResponse(false, "error-invalid-model", "", errorMessage, ModelSerialize{}, w)
		return
	}

	// Check that we have a message body
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		formatModelResponse(false, "error-nil-data", "", "Uninitialized POST data", ModelSerialize{}, w)
		return
	}
	defer r.Body.Close()

	// Decode the JSON body
	mdl := ModelSerialize{}
	err = json.NewDecoder(r.Body).Decode(&mdl)
	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatModelResponse(false, "error-model-data", "", "No model data supplied.", ModelSerialize{}, w)
		return
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatModelResponse(false, "error-decode-json", "", errorMessage, ModelSerialize{}, w)
		return
	}

	// Update the database
	model := Model{ID: modelID, BrandID: mdl.BrandID, Name: mdl.Name, Revision: mdl.Revision}
	errorSubcode, err := Environ.DB.UpdateModel(model)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatModelResponse(false, "error-updating-model", errorSubcode, errorMessage, mdl, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	formatModelResponse(true, "", "", "", mdl, w)
}

// ModelCreateHandler is the API method to create a new model.
func ModelCreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Check that we have a message body
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		formatModelResponse(false, "error-nil-data", "", "Uninitialized POST data", ModelSerialize{}, w)
		return
	}
	defer r.Body.Close()

	// Decode the JSON body
	mdlWithKey := ModelSerialize{}
	err := json.NewDecoder(r.Body).Decode(&mdlWithKey)
	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatModelResponse(false, "error-model-data", "", "No model data supplied", ModelSerialize{}, w)
		return
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatModelResponse(false, "error-decode-json", "", errorMessage, ModelSerialize{}, w)
		return
	}

	// Create a new model, linked to the existing signing-key
	model := Model{BrandID: mdlWithKey.BrandID, Name: mdlWithKey.Name, KeypairID: mdlWithKey.KeypairID, Revision: mdlWithKey.Revision}
	errorSubcode := ""
	model, errorSubcode, err = Environ.DB.CreateModel(model)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatModelResponse(false, "error-creating-model", errorSubcode, errorMessage, ModelSerialize{}, w)
		return
	}

	// Format the model for output and return JSON response
	w.WriteHeader(http.StatusOK)
	formatModelResponse(true, "", "", "", modelForDisplay(model), w)
}
