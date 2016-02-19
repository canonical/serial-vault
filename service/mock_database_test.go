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

// Successful mocks for the database
type mockDB struct{}

// CreateModelTable mock for the create model table method
func (mdb *mockDB) CreateModelTable() error {
	return nil
}

// ModelsList Mock the database response for a list of models
func (mdb *mockDB) ListModels() ([]Model, error) {

	var models []Model
	models = append(models, Model{ID: 1, BrandID: "Vendor", Name: "Alder", SigningKey: "alder", Revision: 1})
	models = append(models, Model{ID: 2, BrandID: "Vendor", Name: "Ash", SigningKey: "ash", Revision: 7})
	models = append(models, Model{ID: 3, BrandID: "Vendor", Name: "Basswood", SigningKey: "basswood", Revision: 23})
	models = append(models, Model{ID: 4, BrandID: "Vendor", Name: "Korina", SigningKey: "korina", Revision: 42})
	models = append(models, Model{ID: 5, BrandID: "Vendor", Name: "Mahogany", SigningKey: "mahogany", Revision: 10})
	models = append(models, Model{ID: 6, BrandID: "Vendor", Name: "Maple", SigningKey: "maple", Revision: 12})
	return models, nil
}

// FindModel mocks the database response for finding a model
func (mdb *mockDB) FindModel(brandID, modelName string, revision int) (*Model, error) {
	model := Model{ID: 1, BrandID: "Vendor", Name: "Alder", SigningKey: "../TestKey.asc", Revision: 1}
	return &model, nil
}

type errorString struct {
	message string
}

func (e *errorString) Error() string {
	return e.message
}

// Unsuccessful mocks for the database
type errorMockDB struct{}

// CreateModelTable mock for the create model table method
func (mdb *errorMockDB) CreateModelTable() error {
	return &errorString{"Error creating the model table."}
}

// ModelsList Mock the database response for a list of models
func (mdb *errorMockDB) ListModels() ([]Model, error) {
	return nil, &errorString{"Error getting the models."}
}

// FindModel mocks the database response for finding a model, returning an invalid signing-key
func (mdb *errorMockDB) FindModel(brandID, modelName string, revision int) (*Model, error) {
	modelNonexistentFile := Model{ID: 1, BrandID: "System", Name: "Bad Path", SigningKey: "not a good path", Revision: 2}
	modelInvalidKeyFile := Model{ID: 1, BrandID: "System", Name: "聖誕快樂", SigningKey: "../README.md", Revision: 2}

	if brandID == modelNonexistentFile.BrandID && modelName == modelNonexistentFile.Name && revision == modelNonexistentFile.Revision {
		return &modelNonexistentFile, nil
	} else if brandID == modelInvalidKeyFile.BrandID && modelName == modelInvalidKeyFile.Name && revision == modelInvalidKeyFile.Revision {
		return &modelInvalidKeyFile, nil
	}
	return nil, &errorString{"Error finding the model."}
}
