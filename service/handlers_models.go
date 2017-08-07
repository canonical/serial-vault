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
	"net/http"
	"strconv"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/gorilla/mux"
)

// ModelSerialize is the JSON version of a model, with the signing key ID
type ModelSerialize struct {
	ID              int    `json:"id"`
	BrandID         string `json:"brand-id"`
	Name            string `json:"model"`
	Type            string `json:"type"`
	KeypairID       int    `json:"keypair-id"`
	APIKey          string `json:"api-key"`
	AuthorityID     string `json:"authority-id"`
	KeyID           string `json:"key-id"`
	KeyActive       bool   `json:"key-active"`
	KeypairIDUser   int    `json:"keypair-id-user"`
	AuthorityIDUser string `json:"authority-id-user"`
	KeyIDUser       string `json:"key-id-user"`
	KeyActiveUser   bool   `json:"key-active-user"`
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

func modelForDisplay(model datastore.Model) ModelSerialize {
	return ModelSerialize{
		ID: model.ID, BrandID: model.BrandID, Name: model.Name, Type: ModelType,
		KeypairID: model.KeypairID, APIKey: model.APIKey, AuthorityID: model.AuthorityID, KeyID: model.KeyID, KeyActive: model.KeyActive,
		KeypairIDUser: model.KeypairIDUser, AuthorityIDUser: model.AuthorityIDUser, KeyIDUser: model.KeyIDUser, KeyActiveUser: model.KeyActiveUser,
	}
}

// ModelsHandler is the API method to list the models
func ModelsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	models := []ModelSerialize{}

	authUser, err := checkIsAdminAndGetAuthUser(w, r)
	if err != nil {
		formatModelsResponse(false, "error-auth", "", "", models, w)
		return
	}

	dbModels, err := listAllowedModels(authUser)
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

func listAllowedModels(authUser datastore.User) ([]datastore.Model, error) {
	switch authUser.Role {
	case 0:
		fallthrough
	case datastore.Superuser:
		return listAllModels()
	case datastore.Admin:
		return listModelsFilteredByUser(authUser.Username)
	}
	return []datastore.Model{}, nil
}

// TODO: Move these two listAllModels and listModelsFilteredByUser to datastore
func listAllModels() ([]datastore.Model, error) {
	return datastore.Environ.DB.ListModels("")
}

func listModelsFilteredByUser(username string) ([]datastore.Model, error) {
	return datastore.Environ.DB.ListModels(username)
}

// ModelGetHandler is the API method to get a model by ID.
func ModelGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetAuthUser(w, r)
	if err != nil {
		formatModelResponse(false, "error-auth", "", "", ModelSerialize{}, w)
		return
	}

	vars := mux.Vars(r)
	modelID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("%v", vars)
		formatModelResponse(false, "error-invalid-model", "", errorMessage, ModelSerialize{}, w)
		return
	}

	model, err := getAllowedModel(modelID, authUser)
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

func getAllowedModel(modelID int, authUser datastore.User) (datastore.Model, error) {
	switch authUser.Role {
	case 0:
		fallthrough
	case datastore.Superuser:
		return getModel(modelID)
	case datastore.Admin:
		return getModelFilteredByUser(modelID, authUser.Username)
	}
	return datastore.Model{}, nil
}

func getModel(modelID int) (datastore.Model, error) {
	return datastore.Environ.DB.GetModel(modelID, "")
}

func getModelFilteredByUser(modelID int, username string) (datastore.Model, error) {
	return datastore.Environ.DB.GetModel(modelID, username)
}

