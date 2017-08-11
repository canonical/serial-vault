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

import "errors"

// ListAllowedKeypairs return the list of keypairs allowed to the user
func (db *DB) ListAllowedKeypairs(authorization User) ([]Keypair, error) {
	switch authorization.Role {
	case Invalid: // Authentication is disabled
		fallthrough
	case Superuser:
		return db.listAllKeypairs()
	case Admin:
		return db.listKeypairsFilteredByUser(authorization.Username)
	default:
		return []Keypair{}, nil
	}
}

// UpdateAllowedKeypairActive updates active enable/disable flag if user is authorized
func (db *DB) UpdateAllowedKeypairActive(keypairID int, active bool, authorization User) error {
	switch authorization.Role {
	case Invalid: // Authentication is disabled
		fallthrough
	case Superuser:
		return db.updateKeypairActive(keypairID, active)
	case Admin:
		return db.updateKeypairActiveFilteredByUser(keypairID, active, authorization.Username)
	default:
		return nil
	}
}

// UpdateKeypairAssertion validates user can update and sets the account-key assertion of a keypair
func (db *DB) UpdateKeypairAssertion(keypair Keypair, authorization User) (string, error) {

	err := validateAuthorityID(keypair.AuthorityID)
	if err != nil {
		return "invalid-assertion", err
	}

	err = db.validateAssertionHeaders(keypair)
	if err != nil {
		return "invalid-assertion", err
	}

	if authorization.Role == Admin {
		// Check that the user has permissions for the account
		if !db.CheckUserInAccount(authorization.Username, keypair.AuthorityID) {
			return "error-auth", errors.New("You do not have permissions for that authority")
		}
	}

	return "", db.updateKeypairAssertion(keypair.ID, keypair.Assertion)
}

func (db *DB) validateAssertionHeaders(keypair Keypair) error {
	oldKeypair, err := db.GetKeypair(keypair.ID)
	if err != nil {
		return err
	}

	if oldKeypair.AuthorityID != keypair.AuthorityID {
		return errors.New("Authority ID does not match the existing account key")
	}

	if oldKeypair.KeyID != keypair.KeyID {
		return errors.New("Authority ID does not match the existing account key")
	}

	return nil
}
