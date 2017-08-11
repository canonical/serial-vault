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
)

// ListAllowedAccounts fetches the available accounts from the database that the user is allowed to see
func (db *DB) ListAllowedAccounts(authorization User) ([]Account, error) {
	switch authorization.Role {
	case Invalid: // Authentication disabled
		fallthrough
	case Superuser:
		return db.listAllAccounts()
	case Admin:
		return db.listAccountsFilteredByUser(authorization.Username)
	default:
		return []Account{}, nil
	}
}

// PutAccount validates permissions and stores an account in the database
func (db *DB) PutAccount(account Account, authorization User) (string, error) {

	err := validateAuthorityID(account.AuthorityID)
	if err != nil {
		return "error-validate-account", err
	}

	if authorization.Role == Admin {
		// Check that the user has permissions for the account
		if !db.CheckUserInAccount(authorization.Username, account.AuthorityID) {
			return "error-auth", errors.New("You do not have permissions for that authority")
		}
	}

	return db.putAccount(account)
}

func validateAuthorityID(AuthorityID string) error {
	return validateNotEmpty("Authority ID", AuthorityID)
}
