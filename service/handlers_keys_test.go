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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAuthorizedKeysHandler(t *testing.T) {
	// Use our test authorized_keys file
	Environ = &Env{AuthorizedKeys: &AuthorizedKeys{path: "../test/authorized_keys"}}

	result, _ := sendListRequest(t)
	if len(result.Keys) != 3 {
		t.Errorf("Expected 3 keys, got: %d", len(result.Keys))
	}
}

func TestAuthorizedKeysHandlerInvalidFile(t *testing.T) {
	// Use our test authorized_keys file
	Environ = &Env{AuthorizedKeys: &AuthorizedKeys{path: "does not exist"}}

	result, _ := sendListRequest(t)
	if len(result.Keys) != 0 {
		t.Errorf("Expected 0 keys, got: %d", len(result.Keys))
	}
}

func sendListRequest(t *testing.T) (AuthorizedKeysResponse, error) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/1.0/keys", nil)
	http.HandlerFunc(AuthorizedKeysHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := AuthorizedKeysResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the models response: %v", err)
	}
	return result, err
}

func sendAddRequest(t *testing.T, data io.Reader) (BooleanResponse, error) {

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/1.0/keys", data)
	http.HandlerFunc(AuthorizedKeyAddHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the models response: %v", err)
	}
	return result, err
}

func TestAuthorizedKeyAddHandler(t *testing.T) {
	// Use a new authorized_keys file
	path := "../test/authorized_keys_new"
	Environ = &Env{AuthorizedKeys: &AuthorizedKeys{path: path}}

	jsonMessage := fmt.Sprintf(`{"device-key": "%s"}`, "rsa-ssh abcdef0123456789 Comment")
	result, _ := sendAddRequest(t, bytes.NewBufferString(jsonMessage))
	if !result.Success {
		t.Error(result.ErrorMessage)
	}

	jsonMessage = fmt.Sprintf(`{"device-key": "%s"}`, "rsa-ssh 0123456789abcdef No Comment")
	result, _ = sendAddRequest(t, bytes.NewBufferString(jsonMessage))
	if !result.Success {
		t.Error(result.ErrorMessage)
	}

	// Check the keys were added correctly
	keys := Environ.AuthorizedKeys.List()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got: %d", len(keys))
	}
	if keys[0] != "rsa-ssh abcdef0123456789 Comment" {
		t.Errorf("Expected 'rsa-ssh abcdef0123456789 Comment', got: %s", keys[0])
	}

	// Delete the created file
	os.Remove(path)
}

func TestAuthorizedKeyAddHandlerNilData(t *testing.T) {
	// Use a new authorized_keys file
	path := "../test/authorized_keys_new"
	Environ = &Env{AuthorizedKeys: &AuthorizedKeys{path: path}}

	result, _ := sendAddRequest(t, nil)
	if result.Success {
		t.Error("Expected an error, got success.")
	}
}

func TestAuthorizedKeyAddHandlerEmptyData(t *testing.T) {
	// Use a new authorized_keys file
	path := "../test/authorized_keys_new"
	Environ = &Env{AuthorizedKeys: &AuthorizedKeys{path: path}}

	result, _ := sendAddRequest(t, bytes.NewBufferString(""))
	if result.Success {
		t.Error("Expected an error, got success.")
	}
}

func TestAuthorizedKeyAddHandlerBadData(t *testing.T) {
	// Use a new authorized_keys file
	path := "../test/authorized_keys_new"
	Environ = &Env{AuthorizedKeys: &AuthorizedKeys{path: path}}

	result, _ := sendAddRequest(t, bytes.NewBufferString("bad"))
	if result.Success {
		t.Error("Expected an error, got success.")
	}
}

func sendDeleteRequest(t *testing.T, data io.Reader) (BooleanResponse, error) {

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/1.0/keys/delete", data)
	http.HandlerFunc(AuthorizedKeyDeleteHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := BooleanResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the models response: %v", err)
	}
	return result, err
}

