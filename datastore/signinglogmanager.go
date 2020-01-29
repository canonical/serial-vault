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

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/CanonicalLtd/serial-vault/service/log"
	sq "github.com/Masterminds/squirrel"
)

const createSigningLogTableSQL = `
	CREATE TABLE IF NOT EXISTS signinglog (
		id             serial primary key not null,
		make           varchar(200) not null,
		model          varchar(200) not null,		
		serial_number  varchar(200) not null,
		fingerprint    varchar(200) not null,
		created        timestamp default current_timestamp,
		revision       int default 1,
		synced         int default 0
	)
`

// Additional columns
const alterSigningLogAddRevisionSQL = "ALTER TABLE signinglog ADD COLUMN revision int default 1"
const alterSigningLogAddSyncedSQL = "ALTER TABLE signinglog ADD COLUMN synced int default 0"

// MaxFromID is the maximum ID value
const MaxFromID = 2147483647

// Indexes
const createSigningLogSerialNumberIndexSQL = "CREATE INDEX IF NOT EXISTS serialnumber_idx ON signinglog (make,model,serial_number)"
const createSigningLogFingerprintIndexSQL = "CREATE INDEX IF NOT EXISTS fingerprint_idx ON signinglog (fingerprint)"
const createSigningLogCreatedIndexSQL = "CREATE INDEX IF NOT EXISTS created_idx ON signinglog (created)"

// Queries
const findMatchingSigningLogSQL = "SELECT EXISTS(SELECT * FROM signinglog where make=$1 and model=$2 and serial_number=$3 and revision=$4)"
const findExistingSigningLogSQL = "SELECT EXISTS(SELECT * FROM signinglog where (make=$1 and model=$2 and serial_number=$3) or fingerprint=$4)"
const findMaxRevisionSigningLogSQL = "SELECT COALESCE(MAX(revision), 0) FROM signinglog where make=$1 and model=$2 and serial_number=$3"
const maxIDSigningLogSQLite = "SELECT COUNT(*)+1 from signinglog"
const createSigningLogSQLite = "INSERT INTO signinglog (id, make, model, serial_number, fingerprint,revision) VALUES ($1, $2, $3, $4, $5, $6)"
const createSigningLogSQL = "INSERT INTO signinglog (make, model, serial_number, fingerprint,revision) VALUES ($1, $2, $3, $4, $5)"
const createSigningLogSyncSQL = "INSERT INTO signinglog (make, model, serial_number, fingerprint,revision,created) VALUES ($1, $2, $3, $4, $5, $6)"

const listSigningLogSQL = `
	SELECT *, count(*) OVER() AS total_count
	FROM signinglog 
	WHERE id < $1 
	ORDER BY id DESC 
	OFFSET $2 LIMIT 50
`

const listSigningLogForUserSQL = `
	SELECT s.*, count(*) OVER() AS total_count
	FROM signinglog s
	WHERE id < $1 and EXISTS(
		SELECT * FROM account acc
		INNER JOIN useraccountlink ua on ua.account_id=acc.id
		INNER JOIN userinfo u on ua.user_id=u.id
		WHERE acc.authority_id=s.make and u.username=$2
	)
	ORDER BY id DESC 
	OFFSET $3 LIMIT 50`

const listSigningLogForAccountForUserSQL = `
	SELECT s.*, count(*) OVER() AS total_count 
	FROM signinglog s
	WHERE id < $1 and EXISTS(
		SELECT * FROM account acc
		INNER JOIN useraccountlink ua on ua.account_id=acc.id
		INNER JOIN userinfo u on ua.user_id=u.id
		WHERE acc.authority_id=s.make and u.username=$2
	)
	AND s.make=$3
	ORDER BY id DESC OFFSET $4 LIMIT 50`

const deleteSigningLogSQL = "DELETE FROM signinglog WHERE id=$1"

const filterValuesModelSigningLogSQL = "SELECT DISTINCT model FROM signinglog WHERE make=$1 ORDER BY model"
const filterValuesModelSigningLogForUserSQL = `
	SELECT DISTINCT model FROM signinglog s
	WHERE EXISTS(
		SELECT * FROM account acc
		INNER JOIN useraccountlink ua on ua.account_id=acc.id
		INNER JOIN userinfo u on ua.user_id=u.id
		WHERE acc.authority_id=s.make and u.username=$1
	)
	AND s.make = $2
	ORDER BY model`
const syncSigningLogSQLite = "SELECT * FROM signinglog WHERE synced = 0"
const syncSigningLogUpdateSQLite = "UPDATE signinglog SET synced=1 WHERE id = $1"

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
	Synced       int       `json:"synced"`
	Total        int       `json:"total_count"`
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
	db.Exec(alterSigningLogAddSyncedSQL)

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

