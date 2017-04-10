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
		id               serial primary key not null,
		brand_id         varchar(200) not null,
		name             varchar(200) not null,
		keypair_id       int references keypair not null,
		user_keypair_id  int references keypair not null
	)
`
const listModelsSQL = `
	select m.id, brand_id, name, keypair_id, k.authority_id, k.key_id, k.active, user_keypair_id, ku.authority_id, ku.key_id, ku.active
	from model m
	inner join keypair k on k.id = m.keypair_id
	inner join keypair ku on ku.id = m.user_keypair_id
	order by name
`
const findModelSQL = `
	select m.id, brand_id, name, keypair_id, k.authority_id, k.key_id, k.active, k.sealed_key, user_keypair_id, ku.authority_id, ku.key_id, ku.active, ku.sealed_key
	from model m
	inner join keypair k on k.id = m.keypair_id
	inner join keypair ku on ku.id = m.user_keypair_id
	where brand_id=$1 and name=$2`
const getModelSQL = `
	select m.id, brand_id, name, keypair_id, k.authority_id, k.key_id, k.active, k.sealed_key, user_keypair_id, ku.authority_id, ku.key_id, ku.active, ku.sealed_key
	from model m
	inner join keypair k on k.id = m.keypair_id
	inner join keypair ku on ku.id = m.user_keypair_id
	where m.id=$1`
const updateModelSQL = "update model set brand_id=$2, name=$3, keypair_id=$4, user_keypair_id=$5 where id=$1"
const createModelSQL = "insert into model (brand_id,name,keypair_id,user_keypair_id) values ($1,$2,$3,$4) RETURNING id"
const deleteModelSQL = "delete from model where id=$1"

// Add the user keypair to the models table (nullable)
const alterModelUserKeypairNullable = "alter table model add column user_keypair_id int references keypair"

// Populate the user keypair and make it not-nullable
const populateModelUserKeypair = "update model set user_keypair_id=keypair_id where user_keypair_id is null"
const alterModelUserKeypairNotNullable = "alter table model alter column user_keypair_id set not null"

// Model holds the model details in the local database
type Model struct {
	ID              int
	BrandID         string
	Name            string
	KeypairID       int
	AuthorityID     string // from the signing keypair
	KeyID           string // from the signing keypair
	KeyActive       bool   // from the signing keypair
	SealedKey       string // from the signing keypair
	KeypairIDUser   int    // from the system-user keypair
	AuthorityIDUser string // from the system-user keypair
	KeyIDUser       string // from the system-user keypair
	KeyActiveUser   bool   // from the system-user keypair
	SealedKeyUser   string // from the system-user keypair
}

// CreateModelTable creates the database table for a model.
func (db *DB) CreateModelTable() error {
	_, err := db.Exec(createModelTableSQL)
	return err
}

// AlterModelTable adds the user keypair link to an existing model table.
func (db *DB) AlterModelTable() error {
	_, err := db.Exec(alterModelUserKeypairNullable)
	if err != nil {
		// Field already exists so skip
		return nil
	}

	// Default the user keypair
	_, err = db.Exec(populateModelUserKeypair)
	if err != nil {
		log.Println("Error defaulting the user keypair")
		return err
	}

	_, err = db.Exec(alterModelUserKeypairNotNullable)
	if err != nil {
		log.Println("Error in making the user keypair not null")
		return err
	}
	return nil
}

// ListModels fetches the full catalogue of models from the database.
func (db *DB) ListModels() ([]Model, error) {
	models := []Model{}

	rows, err := db.Query(listModelsSQL)
	if err != nil {
		log.Printf("Error retrieving database models: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		model := Model{}
		err := rows.Scan(&model.ID, &model.BrandID, &model.Name, &model.KeypairID, &model.AuthorityID, &model.KeyID, &model.KeyActive,
			&model.KeypairIDUser, &model.AuthorityIDUser, &model.KeyIDUser, &model.KeyActiveUser)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}

// FindModel retrieves the model from the database.
func (db *DB) FindModel(brandID, modelName string) (Model, error) {
	model := Model{}

	err := db.QueryRow(findModelSQL, brandID, modelName).Scan(
		&model.ID, &model.BrandID, &model.Name, &model.KeypairID, &model.AuthorityID, &model.KeyID, &model.KeyActive, &model.SealedKey,
		&model.KeypairIDUser, &model.AuthorityIDUser, &model.KeyIDUser, &model.KeyActiveUser, &model.SealedKeyUser)
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

	err := db.QueryRow(getModelSQL, modelID).Scan(&model.ID, &model.BrandID, &model.Name, &model.KeypairID, &model.AuthorityID, &model.KeyID, &model.KeyActive, &model.SealedKey,
		&model.KeypairIDUser, &model.AuthorityIDUser, &model.KeyIDUser, &model.KeyActiveUser, &model.SealedKeyUser)
	if err != nil {
		log.Printf("Error retrieving database model by ID: %v\n", err)
		return model, err
	}

	return model, nil
}

// UpdateModel updates the model.
func (db *DB) UpdateModel(model Model) (string, error) {

	// Validate the data
	if strings.TrimSpace(model.BrandID) == "" || strings.TrimSpace(model.Name) == "" {
		return "error-validate-model", errors.New("The Brand and Model must be supplied")
	}
	if model.KeypairID <= 0 {
		return "error-validate-signingkey", errors.New("The Signing Key must be selected")
	}
	if model.KeypairIDUser <= 0 {
		return "error-validate-userkey", errors.New("The System-User Key must be selected")
	}

	_, err := db.Exec(updateModelSQL, model.ID, model.BrandID, model.Name, model.KeypairID, model.KeypairIDUser)
	if err != nil {
		log.Printf("Error updating the database model: %v\n", err)
		return "", err
	}

	return "", nil
}

// CreateModel updates the model.
func (db *DB) CreateModel(model Model) (Model, string, error) {

	// Validate the data
	if strings.TrimSpace(model.BrandID) == "" || strings.TrimSpace(model.Name) == "" || model.KeypairID <= 0 || model.KeypairIDUser <= 0 {
		return model, "error-validate-new-model", errors.New("The Brand, Model and Signing-Keys must be supplied")
	}

	// Check that the model does not exist
	_, err := db.FindModel(model.BrandID, model.Name)
	if err == nil {
		return model, "error-model-exists", errors.New("A device with the same Brand and Model already exists")
	}

	// Create the model in the database
	var createdModelID int
	err = db.QueryRow(createModelSQL, model.BrandID, model.Name, model.KeypairID, model.KeypairIDUser).Scan(&createdModelID)
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

// DeleteModel deletes a model record.
func (db *DB) DeleteModel(model Model) (string, error) {

	_, err := db.Exec(deleteModelSQL, model.ID)
	if err != nil {
		log.Printf("Error deleting the database model: %v\n", err)
		return "", err
	}

	return "", nil
}
