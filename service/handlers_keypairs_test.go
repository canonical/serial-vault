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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/snapcore/snapd/asserts"
)

func TestKeypairListHandler(t *testing.T) {

	// Mock the database
	Environ = &Env{DB: &MockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/v1/keypairs", nil)
	http.HandlerFunc(KeypairListHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := KeypairsResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the keypairs response: %v", err)
	}
	if len(result.Keypairs) != 4 {
		t.Errorf("Expected 4 keypairs, got %d", len(result.Keypairs))
	}
	if result.Keypairs[0].KeyID != "61abf588e52be7a3" {
		t.Errorf("Expected key ID '61abf588e52be7a3', got %s", result.Keypairs[0].KeyID)
	}
}

func TestKeypairListHandlerWithError(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &ErrorMockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/v1/keypairs", nil)
	http.HandlerFunc(KeypairListHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := KeypairsResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the keypairs response: %v", err)
	}
	if result.Success {
		t.Error("Expected error, got success")
	}
}

func TestKeypairHandlerNilData(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs", nil)
	http.HandlerFunc(KeypairCreateHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the keypair response: %v", err)
	}
	if result.Success {
		t.Error("Expected an error, got success response")
	}
	if result.ErrorCode != "error-nil-data" {
		t.Errorf("Expected an 'nil data' message, got %s", result.ErrorCode)
	}
}

func TestKeypairHandlerNoData(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs", new(bytes.Buffer))
	http.HandlerFunc(KeypairCreateHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the keypair response: %v", err)
	}
	if result.Success {
		t.Error("Expected an error, got success response")
	}
	if result.ErrorCode != "error-keypair-data" {
		t.Errorf("Expected an 'no data' message, got %s", result.ErrorCode)
	}
}

func TestKeypairHandlerBadData(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs", bytes.NewBufferString("bad"))
	http.HandlerFunc(KeypairCreateHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the keypair response: %v", err)
	}
	if result.Success {
		t.Error("Expected an error, got success response")
	}
	if result.ErrorCode != "error-keypair-json" {
		t.Errorf("Expected an 'bad json' message, got %s", result.ErrorCode)
	}
}

func TestKeypairHandlerNoAuthorityID(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs", bytes.NewBufferString("{}"))
	http.HandlerFunc(KeypairCreateHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the keypair response: %v", err)
	}
	if result.Success {
		t.Error("Expected an error, got success response")
	}
	if result.ErrorCode != "error-keypair-json" {
		t.Errorf("Expected a 'bad keypair' message, got %s", result.ErrorCode)
	}
}

func TestKeypairHandlerBadPrivateKeyNotEncoded(t *testing.T) {
	signingKey, err := ioutil.ReadFile("../README.md")
	if err != nil {
		t.Errorf("Error reading the bad signing-key file: %v", err)
	}

	keypair := KeypairWithPrivateKey{PrivateKey: string(signingKey), AuthorityID: "System"}
	data, _ := json.Marshal(keypair)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs", bytes.NewReader(data))
	http.HandlerFunc(KeypairCreateHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the keypair response: %v", err)
	}
	if result.Success {
		t.Error("Expected an error, got success response")
	}
	if result.ErrorCode != "error-keypair-store" {
		t.Errorf("Expected a 'keypair-store' message, got %s", result.ErrorCode)
	}
}

func TestKeypairHandlerBadPrivateKeyEncoded(t *testing.T) {
	signingKey, err := ioutil.ReadFile("../README.md")
	if err != nil {
		t.Errorf("Error reading the bad signing-key file: %v", err)
	}
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	keypair := KeypairWithPrivateKey{PrivateKey: string(encodedSigningKey), AuthorityID: "System"}
	data, _ := json.Marshal(keypair)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs", bytes.NewReader(data))
	http.HandlerFunc(KeypairCreateHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the keypair response: %v", err)
	}
	if result.Success {
		t.Error("Expected an error, got success response")
	}
	if result.ErrorCode != "error-keypair-store" {
		t.Errorf("Expected a 'keypair-store' message, got %s", result.ErrorCode)
	}
}

func TestKeypairHandlerValidPrivateKey(t *testing.T) {
	// Mock the database and the keystore
	config := ConfigSettings{KeyStoreType: "memory"}
	Environ = &Env{DB: &MockDB{}, Config: config}
	Environ.KeypairDB, _ = getMemoryKeyStore(config)

	signingKey, err := ioutil.ReadFile("../keystore/TestKey.asc")
	if err != nil {
		t.Errorf("Error reading the signing-key file: %v", err)
	}
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	keypair := KeypairWithPrivateKey{PrivateKey: string(encodedSigningKey), AuthorityID: "System"}
	data, _ := json.Marshal(keypair)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs", bytes.NewReader(data))
	http.HandlerFunc(KeypairCreateHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the keypair response: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected an success, got error: %s", result.ErrorCode)
	}
}

func TestKeypairHandlerValidPrivateKeyKeyStoreError(t *testing.T) {
	// Mock the database and the keystore
	config := ConfigSettings{KeyStoreType: "memory"}
	Environ = &Env{DB: &MockDB{}, Config: config}
	Environ.KeypairDB, _ = getErrorMockKeyStore(config)

	signingKey, err := ioutil.ReadFile("../keystore/TestKey.asc")
	if err != nil {
		t.Errorf("Error reading the signing-key file: %v", err)
	}
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	keypair := KeypairWithPrivateKey{PrivateKey: string(encodedSigningKey), AuthorityID: "System"}
	data, _ := json.Marshal(keypair)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs", bytes.NewReader(data))
	http.HandlerFunc(KeypairCreateHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	json.NewDecoder(w.Body).Decode(&result)
	if result.Success {
		t.Error("Expected an error, got success response")
	}
	if result.ErrorCode != "error-keypair-store" {
		t.Errorf("Expected a 'keystore error' message, got %s", result.ErrorCode)
	}
}

func TestKeypairHandlerValidPrivateKeyDataStoreError(t *testing.T) {
	// Mock the database and the keystore
	config := ConfigSettings{KeyStoreType: "memory"}
	Environ = &Env{DB: &ErrorMockDB{}, Config: config}
	Environ.KeypairDB, _ = getMemoryKeyStore(config)

	signingKey, err := ioutil.ReadFile("../keystore/TestKey.asc")
	if err != nil {
		t.Errorf("Error reading the signing-key file: %v", err)
	}
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	keypair := KeypairWithPrivateKey{PrivateKey: string(encodedSigningKey), AuthorityID: "System"}
	data, _ := json.Marshal(keypair)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs", bytes.NewReader(data))
	http.HandlerFunc(KeypairCreateHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	json.NewDecoder(w.Body).Decode(&result)
	if result.Success {
		t.Error("Expected an error, got success response")
	}
}

func TestKeypairDisableHandler(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &MockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/1/disable", bytes.NewBufferString("{}"))
	AdminRouter(Environ).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Expected an success, got error: %v", err)
	}
	if !result.Success {
		t.Error("Expected an success, got fail response")
	}
}

func TestKeypairDisableHandlerError(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &ErrorMockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/1/disable", bytes.NewBufferString("{}"))
	AdminRouter(Environ).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Expected an success, got error: %v", err)
	}
	if result.Success {
		t.Error("Expected an failure, got success")
	}
	if result.ErrorCode != "error-keypair-update" {
		t.Errorf("Expected a 'keypair update' message, got %s", result.ErrorCode)
	}
}

func TestKeypairDisableHandlerBadID(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &ErrorMockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/9999999999999999999999999/disable", bytes.NewBufferString("{}"))
	AdminRouter(Environ).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Expected an success, got error: %v", err)
	}
	if result.Success {
		t.Error("Expected an failure, got success")
	}
	if result.ErrorCode != "error-invalid-keypair" {
		t.Errorf("Expected a 'invalid keypair' message, got %s", result.ErrorCode)
	}
}

