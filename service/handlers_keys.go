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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

// AuthorizedKeysResponse is the JSON response from the API AuthorizedKeys call.
// This returns a slice containing the list of ssh keys that can access the vault.
type AuthorizedKeysResponse struct {
	Keys []string `json:"keys"`
}

// BooleanResponse is the JSON response from an API method, indicating success or failure.
type BooleanResponse struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"message"`
}

// AuthorizedKey is the JSON body with the public key.
type AuthorizedKey struct {
	PublicKey string `json:"device-key"`
}

// AuthorizedKeysHandler is the API method to list the authorized keys
func AuthorizedKeysHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	keys := Environ.AuthorizedKeys.List()

	response := AuthorizedKeysResponse{Keys: keys}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding the AuthorizedKeys response: %v\n", err)
	}
}

// AuthorizedKeyAddHandler adds a new authorized key
func AuthorizedKeyAddHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	deviceKey, err := decodeKey(w, r)
	if err != nil {
		return
	}

	err = Environ.AuthorizedKeys.Add(deviceKey)
	if err != nil {
		message := fmt.Sprintf("Error adding new public key: %v", err)
		log.Printf(message)
		formatBooleanResponse(false, message, w)
	} else {
		formatBooleanResponse(true, "", w)
	}
}

// AuthorizedKeyDeleteHandler removes a key from the authorized keys file.
func AuthorizedKeyDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	deviceKey, err := decodeKey(w, r)
	if err != nil {
		return
	}

	err = Environ.AuthorizedKeys.Delete(deviceKey)
	if err != nil {
		message := fmt.Sprintf("Error deleting a public key: %v", err)
		log.Printf(message)
		formatBooleanResponse(false, message, w)
	} else {
		formatBooleanResponse(true, "", w)
	}
}

// AuthorizedKeyDeleteHandler removes a key from the authorized keys file.
func decodeKey(w http.ResponseWriter, r *http.Request) (string, error) {

	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "Uninitialized POST data.", w)
		return "", errors.New("Uninitialized POST data.")
	}
	defer r.Body.Close()

	// Decode the JSON body
	deviceKey := new(AuthorizedKey)
	err := json.NewDecoder(r.Body).Decode(&deviceKey)
	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "No data supplied for signing.", w)
		return "", err
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("Error decoding JSON: %v", err)
		formatBooleanResponse(false, errorMessage, w)
		return "", err
	}

	return deviceKey.PublicKey, nil
}
