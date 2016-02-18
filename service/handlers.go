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
	"log"
	"net/http"
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

// VersionResponse is the JSON response from the API Version method
type VersionResponse struct {
	Version string `json:"version"`
}

// SignResponse is the JSON response from the API Sign method
type SignResponse struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"message"`
	Signature    string `json:"identity"`
}

// ModelsResponse is the JSON response from the API Models method
type ModelsResponse struct {
	Success      bool           `json:"success"`
	ErrorMessage string         `json:"message"`
	Models       []ModelDisplay `json:"models"`
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
		formatSignResponse(false, "Not initialized post data.", "", w)
		return
	}

	assertions := new(Assertions)
	err := json.NewDecoder(r.Body).Decode(&assertions)

	defer r.Body.Close()

	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatSignResponse(false, "No data supplied for signing.", "", w)
		return
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("Error decoding JSON: %v", err)
		formatSignResponse(false, errorMessage, "", w)
		return
	}

	// Validate the model by checking that it exists on the database
	model, err := Environ.DB.FindModel(assertions.Brand, assertions.Model, assertions.Revision)
	if err != nil {
		errorMessage := "Cannot find model with the matching brand, model and revision."
		formatSignResponse(false, errorMessage, "", w)
		return
	}

	// Format the assertions string
	dataToSign, err := formatAssertion(assertions)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("Error formatting the assertions: %v", err)
		formatSignResponse(false, errorMessage, "", w)
		return
	}

	// Read the private key into a string using the model's signing key
	privateKey, err := getPrivateKey(model.SigningKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("Error reading the private key: %v", err)
		formatSignResponse(false, errorMessage, "", w)
		return
	}

	// Sign the assertions
	signedText, err := ClearSign(dataToSign, string(privateKey), "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorMessage := fmt.Sprintf("Error signing the assertions: %v\n", err)
		formatSignResponse(false, errorMessage, "", w)
		return
	}

	// Return successful JSON response with the signed text
	w.WriteHeader(http.StatusOK)
	formatSignResponse(true, "", string(signedText), w)
}

// ModelsHandler is the API method to list the models
func ModelsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var models []ModelDisplay

	dbModels, err := Environ.DB.ListModels()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorMessage := fmt.Sprintf("Error fetching the models: %v", err)
		formatModelsResponse(false, errorMessage, nil, w)
		return
	}

	w.WriteHeader(http.StatusOK)

	// Format the database records for output
	for _, model := range dbModels {
		mdl := ModelDisplay{ID: model.ID, BrandID: model.BrandID, Name: model.Name, Type: ModelType, Revision: model.Revision}
		models = append(models, mdl)
	}

	// Return successful JSON response with the list of models
	formatModelsResponse(true, "", models, w)
}