func TestKeypairEnableHandler(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &MockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/1/enable", bytes.NewBufferString("{}"))
	AdminRouter(Environ).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Expected an success, got error: %v", err)
	}
	if !result.Success {
		t.Error("Expected an success, got fail response")
	}
}

func TestKeypairEnableHandlerError(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &ErrorMockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/1/enable", bytes.NewBufferString("{}"))
	AdminRouter(Environ).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Expected an success, got error: %v", err)
	}
	if result.Success {
		t.Error("Expected an failure, got success")
	}
	if result.ErrorCode != "error-keypair-update" {
		t.Errorf("Expected a 'keypair update' message, got %s", result.ErrorCode)
	}
}

func TestKeypairEnableHandlerBadID(t *testing.T) {
	// Mock the database
	Environ = &Env{DB: &ErrorMockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/9999999999999999999999999/enable", bytes.NewBufferString("{}"))
	AdminRouter(Environ).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Expected an success, got error: %v", err)
	}
	if result.Success {
		t.Error("Expected an failure, got success")
	}
	if result.ErrorCode != "error-invalid-keypair" {
		t.Errorf("Expected a 'keypair update' message, got %s", result.ErrorCode)
	}
}

func TestKeypairAssertionHandler(t *testing.T) {

	// Mock the database
	mockDatabase()

	// Create the account key assertion
	assertAcc, err := generateAccountAssertion(asserts.AccountKeyType, "alder", "maple-inc")
	if err != nil {
		t.Errorf("Error generating the assertion: %v", err)
	}

	// Encode the assertion and create the request
	encodedAssert := base64.StdEncoding.EncodeToString([]byte(assertAcc))
	request, err := json.Marshal(AssertionRequest{ID: 1, Assertion: encodedAssert})
	if err != nil {
		t.Errorf("Error marshalling the assertion to JSON: %v", err)
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/assertion", bytes.NewBuffer(request))
	http.HandlerFunc(KeypairAssertionHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the account key assertion response: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.ErrorMessage)
	}
}

