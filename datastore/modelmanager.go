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
	"regexp"
	"strings"

	"github.com/CanonicalLtd/serial-vault/random"
)

const createModelTableSQL = `
	CREATE TABLE IF NOT EXISTS model (
		id               serial primary key not null,
		brand_id         varchar(200) not null,
		name             varchar(200) not null,
		keypair_id       int references keypair not null,
		user_keypair_id  int references keypair not null,
		api_key          varchar(200) not null
	)
`
const listModelsSQL = `
	select m.id, brand_id, name, keypair_id, api_key, k.authority_id, k.key_id, k.active, user_keypair_id, ku.authority_id, ku.key_id, ku.active, ku.assertion
	from model m
	inner join keypair k on k.id = m.keypair_id
	inner join keypair ku on ku.id = m.user_keypair_id
	order by name
`
const listModelsForUserSQL = `
	select m.id, brand_id, m.name, keypair_id, api_key, k.authority_id, k.key_id, k.active, user_keypair_id, ku.authority_id, ku.key_id, ku.active, ku.assertion
	from model m
	inner join keypair k on k.id = m.keypair_id
	inner join keypair ku on ku.id = m.user_keypair_id
	inner join account acc on acc.authority_id=m.brand_id
	inner join useraccountlink ua on ua.account_id=acc.id
	inner join userinfo u on ua.user_id=u.id
	where u.username=$1 and u.userrole >= $2
	order by name
`
const findModelSQL = `
	select m.id, brand_id, name, keypair_id, api_key, k.authority_id, k.key_id, k.active, k.sealed_key, user_keypair_id, ku.authority_id, ku.key_id, ku.active, ku.sealed_key, ku.assertion
	from model m
	inner join keypair k on k.id = m.keypair_id
	inner join keypair ku on ku.id = m.user_keypair_id
	where brand_id=$1 and name=$2 and api_key=$3`
const getModelSQL = `
	select m.id, brand_id, name, keypair_id, api_key, k.authority_id, k.key_id, k.active, k.sealed_key, user_keypair_id, ku.authority_id, ku.key_id, ku.active, ku.sealed_key, ku.assertion
	from model m
	inner join keypair k on k.id = m.keypair_id
	inner join keypair ku on ku.id = m.user_keypair_id
	where m.id=$1`
const getModelForUserSQL = `
	select m.id, m.brand_id, m.name, m.keypair_id, api_key, k.authority_id, k.key_id, k.active, k.sealed_key, user_keypair_id, ku.authority_id, ku.key_id, ku.active, ku.sealed_key, ku.assertion
	from model m
	inner join keypair k on k.id = m.keypair_id
	inner join keypair ku on ku.id = m.user_keypair_id
	inner join account acc on acc.authority_id=m.brand_id
	inner join useraccountlink ua on ua.account_id=acc.id
	inner join userinfo u on ua.user_id=u.id
	where m.id=$1 and u.username=$2 and u.userrole >= $3`
const updateModelSQL = "update model set brand_id=$2, name=$3, keypair_id=$4, user_keypair_id=$5, api_key=$6 where id=$1"
const updateModelForUserSQL = `
	update model m set brand_id=$2, name=$3, keypair_id=$4, user_keypair_id=$5, api_key=$6
	from account acc
	inner join useraccountlink ua on ua.account_id=acc.id
	inner join userinfo u on ua.user_id=u.id
	where acc.authority_id=m.brand_id and m.id=$1 and u.username=$7 and u.userrole >= $8`
const createModelSQL = "insert into model (brand_id,name,keypair_id,user_keypair_id,api_key) values ($1,$2,$3,$4,$5) RETURNING id"
const deleteModelSQL = "delete from model where id=$1"
const deleteModelForUserSQL = `
	delete from model m
	using account acc
	inner join useraccountlink ua on ua.account_id=acc.id
	inner join userinfo u on ua.user_id=u.id
	where m.id=$1 and acc.authority_id=m.brand_id and u.username=$2 and u.userrole >= $3`

const checkBrandsMatchSQL = `
	select count(*) from model m
	inner join keypair k on k.authority_id = m.brand_id
	inner join keypair ku on ku.authority_id = m.brand_id
	where m.brand_id=$1 and k.id=$2 and ku.id=$3
`

const checkAPIKeyExistsSQL = `
	select exists(
		select * from model where api_key=$1
	)
`

const checkModelExistsSQL = `
	select exists(
		select * from model where brand_id=$1 and name=$2
	)
`

// Add the user keypair to the models table (nullable)
const alterModelUserKeypairNullable = "alter table model add column user_keypair_id int references keypair"

// Populate the user keypair and make it not-nullable
const populateModelUserKeypair = "update model set user_keypair_id=keypair_id where user_keypair_id is null"
const alterModelUserKeypairNotNullable = "alter table model alter column user_keypair_id set not null"

