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
)

func TestModelsHandler(t *testing.T) {

	// Mock the database
	Environ = &Env{DB: &mockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/1.0/models", nil)
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
	if result.Models[0].Name != "Alder" {
		t.Errorf("Expected model name 'Alder', got %s", result.Models[0].Name)
	}
}

func TestModelsHandlerWithError(t *testing.T) {

	// Mock the database
	Environ = &Env{DB: &errorMockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/1.0/models", nil)
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
	Environ = &Env{DB: &mockDB{}}

	result, _ := sendRequest(t, "GET", "/1.0/models/1", nil)

	if result.Model.ID != 1 {
		t.Errorf("Expected model with ID 1, got %d", result.Model.ID)
	}
	if result.Model.Name != "Alder" {
		t.Errorf("Expected model name 'Alder', got %s", result.Model.Name)
	}
}

func TestModelGetHandlerWithError(t *testing.T) {

	// Mock the database
	Environ = &Env{DB: &mockDB{}}

	sendRequestExpectError(t, "GET", "/1.0/models/999999", nil)
}

func TestModelGetHandlerWithBadID(t *testing.T) {

	// Mock the database
	Environ = &Env{DB: &mockDB{}}

	sendRequestExpectError(t, "GET", "/1.0/models/999999999999999999999999999999", nil)
}

func TestModelUpdateHandler(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &mockDB{}}

	// Update a model
	data := `
	{
	  "id": 1,
	  "brand-id": "System",
    "model":"聖誕快樂",
    "serial":"A1234-L",
		"revision": 2,
    "device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
  }`

	result, _ := sendRequest(t, "PUT", "/1.0/models/1", bytes.NewBufferString(data))

	if result.Model.ID != 1 {
		t.Errorf("Expected model with ID 1, got %d", result.Model.ID)
	}
	if result.Model.Name != "聖誕快樂" {
		t.Errorf("Expected model name '聖誕快樂', got %s", result.Model.Name)
	}
}

func TestModelUpdateHandlerWithErrors(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &errorMockDB{}}

	// Update a model
	data := `{}`

	sendRequestExpectError(t, "PUT", "/1.0/models/1", bytes.NewBufferString(data))
}

func TestModelUpdateHandlerWithNilData(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &errorMockDB{}}

	sendRequestExpectError(t, "PUT", "/1.0/models/1", nil)
}

func TestModelUpdateHandlerWithEmptyData(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &errorMockDB{}}

	sendRequestExpectError(t, "PUT", "/1.0/models/1", bytes.NewBufferString(""))
}

func TestModelUpdateHandlerWithBadData(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &errorMockDB{}}

	sendRequestExpectError(t, "PUT", "/1.0/models/1", bytes.NewBufferString("bad"))
}

func TestModelUpdateHandlerWithBadID(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &errorMockDB{}}

	sendRequestExpectError(t, "PUT", "/1.0/models/999999999999999999999999999999", bytes.NewBufferString("bad"))
}

func TestModelDeleteHandler(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &mockDB{}}

	// Delete a model
	data := "{}"
	sendRequest(t, "DELETE", "/1.0/models/1", bytes.NewBufferString(data))
}

func TestModelDeleteHandlerWithErrors(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &errorMockDB{}}

	// Delete a model
	data := `{}`

	sendRequestExpectError(t, "DELETE", "/1.0/models/1", bytes.NewBufferString(data))
}

func TestModelDeleteHandlerWithBadID(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &errorMockDB{}}

	sendRequestExpectError(t, "DELETE", "/1.0/models/999999999999999999999999999999", bytes.NewBufferString("bad"))
}

func TestModelCreateHandler(t *testing.T) {
	// Mock the database
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore"}
	Environ = &Env{DB: &mockDB{}, Config: config}

	// Define a model linked with the signing-key as JSON
	model := ModelSerialize{BrandID: "System", Name: "聖誕快樂", Revision: 2, KeypairID: 1}
	data, _ := json.Marshal(model)

	result, _ := sendRequest(t, "POST", "/1.0/models", bytes.NewReader(data))
	if result.Model.ID != 7 {
		t.Errorf("Expected model with ID 7, got %d", result.Model.ID)
	}
	if result.Model.Name != "聖誕快樂" {
		t.Errorf("Expected model name '聖誕快樂', got %s", result.Model.Name)
	}
}

func TestModelCreateHandlerWithError(t *testing.T) {
	// Mock the database
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore"}
	Environ = &Env{DB: &errorMockDB{}, Config: config}

	// Define a model linked with the signing-key as JSON
	model := ModelSerialize{BrandID: "System", Name: "聖誕快樂", Revision: 2, KeypairID: 1}
	data, _ := json.Marshal(model)

	sendRequestExpectError(t, "POST", "/1.0/models", bytes.NewReader(data))
}

func TestModelCreateHandlerWithBase64Error(t *testing.T) {
	// Mock the database
	config := ConfigSettings{KeyStoreType: "filesystem"}
	Environ = &Env{DB: &errorMockDB{}, Config: config}

	// Define a model linked with the signing-key as JSON
	model := ModelSerialize{BrandID: "System", Name: "聖誕快樂", Revision: 2, KeypairID: 1}
	data, _ := json.Marshal(model)

	sendRequestExpectError(t, "POST", "/1.0/models", bytes.NewReader(data))
}

func TestModelCreateHandlerWithNilData(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &errorMockDB{}}

	sendRequestExpectError(t, "POST", "/1.0/models", nil)
}

func TestModelCreateHandlerWithEmptyData(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &errorMockDB{}}

	sendRequestExpectError(t, "POST", "/1.0/models", bytes.NewBufferString(""))
}

func TestModelCreateHandlerWithBadData(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &errorMockDB{}}

	sendRequestExpectError(t, "POST", "/1.0/models", bytes.NewBufferString("bad"))
}

func sendRequest(t *testing.T, method, url string, data io.Reader) (ModelResponse, error) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)
	AdminRouter(Environ).ServeHTTP(w, r)

	// Check the JSON response
	result := ModelResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
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
	AdminRouter(Environ).ServeHTTP(w, r)

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
