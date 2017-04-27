// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bytes"

	"encoding/base64"

	"github.com/snapcore/snapd/asserts"
)

func generateAccountAssertion(assertType *asserts.AssertionType, accountID, username string) (string, error) {
	privateKey, _ := generatePrivateKey()

	headers := map[string]interface{}{
		"authority-id": "system",
		"account-id":   accountID,
		"username":     username,
		"display-name": username,
		"mail":         "test@example.com",
		"revision":     "1",
		"timestamp":    "2016-01-02T15:04:05Z",
		"validation":   "unproven",
	}

	var body []byte

	switch assertType {
	case asserts.AccountType:
		headers["sign-key-sha3-384"] = privateKey.PublicKey().ID()
		body = nil
	case asserts.AccountKeyType:
		headers["public-key-sha3-384"] = privateKey.PublicKey().ID()
		headers["since"] = "2016-01-02T15:04:05Z"
		body, _ = asserts.EncodePublicKey(privateKey.PublicKey())
	}

	accAssert, err := Environ.KeypairDB.Sign(assertType, headers, body, "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO")
	if err != nil {
		return "", err
	}

	assertAcc := asserts.Encode(accAssert)
	return string(assertAcc), nil
}

func TestAccountsHandler(t *testing.T) {

	// Mock the database
	Environ = &Env{DB: &MockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/v1/accounts", nil)
	http.HandlerFunc(AccountsHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := AccountsResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the accounts response: %v", err)
	}
	if len(result.Accounts) != 1 {
		t.Errorf("Expected 1 accounts, got %d", len(result.Accounts))
	}
}

func TestAccountsHandlerError(t *testing.T) {

	// Mock the database
	Environ = &Env{DB: &ErrorMockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/v1/accounts", nil)
	http.HandlerFunc(AccountsHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := AccountsResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the accounts response: %v", err)
	}
	if result.Success {
		t.Error("Expected error, got success response")
	}
}

func TestAccountsUpsertHandler(t *testing.T) {

	// Mock the database
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	Environ = &Env{DB: &MockDB{}, Config: config}
	Environ.KeypairDB, _ = GetKeyStore(config)

	// Create the account assertion
	assertAcc, err := generateAccountAssertion(asserts.AccountType, "alder", "maple-inc")
	if err != nil {
		t.Errorf("Error generating the assertion: %v", err)
	}

	// Encode the assertion and create the request
	encodedAssert := base64.StdEncoding.EncodeToString([]byte(assertAcc))
	request, err := json.Marshal(AccountAssertionRequest{Assertion: encodedAssert})
	if err != nil {
		t.Errorf("Error marshalling the assertion to JSON: %v", err)
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/accounts", bytes.NewBuffer(request))
	http.HandlerFunc(AccountsUpsertHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the accounts response: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.ErrorMessage)
	}
}

func sendAccountsUpsertError(request []byte, t *testing.T) {

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/accounts", bytes.NewBuffer(request))
	http.HandlerFunc(AccountsUpsertHandler).ServeHTTP(w, r)

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

func TestAccountsUpsertNilRequest(t *testing.T) {

	// Mock the database
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	Environ = &Env{DB: &MockDB{}, Config: config}
	Environ.KeypairDB, _ = GetKeyStore(config)

	sendAccountsUpsertError(nil, t)
}

func TestAccountsUpsertInvalidRequest(t *testing.T) {

	// Mock the database
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	Environ = &Env{DB: &MockDB{}, Config: config}
	Environ.KeypairDB, _ = GetKeyStore(config)

	sendAccountsUpsertError([]byte("InvalidData"), t)
}

func TestAccountsUpsertInvalidEncoding(t *testing.T) {

	// Mock the database
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	Environ = &Env{DB: &MockDB{}, Config: config}
	Environ.KeypairDB, _ = GetKeyStore(config)

	request, err := json.Marshal(AccountAssertionRequest{Assertion: "InvalidData"})
	if err != nil {
		t.Errorf("Error marshalling the assertion to JSON: %v", err)
	}
	sendAccountsUpsertError(request, t)
}

func TestAccountsUpsertInvalidAssertion(t *testing.T) {

	// Mock the database
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	Environ = &Env{DB: &MockDB{}, Config: config}
	Environ.KeypairDB, _ = GetKeyStore(config)

	// Encode the assertion and create the request
	encodedAssert := base64.StdEncoding.EncodeToString([]byte("InvalidData"))
	request, err := json.Marshal(AccountAssertionRequest{Assertion: encodedAssert})
	if err != nil {
		t.Errorf("Error marshalling the assertion to JSON: %v", err)
	}
	sendAccountsUpsertError(request, t)
}

func TestAccountsUpsertInvalidAssertionType(t *testing.T) {

	// Mock the database
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	Environ = &Env{DB: &MockDB{}, Config: config}
	Environ.KeypairDB, _ = GetKeyStore(config)

	// Encode the assertion and create the request (account-key instead of an account assertion)
	assertion, err := generateAccountAssertion(asserts.AccountKeyType, "alder", "maple-inc")
	if err != nil {
		t.Errorf("Error generating the assertion: %v", err)
	}
	encodedAssert := base64.StdEncoding.EncodeToString([]byte(assertion))
	request, err := json.Marshal(AccountAssertionRequest{Assertion: encodedAssert})
	if err != nil {
		t.Errorf("Error marshalling the assertion to JSON: %v", err)
	}
	sendAccountsUpsertError(request, t)
}

func TestAccountsUpsertPutError(t *testing.T) {

	// Mock the database
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	Environ = &Env{DB: &ErrorMockDB{}, Config: config}
	Environ.KeypairDB, _ = GetKeyStore(config)

	// Encode the assertion and create the request
	assertion, err := generateAccountAssertion(asserts.AccountType, "alder", "maple-inc")
	if err != nil {
		t.Errorf("Error generating the assertion: %v", err)
	}
	encodedAssert := base64.StdEncoding.EncodeToString([]byte(assertion))
	request, err := json.Marshal(AccountAssertionRequest{Assertion: encodedAssert})
	if err != nil {
		t.Errorf("Error marshalling the assertion to JSON: %v", err)
	}
	sendAccountsUpsertError(request, t)
}
