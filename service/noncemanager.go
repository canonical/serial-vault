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

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strconv"
	"time"
)

const createDeviceNonceTableSQL = `
	CREATE TABLE IF NOT EXISTS devicenonce (
		id             serial primary key not null,
		nonce          varchar(200) not null,
		timestamp      int not null,		
		created        timestamp default current_timestamp
	)
`

// Queries
const createDeviceNonceSQL = "INSERT INTO devicenonce (nonce, timestamp) VALUES ($1, $2)"

// DeviceNonce holds the details of the nonce, combining a timestamp and random text
type DeviceNonce struct {
	ID        int
	Nonce     string
	TimeStamp int64
	Created   time.Time
}

// CreateDeviceNonceTable creates the database table for nonces with its indexes.
func (db *DB) CreateDeviceNonceTable() error {
	_, err := db.Exec(createDeviceNonceTableSQL)
	if err != nil {
		return err
	}
	return err
}

// CreateDeviceNonce stores a new nonce entry
func (db *DB) CreateDeviceNonce() (DeviceNonce, error) {
	// Generate a nonce with a timestamp and random string
	nonce := generateNonce()

	// Create the nonce in the database
	_, err := db.Exec(createDeviceNonceSQL, nonce.Nonce, nonce.TimeStamp)
	if err != nil {
		log.Printf("Error creating the nonce: %v\n", err)
		return DeviceNonce{}, err
	}

	return nonce, nil
}

func generateNonce() DeviceNonce {
	h := sha1.New()
	timestamp := time.Now().Unix()
	io.WriteString(h, strconv.FormatInt(rand.Int63(), 10))
	io.WriteString(h, strconv.FormatInt(timestamp, 10))
	nonce := fmt.Sprintf("%x", h.Sum(nil))

	return DeviceNonce{Nonce: nonce, TimeStamp: timestamp}
}
