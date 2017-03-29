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

func TestUserIndexHandler(t *testing.T) {

	userIndexTemplate = "../static/app_user.html"

	config := ConfigSettings{Title: "Site Title", Logo: "/url"}
	Environ = &Env{Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	http.HandlerFunc(UserIndexHandler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got: %d", http.StatusOK, w.Code)
	}
}

func TestUserIndexHandlerInvalidTemplate(t *testing.T) {

	userIndexTemplate = "../static/does_not_exist.html"

	config := ConfigSettings{Title: "Site Title", Logo: "/url"}
	Environ = &Env{Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	http.HandlerFunc(UserIndexHandler).ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got: %d", http.StatusInternalServerError, w.Code)
	}
}

func generateSystemUserRequest() string {

	request := SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 1, Since: "20170324T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInvalidModel() string {

	request := SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 99, Since: "20170324T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInactiveModel() string {

	request := SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 2, Since: "20170324T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInvalidSince() string {

	request := SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 1, Since: "2024T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInvalidAssertion() string {

	request := SystemUserRequest{Email: "test", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 1, Since: "20170324T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func TestSystemUserAssertionHandler(t *testing.T) {
	type TestType struct {
		data       string
		statusCode int
		expected   bool
	}

	tests := []TestType{
		{generateSystemUserRequest(), 200, true},
		{"", 400, false},
		{"<invalid\\", 400, false},
		{generateSystemUserRequestInvalidModel(), 200, false},
		{generateSystemUserRequestInactiveModel(), 200, false},
		{generateSystemUserRequestInvalidAssertion(), 200, false},
		{generateSystemUserRequestInvalidSince(), 200, true},
	}

	for _, test := range tests {
		statusCode, result, message := sendSystemUserAssertion(test.data, t)
		if statusCode != test.statusCode {
			t.Errorf("Unexpected status code from request '%v': %d", test.data, statusCode)
		}
		if result != test.expected {
			t.Errorf("Unexpected result from request '%v': %s", test.data, message)
		}
	}
}

func sendSystemUserAssertion(request string, t *testing.T) (int, bool, string) {
	// Mock the database
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	Environ = &Env{DB: &mockDB{}, Config: config}
	Environ.KeypairDB, _ = GetKeyStore(config)

	// Submit the serial-request assertion for signing
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/assertions", bytes.NewBufferString(request))
	SystemUserRouter(Environ).ServeHTTP(w, r)

	// Check the JSON response
	t.Log(w.Body.String())
	result := SystemUserResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the version response: %v", err)
	}

	return w.Code, result.Success, result.ErrorMessage
}