func TestAuthorizedKeyDeleteHandler(t *testing.T) {
	// Use a new authorized_keys file
	path := "../test/authorized_keys_new"
	Environ = &Env{AuthorizedKeys: &AuthorizedKeys{path: path}}

	// Add some ssh keys
	Environ.AuthorizedKeys.Add("rsa-ssh abcdef0123456789 Comment")
	Environ.AuthorizedKeys.Add("rsa-ssh 0123456789abcdef No Comment")
	Environ.AuthorizedKeys.Add("rsa-ssh abc0123456789def Another Comment")

	// Request the deletion of the last key
	jsonMessage := fmt.Sprintf(`{"device-key": "%s"}`, "rsa-ssh abc0123456789def Another Comment")
	result, _ := sendDeleteRequest(t, bytes.NewBufferString(jsonMessage))
	if !result.Success {
		t.Error(result.ErrorMessage)
	}

	// Check the key was deleted correctly
	keys := Environ.AuthorizedKeys.List()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got: %d", len(keys))
	}
	if keys[0] != "rsa-ssh abcdef0123456789 Comment" {
		t.Errorf("Expected 'rsa-ssh abcdef0123456789 Comment', got: %s", keys[0])
	}
	if keys[1] != "rsa-ssh 0123456789abcdef No Comment" {
		t.Errorf("Expected 'rsa-ssh 0123456789abcdef No Comment', got: %s", keys[1])
	}

	// Delete the created file
	os.Remove(path)
}

func TestAuthorizedKeyDeleteHandlerFirst(t *testing.T) {
	// Use a new authorized_keys file
	path := "../test/authorized_keys_new"
	Environ = &Env{AuthorizedKeys: &AuthorizedKeys{path: path}}

	// Add some ssh keys
	Environ.AuthorizedKeys.Add("rsa-ssh abcdef0123456789 Comment")
	Environ.AuthorizedKeys.Add("rsa-ssh 0123456789abcdef No Comment")
	Environ.AuthorizedKeys.Add("rsa-ssh abc0123456789def Another Comment")

	// Request the deletion of the first key
	jsonMessage := fmt.Sprintf(`{"device-key": "%s"}`, "rsa-ssh abcdef0123456789 Comment")
	result, _ := sendDeleteRequest(t, bytes.NewBufferString(jsonMessage))
	if !result.Success {
		t.Error(result.ErrorMessage)
	}

	// Check the key was deleted correctly
	keys := Environ.AuthorizedKeys.List()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got: %d", len(keys))
	}
	if keys[0] != "rsa-ssh 0123456789abcdef No Comment" {
		t.Errorf("Expected 'rsa-ssh 0123456789abcdef No Comment', got: %s", keys[0])
	}
	if keys[1] != "rsa-ssh abc0123456789def Another Comment" {
		t.Errorf("Expected 'rsa-ssh abc0123456789def Another Comment', got: %s", keys[1])
	}

	// Delete the created file
	os.Remove(path)
}

func TestAuthorizedKeyDeleteHandlerWrongKey(t *testing.T) {
	// Use a new authorized_keys file
	path := "../test/authorized_keys_new"
	Environ = &Env{AuthorizedKeys: &AuthorizedKeys{path: path}}

	// Request the deletion of a non-existent key
	jsonMessage := fmt.Sprintf(`{"device-key": "%s"}`, "does not exist")
	result, _ := sendDeleteRequest(t, bytes.NewBufferString(jsonMessage))
	if result.Success {
		t.Error("Expected an error, got success.")
	}
}

func TestAuthorizedKeyDeleteHandlerBadJSON(t *testing.T) {
	// Use a new authorized_keys file
	path := "../test/authorized_keys_new"
	Environ = &Env{AuthorizedKeys: &AuthorizedKeys{path: path}}

	result, _ := sendDeleteRequest(t, bytes.NewBufferString("bad"))
	if result.Success {
		t.Error("Expected an error, got success.")
	}
}
