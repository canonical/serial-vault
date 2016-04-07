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
		keypair_id  int references keypair not null,
		revision    int
	)
`
const listModelsSQL = `
	select m.id, brand_id, name, keypair_id, revision, authority_id, key_id
	from model m
	inner join keypair k on k.id = m.keypair_id and k.active
	order by name
`
const findModelSQL = `
	select m.id, brand_id, name, keypair_id, revision, authority_id, key_id
	from model m
	inner join keypair k on k.id = m.keypair_id and k.active
	where brand_id=$1 and name=$2 and revision=$3`
const getModelSQL = `
	select m.id, brand_id, name, keypair_id, revision, authority_id, key_id
	from model m
	inner join keypair k on k.id = m.keypair_id and k.active
	where m.id=$1`
const updateModelSQL = "update model set brand_id=$2, name=$3, revision=$4, keypair_id=$5 where id=$1"
const createModelSQL = "insert into model (brand_id,name,revision,keypair_id) values ($1,$2,$3,$4) RETURNING id"

// Model holds the model details in the local database
type Model struct {
	ID          int
	BrandID     string
	Name        string
	KeypairID   int
	Revision    int
	AuthorityID string // from the keypair
	KeyID       string // from the keypair
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
		err := rows.Scan(&model.ID, &model.BrandID, &model.Name, &model.KeypairID, &model.Revision, &model.AuthorityID, &model.KeyID)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}

// FindModel retrieves the model from the database.
func (db *DB) FindModel(brandID, modelName string, revision int) (Model, error) {
	model := Model{}

	err := db.QueryRow(findModelSQL, brandID, modelName, revision).Scan(
		&model.ID, &model.BrandID, &model.Name, &model.KeypairID, &model.Revision, &model.AuthorityID, &model.KeyID)
	switch {
	case err == sql.ErrNoRows:
		return model, err
	case err != nil:
		log.Printf("Error retrieving database model: %v\n", err)
		return model, err
	}

	return model, nil
}

// GetModel retrieves the model from the database by ID.
func (db *DB) GetModel(modelID int) (Model, error) {
	model := Model{}

	err := db.QueryRow(getModelSQL, modelID).Scan(&model.ID, &model.BrandID, &model.Name, &model.KeypairID, &model.Revision, &model.AuthorityID, &model.KeyID)
	if err != nil {
		log.Printf("Error retrieving database model by ID: %v\n", err)
		return model, err
	}

	return model, nil
}

// UpdateModel updates the model.
func (db *DB) UpdateModel(model Model) (string, error) {

	// Validate the data
	if strings.TrimSpace(model.BrandID) == "" || strings.TrimSpace(model.Name) == "" || model.Revision <= 0 {
		return "error-validate-model", errors.New("The Brand and Model must be supplied and Revision must be greater than zero")
	}
	if model.KeypairID <= 0 {
		return "error-validate-signingkey", errors.New("The Signing Key must be selected")
	}

	_, err := db.Exec(updateModelSQL, model.ID, model.BrandID, model.Name, model.Revision, model.KeypairID)
	if err != nil {
		log.Printf("Error updating the database model: %v\n", err)
		return "", err
	}

	return "", nil
}

// CreateModel updates the model.
func (db *DB) CreateModel(model Model) (Model, string, error) {

	// Validate the data
	if strings.TrimSpace(model.BrandID) == "" || strings.TrimSpace(model.Name) == "" || model.Revision <= 0 || model.KeypairID <= 0 {
		return model, "error-validate-new-model", errors.New("The Brand, Model and Signing-Key must be supplied and Revision must be greater than zero")
	}

	// Check that the model does not exist
	_, err := db.FindModel(model.BrandID, model.Name, model.Revision)
	if err == nil {
		return model, "error-model-exists", errors.New("A device with the same Brand, Model and Revision already exists")
	}

	// Create the model in the database
	var createdModelID int
	err = db.QueryRow(createModelSQL, model.BrandID, model.Name, model.Revision, model.KeypairID).Scan(&createdModelID)
	if err != nil {
		log.Printf("Error creating the database model: %v\n", err)
		return model, "", err
	}

	// Return the created model
	mdl, err := db.GetModel(createdModelID)
	if err != nil {
		log.Printf("Error creating the database model: %v\n", err)
		return model, "", err
	}
	return mdl, "", nil
}
