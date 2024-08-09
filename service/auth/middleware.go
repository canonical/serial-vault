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

package auth

import (
	"errors"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/service/log"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/usso"
	jwt "github.com/golang-jwt/jwt/v5"
)

// JWTCheck extracts the JWT from the request, validates it and returns the token
func JWTCheck(w http.ResponseWriter, r *http.Request) (*jwt.Token, error) {

	// Do not validate access if user authentication is off (default)
	if !datastore.Environ.Config.EnableUserAuth {
		return nil, nil
	}

	// Get the JWT from the header or cookie
	jwtToken, err := usso.JWTExtractor(r)
	if err != nil {
		log.Println("Error in JWT extraction:", err.Error())
		return nil, errors.New("Error in retrieving the authentication token")
	}

	// Verify the JWT string
	token, err := usso.VerifyJWT(jwtToken)
	if err != nil {
		log.Printf("JWT fails verification: %v", err.Error())
		return nil, errors.New("The authentication token is invalid")
	}

	if !token.Valid {
		log.Println("Invalid JWT")
		return nil, errors.New("The authentication token is invalid")
	}

	// Set up the bearer token in the header
	w.Header().Set("Authorization", "Bearer "+jwtToken)

	return token, nil
}
