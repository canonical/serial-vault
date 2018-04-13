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
	"errors"
	"log"
)

// Understood settings codes
var (
	SettingParentContext = "parent"
	SettingKeyContext    = "key"
)

const createSettingsTableSQL = `
	CREATE TABLE IF NOT EXISTS settings (
		id    serial primary key not null,
		code  varchar(200) not null,
		data  text
	)
`

const upsertSettingsSQL = `
	WITH upsert AS (
		update settings set code=$1, data=$2
		where code=$1
		RETURNING *
	)
	insert into settings (code,data)
	select $1, $2
	where not exists (select * from upsert)
`

// sqlite3 syntax for syncing data locally
const upsertSettingsSQLite = `
	INSERT INTO settings
	(id,code,data)
	VALUES ($1, $2, $3)
`

const maxIDSettingsSQLite = `
	SELECT COUNT(*)+1 from settings
`

const getSettingSQL = "select id, code, data from settings where code=$1"

// Setting holds the keypair reference details in the local database
type Setting struct {
	ID   int
	Code string
	Data string
}

// CreateSettingsTable creates the database table for a setting.
func (db *DB) CreateSettingsTable() error {
	_, err := db.Exec(createSettingsTableSQL)
	return err
}

// PutSetting stores a setting into the database
func (db *DB) PutSetting(setting Setting) error {
	var err error
	// Validate the data
	if err := validateNotEmpty("code", setting.Code); err != nil {
		return errors.New("The code must be entered to store a Setting")
	}

	if InFactory() {
		// We only add new settings for the factory, we don't ever update a setting
		// Need to generate our own ID
		var nextID int
		err = db.QueryRow(maxIDSettingsSQLite).Scan(&nextID)
		if err != nil {
			log.Printf("Error retrieving next setting ID: %v\n", err)
			return err
		}

		_, err = db.Exec(upsertSettingsSQLite, nextID, setting.Code, setting.Data)
	} else {
		_, err = db.Exec(upsertSettingsSQL, setting.Code, setting.Data)
	}

	if err != nil {
		log.Printf("Error updating the database setting: %v\n", err)
		return err
	}

	return nil
}

// GetSetting fetches a single setting from the database by code
func (db *DB) GetSetting(code string) (Setting, error) {
	setting := Setting{}

	err := db.QueryRow(getSettingSQL, code).Scan(&setting.ID, &setting.Code, &setting.Data)
	if err != nil {
		log.Printf("Error retrieving setting by code: %v\n", err)
		return setting, err
	}

	return setting, nil
}
