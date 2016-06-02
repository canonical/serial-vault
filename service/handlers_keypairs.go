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
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/snapcore/snapd/asserts"

	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

// KeypairWithPrivateKey is the JSON version of a keypair, including the base64 armored, signing-key
type KeypairWithPrivateKey struct {
	ID          int    `json:"id"`
	AuthorityID string `json:"authority-id"`
	PrivateKey  string `json:"private-key"`
}

// KeypairListHandler fetches the available keypairs for display from the database.
// Only viewable reference data is stored in the database, not the restricted private key.
func KeypairListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	keypairs, err := Environ.DB.ListKeypairs()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		formatKeypairsResponse(false, "error-fetch-keypairs", "", err.Error(), nil, w)
		return
	}

	// Return successful JSON response with the list of models
	w.WriteHeader(http.StatusOK)
	formatKeypairsResponse(true, "", "", "", keypairs, w)
}

// KeypairCreateHandler is the API method to create a new keypair that can be used
// for signing serial assertions. The keypairs are stored in the signing database
// and the authority-id/key-id is stored in the models database. Models can then be
// linked to one of the existing signing-keys.
func KeypairCreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Check that we have a message body
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-nil-data", "", "Uninitialized POST data", w)
		return
	}
	defer r.Body.Close()

	// Decode the JSON body
	keypairWithKey := KeypairWithPrivateKey{}
	err := json.NewDecoder(r.Body).Decode(&keypairWithKey)
	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-keypair-data", "", "No keypair data supplied", w)
		return
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-keypair-json", "", err.Error(), w)
		return
	}

	// Validate the authority-id
	keypairWithKey.AuthorityID = strings.TrimSpace(keypairWithKey.AuthorityID)
	if len(keypairWithKey.AuthorityID) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-keypair-json", "", "The authority-id is mandatory", w)
		return
	}

	// Deserialize and validate the private key, converting it into a crypto object
	privateKey, errorCode, err := deserializePrivateKey(keypairWithKey.PrivateKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, errorCode, "", err.Error(), w)
		return
	}

	// Store the signing-key in the keypair store using the asserts module
	err = Environ.KeypairDB.ImportKey(keypairWithKey.AuthorityID, privateKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-keypair-store", "", err.Error(), w)
		return
	}

	// Store the signing-key in the database
	keypair := Keypair{
		AuthorityID: keypairWithKey.AuthorityID,
		KeyID:       privateKey.PublicKey().ID(),
	}
	errorCode, err = Environ.DB.PutKeypair(keypair)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, errorCode, "", err.Error(), w)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	formatBooleanResponse(true, "", "", "", w)
}

// deserializePrivateKey decodes a base64 encoded private key file and converts
// it to a private key that can be used for storage in the keypair store
func deserializePrivateKey(base64PrivateKey string) (asserts.PrivateKey, string, error) {
	// The private-key is base64 encoded, so we need to decode it
	decodedPrivateKey, err := base64.StdEncoding.DecodeString(base64PrivateKey)
	if err != nil {
		return nil, "error-decode-key", err
	}

	// Validate the signing-key
	block, err := armor.Decode(bytes.NewReader(decodedPrivateKey))
	if err != nil {
		return nil, "error-invalid-key", err
	}

	p, err := packet.Read(block.Body)
	if err != nil {
		return nil, "error-invalid-key", err
	}

	privateKey, ok := p.(*packet.PrivateKey)
	if !ok {
		return nil, "error-invalid-key", errors.New("Not a private key")
	}
	return asserts.OpenPGPPrivateKey(privateKey), "", nil
}

// KeypairDisableHandler disables an existing keypair, which will mean that any
// linked Models will not be able to be signed. The asserts module does not allow
// a keypair to be deleted, so the keypair will just be disabled in the local database.
func KeypairDisableHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get the keypair primary key
	vars := mux.Vars(r)
	keypairID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("%v", vars["id"])
		formatBooleanResponse(false, "error-invalid-keypair", "", errorMessage, w)
		return
	}

	// Update the keypair in the local database
	err = Environ.DB.UpdateKeypairActive(keypairID, false)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-keypair-update", "", err.Error(), w)
		return
	}

	formatBooleanResponse(true, "", "", "", w)
}

// KeypairEnableHandler enables an existing keypair, which will mean that any
// linked Models will be able to be signed. The asserts module does not allow
// a keypair to be deleted, so the keypair will just be enabled in the local database.
func KeypairEnableHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get the keypair primary key
	vars := mux.Vars(r)
	keypairID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("%v", vars["id"])
		formatBooleanResponse(false, "error-invalid-keypair", "", errorMessage, w)
		return
	}

	// Update the keypair in the local database
	err = Environ.DB.UpdateKeypairActive(keypairID, true)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-keypair-update", "", err.Error(), w)
		return
	}
	formatBooleanResponse(true, "", "", "", w)
}
