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

package usso

import (
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/juju/usso/openid"
)

// NewJWTToken creates a new JWT from the verified OpenID response
func NewJWTToken(resp *openid.Response) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	token.Claims[ClaimsUsername] = resp.SReg["nickname"]
	token.Claims[ClaimsName] = resp.SReg["fullname"]
	token.Claims[ClaimsEmail] = resp.SReg["email"]
	token.Claims[ClaimsIdentity] = resp.ID
	token.Claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("Error signing the JWT: %v", err.Error())
	}
	return tokenString, err
}

// VerifyJWT checks that we have a valid token
func VerifyJWT(jwtToken string) (interface{}, error) {

	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	return token, err
}
