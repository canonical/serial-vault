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
		t.Error("Expected error, got sucess")
	}
}
