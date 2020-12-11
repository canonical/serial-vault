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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/snapcore/snapd/asserts"
)

// In the modelassertion table we store only the latest assertion.
// So it's enough to store in signed_modelassertion table only the latest signed assertion
const createSignedModelAssertTableSQL = `
	CREATE TABLE IF NOT EXISTS signed_modelassertion (
		model_id INT references model NOT NULL,
		revision INT NOT NULL,
		body BYTEA NOT NULL,
		headers JSONB NOT NULL,
	    content BYTEA NOT NULL,
	    signature BYTEA NOT NULL,
		ts TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
		UNIQUE (model_id)
	);
`

// Insert or overwrite signed assertion for a given model_id
const upsertSignedModelAssertSQL = `
INSERT INTO signed_modelassertion (model_id, revision, body, headers, content, signature) 
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (model_id) DO UPDATE 
SET revision=$2, body=$3, headers=$4, content=$5, signature=$6
`
const deleteSignedModelAssertSQL = `
DELETE FROM signed_modelassertion
WHERE model_id=$1
`

const getSignedModelAssertSQL = `
SELECT body, headers, content, signature 
FROM signed_modelassertion
WHERE model_id=$1`

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

// ModelAssertionHeadersForModel returns the model assertion headers for a model
// if the assertion part of the model is empty, the function will return the
// model assertion from the database
func ModelAssertionHeadersForModel(m Model) (map[string]interface{}, Keypair, error) {
	// Get the assertion headers for the model
	var err error
	assert := m.ModelAssertion
	if assert.ID == 0 {
		assert, err = Environ.DB.GetModelAssert(m.ID)
		if err != nil {
			return nil, Keypair{}, err
		}
	}
	// Get the keypair for the model assertion
	keypair, err := Environ.DB.GetKeypair(assert.KeypairID)
	if err != nil {
		return nil, keypair, err
	}

	// Create the model assertion header
	headers := map[string]interface{}{
		"type":              asserts.ModelType.Name,
		"authority-id":      m.BrandID,
		"brand-id":          m.BrandID,
		"series":            fmt.Sprintf("%d", assert.Series),
		"model":             m.Name,
		"store":             assert.Store,
		"sign-key-sha3-384": keypair.KeyID,
		"timestamp":         time.Now().Format(time.RFC3339),
	}

	// Add the optional fields as needed
	assert.Classic = formatClassic(assert.Classic)
	if assert.Classic != "" {
		headers["classic"] = assert.Classic
	}

	if assert.DisplayName != "" {
		headers["display-name"] = assert.DisplayName
	}

	// Some headers are required for Ubuntu Core, whilst optional or invalid for Classic
	if headers["classic"] == "true" {
		// Classic
		if assert.Architecture != "" {
			headers["architecture"] = assert.Architecture
		}
		if assert.Gadget != "" {
			headers["gadget"] = assert.Gadget
		}
	} else {
		// Core
		headers["kernel"] = assert.Kernel
		headers["architecture"] = assert.Architecture
		headers["gadget"] = assert.Gadget

		if len(assert.Base) != 0 {
			headers["base"] = assert.Base
		}
	}

	// Check if the optional fields as needed
	if assert.RequiredSnaps == "" {
		return headers, keypair, nil
	}

	snapList := strings.Split(assert.RequiredSnaps, ",")
	reqdSnaps := []interface{}{}
	for _, s := range snapList {
		reqdSnaps = append(reqdSnaps, strings.TrimSpace(s))
	}
	headers["required-snaps"] = reqdSnaps
	return headers, keypair, nil
}

func formatClassic(value string) string {
	classic := strings.ToLower(value)
	if classic != "true" && classic != "false" {
		classic = ""
	}
	return classic
}

// TODO: remove me after migration is done
// run over all models with assertion and store a
// signed model assertion in the new table
func (db *DB) runSignedModelAssertTableMigration() error {
	// fetch the full catalogue of models from the database
	// together with linked model assertion headers
	models, err := db.listModelsFilteredByUser("")
	if err != nil {
		return err
	}

	for _, model := range models {
		if model.ModelAssertion.ID == 0 {
			continue
		}
		assertionHeaders, keypair, err := ModelAssertionHeadersForModel(model)
		if err != nil {
			return fmt.Errorf("runSignedModelAssertTableMigration(): %v", err)
		}
		signedAssertion, err := Environ.KeypairDB.SignAssertion(asserts.ModelType,
			assertionHeaders,
			[]byte(""),
			model.BrandID,
			keypair.KeyID,
			keypair.SealedKey)
		if err != nil {
			return fmt.Errorf("runSignedModelAssertTableMigration(): %v", err)
		}

		err = db.UpsertSignedModelAssert(model.ID, model.ModelAssertion.Revision, signedAssertion)
		if err != nil {
			return fmt.Errorf("runSignedModelAssertTableMigration(): modelID=%d, revision=%d: err=%v", model.ID, model.ModelAssertion.Revision, err)
		}
	}
	return nil
}

