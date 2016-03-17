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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Assertions are the details of the device
type Assertions struct {
	Brand        string `json:"brand-id"`
	Model        string `json:"model"`
	SerialNumber string `json:"serial"`
	Type         string `json:"type"`
	Revision     int    `json:"revision"`
	PublicKey    string `json:"device-key"`
}

// ModelDisplay is the JSON version of a model, excluding the signing-key
type ModelDisplay struct {
	ID       int    `json:"id"`
	BrandID  string `json:"brand-id"`
	Name     string `json:"model"`
	Type     string `json:"type"`
	Revision int    `json:"revision"`
}

// ModelWithKey is the JSON version of a model, including the signing-key
type ModelWithKey struct {
	ID         int    `json:"id"`
	BrandID    string `json:"brand-id"`
	Name       string `json:"model"`
	Type       string `json:"type"`
	SigningKey string `json:"signing-key"`
	Revision   int    `json:"revision"`
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
	Success      bool           `json:"success"`
	ErrorCode    string         `json:"error_code"`
	ErrorSubcode string         `json:"error_subcode"`
	ErrorMessage string         `json:"message"`
	Models       []ModelDisplay `json:"models"`
}

// ModelResponse is the JSON response from the API Get Model method
type ModelResponse struct {
	Success      bool         `json:"success"`
	ErrorCode    string       `json:"error_code"`
	ErrorSubcode string       `json:"error_subcode"`
	ErrorMessage string       `json:"message"`
	Model        ModelDisplay `json:"model"`
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
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		formatSignResponse(false, "error-nil-data", "", "Uninitialized POST data", "", w)
		return
	}

	assertions := new(Assertions)
	err := json.NewDecoder(r.Body).Decode(&assertions)

	defer r.Body.Close()

	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatSignResponse(false, "error-sign-empty", "", "No data supplied for signing", "", w)
		return
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatSignResponse(false, "error-decode-json", "", errorMessage, "", w)
		return
	}

	// Validate the model by checking that it exists on the database
	model, err := Environ.DB.FindModel(assertions.Brand, assertions.Model, assertions.Revision)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		formatSignResponse(false, "error-model-not-found", "", "Cannot find model with the matching brand, model and revision", "", w)
		return
	}

	// Format the assertions string
	dataToSign, err := formatAssertion(assertions)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatSignResponse(false, "error-format-assertions", "", errorMessage, "", w)
		return
	}

	// Read the private key into a string using the model's signing key
	privateKey, err := getPrivateKey(model.SigningKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatSignResponse(false, "error-read-private-key", "", errorMessage, "", w)
		return
	}

	// Sign the assertions
	signedText, err := ClearSign(dataToSign, string(privateKey), "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorMessage := fmt.Sprintf("%v", err)
		formatSignResponse(false, "error-signing-assertions", "", errorMessage, "", w)
		return
	}

	// Return successful JSON response with the signed text
	w.WriteHeader(http.StatusOK)
	formatSignResponse(true, "", "", "", string(signedText), w)
}

func modelForDisplay(model Model) ModelDisplay {
	return ModelDisplay{ID: model.ID, BrandID: model.BrandID, Name: model.Name, Type: ModelType, Revision: model.Revision}
}

// ModelsHandler is the API method to list the models
func ModelsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var models []ModelDisplay

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
		formatModelResponse(false, "error-invalid-model", "", errorMessage, ModelDisplay{}, w)
		return
	}

	model, err := Environ.DB.GetModel(modelID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("Model ID: %d.", modelID)
		formatModelResponse(false, "error-get-model", "", errorMessage, ModelDisplay{ID: modelID}, w)
		return
	}

	// Format the model for output and return JSON response
	w.WriteHeader(http.StatusOK)
	mdl := modelForDisplay(*model)
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
		formatModelResponse(false, "error-invalid-model", "", errorMessage, ModelDisplay{}, w)
		return
	}

	// Check that we have a message body
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		formatModelResponse(false, "error-nil-data", "", "Uninitialized POST data", ModelDisplay{}, w)
		return
	}
	defer r.Body.Close()

	// Decode the JSON body
	mdl := ModelDisplay{}
	err = json.NewDecoder(r.Body).Decode(&mdl)
	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatModelResponse(false, "error-model-data", "", "No model data supplied.", ModelDisplay{}, w)
		return
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatModelResponse(false, "error-decode-json", "", errorMessage, ModelDisplay{}, w)
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
		formatModelResponse(false, "error-nil-data", "", "Uninitialized POST data", ModelDisplay{}, w)
		return
	}
	defer r.Body.Close()

	// Decode the JSON body
	mdlWithKey := ModelWithKey{}
	err := json.NewDecoder(r.Body).Decode(&mdlWithKey)
	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatModelResponse(false, "error-model-data", "", "No model data supplied", ModelDisplay{}, w)
		return
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatModelResponse(false, "error-decode-json", "", errorMessage, ModelDisplay{}, w)
		return
	}

	// The signing-key is base64 encoded, so we need to decode it
	decodedSigningKey, err := base64.StdEncoding.DecodeString(mdlWithKey.SigningKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatModelResponse(false, "error-decode-key", "", errorMessage, ModelDisplay{}, w)
		return
	}
	mdlWithKey.SigningKey = string(decodedSigningKey)

	// Store the signing-key in the keystore and create a new model
	model := Model{BrandID: mdlWithKey.BrandID, Name: mdlWithKey.Name, SigningKey: mdlWithKey.SigningKey, Revision: mdlWithKey.Revision}
	errorSubcode := ""
	model.ID, errorSubcode, err = Environ.DB.CreateModel(model)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatModelResponse(false, "error-creating-model", errorSubcode, errorMessage, ModelDisplay{}, w)
		return
	}

	// Format the model for output and return JSON response
	w.WriteHeader(http.StatusOK)
	mdl := modelForDisplay(model)
	formatModelResponse(true, "", "", "", mdl, w)
}
