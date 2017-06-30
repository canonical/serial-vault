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
	"log"
	"time"
)

// OpenidNonceMaxAge is the maximum age of stored nonces. Any nonces older
// than this will automatically be rejected. Stored nonces older
// than this will periodically be purged from the database.
const OpenidNonceMaxAge = maxNonceAgeInSeconds * time.Second
const maxNonceAgeInSeconds = 60

// https://openid.net/specs/openid-authentication-2_0.html for more information
const createOpenidNonceTableSQL = `
	CREATE TABLE IF NOT EXISTS openidnonce (
		id             serial primary key not null,
		nonce          varchar(255) not null,
		endpoint       varchar(255) not null,
		timestamp      int not null
	)
`

// Indexes
const createOpenidNonceIndexSQL = "CREATE UNIQUE INDEX IF NOT EXISTS nonce_endpoint_idx ON openidnonce (nonce, endpoint)"

// Queries
const createOpenidNonceSQL = "INSERT INTO openidnonce (nonce, endpoint, timestamp) VALUES ($1, $2, $3)"
const deleteExpiredOpenidNonceSQL = "DELETE FROM openidnonce where timestamp<$1"

// OpenidNonce holds the details of the nonce, combining a timestamp and random text
type OpenidNonce struct {
	ID        int
	Nonce     string
	Endpoint  string
	TimeStamp int64
}

// CreateOpenidNonceTable creates the database table for nonces with its indexes.
func (db *DB) CreateOpenidNonceTable() error {
	// Create the table
	_, err := db.Exec(createOpenidNonceTableSQL)
	if err != nil {
		return err
	}

	// Create the index
	_, err = db.Exec(createOpenidNonceIndexSQL)
	return err
}

// CreateOpenidNonce stores a new nonce entry
func (db *DB) CreateOpenidNonce(nonce OpenidNonce) error {

	// Delete the expired nonces
	err := db.deleteExpiredOpenidNonces()
	if err != nil {
		log.Printf("Error checking expired openid nonces: %v\n", err)
		return err
	}

	// Create the nonce in the database
	_, err = db.Exec(createOpenidNonceSQL, nonce.Nonce, nonce.Endpoint, nonce.TimeStamp)
	if err != nil {
		log.Printf("Error creating the openid nonce: %v\n", err)
		return err
	}

	return nil
}

// deleteExpiredOpenidNonces removes nonces with timestamp older than max allowed lifetime
func (db *DB) deleteExpiredOpenidNonces() error {
	// Remove expired nonces from the table
	timestamp := time.Now().Unix() - maxNonceAgeInSeconds
	_, err := db.Exec(deleteExpiredOpenidNonceSQL, timestamp)
	if err != nil {
		log.Printf("Error deleting expired openid nonces: %v\n", err)
		return errors.New("Error communicating with the database")
	}

	return nil
}
