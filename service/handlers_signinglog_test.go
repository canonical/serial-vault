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

	"github.com/CanonicalLtd/serial-vault/datastore"
)

func TestSigningLogListHandler(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}

	response, _ := sendSigningLogRequest(t, "GET", "/v1/signinglog", nil)
	if len(response.SigningLog) != 10 {
		t.Errorf("Expected 10 signing logs, got: %d", len(response.SigningLog))
	}
}

func TestSigningLogListHandlerError(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}}

	sendSigningLogRequestExpectError(t, "GET", "/v1/signinglog", nil)
}

func TestSigningLogDeleteHandler(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}

	// Delete a signing log
	data := "{}"
	sendSigningLogRequest(t, "DELETE", "/v1/signinglog/1", bytes.NewBufferString(data))
}

func TestSigningLogDeleteHandlerWrongID(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}

	// Delete a signing log
	data := "{}"
	sendSigningLogRequestExpectError(t, "DELETE", "/v1/signinglog/22", bytes.NewBufferString(data))
}

func TestSigningLogDeleteHandlerError(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}}

	// Delete a signing log
	data := "{}"
	sendSigningLogRequestExpectError(t, "DELETE", "/v1/signinglog/1", bytes.NewBufferString(data))
}

func TestSigningLogDeleteHandlerBadID(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}}

	// Delete a signing log
	data := "{}"
	sendSigningLogRequestExpectError(t, "DELETE", "/v1/signinglog/99999999999999999999999999999999999999999999999", bytes.NewBufferString(data))
}

func TestSigningLogFilterValues(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/v1/signinglog/filters", nil)
	AdminRouter().ServeHTTP(w, r)

	// Check the JSON response
	result := SigningLogFiltersResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the signing log filters response: %v", err)
	}
	if !result.Success {
		t.Error("Expected success, got error")
	}
}

func TestSigningLogFilterValuesError(t *testing.T) {
	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/v1/signinglog/filters", nil)
	AdminRouter().ServeHTTP(w, r)

	// Check the JSON response
	result := SigningLogFiltersResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the signing log filters response: %v", err)
	}
	if result.Success {
		t.Error("Expected error, got success")
	}
}

func sendSigningLogRequest(t *testing.T, method, url string, data io.Reader) (SigningLogResponse, error) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)
	AdminRouter().ServeHTTP(w, r)

	// Check the JSON response
	result := SigningLogResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the signing log response: %v", err)
	}
	if !result.Success {
		t.Error("Expected success, got error")
	}

	return result, err
}

func sendSigningLogRequestExpectError(t *testing.T, method, url string, data io.Reader) (SigningLogResponse, error) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)
	AdminRouter().ServeHTTP(w, r)

	// Check the JSON response
	result := SigningLogResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the signing log response: %v", err)
	}
	if result.Success {
		t.Error("Expected error, got success")
	}

	return result, err
}
