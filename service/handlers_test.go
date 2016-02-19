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

var sign_cases = []struct {
	assertions        string
	mockDB            string
	expectedSuccess   bool
	expectedError     string
	expectedSignature bool
}{
	{`{"brand-id": "System",
    "model":"聖誕快樂",
    "serial":"A1234/L",
	"revision": 2,
    "device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"}`, "MockDB", true, "", true},
	{`{"bad json"}`, "MockDB", false, "Error decoding JSON: invalid character '}' after object key", false},
	{`
  	{
	"brand-id": "System",
    "model": 999
    "serial":"A1234/L",
	"revision": "This should be numeric",
    "device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"}`, "MockDB", false, `Error decoding JSON: invalid character '"' after object key:value pair`, false},
	{`

  {
	"brand-id": "System",
    "model":"Bad Path",
    "serial":"A1234/L",
	"revision": 2,
    "device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
}`, "ErrorMockDB", false, `Error reading the private key: open not a good path: no such file or directory`, false},
	{`

  {
	"brand-id": "System",
    "model":"聖誕快樂",
    "serial":"A1234/L",
	"revision": 2,
    "device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
}`, "ErrorMockDB", false, `Error signing the assertions: openpgp: invalid argument: no armored data found`, false},
	{`
  {
	"brand-id": "System",
    "model":"Cannot Find This",
    "serial":"A1234/L",
	"revision": 2,
    "device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
  }`, "ErrorMockDB", false, `Cannot find model with the matching brand, model and revision.`, false},
}

func TestSignHandlerGeneric(t *testing.T) {

	var result SignResponse
	var w *httptest.ResponseRecorder
	var r *http.Request
	var err error

	for _, tt := range sign_cases {

		if tt.mockDB == "MockDB" {
			Environ = &Env{DB: &mockDB{}}
		}
		if tt.mockDB == "ErrorMockDB" {
			Environ = &Env{DB: &errorMockDB{}}
		}
		result = SignResponse{}
		w = httptest.NewRecorder()
		r, _ = http.NewRequest("POST", "/1.0/sign", bytes.NewBufferString(tt.assertions))

		http.HandlerFunc(SignHandler).ServeHTTP(w, r)

		// Check the JSON response
		err = json.NewDecoder(w.Body).Decode(&result)

		if err != nil {
			t.Errorf("Error decoding the signed response: %v", err)
		}

		if result.Success != tt.expectedSuccess {
			t.Errorf("Success. Expected: %t; got: %t", tt.expectedSuccess, result.Success)
		}

		if result.ErrorMessage != tt.expectedError {
			t.Errorf("Error message. Expected: %s; got: %s", tt.expectedError, result.ErrorMessage)
		}

		if result.Signature != "" != tt.expectedSignature {
			t.Errorf("Non-empty signature: Expected: %t; got: %t", tt.expectedSignature, result.Signature != "")
		}
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

var models_cases = []struct {
	mockDB         string
	success        bool
	numberModels   int
	firstModelName string
}{
	{"MockDB", true, 6, "Alder"},
	{"ErrorMockDB", false, 0, ""},
}

func TestModelsHandlerGeneric(t *testing.T) {

	// Mock the database
	var result ModelsResponse
	var w *httptest.ResponseRecorder
	var r *http.Request
	var err error

	for _, tt := range models_cases {
		if tt.mockDB == "MockDB" {
			Environ = &Env{DB: &mockDB{}}
		}
		if tt.mockDB == "ErrorMockDB" {
			Environ = &Env{DB: &errorMockDB{}}
		}

		w = httptest.NewRecorder()
		r, _ = http.NewRequest("GET", "/1.0/models", nil)
		http.HandlerFunc(ModelsHandler).ServeHTTP(w, r)

		// Check the JSON response
		result = ModelsResponse{}
		err = json.NewDecoder(w.Body).Decode(&result)
		if err != nil {
			t.Errorf("Error decoding the models response: %v", err)
		}
		if result.Success != tt.success {
			t.Errorf("Expected success: %t; got: %t", tt.success, result.Success)
		}
		if len(result.Models) != tt.numberModels {
			t.Errorf("Expected number of models: %d; got: %d", tt.numberModels, len(result.Models))
		}
		if len(result.Models) > 0 && result.Models[0].Name != tt.firstModelName {
			t.Errorf("Expected model name: %s; got: %s", result.Models[0].Name)
		}
	}
}
