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

package model

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/auth"
	"github.com/CanonicalLtd/serial-vault/service/response"
)

// ListResponse is the JSON response from the API Models method
type ListResponse struct {
	Success      bool              `json:"success"`
	ErrorCode    string            `json:"error_code"`
	ErrorSubcode string            `json:"error_subcode"`
	ErrorMessage string            `json:"message"`
	Models       []datastore.Model `json:"models"`
}

// GetResponse is the JSON response from the API Get Model method
type GetResponse struct {
	Success      bool            `json:"success"`
	ErrorCode    string          `json:"error_code"`
	ErrorSubcode string          `json:"error_subcode"`
	ErrorMessage string          `json:"message"`
	Model        datastore.Model `json:"model"`
}

// listHandler is the API method to fetch the user records
func listHandler(w http.ResponseWriter, user datastore.User, apiCall bool) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Standard, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	dbModels, err := datastore.Environ.DB.ListAllowedModels(user)
	if err != nil {
		response.FormatStandardResponse(false, "error-fetch-models", "", err.Error(), w)
		return
	}

	// Return successful JSON response with the list of models
	w.WriteHeader(http.StatusOK)
	formatListResponse(dbModels, w)
}

// getHandler is the API method to fetch the models
func getHandler(w http.ResponseWriter, user datastore.User, apiCall bool, modelID int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	model, err := datastore.Environ.DB.GetAllowedModel(modelID, user)
	if err != nil {
		response.FormatStandardResponse(false, "error-fetch-model", "", err.Error(), w)
		return
	}

	// Return successful JSON response with the list of models
	w.WriteHeader(http.StatusOK)
	formatGetResponse(model, w)
}

func updateHandler(w http.ResponseWriter, user datastore.User, apiCall bool, modelID int, mdl datastore.Model) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	if modelID != mdl.ID {
		response.FormatStandardResponse(false, "error-stores-json", "", "The model IDs do not match", w)
		return
	}

	errorSubcode, err := datastore.Environ.DB.UpdateAllowedModel(mdl, user)
	if err != nil {
		log.Println("Error updating the store:", err)
		response.FormatStandardResponse(false, "error-updating-model", errorSubcode, "Error updating the model", w)
		return
	}

	// Return successful JSON response
	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", "", w)
}

func deleteHandler(w http.ResponseWriter, user datastore.User, apiCall bool, modelID int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	mdl := datastore.Model{ID: modelID}
	errorSubcode, err := datastore.Environ.DB.DeleteAllowedModel(mdl, user)
	if err != nil {
		response.FormatStandardResponse(false, "error-deleting-model", errorSubcode, err.Error(), w)
		return
	}

	// Return successful JSON response
	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", "", w)
}

func createHandler(w http.ResponseWriter, user datastore.User, apiCall bool, mdl datastore.Model) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	_, errorSubcode, err := datastore.Environ.DB.CreateAllowedModel(mdl, user)
	if err != nil {
		response.FormatStandardResponse(false, "error-model-json", errorSubcode, "", w)
		return
	}

	// Return successful JSON response
	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", "", w)
}

func assertionHeaders(w http.ResponseWriter, user datastore.User, apiCall bool, assert datastore.ModelAssertion) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	// Check that the user has permissions to access the model
	_, err = datastore.Environ.DB.GetAllowedModel(assert.ModelID, user)
	if err != nil {
		response.FormatStandardResponse(false, "error-get-model", "", "Cannot find model with the selected ID", w)
		return
	}

	err = datastore.Environ.DB.UpsertModelAssert(assert)
	if err != nil {
		response.FormatStandardResponse(false, "create-assertion", "", err.Error(), w)
		return
	}

	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", "", w)
}

func formatListResponse(models []datastore.Model, w http.ResponseWriter) error {
	response := ListResponse{Success: true, Models: models}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the models response.")
		return err
	}
	return nil
}

func formatGetResponse(model datastore.Model, w http.ResponseWriter) error {
	response := GetResponse{Success: true, Model: model}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the model response.")
		return err
	}
	return nil
}
