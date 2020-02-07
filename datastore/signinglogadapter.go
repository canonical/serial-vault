// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2018 Canonical Ltd
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

// ListAllowedSigningLog return signing logs the user is authorized to see
func (db *DB) ListAllowedSigningLog(authorization User) ([]SigningLog, error) {
	switch authorization.Role {
	case Invalid: // Authentication disabled
		fallthrough
	case Superuser:
		return db.listAllSigningLog()
	case SyncUser:
		fallthrough
	case Admin:
		return db.listSigningLogFilteredByUser(authorization.Username)
	default:
		return []SigningLog{}, nil
	}
}

// ListAllowedSigningLogForAccount return signing logs the user is authorized to see
func (db *DB) ListAllowedSigningLogForAccount(authorization User, authorityID string, params *SigningLogParams) ([]SigningLog, error) {
	switch authorization.Role {
	case Invalid: // Authentication disabled
		fallthrough
	case Superuser:
		return db.listAllSigningLogForAccount(authorityID, params)
	case SyncUser:
		fallthrough
	case Admin:
		return db.listSigningLogForAccountFilteredByUser(authorization.Username, authorityID, params)
	default:
		return []SigningLog{}, nil
	}
}

// AllowedSigningLogFilterValues return signing log filters authorized for the user
func (db *DB) AllowedSigningLogFilterValues(authorization User, authorityID string) (SigningLogFilters, error) {
	switch authorization.Role {
	case Invalid: // Authentication disabled
		fallthrough
	case Superuser:
		return db.allSigningLogFilterValues(authorityID)
	case Admin:
		return db.signingLogFilterValuesFilteredByUser(authorization.Username, authorityID)
	default:
		return SigningLogFilters{}, nil
	}
}