// ModelUpdateHandler is the API method to update a model.
func ModelUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetAuthUser(w, r)
	if err != nil {
		formatModelResponse(false, "error-auth", "", "", ModelSerialize{}, w)
		return
	}

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

	// Validate model name; the rule is: lowercase with no spaces
	err = validateModelName(mdl.Name)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatModelResponse(false, "error-updating-model", "", err.Error(), ModelSerialize{}, w)
		return
	}

	// Update the database
	model := datastore.Model{ID: modelID, BrandID: mdl.BrandID, Name: mdl.Name, KeypairID: mdl.KeypairID, KeypairIDUser: mdl.KeypairIDUser}
	errorSubcode, err := updateAllowedModel(model, authUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatModelResponse(false, "error-updating-model", errorSubcode, err.Error(), mdl, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	formatModelResponse(true, "", "", "", mdl, w)
}

func updateAllowedModel(model datastore.Model, authUser datastore.User) (string, error) {
	switch authUser.Role {
	case 0:
		fallthrough
	case datastore.Superuser:
		return updateModel(model)
	case datastore.Admin:
		return updateModelFilteredByUser(model, authUser.Username)
	}
	return "", nil
}

func updateModel(model datastore.Model) (string, error) {
	return datastore.Environ.DB.UpdateModel(model, "")
}

func updateModelFilteredByUser(model datastore.Model, username string) (string, error) {
	return datastore.Environ.DB.UpdateModel(model, username)
}

// ModelDeleteHandler is the API method to delete a model.
func ModelDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetAuthUser(w, r)
	if err != nil {
		formatModelResponse(false, "error-auth", "", "", ModelSerialize{}, w)
		return
	}

	// Get the model primary key
	vars := mux.Vars(r)
	modelID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("%v", vars["id"])
		formatModelResponse(false, "error-invalid-model", "", errorMessage, ModelSerialize{}, w)
		return
	}

	// Update the database
	model := datastore.Model{ID: modelID}
	errorSubcode, err := deleteAllowedModel(model, authUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatModelResponse(false, "error-deleting-model", errorSubcode, errorMessage, ModelSerialize{}, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	formatModelResponse(true, "", "", "", ModelSerialize{}, w)
}

func deleteAllowedModel(model datastore.Model, authUser datastore.User) (string, error) {
	switch authUser.Role {
	case 0:
		fallthrough
	case datastore.Superuser:
		return deleteModel(model)
	case datastore.Admin:
		return deleteModelFilteredByUser(model, authUser.Username)
	}
	return "", nil
}

func deleteModel(model datastore.Model) (string, error) {
	return datastore.Environ.DB.DeleteModel(model, "")
}

func deleteModelFilteredByUser(model datastore.Model, username string) (string, error) {
	return datastore.Environ.DB.DeleteModel(model, username)
}

// ModelCreateHandler is the API method to create a new model.
func ModelCreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetAuthUser(w, r)
	if err != nil {
		formatModelResponse(false, "error-auth", "", "", ModelSerialize{}, w)
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
	mdlWithKey := ModelSerialize{}
	err = json.NewDecoder(r.Body).Decode(&mdlWithKey)
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

	// Validate model name; the rule is: lowercase with no spaces
	err = validateModelName(mdlWithKey.Name)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatModelResponse(false, "error-creating-model", "", err.Error(), ModelSerialize{}, w)
		return
	}

	// Create a new model, linked to the existing signing-key
	model := datastore.Model{BrandID: mdlWithKey.BrandID, Name: mdlWithKey.Name, KeypairID: mdlWithKey.KeypairID, KeypairIDUser: mdlWithKey.KeypairIDUser}
	errorSubcode := ""
	model, errorSubcode, err = createAllowedModel(model, authUser)
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

func createAllowedModel(model datastore.Model, authUser datastore.User) (datastore.Model, string, error) {
	switch authUser.Role {
	case 0:
		fallthrough
	case datastore.Superuser:
		return createModel(model)
	case datastore.Admin:
		return createModelFilteredByUser(model, authUser.Username)
	}
	return datastore.Model{}, "", nil
}

func createModel(model datastore.Model) (datastore.Model, string, error) {
	return datastore.Environ.DB.CreateModel(model, "")
}

func createModelFilteredByUser(model datastore.Model, username string) (datastore.Model, string, error) {
	return datastore.Environ.DB.CreateModel(model, username)
}
