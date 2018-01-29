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

package datastore

import (
	"errors"
	"regexp"
)

var validStoreNameRegexp = regexp.MustCompile(defaultNicknamePattern)
var validSerialNumberRegexp = regexp.MustCompile(defaultNicknamePattern)

// ListSubstores return account sub-stores the user is authorized to see
func (db *DB) ListSubstores(accountID int, authorization User) ([]Substore, error) {
	switch authorization.Role {
	case Invalid: // Authentication disabled
		fallthrough
	case Superuser:
		return db.listSubstores(accountID)
	case Admin:
		return db.listSubstoresFilteredByUser(accountID, authorization.Username)
	default:
		return []Substore{}, nil
	}
}

// UpdateAllowedSubstore updates the sub-store if authorization is allowed to do it
func (db *DB) UpdateAllowedSubstore(store Substore, authorization User) (string, error) {

	errorSubcode, err := validateSubstore(store, "error-validate-store")
	if err != nil {
		return errorSubcode, err
	}

	switch authorization.Role {
	case Invalid: // Authentication is disabled
		fallthrough
	case Superuser:
		return db.updateSubstore(store)
	case Admin:
		return db.updateSubstoreFilteredByUser(store, authorization.Username)
	default:
		return "", nil
	}
}

func validateSubstore(store Substore, validateStoreLabel string) (string, error) {

	err := validateModelID(store.FromModelID)
	if err != nil {
		return validateStoreLabel, err
	}

	err = validateModelID(store.ToModelID)
	if err != nil {
		return validateStoreLabel, err
	}

	err = validateSyntax("Sub-store name", store.Store, validStoreNameRegexp)
	if err != nil {
		return validateStoreLabel, err
	}

	err = validateSyntax("Serial-number", store.SerialNumber, validSerialNumberRegexp)
	if err != nil {
		return validateStoreLabel, err
	}

	return "", nil
}

func validateModelID(modelID int) error {
	if modelID <= 0 {
		return errors.New("The Model must be selected")
	}
	return nil
}
