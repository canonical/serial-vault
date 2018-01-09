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
	"log"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/store"
)

const (
	ssoBaseURL   = "https://login.ubuntu.com/api/v2/"
	storeBaseURL = "https://dashboard.snapcraft.io/dev/api/"
)

// StorePermissions is the SSO authorization for the store
type StorePermissions struct {
	Permissions []string `json:"permissions"`
}

// StoreACL is the SSO authorization for the store
type StoreACL struct {
	Macaroon string `json:"macaroon"`
}

// StoreAuth is the SSO authorization for the store
type StoreAuth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	OTP      string `json:"otp"`
}

// StoreKeyRegister is the request to submit a signing-key to the store
type StoreKeyRegister struct {
	StoreAuth
	AuthorityID string `json:"authority-id"`
	KeyName     string `json:"key-name"`
}

// StoreKeyRegisterHandler is the API method to upload a signing-key to the store
func StoreKeyRegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
	if err != nil {
		formatBooleanResponse(false, "error-auth", "", "", w)
		return
	}

	// Decode the JSON body
	keyAuth := store.KeyRegister{}
	err = json.NewDecoder(r.Body).Decode(&keyAuth)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Error in store key request: %v", err)
		formatBooleanResponse(false, "error-decode-json", "", "", w)
		return
	}

	// Check that the user has permissions to this authority-id
	if !datastore.Environ.DB.CheckUserInAccount(authUser.Username, keyAuth.AuthorityID) {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-auth", "", "Your user does not have permissions for the Signing Authority", w)
		return
	}

	keypair, err := datastore.Environ.DB.GetKeypairByName(keyAuth.AuthorityID, keyAuth.KeyName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Error in fetching the keypair: %v", err)
		formatBooleanResponse(false, "error-invalid-keypair", "", "", w)
		return
	}

	// Register the account key with the store
	err = store.RegisterKey(keyAuth, keypair)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Error in submitting the keypair: %v", err)
		formatBooleanResponse(false, "error-store-keypair", "", err.Error(), w)
		return
	}
	formatBooleanResponse(true, "", "", "", w)
}
