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

import (
	"database/sql"
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

// Additional columns
const alterSigningLogAddRevisionSQL = "ALTER TABLE signinglog ADD COLUMN revision int default 1"

// MaxFromID is the maximum ID value
const MaxFromID = 2147483647

// Indexes
const createSigningLogSerialNumberIndexSQL = "CREATE INDEX IF NOT EXISTS serialnumber_idx ON signinglog (make,model,serial_number)"
const createSigningLogFingerprintIndexSQL = "CREATE INDEX IF NOT EXISTS fingerprint_idx ON signinglog (fingerprint)"
const createSigningLogCreatedIndexSQL = "CREATE INDEX IF NOT EXISTS created_idx ON signinglog (created)"

// Queries
const findExistingSigningLogSQL = "SELECT EXISTS(SELECT * FROM signinglog where (make=$1 and model=$2 and serial_number=$3) or fingerprint=$4)"
const findMaxRevisionSigningLogSQL = "SELECT COALESCE(MAX(revision), 0) FROM signinglog where make=$1 and model=$2 and serial_number=$3"
const createSigningLogSQL = "INSERT INTO signinglog (make, model, serial_number, fingerprint,revision) VALUES ($1, $2, $3, $4, $5)"
const listSigningLogSQL = "SELECT * FROM signinglog WHERE id < $1 ORDER BY id DESC LIMIT 10000"
const listSigningLogForUserSQL = `
	SELECT s.* FROM signinglog s
	WHERE id < $1 and EXISTS(
		SELECT * FROM account acc
		INNER JOIN useraccountlink ua on ua.account_id=acc.id
		INNER JOIN userinfo u on ua.user_id=u.id
		WHERE acc.authority_id=s.make and u.username=$2 and u.userrole >= $3
	)
	ORDER BY id DESC LIMIT 10000`
const deleteSigningLogSQL = "DELETE FROM signinglog WHERE id=$1"
const filterValuesMakeSigningLogSQL = "SELECT DISTINCT make FROM signinglog ORDER BY make"
const filterValuesMakeSigningLogForUserSQL = `
	SELECT DISTINCT make FROM signinglog s
	WHERE EXISTS(
		SELECT * FROM account acc
		INNER JOIN useraccountlink ua on ua.account_id=acc.id
		INNER JOIN userinfo u on ua.user_id=u.id
		WHERE acc.authority_id=s.make and u.username=$1 and u.userrole >= $2
	)
	ORDER BY make`
const filterValuesModelSigningLogSQL = "SELECT DISTINCT model FROM signinglog ORDER BY model"
const filterValuesModelSigningLogForUserSQL = `
	SELECT DISTINCT model FROM signinglog s
	WHERE EXISTS(
		SELECT * FROM account acc
		INNER JOIN useraccountlink ua on ua.account_id=acc.id
		INNER JOIN userinfo u on ua.user_id=u.id
		WHERE acc.authority_id=s.make and u.username=$1 and u.userrole >= $2
	)
	ORDER BY model`

// SigningLog holds the details of the serial number and public key fingerprint that were supplied
// in a serial assertion for signing. The details are stored in the local database,
type SigningLog struct {
	ID           int       `json:"id"`
	Make         string    `json:"make"`
	Model        string    `json:"model"`
	SerialNumber string    `json:"serialnumber"`
	Fingerprint  string    `json:"fingerprint"`
	Created      time.Time `json:"created"`
	Revision     int       `json:"revision"`
}

// SigningLogFilters holds the values of the filters for the searchable columns
type SigningLogFilters struct {
	Makes  []string `json:"makes"`
	Models []string `json:"models"`
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
	_, err = db.Exec(createSigningLogCreatedIndexSQL)
	if err != nil {
		return err
	}
	_, err = db.Exec(createSigningLogFingerprintIndexSQL)
	if err != nil {
		return err
	}

	// Ignoring the error when adding the column
	db.Exec(alterSigningLogAddRevisionSQL)

	return nil
}

// CheckForDuplicate verifies that the serial number and the device-key fingerprint have not be used previously.
// If a duplicate serial number does exist, it returns the maximum revision number for the serial number.
func (db *DB) CheckForDuplicate(signLog *SigningLog) (bool, int, error) {
	var duplicateExists bool
	var maxRevision int
	err := db.QueryRow(findExistingSigningLogSQL, signLog.Make, signLog.Model, signLog.SerialNumber, signLog.Fingerprint).Scan(&duplicateExists)
	if err != nil {
		log.Printf("Error checking signinglog for duplicate: %v\n", err)
		return false, 0, errors.New("Error communicating with the database")
	}

	// If we do have a duplicate, we need to find the maximum revision number
	err = db.QueryRow(findMaxRevisionSigningLogSQL, signLog.Make, signLog.Model, signLog.SerialNumber).Scan(&maxRevision)
	if err != nil {
		log.Printf("Error checking signinglog for maximum revision number of the serial: %v\n", err)
		return false, 0, errors.New("Error communicating with the database")
	}

	return duplicateExists, maxRevision, nil
}

// CreateSigningLog logs that a specific serial number has been used, along with the device-key fingerprint.
func (db *DB) CreateSigningLog(signLog SigningLog) error {

	// Validate the data
	if strings.TrimSpace(signLog.Make) == "" || strings.TrimSpace(signLog.Model) == "" || strings.TrimSpace(signLog.SerialNumber) == "" || strings.TrimSpace(signLog.Fingerprint) == "" {
		return errors.New("The Make, Model, Serial Number and device-key Fingerprint must be supplied")
	}

	// Create the log in the database
	_, err := db.Exec(createSigningLogSQL, signLog.Make, signLog.Model, signLog.SerialNumber, signLog.Fingerprint, signLog.Revision)
	if err != nil {
		log.Printf("Error creating the signing log: %v\n", err)
		return err
	}

	return nil
}

// ListSigningLog returns a list of signing log records from a specific date/time.
// The fromId parameter is used enables the use of indexes for more efficient pagination.
func (db *DB) ListSigningLog(username string) ([]SigningLog, error) {
	signingLogs := []SigningLog{}

	var (
		rows *sql.Rows
		err  error
	)

	if len(username) == 0 {
		rows, err = db.Query(listSigningLogSQL, MaxFromID)
	} else {
		rows, err = db.Query(listSigningLogForUserSQL, MaxFromID, username, Admin)
	}
	if err != nil {
		log.Printf("Error retrieving signing logs: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		signingLog := SigningLog{}
		err := rows.Scan(&signingLog.ID, &signingLog.Make, &signingLog.Model, &signingLog.SerialNumber, &signingLog.Fingerprint, &signingLog.Created, &signingLog.Revision)
		if err != nil {
			return nil, err
		}
		signingLogs = append(signingLogs, signingLog)
	}

	return signingLogs, nil
}

// DeleteSigningLog deletes a signing log record.
func (db *DB) DeleteSigningLog(signingLog SigningLog) (string, error) {

	_, err := db.Exec(deleteSigningLogSQL, signingLog.ID)
	if err != nil {
		log.Printf("Error deleting the database signing log: %v\n", err)
		return "", err
	}

	return "", nil
}

// SigningLogFilterValues returns the unique values of the main filterable columns
func (db *DB) SigningLogFilterValues(username string) (SigningLogFilters, error) {
	filters := SigningLogFilters{}

	var (
		makesSQL  string
		modelsSQL string
	)

	if len(username) == 0 {
		makesSQL = filterValuesMakeSigningLogSQL
		modelsSQL = filterValuesModelSigningLogSQL
	} else {
		makesSQL = filterValuesMakeSigningLogForUserSQL
		modelsSQL = filterValuesModelSigningLogForUserSQL
	}

	err := db.filterValuesForField(username, makesSQL, &filters.Makes)
	if err != nil {
		log.Printf("Error retrieving filter values: %v\n", err)
		return filters, err
	}

	err = db.filterValuesForField(username, modelsSQL, &filters.Models)
	if err != nil {
		log.Printf("Error retrieving filter values: %v\n", err)
		return filters, err
	}

	return filters, nil
}

func (db *DB) filterValuesForField(username string, sqlQuery string, fieldValues *[]string) error {

	var (
		rows *sql.Rows
		err  error
	)
	values := []string{}

	if len(username) == 0 {
		rows, err = db.Query(sqlQuery)
	} else {
		rows, err = db.Query(sqlQuery, username, Admin)
	}

	if err != nil {
		log.Printf("Error retrieving filter values: %v\n", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var fieldValue string
		err := rows.Scan(&fieldValue)
		if err != nil {
			return err
		}
		values = append(values, fieldValue)
	}

	*fieldValues = values

	return nil
}
