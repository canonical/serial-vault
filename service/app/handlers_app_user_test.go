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

package app_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/app"
)

func generateSystemUserRequest() string {
	request := app.SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 1, Since: "20170324T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInvalidModel() string {
	request := app.SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 99, Since: "20170324T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInactiveModel() string {
	request := app.SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 2, Since: "20170324T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInvalidSince() string {
	request := app.SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 1, Since: "2024T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInvalidAssertion() string {
	request := app.SystemUserRequest{Email: "test", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 1, Since: "20170324T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func TestSystemUserAssertionHandler(t *testing.T) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)

	service.MiddlewareWithCSRF = service.Middleware

	type TestType struct {
		data       string
		statusCode int
		expected   bool
	}

	tests := []TestType{
		{generateSystemUserRequest(), 200, true},
		{"", 400, false},
		{"<invalid\\", 400, false},
		{generateSystemUserRequestInvalidModel(), 400, false},
		{generateSystemUserRequestInactiveModel(), 400, false},
		{generateSystemUserRequestInvalidAssertion(), 400, false},
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
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)

	// Submit the serial-request assertion for signing
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/assertions", bytes.NewBufferString(request))
	service.AdminRouter().ServeHTTP(w, r)

	// Check the JSON response
	result := app.SystemUserResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the version response: %v", err)
		t.Log(w.Body.String())
	}

	return w.Code, result.Success, result.ErrorMessage
}
