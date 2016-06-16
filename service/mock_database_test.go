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

import "errors"

// Successful mocks for the database
type mockDB struct {
	encryptedAuthKeyHash string
}

// CreateModelTable mock for the create model table method
func (mdb *mockDB) CreateModelTable() error {
	return nil
}

// CreateKeypairTable mock for the create keypair table method
func (mdb *mockDB) CreateKeypairTable() error {
	return nil
}

// CreateSettingsTable mock for the create settings table method
func (mdb *mockDB) CreateSettingsTable() error {
	return nil
}

// ModelsList Mock the database response for a list of models
func (mdb *mockDB) ListModels() ([]Model, error) {

	var models []Model
	models = append(models, Model{ID: 1, BrandID: "Vendor", Name: "Alder", KeypairID: 1, Revision: 1, AuthorityID: "System", KeyID: "61abf588e52be7a3"})
	models = append(models, Model{ID: 2, BrandID: "Vendor", Name: "Ash", KeypairID: 1, Revision: 7, AuthorityID: "System", KeyID: "61abf588e52be7a3"})
	models = append(models, Model{ID: 3, BrandID: "Vendor", Name: "Basswood", KeypairID: 1, Revision: 23, AuthorityID: "System", KeyID: "61abf588e52be7a3"})
	models = append(models, Model{ID: 4, BrandID: "Vendor", Name: "Korina", KeypairID: 1, Revision: 42, AuthorityID: "System", KeyID: "61abf588e52be7a3"})
	models = append(models, Model{ID: 5, BrandID: "Vendor", Name: "Mahogany", KeypairID: 1, Revision: 10, AuthorityID: "System", KeyID: "61abf588e52be7a3"})
	models = append(models, Model{ID: 6, BrandID: "Vendor", Name: "Maple", KeypairID: 1, Revision: 12, AuthorityID: "System", KeyID: "61abf588e52be7a3"})
	return models, nil
}

// FindModel mocks the database response for finding a model
func (mdb *mockDB) FindModel(brandID, modelName string, revision int) (Model, error) {
	model := Model{ID: 1, BrandID: "System", Name: "Alder", KeypairID: 1, Revision: 1, AuthorityID: "System", KeyID: "61abf588e52be7a3", KeyActive: true}
	if modelName == "Inactive" {
		model = Model{ID: 1, BrandID: "System", Name: "Inactive", KeypairID: 1, Revision: 1, AuthorityID: "System", KeyID: "61abf588e52be7a3", KeyActive: false}
	}
	if model.BrandID != brandID || model.Name != modelName || model.Revision != revision {
		return model, errors.New("Cannot find a model for that brand, model and revision")
	}
	return model, nil
}

// GetModel mocks the model from the database by ID.
func (mdb *mockDB) GetModel(modelID int) (Model, error) {

	var model Model
	found := false
	models, _ := mdb.ListModels()

	for _, mdl := range models {
		if mdl.ID == modelID {
			model = mdl
			found = true
			break
		}
	}

	if !found {
		return model, errors.New("Cannot find the model.")
	}

	return model, nil
}

// UpdateModel mocks the model update.
func (mdb *mockDB) UpdateModel(model Model) (string, error) {
	models, _ := mdb.ListModels()
	found := false

	for _, mdl := range models {
		if mdl.ID == model.ID {
			found = true
			break
		}
	}

	if !found {
		return "", errors.New("Cannot find the model.")
	}
	return "", nil
}

// DeleteModel mocks the model deletion.
func (mdb *mockDB) DeleteModel(model Model) (string, error) {
	models, _ := mdb.ListModels()
	found := false

	for _, mdl := range models {
		if mdl.ID == model.ID {
			found = true
			break
		}
	}

	if !found {
		return "", errors.New("Cannot find the model.")
	}
	return "", nil
}

// CreateModel mocks creating a new model.
func (mdb *mockDB) CreateModel(model Model) (Model, string, error) {
	model = Model{ID: 7, BrandID: "System", Name: "聖誕快樂", KeypairID: 1, Revision: 2, AuthorityID: "system", KeyID: "61abf588e52be7a3"}

	return model, "", nil
}

