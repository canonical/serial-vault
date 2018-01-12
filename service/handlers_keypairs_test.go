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

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/snapcore/snapd/asserts"
)

func TestKeypairListHandler(t *testing.T) {

	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

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
	if result.Keypairs[0].KeyID != "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO" {
		t.Errorf("Expected key ID 'UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO', got %s", result.Keypairs[0].KeyID)
	}
}

func TestKeypairListHandlerWithPermissions(t *testing.T) {

	// Mock the database
	c := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/v1/keypairs", nil)

	// Create a JWT and add it to the request
	err := createJWTWithRole(r, datastore.Admin)
	if err != nil {
		t.Errorf("Error creating a JWT: %v", err)
	}

	http.HandlerFunc(KeypairListHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := KeypairsResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the keypairs response: %v", err)
	}
	if len(result.Keypairs) != 2 {
		t.Errorf("Expected 2 keypairs, got %d", len(result.Keypairs))
	}
	if result.Keypairs[0].KeyID != "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO" {
		t.Errorf("Expected key ID 'UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO', got %s", result.Keypairs[0].KeyID)
	}
}

func TestKeypairListHandlerWithoutPermissions(t *testing.T) {

	// Mock the database
	c := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/v1/keypairs", nil)

	// Create a JWT and add it to the request
	err := createJWTWithRole(r, datastore.Standard)
	if err != nil {
		t.Errorf("Error creating a JWT: %v", err)
	}

	http.HandlerFunc(KeypairListHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := KeypairsResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the keypairs response: %v", err)
	}
	if result.Success {
		t.Error("Expected failure, got success")
	}
	if result.ErrorCode != "error-auth" {
		t.Error("Expected error-auth code")
	}
}

