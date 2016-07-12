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
	"errors"
	"log"
	"strings"
	"time"
)

const createSigningLogTableSQL = `
	CREATE TABLE IF NOT EXISTS signinglog (
		id             serial primary key not null,
		make           varchar(200) not null,
		model          varchar(200) not null,		
		serial_number  varchar(200) not null,
		fingerprint    varchar(200) not null,
		created        timestamp default current_timestamp
	)
`

const createSigningLogSerialNumberIndexSQL = "CREATE INDEX IF NOT EXISTS serialnumber_idx ON signinglog (make,model,serial_number)"
const createSigningLogFingerprintIndexSQL = "CREATE INDEX IF NOT EXISTS fingerprint_idx ON signinglog (fingerprint)"
const findExistingSigningLogSQL = "SELECT EXISTS(SELECT * FROM signinglog where (make=$1 and model=$2 and serial_number=$3) or fingerprint=$4)"
const createSigningLogSQL = "INSERT INTO signinglog (make, model, serial_number, fingerprint) VALUES ($1, $2, $3, $4)"

// SigningLog holds the details of the serial number and public key fingerprint that were supplied
// in a serial assertion for signing. The details are stored in the local database,
type SigningLog struct {
	ID           int
	Make         string
	Model        string
	SerialNumber string
	Fingerprint  string
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
	_, err = db.Exec(createSigningLogFingerprintIndexSQL)

	return err
}

// CheckForDuplicate verifies that the serial number and the device-key fingerprint have not be used previously
func (db *DB) CheckForDuplicate(signLog SigningLog) (bool, error) {
	var duplicateExists bool
	err := db.QueryRow(findExistingSigningLogSQL, signLog.Make, signLog.Model, signLog.SerialNumber, signLog.Fingerprint).Scan(&duplicateExists)
	if err != nil {
		log.Printf("Error checking signinglog for duplicate: %v\n", err)
		return false, errors.New("Error communicating with the database")
	}
	return duplicateExists, nil
}

// CreateSigningLog logs that a specific serial number has been used, along with the device-key fingerprint.
func (db *DB) CreateSigningLog(signLog SigningLog) error {

	// Validate the data
	if strings.TrimSpace(signLog.Make) == "" || strings.TrimSpace(signLog.Model) == "" || strings.TrimSpace(signLog.SerialNumber) == "" || strings.TrimSpace(signLog.Fingerprint) == "" {
		return errors.New("The Make, Model, Serial Number and device-key Fingerprint must be supplied")
	}

	// Create the log in the database
	_, err := db.Exec(createSigningLogSQL, signLog.Make, signLog.Model, signLog.SerialNumber, signLog.Fingerprint)
	if err != nil {
		log.Printf("Error creating the signing log: %v\n", err)
		return err
	}

	return nil
}
