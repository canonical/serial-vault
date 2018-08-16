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
	"errors"
	"log"
	"time"
)

const createModelAssertTableSQL = `
	CREATE TABLE IF NOT EXISTS modelassertion (
		id               serial primary key not null,
		model_id         int references model not null,
		keypair_id       int references keypair not null,
		series           int not null,
		architecture     varchar(20) not null,
		revision         int not null default 0,
		gadget           varchar(60) not null,
		kernel           varchar(60) not null,
		store            varchar(60),
		required_snaps   text default '',
		base             varchar(20) default '',
		classic          varchar(10) default '',
		display_name     varchar(200) default '',
		created          timestamp default current_timestamp,
		modified         timestamp default current_timestamp
	)
`
const createModelAssertSQL = `
INSERT INTO modelassertion 
(model_id,keypair_id,series,architecture,revision,gadget,kernel,store,required_snaps,base,classic,display_name) 
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) 
RETURNING id`

const updateModelAssertSQL = `
UPDATE modelassertion
SET model_id=$2, keypair_id=$3, series=$4, architecture=$5, revision=$6, gadget=$7, kernel=$8, store=$9, modified=$10, required_snaps=$11, base=$12, classic=$13, display_name=$14 
WHERE id=$1`

const getModelAssertSQL = `
SELECT id,model_id,keypair_id,series,architecture,revision,gadget,kernel,store,required_snaps,base,classic,display_name,created,modified
FROM modelassertion
WHERE model_id=$1
`

const deleteModelAssertSQL = `
DELETE FROM modelassertion
WHERE model_id=$1
`

// Add the UC18 fields to the model assertion
const alterModelAssertUC18Fields = `
ALTER TABLE modelassertion 
ADD COLUMN base varchar(20) default '',
ADD COLUMN classic varchar(10) default '',
ADD COLUMN display_name varchar(200) default ''
`

// ModelAssertion holds the model assertion details in the local database
type ModelAssertion struct {
	ID            int       `json:"id"`
	ModelID       int       `json:"model_id"`
	KeypairID     int       `json:"keypair_id"`
	Series        int       `json:"series"`
	Architecture  string    `json:"architecture"`
	Revision      int       `json:"revision"`
	Gadget        string    `json:"gadget"`
	Kernel        string    `json:"kernel"`
	Store         string    `json:"store"`
	RequiredSnaps string    `json:"required_snaps"`
	Base          string    `json:"base"`
	Classic       string    `json:"classic"`
	DisplayName   string    `json:"display_name"`
	Created       time.Time `json:"created"`
	Modified      time.Time `json:"modified"`
}

// CreateModelAssertTable creates the database table for a model assertion
func (db *DB) CreateModelAssertTable() error {
	_, err := db.Exec(createModelAssertTableSQL)
	return err
}

// AlterModelAssertTable updates an existing database model assertion table with additional fields
func (db *DB) AlterModelAssertTable() error {
	// Ignore error as the fields may already exist
	db.Exec(alterModelAssertUC18Fields)

	return nil
}

// CreateModelAssert adds a model assertion record to allow generation of a signed assertion
func (db *DB) CreateModelAssert(m ModelAssertion) (int, error) {
	var createdID int
	err := db.QueryRow(createModelAssertSQL, m.ModelID, m.KeypairID, m.Series, m.Architecture, m.Revision, m.Gadget, m.Kernel, m.Store, m.RequiredSnaps, m.Base, m.Classic, m.DisplayName).Scan(&createdID)
	if err != nil {
		log.Printf("Error creating the model assertion: %v\n", err)
	}
	return createdID, err
}

// UpdateModelAssert updates the model assertion details
func (db *DB) UpdateModelAssert(m ModelAssertion) error {
	var err error

	_, err = db.Exec(updateModelAssertSQL, m.ID, m.ModelID, m.KeypairID, m.Series, m.Architecture, m.Revision, m.Gadget, m.Kernel, m.Store, time.Now().UTC(), m.RequiredSnaps, m.Base, m.Classic, m.DisplayName)

	if err != nil {
		log.Printf("Error updating the model assertion: %v\n", err)
	}

	return err

}

// UpsertModelAssert creates or updates the model assertion headers
func (db *DB) UpsertModelAssert(m ModelAssertion) error {
	var err error

	if err = validateModelAssertion(m); err != nil {
		return err
	}

	if m.ID > 0 {
		err = db.UpdateModelAssert(m)
	} else {
		_, err = db.CreateModelAssert(m)
	}

	return err

}

// deleteModelAssert deletes the model assertion details
func (db *DB) deleteModelAssert(modelID int) error {
	var err error

	_, err = db.Exec(deleteModelAssertSQL, modelID)
	if err != nil {
		log.Printf("Error deleting the model assertion: %v\n", err)
	}
	return err
}

// GetModelAssert fetches the model assertion
func (db *DB) GetModelAssert(modelID int) (ModelAssertion, error) {
	m := ModelAssertion{}
	err := db.QueryRow(getModelAssertSQL, modelID).Scan(&m.ID, &m.ModelID, &m.KeypairID, &m.Series, &m.Architecture, &m.Revision, &m.Gadget, &m.Kernel, &m.Store, &m.RequiredSnaps, &m.Base, &m.Classic, &m.DisplayName, &m.Created, &m.Modified)
	if err != nil {
		return m, err
	}

	return m, nil
}

func validateModelAssertion(m ModelAssertion) error {
	if m.ModelID <= 0 {
		return errors.New("Model must be provided")
	}
	if m.KeypairID <= 0 {
		return errors.New("Signing Key must be provided")
	}
	if m.Series < 16 {
		return errors.New("Series must be at least 16")
	}

	if err := validateNotEmpty("Architecture", m.Architecture); err != nil {
		return err
	}
	if err := validateNotEmpty("Gadget", m.Gadget); err != nil {
		return err
	}
	if err := validateNotEmpty("Kernel", m.Kernel); err != nil {
		return err
	}
	return validateNotEmpty("Store", m.Store)
}
