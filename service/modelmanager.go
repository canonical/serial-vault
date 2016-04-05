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
	"database/sql"
	"errors"
	"log"
	"strings"
)

const createModelTableSQL = `
	CREATE TABLE IF NOT EXISTS model (
		id          serial primary key not null,
		brand_id    varchar(200) not null,
		name        varchar(200) not null,
		signing_key text default '',
		revision    int
	)
`
const listModelsSQL = "select id, brand_id, name, signing_key, revision from model order by name"
const findModelSQL = "select id, brand_id, name, signing_key, revision from model where brand_id=$1 and name=$2 and revision=$3"
const getModelSQL = "select id, brand_id, name, signing_key, revision from model where id=$1"
const updateModelSQL = "update model set brand_id=$2, name=$3, revision=$4 where id=$1"
const createModelSQL = "insert into model (brand_id,name,revision) values ($1,$2,$3)"
const updateKeyModelSQL = "update model set signing_key=$2 where id=$1"

// Model holds the model details in the local database
type Model struct {
	ID         int
	BrandID    string
	Name       string
	SigningKey string
	Revision   int
}

// CreateModelTable creates the database table for a model.
func (db *DB) CreateModelTable() error {
	_, err := db.Exec(createModelTableSQL)
	return err
}

// ListModels fetches the full catalogue of models from the database.
func (db *DB) ListModels() ([]Model, error) {
	var models []Model

	rows, err := db.Query(listModelsSQL)
	if err != nil {
		log.Printf("Error retrieving database models: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		model := Model{}
		err := rows.Scan(&model.ID, &model.BrandID, &model.Name, &model.SigningKey, &model.Revision)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}

// FindModel retrieves the model from the database.
func (db *DB) FindModel(brandID, modelName string, revision int) (*Model, error) {
	model := Model{}

	err := db.QueryRow(findModelSQL, brandID, modelName, revision).Scan(&model.ID, &model.BrandID, &model.Name, &model.SigningKey, &model.Revision)
	switch {
	case err == sql.ErrNoRows:
		return nil, err
	case err != nil:
		log.Printf("Error retrieving database model: %v\n", err)
		return nil, err
	}

	return &model, nil
}

// GetModel retrieves the model from the database by ID.
func (db *DB) GetModel(modelID int) (*Model, error) {
	model := Model{}

	err := db.QueryRow(getModelSQL, modelID).Scan(&model.ID, &model.BrandID, &model.Name, &model.SigningKey, &model.Revision)
	if err != nil {
		log.Printf("Error retrieving database model by ID: %v\n", err)
		return nil, err
	}

	return &model, nil
}

// UpdateModel updates the model.
func (db *DB) UpdateModel(model Model) (string, error) {

	// Validate the data
	if strings.TrimSpace(model.BrandID) == "" || strings.TrimSpace(model.Name) == "" || model.Revision <= 0 {
		return "error-validate-model", errors.New("The Brand and Model must be supplied and Revision must be greater than zero")
	}

	_, err := db.Exec(updateModelSQL, model.ID, model.BrandID, model.Name, model.Revision)
	if err != nil {
		log.Printf("Error updating the database model: %v\n", err)
		return "", err
	}

	return "", nil
}

// CreateModel updates the model.
func (db *DB) CreateModel(model Model) (int, string, error) {

	// Validate the data
	if strings.TrimSpace(model.BrandID) == "" || strings.TrimSpace(model.Name) == "" || model.Revision <= 0 || strings.TrimSpace(model.SigningKey) == "" {
		return 0, "error-validate-new-model", errors.New("The Brand, Model and Signing-Key must be supplied and Revision must be greater than zero")
	}

	// Check that the model does not exist
	_, err := db.FindModel(model.BrandID, model.Name, model.Revision)
	if err == nil {
		return 0, "error-model-exists", errors.New("A device with the same Brand, Model and Revision already exists")
	}

	// TODO: Verify that the signing-key is valid
	// _, err = ClearSign("Text to Sign", model.SigningKey, "")
	// if err != nil {
	// 	return 0, "error-invalid-key", errors.New("The Signing-key is invalid")
	// }

	// Create the model in the database
	_, err = db.Exec(createModelSQL, model.BrandID, model.Name, model.Revision)
	if err != nil {
		log.Printf("Error creating the database model: %v\n", err)
		return 0, "", err
	}

	// Get the created model
	mdl, err := db.FindModel(model.BrandID, model.Name, model.Revision)
	if err != nil {
		return 0, "error-created-model", errors.New("Cannot find the created model")
	}

	// Store the signing-key in the keystore
	//TODO: use the asserts module to store the new signing-key
	// create the privateKey object
	//Environ.KeypairDB.Put(authorityID, privateKey)

	// keystore, err := GetKeyStore()
	// if err != nil {
	// 	return 0, "", err
	// }
	// keyLocation, err := keystore.Put([]byte(model.SigningKey), *mdl)
	// if err != nil {
	// 	return 0, "", err
	// }

	// Update the reference to the stored signing-key in the model
	// TODO: remove as we use keyID instead
	// err = db.updateModelKey(mdl.ID, keyLocation)
	// if err != nil {
	// 	return 0, "", err
	// }

	return mdl.ID, "", nil
}

// updateModelKey updates the reference to the signing-key location
func (db *DB) updateModelKey(modelID int, keyPath string) error {
	_, err := db.Exec(updateKeyModelSQL, modelID, keyPath)
	if err != nil {
		log.Printf("Error retrieving database model: %v\n", err)
	}
	return err
}