// Add the API key field to the models table (nullable)
const alterModelAPIKey = "alter table model add column api_key varchar(200) default ''"

// Make the API key not-nullable
const alterModelAPIKeyNotNullable = `alter table model
	alter column api_key set not null,
	alter column api_key drop default
`

// Indexes
const createModelAPIKeyIndexSQL = "CREATE INDEX IF NOT EXISTS api_key_idx ON model (api_key)"

// Model holds the model details in the local database
type Model struct {
	ID              int
	BrandID         string
	Name            string
	KeypairID       int
	APIKey          string
	AuthorityID     string // from the signing keypair
	KeyID           string // from the signing keypair
	KeyActive       bool   // from the signing keypair
	SealedKey       string // from the signing keypair
	KeypairIDUser   int    // from the system-user keypair
	AuthorityIDUser string // from the system-user keypair
	KeyIDUser       string // from the system-user keypair
	KeyActiveUser   bool   // from the system-user keypair
	SealedKeyUser   string // from the system-user keypair
	AssertionUser   string // from the system-user keypair
}

// CreateModelTable creates the database table for a model.
func (db *DB) CreateModelTable() error {
	_, err := db.Exec(createModelTableSQL)
	return err
}

// AlterModelTable updates an existing database model table with additional fields
func (db *DB) AlterModelTable() error {
	err := db.addUserKeypairFields()
	if err != nil {
		return err
	}

	err = db.addAPIKeyField()
	if err != nil {
		return err
	}

	// Create the index on the API key
	_, err = db.Exec(createModelAPIKeyIndexSQL)
	if err != nil {
		return err
	}

	return nil
}