// CheckForMatching checks to see if a matching signing-log entry exists
// (same brand, model, serial number and revision)
func (db *DB) CheckForMatching(signLog SigningLog) (bool, error) {
	var duplicateExists bool
	err := db.QueryRow(findMatchingSigningLogSQL, signLog.Make, signLog.Model, signLog.SerialNumber, signLog.Revision).Scan(&duplicateExists)
	if err != nil {
		log.Printf("Error checking signinglog for matching record: %v\n", err)
		return false, errors.New("Error communicating with the database")
	}

	return duplicateExists, nil
}

// CreateSigningLog logs that a specific serial number has been used, along with the device-key fingerprint.
func (db *DB) CreateSigningLog(signLog SigningLog) error {
	var err error
	// Validate the data
	if !validateStringsNotEmpty(signLog.Make, signLog.Model, signLog.SerialNumber, signLog.Fingerprint) {
		return errors.New("The Make, Model, Serial Number and device-key Fingerprint must be supplied")
	}

	// Create the signing log in the database
	if InFactory() {
		// Need to generate our own ID
		var nextID int
		err = db.QueryRow(maxIDSigningLogSQLite).Scan(&nextID)
		if err != nil {
			log.Printf("Error retrieving next signing-log ID: %v\n", err)
			return err
		}

		_, err = db.Exec(createSigningLogSQLite, nextID, signLog.Make, signLog.Model, signLog.SerialNumber, signLog.Fingerprint, signLog.Revision)
	} else {
		_, err = db.Exec(createSigningLogSQL, signLog.Make, signLog.Model, signLog.SerialNumber, signLog.Fingerprint, signLog.Revision)
	}

	// Create the log in the database
	if err != nil {
		log.Printf("Error creating the signing log: %v\n", err)
		return err
	}

	return nil
}

// CreateSigningLogSync logs that a specific serial number has been used, along with the device-key fingerprint.
func (db *DB) CreateSigningLogSync(signLog SigningLog) error {
	var err error
	// Validate the data
	if !validateStringsNotEmpty(signLog.Make, signLog.Model, signLog.SerialNumber, signLog.Fingerprint) {
		return errors.New("The Make, Model, Serial Number and device-key Fingerprint must be supplied")
	}

	// Create the signing log in the database
	_, err = db.Exec(createSigningLogSyncSQL, signLog.Make, signLog.Model, signLog.SerialNumber, signLog.Fingerprint, signLog.Revision, signLog.Created)
	if err != nil {
		log.Printf("Error creating the signing log: %v\n", err)
		return err
	}

	return nil
}

func (db *DB) listAllSigningLog(params *SigningLogParams) ([]SigningLog, error) {
	return db.listSigningLogFilteredByUser(anyUserFilter, params)
}

