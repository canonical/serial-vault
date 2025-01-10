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
	"fmt"
	"regexp"
)

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

// GetAllowedSubstore return the sub-store if the user is authorized to see it
func (db *DB) GetAllowedSubstore(modelID int, serial string, authorization User) (Substore, error) {
	switch authorization.Role {
	case Invalid: // Authentication disabled
		fallthrough
	case Superuser:
		return db.GetSubstore(modelID, serial)
	case Admin:
		return db.GetSubstoreFilteredByUser(modelID, serial, authorization.Username)
	default:
		return Substore{}, nil
	}
}

// UpdateAllowedSubstore updates the sub-store if authorization is allowed to do it
func (db *DB) UpdateAllowedSubstore(store Substore, authorization User) error {

	_, err := validateSubstore(store, "error-validate-store")
	if err != nil {
		return err
	}

	switch authorization.Role {
	case Invalid: // Authentication is disabled
		fallthrough
	case Superuser:
		return db.updateSubstore(store)
	case Admin:
		return db.updateSubstoreFilteredByUser(store, authorization.Username)
	default:
		return nil
	}
}

// CreateAllowedSubstore creates a new model in case authorization is allowed to do it
func (db *DB) CreateAllowedSubstore(store Substore, authorization User) (Substore, error) {
	// Validate the substore record
	_, err := validateSubstore(store, "")
	if err != nil {
		return store, err
	}

	// Validate that the user has access to the account
	acc, err := db.GetAccountByID(store.AccountID, authorization)
	if err != nil || acc.ID == 0 {
		return store, errors.New("You do not have permissions to this account")
	}

	fromModel, err := db.getModel(store.FromModelID)
	if err != nil || fromModel.BrandID != acc.AuthorityID {
		return store, errors.New("The source model does not exist or does not belong to this account's brand")
	}

	switch authorization.Role {
	case Invalid: // Authentication is disabled
		fallthrough
	case Superuser:
		fallthrough
	case Admin:
		return db.createSubstore(store)
	default:
		return Substore{}, nil
	}
}

// DeleteAllowedSubstore deletes sub-store model if allowed to authorization
func (db *DB) DeleteAllowedSubstore(storeID int, authorization User) (string, error) {
	switch authorization.Role {
	case Invalid: // Authentication is disabled
		fallthrough
	case Superuser:
		return db.deleteSubstore(storeID)
	case Admin:
		return db.deleteSubstoreFilteredByUser(storeID, authorization.Username)
	default:
		return "", nil
	}
}

func validateSubstore(store Substore, validateStoreLabel string) (string, error) {
	errTemplate := "invalid substore %s: %v"

	err := validateModelID("From Model", store.FromModelID)
	if err != nil {
		return validateStoreLabel, fmt.Errorf(errTemplate, store.ModelName, err)
	}

	err = validateNotEmpty("Sub-store name", store.Store)
	if err != nil {
		return validateStoreLabel, fmt.Errorf(errTemplate, store.ModelName, err)
	}

	err = validateNotEmpty("Serial-number", store.SerialNumber)
	if err != nil {
		return validateStoreLabel, fmt.Errorf(errTemplate, store.ModelName, err)
	}

	err = validateModelName(store.ModelName)
	if err != nil {
		return validateStoreLabel, fmt.Errorf(errTemplate, store.ModelName, err)
	}

	return "", nil
}

func validateModelID(fieldname string, modelID int) error {
	if modelID <= 0 {
		return fmt.Errorf("'%s' must be selected", fieldname)
	}
	return nil
}
