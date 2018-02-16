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

package utils

import (
	"errors"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/usso"
	jwt "github.com/dgrijalva/jwt-go"
)

// CheckIsStandardAndGetUserFromJWT verifies that we have an authenticated, standard user
func CheckIsStandardAndGetUserFromJWT(w http.ResponseWriter, r *http.Request) (datastore.User, error) {
	return checkPermissionsAndGetUserFromJWT(w, r, datastore.Standard)
}

// CheckIsAdminAndGetUserFromJWT verifies that we have an authenticated, admin user
func CheckIsAdminAndGetUserFromJWT(w http.ResponseWriter, r *http.Request) (datastore.User, error) {
	return checkPermissionsAndGetUserFromJWT(w, r, datastore.Admin)
}

// CheckIsSuperuserAndGetUserFromJWT verifies that we have an authenticated, superuser user
func CheckIsSuperuserAndGetUserFromJWT(w http.ResponseWriter, r *http.Request) (datastore.User, error) {
	return checkPermissionsAndGetUserFromJWT(w, r, datastore.Superuser)
}

func checkPermissionsAndGetUserFromJWT(w http.ResponseWriter, r *http.Request, minimumAuthorizedRole int) (datastore.User, error) {
	user, err := GetUserFromJWT(w, r)
	if err != nil {
		return user, err
	}
	err = CheckUserPermissions(user, minimumAuthorizedRole)
	if err != nil {
		return user, err
	}
	return user, nil
}

// GetUserFromJWT retrieves the user details from the JSON Web Token
func GetUserFromJWT(w http.ResponseWriter, r *http.Request) (datastore.User, error) {
	token, err := JWTCheck(w, r)
	if err != nil {
		return datastore.User{}, err
	}

	// Null token means that auth is not enabled.
	if token == nil {
		return datastore.User{}, nil
	}

	claims := token.Claims.(jwt.MapClaims)
	username := claims[usso.ClaimsUsername].(string)
	role := int(claims[usso.ClaimsRole].(float64))

	return datastore.User{
		Username: username,
		Role:     role,
	}, nil
}

// CheckUserPermissions verifies that a user has a minimum role
func CheckUserPermissions(user datastore.User, minimumAuthorizedRole int) error {
	// User authentication is turned off
	if !datastore.Environ.Config.EnableUserAuth {
		// Superuser permissions don't allow turned off authentication
		if minimumAuthorizedRole == datastore.Superuser {
			return errors.New("The user is not authorized")
		}
		return nil
	}

	if user.Role < minimumAuthorizedRole {
		return errors.New("The user is not authorized")
	}
	return nil
}