// CreateSignedModelAssertTable creates the database table for a signed model assertion
func (db *DB) CreateSignedModelAssertTable() error {
	_, err := db.Exec(createSignedModelAssertTableSQL)
	if err == nil {
		// TODO: this on time migration and should be removed after release
		return db.runSignedModelAssertTableMigration()
	}
	return err
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
		return 0, fmt.Errorf("error creating the model assertion: %v", err)
	}

	return createdID, nil
}

// UpdateModelAssert updates the model assertion details of the model assertion
func (db *DB) UpdateModelAssert(m ModelAssertion) error {
	var err error

	_, err = db.Exec(updateModelAssertSQL, m.ID, m.ModelID, m.KeypairID, m.Series, m.Architecture, m.Revision, m.Gadget, m.Kernel, m.Store, time.Now().UTC(), m.RequiredSnaps, m.Base, m.Classic, m.DisplayName)

	if err != nil {
		return fmt.Errorf("error updating the model assertion for %d: %v", m.ID, err)
	}

	return nil
}

// UpsertSignedModelAssert creates or updates signed model assertion
func (db *DB) UpsertSignedModelAssert(modelID int, revision int, assertion asserts.Assertion) error {
	var err error
	headers, err := json.Marshal(assertion.Headers())
	if err != nil {
		return err
	}

	content, signature := assertion.Signature()
	_, err = db.Exec(upsertSignedModelAssertSQL, modelID, revision, assertion.Body(), headers, content, signature)
	return err
}

type dbHeaders map[string]interface{}

func (h *dbHeaders) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, h)
	default:
		return fmt.Errorf("unexpected type %T", src)
	}
}

// GetSignedModelAssert returns signed model assertion
func (db *DB) GetSignedModelAssert(modelID int) (asserts.Assertion, error) {
	var headers dbHeaders
	var body, content, signature []byte

	err := db.QueryRow(getSignedModelAssertSQL, modelID).Scan(&body, &headers, &content, &signature)
	if err != nil {
		return nil, fmt.Errorf("error fetching the model assertion for modelID=%d: %v", modelID, err)
	}

	return asserts.Assemble(headers, body, content, signature)
}

// UpsertModelAssert creates or updates the model assertion headers
func (db *DB) UpsertModelAssert(m ModelAssertion) error {
	var err error

	if err = validateModelAssertion(m); err != nil {
		return fmt.Errorf("error upserting the model assertion for model %d: %v", m.ModelID, err)
	}

	if m.ID > 0 {
		err = db.UpdateModelAssert(m)
	} else {
		_, err = db.CreateModelAssert(m)
	}

	return err
}

// deleteModelAssert deletes the signed model assertion details
func (db *DB) deletSignedModelAssert(modelID int) error {
	var err error
	_, err = db.Exec(deleteSignedModelAssertSQL, modelID)
	if err != nil {
		return fmt.Errorf("error deleting the signed model assertion: %v", err)
	}
	return nil
}

// deleteModelAssert deletes the model assertion details
func (db *DB) deleteModelAssert(modelID int) error {
	var err error

	_, err = db.Exec(deleteModelAssertSQL, modelID)
	if err != nil {
		return fmt.Errorf("error deleting the model assertion: %v", err)
	}
	return nil
}

// GetModelAssert fetches the model assertion
func (db *DB) GetModelAssert(modelID int) (ModelAssertion, error) {
	m := ModelAssertion{}
	err := db.QueryRow(getModelAssertSQL, modelID).Scan(&m.ID, &m.ModelID, &m.KeypairID, &m.Series, &m.Architecture, &m.Revision, &m.Gadget, &m.Kernel, &m.Store, &m.RequiredSnaps, &m.Base, &m.Classic, &m.DisplayName, &m.Created, &m.Modified)
	if err != nil {
		return m, fmt.Errorf("error fetching the model assertion for %d: %v", modelID, err)
	}

	return m, nil
}

func validateModelAssertion(m ModelAssertion) error {
	errTemplate := "invalid model assertion: %v "
	if m.ModelID <= 0 {
		return fmt.Errorf(errTemplate, "Model must be provided")
	}
	if m.KeypairID <= 0 {
		return fmt.Errorf(errTemplate, "Signing Key must be provided")
	}
	if m.Series < 16 {
		return fmt.Errorf(errTemplate, "Series must be at least 16")
	}

	if err := validateNotEmpty("Architecture", m.Architecture); err != nil {
		return fmt.Errorf(errTemplate, err)
	}
	if err := validateNotEmpty("Gadget", m.Gadget); err != nil {
		return fmt.Errorf(errTemplate, err)
	}
	if err := validateNotEmpty("Kernel", m.Kernel); err != nil {
		return fmt.Errorf(errTemplate, err)
	}
	if err := validateNotEmpty("Store", m.Store); err != nil {
		return fmt.Errorf(errTemplate, err)
	}

	return nil
}