func sendKeypairAssertionError(request []byte, t *testing.T) {

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/assertion", bytes.NewBuffer(request))
	http.HandlerFunc(KeypairAssertionHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the accounts response: %v", err)
	}
	if result.Success {
		t.Errorf("Expected failure, got success")
	}
}

func mockDatabase() {
	// Mock the database
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	Environ = &Env{DB: &MockDB{}, Config: config}
	Environ.KeypairDB, _ = GetKeyStore(config)
}

func TestKeypairAssertionWithErrors(t *testing.T) {

	mockDatabase()

	sendKeypairAssertionError(nil, t)
	sendKeypairAssertionError([]byte("InvalidData"), t)

	// Invalid encoding
	request, err := json.Marshal(AssertionRequest{ID: 1, Assertion: "InvalidData"})
	if err != nil {
		t.Errorf("Error marshalling the assertion to JSON: %v", err)
	}
	sendKeypairAssertionError(request, t)

	// Encode the assertion and create the request
	encodedAssert := base64.StdEncoding.EncodeToString([]byte("InvalidData"))
	request, err = json.Marshal(AssertionRequest{ID: 1, Assertion: encodedAssert})
	if err != nil {
		t.Errorf("Error marshalling the assertion to JSON: %v", err)
	}
	sendKeypairAssertionError(request, t)
}

func TestKeypairAssertionInvalidAssertionType(t *testing.T) {

	mockDatabase()

	// Encode the assertion and create the request (account instead of an account-key assertion)
	assertion, err := generateAccountAssertion(asserts.AccountType, "alder", "maple-inc")
	if err != nil {
		t.Errorf("Error generating the assertion: %v", err)
	}
	encodedAssert := base64.StdEncoding.EncodeToString([]byte(assertion))
	request, err := json.Marshal(AssertionRequest{ID: 1, Assertion: encodedAssert})
	if err != nil {
		t.Errorf("Error marshalling the assertion to JSON: %v", err)
	}
	sendKeypairAssertionError(request, t)
}

func TestKeypairAssertionInvalidID(t *testing.T) {

	mockDatabase()

	// Encode the assertion and create the request
	assertion, err := generateAccountAssertion(asserts.AccountKeyType, "alder", "maple-inc")
	if err != nil {
		t.Errorf("Error generating the assertion: %v", err)
	}
	encodedAssert := base64.StdEncoding.EncodeToString([]byte(assertion))
	request, err := json.Marshal(AssertionRequest{Assertion: encodedAssert})
	if err != nil {
		t.Errorf("Error marshalling the assertion to JSON: %v", err)
	}
	sendKeypairAssertionError(request, t)
}

func TestKeypairAssertionUpdateError(t *testing.T) {

	// Mock the database
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	Environ = &Env{DB: &ErrorMockDB{}, Config: config}
	Environ.KeypairDB, _ = GetKeyStore(config)

	// Encode the assertion and create the request
	assertion, err := generateAccountAssertion(asserts.AccountKeyType, "alder", "maple-inc")
	if err != nil {
		t.Errorf("Error generating the assertion: %v", err)
	}
	encodedAssert := base64.StdEncoding.EncodeToString([]byte(assertion))
	request, err := json.Marshal(AssertionRequest{ID: 1, Assertion: encodedAssert})
	if err != nil {
		t.Errorf("Error marshalling the assertion to JSON: %v", err)
	}
	sendKeypairAssertionError(request, t)
}