func (mdb *mockDB) GetKeypair(keypairID int) (Keypair, error) {
	keypair := Keypair{ID: 1, AuthorityID: "system", KeyID: "61abf588e52be7a3", Active: true}
	return keypair, nil
}

func (mdb *mockDB) ListKeypairs() ([]Keypair, error) {
	var keypairs []Keypair
	keypairs = append(keypairs, Keypair{ID: 1, AuthorityID: "system", KeyID: "61abf588e52be7a3", Active: true})
	keypairs = append(keypairs, Keypair{ID: 2, AuthorityID: "system", KeyID: "invalidone", Active: true})
	return keypairs, nil
}

func (mdb *mockDB) PutKeypair(keypair Keypair) (string, error) {
	return "", nil
}

func (mdb *mockDB) UpdateKeypairActive(keypairID int, active bool) error {
	return nil
}

func (mdb *mockDB) GetSetting(code string) (Setting, error) {
	switch code {
	case "System/12345678abcdef":
		// Returning the encrypted, base64 encoded HMAC-ed auth-key: fake-hmac-ed-data
		return Setting{Code: "System/12345678abcdef", Data: "pmXt1iwvM5P947KATp24rMQFHEnAf2tUXGl1XXyfhDhf"}, nil

	case "System/abcdef12345678":
		return Setting{Code: "System/abcdef12345678", Data: mdb.encryptedAuthKeyHash}, nil

	case "do-not-find":
		return Setting{}, errors.New("Cannot find 'do-not-find'")

	default:
		return Setting{Code: code, Data: code}, nil
	}
}

func (mdb *mockDB) PutSetting(setting Setting) error {
	if setting.Code == "System/abcdef12345678" {
		mdb.encryptedAuthKeyHash = setting.Data
	}
	return nil
}

// Unsuccessful mocks for the database
type errorMockDB struct{}

// CreateModelTable mock for the create model table method
func (mdb *errorMockDB) CreateModelTable() error {
	return errors.New("Error creating the model table.")
}

// CreateKeypairTable mock for the create keypair table method
func (mdb *errorMockDB) CreateKeypairTable() error {
	return nil
}

// CreateSettingsTable mock for the create settings table method
func (mdb *errorMockDB) CreateSettingsTable() error {
	return nil
}

// ModelsList Mock the database response for a list of models
func (mdb *errorMockDB) ListModels() ([]Model, error) {
	return nil, errors.New("Error getting the models.")
}

// FindModel mocks the database response for finding a model, returning an invalid signing-key
func (mdb *errorMockDB) FindModel(brandID, modelName string, revision int) (Model, error) {
	return Model{}, errors.New("Error finding the model.")
}

// GetModel mocks the model from the database by ID, returning an error.
func (mdb *errorMockDB) GetModel(modelID int) (Model, error) {
	return Model{}, errors.New("Error retrieving the model.")
}

// UpdateModel mocks the model update, returning an error.
func (mdb *errorMockDB) UpdateModel(model Model) (string, error) {
	return "", errors.New("Error updating the database model.")
}

// DeleteModel mocks the model deletion, returning an error.
func (mdb *errorMockDB) DeleteModel(model Model) (string, error) {
	return "", errors.New("Error deleting the database model.")
}

// CreateModel mocks creating a new model, returning an error.
func (mdb *errorMockDB) CreateModel(model Model) (Model, string, error) {
	return Model{}, "", errors.New("Error creating the database model.")
}

func (mdb *errorMockDB) GetKeypair(keypairID int) (Keypair, error) {
	keypair := Keypair{AuthorityID: "system", KeyID: "61abf588e52be7a3", Active: true}
	return keypair, errors.New("Error fetching from the database.")
}

func (mdb *errorMockDB) ListKeypairs() ([]Keypair, error) {
	var keypairs []Keypair
	return keypairs, errors.New("Error fetching from the database.")
}

func (mdb *errorMockDB) PutKeypair(keypair Keypair) (string, error) {
	return "", errors.New("Error updating the database.")
}

func (mdb *errorMockDB) UpdateKeypairActive(keypairID int, active bool) error {
	return errors.New("Error updating the database.")
}

func (mdb *errorMockDB) GetSetting(code string) (Setting, error) {
	return Setting{Code: code, Data: code}, nil
}

func (mdb *errorMockDB) PutSetting(setting Setting) error {
	return nil
}
