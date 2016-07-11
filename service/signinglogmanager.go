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

package service

import "time"

const createSigningLogTableSQL = `
	CREATE TABLE IF NOT EXISTS signinglog (
		id             serial primary key not null,
		serial_number  varchar(200) not null,
		key_id         varchar(200) not null,
		created        timestamp default current_timestamp
	)
`

const createSigningLogSerialNumberIndexSQL = "CREATE INDEX IF NOT EXISTS serialnumber_idx ON signinglog (serial_number)"
const createSigningLogKeyIDIndexSQL = "CREATE INDEX IF NOT EXISTS keyid_idx ON signinglog (key_id)"

// SigningLog holds the details of the serial number and public key fingerprint that were supplied
// in a serial assertion for signing. The details are stored in the local database,
type SigningLog struct {
	ID           int
	SerialNumber string
	KeyID        string
	Created      time.Time
}

// CreateSigningLogTable creates the database table for a signing log with its indexes.
func (db *DB) CreateSigningLogTable() error {
	_, err := db.Exec(createSigningLogTableSQL)
	if err != nil {
		return err
	}
	_, err = db.Exec(createSigningLogSerialNumberIndexSQL)
	if err != nil {
		return err
	}
	_, err = db.Exec(createSigningLogKeyIDIndexSQL)

	return err
}
