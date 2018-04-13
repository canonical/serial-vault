// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
 * License granted by Canonical Limited
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

// ListAllowedTestLog return test logs the user is authorized to see
func (db *DB) ListAllowedTestLog(authorization User) ([]TestLog, error) {
	switch authorization.Role {
	case Invalid: // Authentication disabled
		fallthrough
	case Superuser:
		return db.listAllTestLog()
	case Admin:
		return db.listTestLogFilteredByUser(authorization.Username)
	default:
		return []TestLog{}, nil
	}
}
