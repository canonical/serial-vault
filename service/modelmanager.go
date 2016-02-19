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

import "log"

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
const findModelSQL = "select id, brand_id, name, signing_key, revision from model where brand_id=$1, name=$2, revision=$3"

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
	var model *Model

	err := db.QueryRow(findModelSQL, brandID, modelName, revision).Scan(&model.ID, &model.BrandID, &model.Name, &model.SigningKey, &model.Revision)
	if err != nil {
		log.Printf("Error retrieving database model: %v\n", err)
		return nil, err
	}

	return model, nil
}
