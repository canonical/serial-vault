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

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

const createTestLogTableSQL = `
	CREATE TABLE IF NOT EXISTS testlog (
		id               serial primary key not null,
		brand_id         varchar(200) not null,
		model            varchar(200) not null,
		filename         varchar(200) not null,
		data             text,
		created          timestamp default current_timestamp,
		synced           timestamp
	)
`

const createTestLogSQLite = "INSERT INTO testlog (id,brand_id,model,filename,data) VALUES ($1, $2, $3, $4, $5)"
const createTestLogSQL = "INSERT INTO testlog (brand_id,model,filename,data) VALUES ($1, $2, $3, $4)"

const listTestLogSQL = "SELECT id,brand_id,model,filename,data,created FROM testlog WHERE synced IS NULL"
const listTestLogForUserSQL = `
	SELECT t.id, t.brand_id, t.model, t.filename, t.data, t.created FROM testlog t
	WHERE EXISTS(
		SELECT * FROM account acc
		INNER JOIN useraccountlink ua on ua.account_id=acc.id
		INNER JOIN userinfo u on ua.user_id=u.id
		WHERE acc.authority_id=t.brand_id and u.username=$1
	) AND synced IS NULL
`
const maxIDTestLogSQLite = "SELECT COUNT(*)+1 from testlog"
const deleteTestLogSQL = "DELETE FROM testlog WHERE id = $1"
const updateTestLogSyncedSQL = `
	UPDATE testlog t SET synced=current_timestamp
	WHERE EXISTS(
		SELECT * FROM account acc
		INNER JOIN useraccountlink ua on ua.account_id=acc.id
		INNER JOIN userinfo u on ua.user_id=u.id
		WHERE acc.authority_id=t.brand_id and u.username=$2
	) AND t.id = $1
`

// TestLog holds a test log sync-ed from the factory
type TestLog struct {
	ID       int       `json:"id"`
	Brand    string    `json:"brand_id"`
	Model    string    `json:"model"`
	Filename string    `json:"filename"`
	Data     string    `json:"data"`
	Created  time.Time `json:"created"`
	Synced   time.Time `json:"synced"` // used to indicate it has been synced to an external system
}

// CreateTestLogTable creates the database table for a test log
func (db *DB) CreateTestLogTable() error {
	_, err := db.Exec(createTestLogTableSQL)
	return err
}

// CreateTestLog keeps a record of a test log
func (db *DB) CreateTestLog(testLog TestLog) error {
	var err error
	// Validate the data
	if !validateStringsNotEmpty(testLog.Brand, testLog.Model, testLog.Filename, testLog.Data) {
		return errors.New("The brand, model, filename and file (base64-encoded) must be supplied")
	}

	// Create the signing log in the database
	if InFactory() {
		// Need to generate our own ID
		var nextID int
		err = db.QueryRow(maxIDTestLogSQLite).Scan(&nextID)
		if err != nil {
			log.Printf("Error retrieving next test log ID: %v\n", err)
			return err
		}

		_, err = db.Exec(createTestLogSQLite, nextID, testLog.Brand, testLog.Model, testLog.Filename, testLog.Data)
	} else {
		_, err = db.Exec(createTestLogSQL, testLog.Brand, testLog.Model, testLog.Filename, testLog.Data)
	}

	// Create the log in the database
	if err != nil {
		log.Printf("Error creating the test log: %v\n", err)
		return err
	}

	return nil
}

func (db *DB) listAllTestLog() ([]TestLog, error) {
	return db.listTestLogFilteredByUser(anyUserFilter)
}

func (db *DB) listTestLogFilteredByUser(username string) ([]TestLog, error) {
	testLogs := []TestLog{}

	var (
		rows *sql.Rows
		err  error
	)

	if len(username) == 0 {
		rows, err = db.Query(listTestLogSQL)
	} else {
		rows, err = db.Query(listTestLogForUserSQL, username)
	}
	if err != nil {
		log.Printf("Error retrieving test logs: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		testLog := TestLog{}
		err := rows.Scan(&testLog.ID, &testLog.Brand, &testLog.Model, &testLog.Filename, &testLog.Data, &testLog.Created)
		if err != nil {
			return nil, err
		}
		testLogs = append(testLogs, testLog)
	}

	return testLogs, nil
}

// SyncDeleteTestLog remove a test log from the factory
func (db *DB) SyncDeleteTestLog(ID int) error {
	if Environ.Config.Driver != "sqlite3" {
		return errors.New("Only valid within a factory")
	}

	_, err := db.Exec(deleteTestLogSQL, ID)
	return err
}
