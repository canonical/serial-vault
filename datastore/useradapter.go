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

package datastore

import (
	"errors"
	"regexp"
)

const validUsernamePattern = defaultNicknamePattern
const validEmailPattern = `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`

var validUsernameRegexp = regexp.MustCompile(validUsernamePattern)
var validEmailRegexp = regexp.MustCompile(validEmailPattern)

// CreateUser validates and adds a new record to User database table, Returns new record identifier if success
func (db *DB) CreateUser(user User) (int, error) {

	// Check the API key and default it if it is invalid
	apiKey, err := buildValidOrDefaultAPIKey(user.APIKey)
	if err != nil {
		return 0, errors.New("Error in generating a valid API key")
	}
	user.APIKey = apiKey

	err = validateUser(user)
	if err != nil {
		return 0, err
	}
	return db.createUser(user)
}

// UpdateUser validates and sets user new values for an existing record. Also updates useraccount link. All that in a transaction
func (db *DB) UpdateUser(user User) error {
	// Check the API key and default it if it is invalid
	apiKey, err := buildValidOrDefaultAPIKey(user.APIKey)
	if err != nil {
		return errors.New("Error in generating a valid API key")
	}
	user.APIKey = apiKey

	err = validateUser(user)
	if err != nil {
		return err
	}

	return db.updateUser(user)
}

func validateUser(user User) error {
	// Validate username; the rule is: lowercase with no spaces
	err := validateUsername(user.Username)
	if err != nil {
		return err
	}

	// Validate name; the rule is: not empty
	err = validateUserFullName(user.Name)
	if err != nil {
		return err
	}

	// Validate email; the rule is: not empty
	err = validateUserEmail(user.Email)
	if err != nil {
		return err
	}

	// Validate role; the rule is the role is 100, 200 or 300
	err = validateUserRole(user.Role)
	if err != nil {
		return err
	}
	return nil
}

func validateUsername(username string) error {
	return validateSyntax("Username", username, validUsernameRegexp)
}

func validateUserRole(role int) error {
	if role != Standard && role != SyncUser && role != Admin && role != Superuser {
		return errors.New("Role is not amongst valid ones")
	}
	return nil
}

func validateUserFullName(name string) error {
	return validateNotEmpty("Name", name)
}

func validateUserEmail(email string) error {
	return validateSyntax("Email", email, validEmailRegexp)
}