// TODO: tests this!
func (db *DB) listSigningLogFilteredByUser(username string, params *SigningLogParams) ([]SigningLog, error) {
	signingLogs := []SigningLog{}

	var (
		rows *sql.Rows
		err  error
	)

	if len(username) == 0 {
		rows, err = db.Query(listSigningLogSQL, MaxFromID, params.Offset)
	} else {
		rows, err = db.Query(listSigningLogForUserSQL, MaxFromID, username, params.Offset)
	}
	if err != nil {
		log.Printf("Error retrieving signing logs: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		signingLog := SigningLog{}
		err := rows.Scan(&signingLog.ID, &signingLog.Make, &signingLog.Model,
			&signingLog.SerialNumber, &signingLog.Fingerprint, &signingLog.Created,
			&signingLog.Revision, &signingLog.Synced, &signingLog.Total)
		if err != nil {
			return nil, err
		}
		signingLogs = append(signingLogs, signingLog)
	}

	return signingLogs, nil
}

func (db *DB) listAllSigningLogForAccount(authorityID string, params *SigningLogParams) ([]SigningLog, error) {
	return db.listSigningLogForAccountFilteredByUser(anyUserFilter, authorityID, params)
}

// TODO: add test for the sql!
func signingLogSQLBuilder(username, authorityID string, params *SigningLogParams) sq.SelectBuilder {

	// const listSigningLogForAccountSQL = "SELECT * FROM signinglog WHERE id < $1 AND make=$2 ORDER BY id DESC LIMIT 10000"
	// db.Query(listSigningLogForAccountSQL, MaxFromID, authorityID)
	listSigningLogSQL := sq.
		Select("*", "count(*) OVER() AS total_count").
		From("signinglog s").          // FROM signinglog s
		Where(sq.Lt{"id": MaxFromID}). // WHERE id < $1
		Where("make=?", authorityID).  // AND make=$2
		OrderBy("id DESC").
		Limit(50).Offset(uint64(params.Offset)).
		PlaceholderFormat(sq.Dollar)

	if username != "" {
		nestedBuilder := sq.Select("*").Prefix("EXISTS (").
			From("account acc").
			JoinClause("INNER JOIN useraccountlink ua on ua.account_id=acc.id").
			JoinClause("INNER JOIN userinfo u on ua.user_id=u.id").
			Where("acc.authority_id=s.make AND u.username=?", username).
			Suffix(")").PlaceholderFormat(sq.Dollar)

		listSigningLogSQL = listSigningLogSQL.Where(nestedBuilder)
	}
	if len(params.Filter) > 0 {
		listSigningLogSQL = listSigningLogSQL.Where(sq.Eq{"model": params.Filter})
	}
	if params.Serialnumber != "" {
		listSigningLogSQL = listSigningLogSQL.Where(sq.Like{"serial_number": fmt.Sprintf("%s%%", params.Serialnumber)})
	}

	return listSigningLogSQL
}

func (db *DB) listSigningLogForAccountFilteredByUser(username, authorityID string, params *SigningLogParams) ([]SigningLog, error) {
	signingLogs := []SigningLog{}

	// if len(username) == 0 {
	listSigningLogSQL := signingLogSQLBuilder(username, authorityID, params)
	rows, err := listSigningLogSQL.RunWith(db).Query()
	if err != nil {
		log.Printf("Error retrieving signing logs: %v\n", err)
		return nil, err
	}
	for rows.Next() {
		signingLog := SigningLog{}
		err := rows.Scan(&signingLog.ID, &signingLog.Make, &signingLog.Model,
			&signingLog.SerialNumber, &signingLog.Fingerprint, &signingLog.Created,
			&signingLog.Revision, &signingLog.Synced, &signingLog.Total)
		if err != nil {
			log.Printf("Error retrieving signing logs: %v\n", err)
			return nil, err
		}
		signingLogs = append(signingLogs, signingLog)
	}
	// } else {
	// SELECT s.*, count(*) OVER() AS total_count
	// FROM signinglog s
	// WHERE id < $1 and EXISTS(
	// 	SELECT * FROM account acc
	// 	INNER JOIN useraccountlink ua on ua.account_id=acc.id
	// 	INNER JOIN userinfo u on ua.user_id=u.id
	// 	WHERE acc.authority_id=s.make and u.username=$2
	// )
	// AND s.make=$3
	// ORDER BY id DESC OFFSET $4 LIMIT 50
	// rows, err = db.Query(listSigningLogForAccountForUserSQL, MaxFromID, username, authorityID, offset)
	// }
	// if err != nil {
	// 	log.Printf("Error retrieving signing logs: %v\n", err)
	// 	return nil, err
	// }
	// defer rows.Close()

	// for rows.Next() {
	// 	signingLog := SigningLog{}
	// 	err := rows.Scan(&signingLog.ID, &signingLog.Make, &signingLog.Model,
	// 		&signingLog.SerialNumber, &signingLog.Fingerprint, &signingLog.Created,
	// 		&signingLog.Revision, &signingLog.Synced, &signingLog.Total)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	signingLogs = append(signingLogs, signingLog)
	// }

	return signingLogs, nil
}

func (db *DB) allSigningLogFilterValues(authorityID string) (SigningLogFilters, error) {
	return db.signingLogFilterValuesFilteredByUser(anyUserFilter, authorityID)
}

func (db *DB) signingLogFilterValuesFilteredByUser(username, authorityID string) (SigningLogFilters, error) {
	filters := SigningLogFilters{}

	var modelsSQL string

	if len(username) == 0 {
		modelsSQL = filterValuesModelSigningLogSQL
	} else {
		modelsSQL = filterValuesModelSigningLogForUserSQL
	}

	err := db.filterValuesForField(username, modelsSQL, authorityID, &filters.Models)
	if err != nil {
		log.Printf("Error retrieving filter values: %v\n", err)
		return filters, err
	}

	return filters, nil
}

func (db *DB) filterValuesForField(username, sqlQuery, authorityID string, fieldValues *[]string) error {

	var (
		rows *sql.Rows
		err  error
	)
	values := []string{}

	if len(username) == 0 {
		rows, err = db.Query(sqlQuery, authorityID)
	} else {
		rows, err = db.Query(sqlQuery, username, authorityID)
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

// SyncSigningLog fetches the factory signing logs to sync with the cloud
func (db *DB) SyncSigningLog() ([]SigningLog, error) {
	signingLogs := []SigningLog{}

	rows, err := db.Query(syncSigningLogSQLite)
	if err != nil {
		log.Printf("Error retrieving signing logs: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		signingLog := SigningLog{}
		err := rows.Scan(&signingLog.ID, &signingLog.Make, &signingLog.Model,
			&signingLog.SerialNumber, &signingLog.Fingerprint, &signingLog.Created,
			&signingLog.Revision, &signingLog.Synced, &signingLog.Total)
		if err != nil {
			return nil, err
		}
		signingLogs = append(signingLogs, signingLog)
	}

	return signingLogs, nil
}

// SyncUpdateSigningLog updates the synced status of a signing log
func (db *DB) SyncUpdateSigningLog(id int) error {
	_, err := db.Exec(syncSigningLogUpdateSQLite, id)
	return err
}
