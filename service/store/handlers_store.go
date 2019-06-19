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

package store

import (
	"encoding/json"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/service/log"

	"github.com/CanonicalLtd/serial-vault/service/auth"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/CanonicalLtd/serial-vault/store"
)

const (
	ssoBaseURL   = "https://login.ubuntu.com/api/v2/"
	storeBaseURL = "https://dashboard.snapcraft.io/dev/api/"
)

// Permissions is the SSO authorization for the store
type Permissions struct {
	Permissions []string `json:"permissions"`
}

// ACL is the SSO authorization for the store
type ACL struct {
	Macaroon string `json:"macaroon"`
}

// Auth is the SSO authorization for the store
type Auth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	OTP      string `json:"otp"`
}

// KeyRegister is the API method to upload a signing-key to the store
func KeyRegister(w http.ResponseWriter, r *http.Request) {
	authUser, err := auth.GetUserFromJWT(w, r)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	// Decode the JSON body
	keyAuth := store.KeyRegister{}
	err = json.NewDecoder(r.Body).Decode(&keyAuth)
	if err != nil {
		log.Printf("Error in store key request: %v", err)
		response.FormatStandardResponse(false, "error-decode-json", "", "", w)
		return
	}

	keyRegisterHandler(w, authUser, false, keyAuth)
}