// addUserKeypairFields adds the user keypair link to an existing model table.
func (db *DB) addUserKeypairFields() error {
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

// addAPIKeyField adds and defaults the API key field to the model table
func (db *DB) addAPIKeyField() error {

	// Add the API key field to the model table
	_, err := db.Exec(alterModelAPIKey)
	if err != nil {
		// Field already exists so skip
		return nil
	}

	// Default the API key for any records where it is empty
	models, err := db.ListModels("")
	if err != nil {
		return err
	}
	for _, model := range models {
		if len(model.APIKey) > 0 {
			continue
		}

		// Generate an random API key and update the record
		apiKey, err := db.generateAPIKey()
		if err != nil {
			log.Printf("Could not generate random string for the API key")
			return errors.New("Error generating random string for the API key")
		}

		// Update the API key on the model
		model.APIKey = apiKey
		db.UpdateModel(model, "")
	}

	// Add the constraints to the API key field
	_, err = db.Exec(alterModelAPIKeyNotNullable)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) generateAPIKey() (string, error) {
	reg, _ := regexp.Compile("[^A-Za-z0-9]+")

	// Generate an random API key and update the record
	apiKey, err := random.GenerateRandomString(40)
	if err != nil {
		log.Printf("Could not generate random string for the API key")
		return "", errors.New("Error generating random string for the API key")
	}

	return reg.ReplaceAllString(apiKey, ""), nil
}

// ListModels fetches the full catalogue of models from the database.
// If a username is supplied, then only show the models for the user
// [Permissions: Admin]
func (db *DB) ListModels(username string) ([]Model, error) {
	models := []Model{}

	var (
		rows *sql.Rows
		err  error
	)

	if len(username) == 0 {
		rows, err = db.Query(listModelsSQL)
	} else {
		rows, err = db.Query(listModelsForUserSQL, username, Admin)
	}
	if err != nil {
		log.Printf("Error retrieving database models: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		model := Model{}
		err := rows.Scan(&model.ID, &model.BrandID, &model.Name, &model.KeypairID, &model.APIKey, &model.AuthorityID, &model.KeyID, &model.KeyActive,
			&model.KeypairIDUser, &model.AuthorityIDUser, &model.KeyIDUser, &model.KeyActiveUser, &model.AssertionUser)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}

// FindModel retrieves the model from the database.
func (db *DB) FindModel(brandID, modelName, apiKey string) (Model, error) {
	model := Model{}

	err := db.QueryRow(findModelSQL, brandID, modelName, apiKey).Scan(
		&model.ID, &model.BrandID, &model.Name, &model.KeypairID, &model.APIKey, &model.AuthorityID, &model.KeyID, &model.KeyActive, &model.SealedKey,
		&model.KeypairIDUser, &model.AuthorityIDUser, &model.KeyIDUser, &model.KeyActiveUser, &model.SealedKeyUser, &model.AssertionUser)
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
func (db *DB) GetModel(modelID int, username string) (Model, error) {
	model := Model{}

	var row *sql.Row

	if len(username) == 0 {
		row = db.QueryRow(getModelSQL, modelID)
	} else {
		row = db.QueryRow(getModelForUserSQL, modelID, username, Admin)
	}

	err := row.Scan(&model.ID, &model.BrandID, &model.Name, &model.KeypairID, &model.APIKey, &model.AuthorityID, &model.KeyID, &model.KeyActive, &model.SealedKey,
		&model.KeypairIDUser, &model.AuthorityIDUser, &model.KeyIDUser, &model.KeyActiveUser, &model.SealedKeyUser, &model.AssertionUser)
	if err != nil {
		log.Printf("Error retrieving database model by ID: %v\n", err)
		return model, err
	}

	return model, nil
}

// UpdateModel updates the model.
func (db *DB) UpdateModel(model Model, username string) (string, error) {

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

	if !db.checkBrandsMatch(username, model.BrandID, model.KeypairID, model.KeypairIDUser) {
		return "error-auth", errors.New("The model and the keys must have the same brand")
	}

	var err error

	if len(username) == 0 {
		_, err = db.Exec(updateModelSQL, model.ID, model.BrandID, model.Name, model.KeypairID, model.KeypairIDUser, model.APIKey)
	} else {
		_, err = db.Exec(updateModelForUserSQL, model.ID, model.BrandID, model.Name, model.KeypairID, model.KeypairIDUser, model.APIKey, username, Admin)
	}
	if err != nil {
		log.Printf("Error updating the database model: %v\n", err)
		return "", err
	}

	return "", nil
}

// CreateModel updates the model.
func (db *DB) CreateModel(model Model, username string) (Model, string, error) {

	if db.RoleForUser(username) < Admin {
		return model, "error-auth", errors.New("The user does not have permissions to create a model")
	}

	// Validate the data
	if strings.TrimSpace(model.BrandID) == "" || strings.TrimSpace(model.Name) == "" || model.KeypairID <= 0 || model.KeypairIDUser <= 0 {
		return model, "error-validate-new-model", errors.New("The Brand, Model and Signing-Keys must be supplied")
	}

	if !db.CheckUserInAccount(username, model.BrandID) {
		return model, "error-auth", errors.New("The user does not have permissions to create a model for this account")
	}

	if !db.checkBrandsMatch(username, model.BrandID, model.KeypairID, model.KeypairIDUser) {
		return model, "error-auth", errors.New("The model and the keys must have the same brand")
	}

	// Check that the model does not exist
	if found := db.checkModelExists(model.BrandID, model.Name); found {
		return model, "error-model-exists", errors.New("A device with the same Brand and Model already exists")
	}

	// Create the model in the database
	var createdModelID int
	err := db.QueryRow(createModelSQL, model.BrandID, model.Name, model.KeypairID, model.KeypairIDUser, model.APIKey).Scan(&createdModelID)
	if err != nil {
		log.Printf("Error creating the database model: %v\n", err)
		return model, "", err
	}

	// Return the created model
	mdl, err := db.GetModel(createdModelID, username)
	if err != nil {
		log.Printf("Error creating the database model: %v\n", err)
		return model, "", err
	}
	return mdl, "", nil
}

// DeleteModel deletes a model record.
func (db *DB) DeleteModel(model Model, username string) (string, error) {
	var err error

	if len(username) == 0 {
		_, err = db.Exec(deleteModelSQL, model.ID)
	} else {
		_, err = db.Exec(deleteModelForUserSQL, model.ID, username, Admin)
	}
	if err != nil {
		log.Printf("Error deleting the database model: %v\n", err)
		return "", err
	}

	return "", nil
}

func (db *DB) checkBrandsMatch(username, brandID string, keypairID, keypairIDUser int) bool {
	if username == "" {
		return true
	}

	var count int

	row := db.QueryRow(checkBrandsMatchSQL, brandID, keypairID, keypairIDUser)
	err := row.Scan(&count)
	if err != nil {
		log.Printf("Error checking that the account matches for a model: %v\n", err)
		return false
	}

	return count > 0
}

// CheckAPIKey validates that there is a model for the supplied API key
func (db *DB) CheckAPIKey(apiKey string) bool {
	row := db.QueryRow(checkAPIKeyExistsSQL, apiKey)
	return db.checkBoolQuery(row)
}

// CheckAPIKey validates that there is a model for the supplied API key
func (db *DB) checkModelExists(brandID, name string) bool {
	row := db.QueryRow(checkModelExistsSQL, brandID, name)
	return db.checkBoolQuery(row)
}

func (db *DB) checkBoolQuery(row *sql.Row) bool {
	var found bool

	err := row.Scan(&found)
	if err != nil {
		log.Printf("Error with the boolean query: %v\n", err)
		return false
	}

	return found
}
