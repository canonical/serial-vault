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
	"regexp"
	"strings"

	"github.com/CanonicalLtd/serial-vault/random"
)

const validModelNamePattern = defaultNicknamePattern

var validModelNameRegexp = regexp.MustCompile(validModelNamePattern)

// ListAllowedModels returns the models allowed to be seen to the authorization
func (db *DB) ListAllowedModels(authorization User) ([]Model, error) {
	switch authorization.Role {
	case Invalid: // Authentication is disabled
		fallthrough
	case Superuser:
		return db.listAllModels()
	case Standard:
		fallthrough
	case Admin:
		return db.listModelsFilteredByUser(authorization.Username)
	default:
		return []Model{}, nil
	}
}

// GetAllowedModel returns the model allowed to be seen by the authorization
func (db *DB) GetAllowedModel(modelID int, authorization User) (Model, error) {
	switch authorization.Role {
	case Invalid: // Authentication is disabled
		fallthrough
	case Superuser:
		return db.getModel(modelID)
	case Admin:
		return db.getModelFilteredByUser(modelID, authorization.Username)
	default:
		return Model{}, nil
	}
}

// UpdateAllowedModel updates the model if authorization is allowed to do it
func (db *DB) UpdateAllowedModel(model Model, authorization User) (string, error) {

	errorSubcode, err := validateModel(model, "error-validate-model")
	if err != nil {
		return errorSubcode, err
	}

	if !db.checkBrandsMatch(model.BrandID, model.KeypairID, model.KeypairIDUser) {
		return "error-auth", errors.New("The model and the keys must have the same brand")
	}

	// Check the API key and default it if it is invalid
	err = buildValidOrDefaultAPIKey(&model)
	if err != nil {
		return "error-model-apikey", errors.New("Error in generating a valid API key")
	}

	switch authorization.Role {
	case Invalid: // Authentication is disabled
		fallthrough
	case Superuser:
		return db.updateModel(model)
	case Admin:
		return db.updateModelFilteredByUser(model, authorization.Username)
	default:
		return "", nil
	}
}

// DeleteAllowedModel deletes model if allowed to authorization
func (db *DB) DeleteAllowedModel(model Model, authorization User) (string, error) {
	switch authorization.Role {
	case Invalid: // Authentication is disabled
		fallthrough
	case Superuser:
		return db.deleteModel(model)
	case Admin:
		return db.deleteModelFilteredByUser(model, authorization.Username)
	default:
		return "", nil
	}
}

// CreateAllowedModel creates a new model in case authorization is allowed to do it
func (db *DB) CreateAllowedModel(model Model, authorization User) (Model, string, error) {

	errorSubcode, err := validateModel(model, "error-validate-new-model")
	if err != nil {
		return model, errorSubcode, err
	}

	if !db.CheckUserInAccount(authorization.Username, model.BrandID) {
		return model, "error-auth", errors.New("The user does not have permissions to create a model for this account")
	}

	if !db.checkBrandsMatch(model.BrandID, model.KeypairID, model.KeypairIDUser) {
		return model, "error-auth", errors.New("The model and the keys must have the same brand")
	}

	// Check the API key and default it if it is invalid
	err = buildValidOrDefaultAPIKey(&model)
	if err != nil {
		return model, "error-model-apikey", errors.New("Error in generating a valid API key")
	}

	// Check that the model does not exist
	if found := db.checkModelExists(model.BrandID, model.Name); found {
		return model, "error-model-exists", errors.New("A device with the same Brand and Model already exists")
	}

	switch authorization.Role {
	case Invalid: // Authentication is disabled
		fallthrough
	case Superuser:
		return db.createModel(model)
	case Admin:
		return db.createModelFilteredByUser(model, authorization.Username)
	default:
		return Model{}, "", nil
	}
}

func validateModel(model Model, validateModelLabel string) (string, error) {

	err := validateBrandID(model.BrandID)
	if err != nil {
		return validateModelLabel, err
	}

	err = validateModelName(model.Name)
	if err != nil {
		return validateModelLabel, err
	}

	err = validateKeypairID(model.KeypairID)
	if err != nil {
		return "error-validate-signingkey", err
	}

	err = validateKeypairIDUser(model.KeypairIDUser)
	if err != nil {
		return "error-validate-userkey", err
	}

	return "", nil
}

func validateBrandID(brandID string) error {
	return validateNotEmpty("Brand ID", brandID)
}

// validateModelName validates name for the model; the rule is: lowercase with no spaces
func validateModelName(name string) error {
	return validateSyntax("Model name", name, validModelNameRegexp)
}

func validateKeypairID(keypairID int) error {
	if keypairID <= 0 {
		return errors.New("The Signing Key must be selected")
	}
	return nil
}

func validateKeypairIDUser(keypairIDUser int) error {
	if keypairIDUser <= 0 {
		return errors.New("The System-User Key must be selected")
	}
	return nil
}

// buildValidOrDefaultAPIKey checks the API key and creates a default API key if the field is empty
func buildValidOrDefaultAPIKey(model *Model) error {
	// Remove all whitespace from the API key
	model.APIKey = strings.Replace(model.APIKey, " ", "", -1)

	// Check we have a minimum API key size
	if len(model.APIKey) > minAPIKeyLength {
		return nil
	}

	apiKey, err := generateAPIKey()
	if err != nil {
		return err
	}
	model.APIKey = apiKey
	return nil
}

func generateAPIKey() (string, error) {
	reg, _ := regexp.Compile("[^A-Za-z0-9]+")

	// Generate an random API key and update the record
	apiKey, err := random.GenerateRandomString(40)
	if err != nil {
		log.Printf("Could not generate random string for the API key")
		return "", errors.New("Error generating random string for the API key")
	}

	return reg.ReplaceAllString(apiKey, ""), nil
}