func TestKeypairListHandlerWithError(t *testing.T) {
	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

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

func TestKeypairHandlerWithPermissions(t *testing.T) {

	// Mock the database and the keystore
	config := config.Settings{KeyStoreType: "memory", EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.Environ.KeypairDB, _ = datastore.GetMemoryKeyStore(config)

	signingKey, err := ioutil.ReadFile("../keystore/TestKey.asc")
	if err != nil {
		t.Errorf("Error reading the signing-key file: %v", err)
	}
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	keypair := KeypairWithPrivateKey{PrivateKey: string(encodedSigningKey), AuthorityID: "System"}
	data, _ := json.Marshal(keypair)

	// Check the JSON response
	result, err := sendKeypairCreate(t, bytes.NewReader(data))
	if err != nil {
		t.Errorf("Error decoding the keypair response: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected an success, got error: %s", result.ErrorCode)
	}
}

func TestKeypairHandlerWithoutPermissions(t *testing.T) {

	// Mock the database and the keystore
	config := config.Settings{KeyStoreType: "memory", EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.Environ.KeypairDB, _ = datastore.GetMemoryKeyStore(config)

	signingKey, err := ioutil.ReadFile("../keystore/TestKey.asc")
	if err != nil {
		t.Errorf("Error reading the signing-key file: %v", err)
	}
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	keypair := KeypairWithPrivateKey{PrivateKey: string(encodedSigningKey), AuthorityID: "System"}
	data, _ := json.Marshal(keypair)

	// Check the JSON response
	result, err := sendKeypairCreateWithoutPermissions(t, bytes.NewReader(data))
	if err != nil {
		t.Errorf("Error decoding the keypair response: %v", err)
	}
	if result.Success {
		t.Error("Expected failure, got success")
	}
	if result.ErrorCode != "error-auth" {
		t.Error("Expected error-auth code")
	}
}

func TestKeypairHandlerNilData(t *testing.T) {
	// Mock the database and the keystore
	config := config.Settings{KeyStoreType: "memory", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.Environ.KeypairDB, _ = datastore.GetMemoryKeyStore(config)

	// Check the JSON response
	result, err := sendKeypairCreate(t, nil)
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
	// Mock the database and the keystore
	config := config.Settings{KeyStoreType: "memory", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.Environ.KeypairDB, _ = datastore.GetMemoryKeyStore(config)

	// Check the JSON response
	result, err := sendKeypairCreate(t, new(bytes.Buffer))
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
	// Mock the database and the keystore
	config := config.Settings{KeyStoreType: "memory", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.Environ.KeypairDB, _ = datastore.GetMemoryKeyStore(config)

	// Check the JSON response
	result, err := sendKeypairCreate(t, bytes.NewBufferString("bad"))
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
	// Mock the database and the keystore
	config := config.Settings{KeyStoreType: "memory", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.Environ.KeypairDB, _ = datastore.GetMemoryKeyStore(config)

	// Check the JSON response
	result, err := sendKeypairCreate(t, bytes.NewBufferString("{}"))
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
	// Mock the database and the keystore
	config := config.Settings{KeyStoreType: "memory", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.Environ.KeypairDB, _ = datastore.GetMemoryKeyStore(config)

	signingKey, err := ioutil.ReadFile("../README.md")
	if err != nil {
		t.Errorf("Error reading the bad signing-key file: %v", err)
	}

	keypair := KeypairWithPrivateKey{PrivateKey: string(signingKey), AuthorityID: "System"}
	data, _ := json.Marshal(keypair)

	// Check the JSON response
	result, err := sendKeypairCreate(t, bytes.NewReader(data))
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
	// Mock the database and the keystore
	config := config.Settings{KeyStoreType: "memory", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.Environ.KeypairDB, _ = datastore.GetMemoryKeyStore(config)

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
	result, err := sendKeypairCreate(t, bytes.NewReader(data))
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
	config := config.Settings{KeyStoreType: "memory", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.Environ.KeypairDB, _ = datastore.GetMemoryKeyStore(config)

	signingKey, err := ioutil.ReadFile("../keystore/TestKey.asc")
	if err != nil {
		t.Errorf("Error reading the signing-key file: %v", err)
	}
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	keypair := KeypairWithPrivateKey{PrivateKey: string(encodedSigningKey), AuthorityID: "System"}
	data, _ := json.Marshal(keypair)

	// Check the JSON response
	result, err := sendKeypairCreate(t, bytes.NewReader(data))
	if err != nil {
		t.Errorf("Error decoding the keypair response: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected an success, got error: %s", result.ErrorCode)
	}
}

func TestKeypairHandlerValidPrivateKeyKeyStoreError(t *testing.T) {
	// Mock the database and the keystore
	config := config.Settings{KeyStoreType: "memory", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.Environ.KeypairDB, _ = datastore.GetErrorMockKeyStore(config)

	signingKey, err := ioutil.ReadFile("../keystore/TestKey.asc")
	if err != nil {
		t.Errorf("Error reading the signing-key file: %v", err)
	}
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	keypair := KeypairWithPrivateKey{PrivateKey: string(encodedSigningKey), AuthorityID: "System"}
	data, _ := json.Marshal(keypair)

	// Send the request and check the JSON response
	result, err := sendKeypairCreate(t, bytes.NewReader(data))
	if err != nil {
		t.Errorf("Error decoding the keypair response: %v", err)
	}
	if result.Success {
		t.Error("Expected an error, got success response")
	}
	if result.ErrorCode != "error-keypair-store" {
		t.Errorf("Expected a 'keystore error' message, got %s", result.ErrorCode)
	}
}

func sendKeypairCreate(t *testing.T, body io.Reader) (BooleanResponse, error) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs", body)

	// Create a JWT and add it to the request
	err := createJWTWithRole(r, datastore.Admin)
	if err != nil {
		t.Errorf("Error creating a JWT: %v", err)
	}

	http.HandlerFunc(KeypairCreateHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func sendKeypairCreateWithoutPermissions(t *testing.T, body io.Reader) (BooleanResponse, error) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs", body)

	// Create a JWT and add it to the request
	err := createJWTWithRole(r, datastore.Standard)
	if err != nil {
		t.Errorf("Error creating a JWT: %v", err)
	}

	http.HandlerFunc(KeypairCreateHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func TestKeypairHandlerValidPrivateKeyDataStoreError(t *testing.T) {
	// Mock the database and the keystore
	config := config.Settings{KeyStoreType: "memory", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}
	datastore.Environ.KeypairDB, _ = datastore.GetMemoryKeyStore(config)

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
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/1/disable", bytes.NewBufferString("{}"))
	AdminRouter().ServeHTTP(w, r)

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

func TestKeypairDisableHandlerWithPermissions(t *testing.T) {
	// Mock the database
	config := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/1/disable", bytes.NewBufferString("{}"))

	// Create a JWT and add it to the request
	err := createJWTWithRole(r, datastore.Admin)
	if err != nil {
		t.Errorf("Error creating a JWT: %v", err)
	}

	AdminRouter().ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Expected an success, got error: %v", err)
	}
	if !result.Success {
		t.Error("Expected an success, got fail response")
	}
}

func TestKeypairDisableHandlerWithoutPermissions(t *testing.T) {
	// Mock the database
	config := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/1/disable", bytes.NewBufferString("{}"))
	AdminRouter().ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Expected a success, got error: %v", err)
	}
	if result.Success {
		t.Error("Expected a fail, got success response")
	}
	if result.ErrorCode != "error-auth" {
		t.Error("Expected error-auth code")
	}
}

func TestKeypairDisableHandlerError(t *testing.T) {
	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/1/disable", bytes.NewBufferString("{}"))
	AdminRouter().ServeHTTP(w, r)

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
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/9999999999999999999999999/disable", bytes.NewBufferString("{}"))
	AdminRouter().ServeHTTP(w, r)

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
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/1/enable", bytes.NewBufferString("{}"))
	AdminRouter().ServeHTTP(w, r)

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

func TestKeypairEnableHandlerWithPermissions(t *testing.T) {
	// Mock the database
	config := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/1/enable", bytes.NewBufferString("{}"))

	// Create a JWT and add it to the request
	err := createJWTWithRole(r, datastore.Admin)
	if err != nil {
		t.Errorf("Error creating a JWT: %v", err)
	}

	AdminRouter().ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Expected an success, got error: %v", err)
	}
	if !result.Success {
		t.Error("Expected an success, got fail response")
	}
}

func TestKeypairEnableHandlerWithoutPermissions(t *testing.T) {
	// Mock the database
	config := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/1/enable", bytes.NewBufferString("{}"))
	AdminRouter().ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Expected a success, got error: %v", err)
	}
	if result.Success {
		t.Error("Expected a fail, got success response")
	}
	if result.Success {
		t.Error("Expected failure, got success")
	}
	if result.ErrorCode != "error-auth" {
		t.Error("Expected error-auth code")
	}
}

func TestKeypairEnableHandlerError(t *testing.T) {
	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/1/enable", bytes.NewBufferString("{}"))
	AdminRouter().ServeHTTP(w, r)

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
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/keypairs/9999999999999999999999999/enable", bytes.NewBufferString("{}"))
	AdminRouter().ServeHTTP(w, r)

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
	datastore.Environ.Config.JwtSecret = "SomeTestSecretValue"

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

func TestKeypairAssertionHandlerWithPermissions(t *testing.T) {

	// Mock the database
	mockDatabase()
	datastore.Environ.Config.EnableUserAuth = true
	datastore.Environ.Config.JwtSecret = "SomeTestSecretValue"

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

	// Create a JWT and add it to the request
	err = createJWTWithRole(r, datastore.Admin)
	if err != nil {
		t.Errorf("Error creating a JWT: %v", err)
	}

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

func TestKeypairAssertionHandlerWithoutPermissions(t *testing.T) {

	// Mock the database
	mockDatabase()
	datastore.Environ.Config.EnableUserAuth = true
	datastore.Environ.Config.JwtSecret = "SomeTestSecretValue"

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
	if result.Success {
		t.Error("Expected failure, got success")
	}
	if result.ErrorCode != "error-auth" {
		t.Error("Expected error-auth code")
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
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)
}

func TestKeypairAssertionWithErrors(t *testing.T) {

	mockDatabase()
	datastore.Environ.Config.JwtSecret = "SomeTestSecretValue"

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
	datastore.Environ.Config.JwtSecret = "SomeTestSecretValue"

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
	datastore.Environ.Config.JwtSecret = "SomeTestSecretValue"

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
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}
	datastore.OpenKeyStore(config)

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

func TestKeypairStatusProgressHandlerWithPermissions(t *testing.T) {
	// Mock the database
	c := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	// List the keypair status records
	data := ""
	sendRequest(t, "GET", "/v1/keypairs/status", bytes.NewBufferString(data))
}

func TestKeypairStatusProgressHandlerWithoutPermissions(t *testing.T) {
	// Mock the database
	c := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	// List the keypair status records
	data := ""
	sendRequestWithoutPermissions(t, "GET", "/v1/keypairs/status", bytes.NewBufferString(data))
}

func TestKeypairStatusHandlerWithPermissions(t *testing.T) {
	// Mock the database
	c := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	// List the keypair status records
	data := ""
	sendRequest(t, "GET", "/v1/keypairs/status/system/key1", bytes.NewBufferString(data))
}

func TestKeypairStatusHandlerWithoutPermissions(t *testing.T) {
	// Mock the database
	c := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	// List the keypair status records
	data := ""
	sendRequestWithoutPermissions(t, "GET", "/v1/keypairs/status/system/key1", bytes.NewBufferString(data))
}
