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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
)

func TestModelsHandler(t *testing.T) {

	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/v1/models", nil)
	http.HandlerFunc(ModelsHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := ModelsResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the models response: %v", err)
	}
	if len(result.Models) != 6 {
		t.Errorf("Expected 6 models, got %d", len(result.Models))
	}
	if result.Models[0].Name != "alder" {
		t.Errorf("Expected model name 'alder', got %s", result.Models[0].Name)
	}
}

func TestModelsHandlerWithPermissions(t *testing.T) {

	// Mock the database
	c := config.Settings{EnableUserAuth: true}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/v1/models", nil)

	// Create a JWT and add it to the request
	jwtToken, err := createJWT()
	if err != nil {
		t.Errorf("Error creating a JWT: %v", err)
	}
	r.Header.Set("Authorization", "Bearer "+jwtToken)

	http.HandlerFunc(ModelsHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := ModelsResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the models response: %v", err)
	}
	if len(result.Models) != 3 {
		t.Errorf("Expected 3 models, got %d", len(result.Models))
	}
	if result.Models[0].Name != "alder" {
		t.Errorf("Expected model name 'alder', got %s", result.Models[0].Name)
	}
}

func TestModelsHandlerWithError(t *testing.T) {

	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/v1/models", nil)
	http.HandlerFunc(ModelsHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := ModelsResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the models response: %v", err)
	}
	if result.Success {
		t.Error("Expected error, got success response")
	}

}

func TestModelGetHandler(t *testing.T) {

	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}

	result, _ := sendRequest(t, "GET", "/v1/models/1", nil)

	if result.Model.ID != 1 {
		t.Errorf("Expected model with ID 1, got %d", result.Model.ID)
	}
	if result.Model.Name != "alder" {
		t.Errorf("Expected model name 'alder', got %s", result.Model.Name)
	}
}

func TestModelGetHandlerWithPermissions(t *testing.T) {

	// Mock the database
	c := config.Settings{EnableUserAuth: true}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	result, _ := sendRequest(t, "GET", "/v1/models/1", nil)

	if result.Model.ID != 1 {
		t.Errorf("Expected model with ID 1, got %d", result.Model.ID)
	}
	if result.Model.Name != "alder" {
		t.Errorf("Expected model name 'alder', got %s", result.Model.Name)
	}
}

func TestModelGetHandlerWithError(t *testing.T) {

	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}

	sendRequestExpectError(t, "GET", "/v1/models/999999", nil)
}

func TestModelGetHandlerWithBadID(t *testing.T) {

	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}

	sendRequestExpectError(t, "GET", "/v1/models/999999999999999999999999999999", nil)
}

func TestModelUpdateHandler(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}

	// Update a model
	data := `
	{
		"id": 1,
		"brand-id": "System",
		"model":"the-model",
		"serial":"A1234-L",
		"device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
	}`

	result, _ := sendRequest(t, "PUT", "/v1/models/1", bytes.NewBufferString(data))

	if result.Model.ID != 1 {
		t.Errorf("Expected model with ID 1, got %d", result.Model.ID)
	}
	if result.Model.Name != "the-model" {
		t.Errorf("Expected model name 'the-model', got %s", result.Model.Name)
	}
}

func TestModelUpdateHandlerWithErrors(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}}

	// Update a model
	data := `{}`

	sendRequestExpectError(t, "PUT", "/v1/models/1", bytes.NewBufferString(data))
}

func TestModelUpdateHandlerWithNilData(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}}

	sendRequestExpectError(t, "PUT", "/v1/models/1", nil)
}

func TestModelUpdateHandlerWithEmptyData(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}}

	sendRequestExpectError(t, "PUT", "/v1/models/1", bytes.NewBufferString(""))
}

func TestModelUpdateHandlerWithBadData(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}}

	sendRequestExpectError(t, "PUT", "/v1/models/1", bytes.NewBufferString("bad"))
}

func TestModelUpdateHandlerWithBadID(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}}

	sendRequestExpectError(t, "PUT", "/v1/models/999999999999999999999999999999", bytes.NewBufferString("bad"))
}

func TestModelDeleteHandler(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}

	// Delete a model
	data := "{}"
	sendRequest(t, "DELETE", "/v1/models/1", bytes.NewBufferString(data))
}

func TestModelDeleteHandlerWithErrors(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}}

	// Delete a model
	data := `{}`

	sendRequestExpectError(t, "DELETE", "/v1/models/1", bytes.NewBufferString(data))
}

func TestModelDeleteHandlerWithBadID(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}}

	sendRequestExpectError(t, "DELETE", "/v1/models/999999999999999999999999999999", bytes.NewBufferString("bad"))
}

func TestModelCreateHandler(t *testing.T) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	// Define a model linked with the signing-key as JSON
	model := ModelSerialize{BrandID: "System", Name: "the-model", KeypairID: 1}
	data, _ := json.Marshal(model)

	result, _ := sendRequest(t, "POST", "/v1/models", bytes.NewReader(data))
	if result.Model.ID != 7 {
		t.Errorf("Expected model with ID 7, got %d", result.Model.ID)
	}
	if result.Model.Name != "the-model" {
		t.Errorf("Expected model name 'the-model', got %s", result.Model.Name)
	}
}

func TestModelCreateHandlerWithError(t *testing.T) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	// Define a model linked with the signing-key as JSON
	model := ModelSerialize{BrandID: "System", Name: "the-model", KeypairID: 1}
	data, _ := json.Marshal(model)

	sendRequestExpectError(t, "POST", "/v1/models", bytes.NewReader(data))
}

func TestModelCreateHandlerWithBase64Error(t *testing.T) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	// Define a model linked with the signing-key as JSON
	model := ModelSerialize{BrandID: "System", Name: "the-model", KeypairID: 1}
	data, _ := json.Marshal(model)

	sendRequestExpectError(t, "POST", "/v1/models", bytes.NewReader(data))
}

func TestModelCreateHandlerWithNilData(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}}

	sendRequestExpectError(t, "POST", "/v1/models", nil)
}

func TestModelCreateHandlerWithEmptyData(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}}

	sendRequestExpectError(t, "POST", "/v1/models", bytes.NewBufferString(""))
}

func TestModelCreateHandlerWithBadData(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}}

	sendRequestExpectError(t, "POST", "/v1/models", bytes.NewBufferString("bad"))
}

func sendRequest(t *testing.T, method, url string, data io.Reader) (ModelResponse, error) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	// Create a JWT and add it to the request
	jwtToken, err := createJWT()
	if err != nil {
		t.Errorf("Error creating a JWT: %v", err)
	}
	r.Header.Set("Authorization", "Bearer "+jwtToken)

	AdminRouter().ServeHTTP(w, r)

	// Check the JSON response
	result := ModelResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the model response: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.ErrorMessage)
	}

	return result, err
}

func sendRequestExpectError(t *testing.T, method, url string, data io.Reader) (ModelResponse, error) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)
	AdminRouter().ServeHTTP(w, r)

	// Check the JSON response
	result := ModelResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the model response: %v", err)
	}
	if result.Success {
		t.Error("Expected error, got success")
	}

	return result, err
}
