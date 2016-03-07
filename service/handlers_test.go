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
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignHandlerNilData(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/1.0/sign", nil)
	http.HandlerFunc(SignHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := SignResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the signed response: %v", err)
	}
	if result.Success {
		t.Error("Expected an error, got success response")
	}
}

func TestSignHandlerNoData(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/1.0/sign", new(bytes.Buffer))
	http.HandlerFunc(SignHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := SignResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the signed response: %v", err)
	}
	if result.Success {
		t.Error("Expected an error, got success response")
	}
}

func TestSignHandler(t *testing.T) {
	// Mock the database
	config := ConfigSettings{PrivateKeyPath: "../TestKey.asc"}
	Environ = &Env{DB: &mockDB{}, Config: config}

	const assertions = `
  {
	  "brand-id": "System",
    "model":"聖誕快樂",
    "serial":"A1234/L",
		"revision": 2,
    "device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
  }`

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/1.0/sign", bytes.NewBufferString(assertions))
	http.HandlerFunc(SignHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := SignResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the signed response: %v", err)
	}
	if !result.Success {
		t.Errorf("Error generated in signing the device: %s", result.ErrorMessage)
	}
	if result.Signature == "" {
		t.Errorf("Empty signed data returned.")
	}
}

func TestSignHandlerBadJson(t *testing.T) {
	const assertions = `
  {
	  "bad json"
  }`

	config := ConfigSettings{PrivateKeyPath: "../TestKey.asc"}
	Environ = &Env{Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/1.0/sign", bytes.NewBufferString(assertions))
	http.HandlerFunc(SignHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := SignResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the signed response: %v", err)
	}
	if result.Success {
		t.Error("Expected failure when sending invalid JSON, got success")
	}
}

func TestSignHandlerBadAssertion(t *testing.T) {
	const assertions = `
  {
	  "brand-id": "System",
    "model": 999
    "serial":"A1234/L",
		"revision": "This should be numeric",
    "device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
  }`

	config := ConfigSettings{PrivateKeyPath: "../TestKey.asc"}
	Environ = &Env{Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/1.0/sign", bytes.NewBufferString(assertions))
	http.HandlerFunc(SignHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := SignResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the signed response: %v", err)
	}
	if result.Success {
		t.Error("Expected failure when sending invalid JSON, got success")
	}
}

func TestSignHandlerBadPrivateKeyPath(t *testing.T) {
	// Mock the database using an incorrect signing-key (invalid path)
	config := ConfigSettings{PrivateKeyPath: "Not a good path"}
	Environ = &Env{DB: &errorMockDB{}, Config: config}

	const assertions = `
  {
	  "brand-id": "System",
    "model":"Bad Path",
    "serial":"A1234/L",
		"revision": 2,
    "device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
  }`

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/1.0/sign", bytes.NewBufferString(assertions))
	http.HandlerFunc(SignHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := SignResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the signed response: %v", err)
	}
	if result.Success {
		t.Error("Expected failure with an invalid private key path, got success")
	}
}

func TestSignHandlerBadPrivateKeyFile(t *testing.T) {
	// Mock the database using an incorrect signing-key (README.md)
	Environ = &Env{DB: &errorMockDB{}}

	const assertions = `
  {
	  "brand-id": "System",
    "model":"聖誕快樂",
    "serial":"A1234/L",
		"revision": 2,
    "device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
  }`

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/1.0/sign", bytes.NewBufferString(assertions))
	http.HandlerFunc(SignHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := SignResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the signed response: %v", err)
	}
	if result.Success {
		t.Error("Expected failure with an invalid private key file, got success")
	}
}

func TestSignHandlerNonExistentModel(t *testing.T) {
	// Mock the database, ot finding the model
	Environ = &Env{DB: &errorMockDB{}}

	const assertions = `
  {
	  "brand-id": "System",
    "model":"Cannot Find This",
    "serial":"A1234/L",
		"revision": 2,
    "device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
  }`

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/1.0/sign", bytes.NewBufferString(assertions))
	http.HandlerFunc(SignHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := SignResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the signed response: %v", err)
	}
	if result.Success {
		t.Error("Expected failure with an invalid model, got success")
	}
}

func TestVersionHandler(t *testing.T) {

	config := ConfigSettings{Version: "1.2.5"}
	Environ = &Env{Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/1.0/version", nil)
	http.HandlerFunc(VersionHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := VersionResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the version response: %v", err)
	}
	if result.Version != Environ.Config.Version {
		t.Errorf("Incorrect version returned. Expected '%s' got: %v", Environ.Config.Version, result.Version)
	}

}

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

	sendRequestExpectError(t, "GET", "/1.0/models", nil)
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
    "serial":"A1234/L",
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

func TestModelCreateHandler(t *testing.T) {
	// Mock the database
	config := ConfigSettings{PrivateKeyPath: "../TestKey.asc", KeyStoreType: "filesystem"}
	Environ = &Env{DB: &mockDB{}, Config: config}

	// Read the test signing-key file
	signingKey, err := ioutil.ReadFile(config.PrivateKeyPath)
	if err != nil {
		t.Errorf("Error reading the test signing-key: %v", err)
	}

	// base64 encode the signing-key
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	// Create a model with the signing-key and convert it to JSON
	model := ModelWithKey{BrandID: "System", Name: "聖誕快樂", Revision: 2, SigningKey: string(encodedSigningKey)}
	data, err := json.Marshal(model)

	result, err := sendRequest(t, "POST", "/1.0/models", bytes.NewReader(data))
	if result.Model.ID != 7 {
		t.Errorf("Expected model with ID 7, got %d", result.Model.ID)
	}
	if result.Model.Name != "聖誕快樂" {
		t.Errorf("Expected model name '聖誕快樂', got %s", result.Model.Name)
	}
}

func TestModelCreateHandlerWithError(t *testing.T) {
	// Mock the database
	config := ConfigSettings{PrivateKeyPath: "../TestKey.asc", KeyStoreType: "filesystem"}
	Environ = &Env{DB: &errorMockDB{}, Config: config}

	// Read the test signing-key file
	signingKey, err := ioutil.ReadFile(config.PrivateKeyPath)
	if err != nil {
		t.Errorf("Error reading the test signing-key: %v", err)
	}

	// base64 encode the signing-key
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	// Create a model with the signing-key and convert it to JSON
	model := ModelWithKey{BrandID: "System", Name: "聖誕快樂", Revision: 2, SigningKey: string(encodedSigningKey)}
	data, err := json.Marshal(model)

	sendRequestExpectError(t, "POST", "/1.0/models", bytes.NewReader(data))
}

func TestModelCreateHandlerWithBase64Error(t *testing.T) {
	// Mock the database
	config := ConfigSettings{PrivateKeyPath: "../TestKey.asc", KeyStoreType: "filesystem"}
	Environ = &Env{DB: &errorMockDB{}, Config: config}

	// Read the test signing-key file
	signingKey, err := ioutil.ReadFile(config.PrivateKeyPath)
	if err != nil {
		t.Errorf("Error reading the test signing-key: %v", err)
	}

	// Create a model with the signing-key and convert it to JSON (no base64 encoding)
	model := ModelWithKey{BrandID: "System", Name: "聖誕快樂", Revision: 2, SigningKey: string(signingKey)}
	data, err := json.Marshal(model)

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
	Router(Environ).ServeHTTP(w, r)

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
	Router(Environ).ServeHTTP(w, r)

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
