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
	"strings"
	"time"
)

const createTestLogTableSQL = `
	CREATE TABLE IF NOT EXISTS testlog (
		id               serial primary key not null,
		brand_id         varchar(200) not null,
		model            varchar(200) not null,
		filename         varchar(200) not null,
		data             text,
		logged           int,
		created          timestamp default current_timestamp,
		synced           timestamp
	)
`

const createTestLogSQLite = "INSERT INTO testlog (id,brand_id,model,filename,data,logged) VALUES ($1, $2, $3, $4, $5, $6)"
const createTestLogSQL = "INSERT INTO testlog (brand_id,model,filename,data,logged) VALUES ($1, $2, $3, $4, $5)"

const listTestLogSQL = "SELECT * FROM testlog WHERE not synced"
const listTestLogForUserSQL = `
	SELECT t.* FROM testlog t
	WHERE EXISTS(
		SELECT * FROM account acc
		INNER JOIN useraccountlink ua on ua.account_id=acc.id
		INNER JOIN userinfo u on ua.user_id=u.id
		WHERE acc.authority_id=t.brand_id and u.username=$1
	)
`
const maxIDTestLogSQLite = "SELECT COUNT(*)+1 from testlog"

// TestLog holds a test log sync-ed from the factory
type TestLog struct {
	ID       int       `json:"id"`
	Brand    string    `json:"brand_id"`
	Model    string    `json:"model"`
	Filename string    `json:"filename"`
	Data     string    `json:"data"`
	Logged   int       `json:"logged"`
	Created  time.Time `json:"created"`
	Synced   time.Time `json:"synced"` // used to indicate it has been synced to an external system
}

// CreateTestLogTable creates the database table for a test log
func (db *DB) CreateTestLogTable() error {
	_, err := db.Exec(createTestLogTableSQL)
	if err != nil {
		return err
	}

	return nil
}

// CreateTestLog keeps a record of a test log
func (db *DB) CreateTestLog(testLog TestLog) error {
	var err error
	// Validate the data
	if strings.TrimSpace(testLog.Brand) == "" || strings.TrimSpace(testLog.Model) == "" || strings.TrimSpace(testLog.Filename) == "" || strings.TrimSpace(testLog.Data) == "" {
		return errors.New("The brand, model, filename and file (base64-encoded) must be supplied")
	}

	// Create the signing log in the database
	if Environ.Config.Driver == "sqlite3" {
		// Need to generate our own ID
		var nextID int
		err = db.QueryRow(maxIDTestLogSQLite).Scan(&nextID)
		if err != nil {
			log.Printf("Error retrieving next test log ID: %v\n", err)
			return err
		}

		_, err = db.Exec(createTestLogSQLite, nextID, testLog.Brand, testLog.Model, testLog.Filename, testLog.Data, testLog.Logged)
	} else {
		_, err = db.Exec(createTestLogSQL, testLog.Brand, testLog.Model, testLog.Filename, testLog.Data, testLog.Logged)
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
		err := rows.Scan(&testLog.ID, &testLog.Brand, &testLog.Model, &testLog.Filename, &testLog.Data, &testLog.Logged, &testLog.Created, &testLog.Synced)
		if err != nil {
			return nil, err
		}
		testLogs = append(testLogs, testLog)
	}

	return testLogs, nil
}
