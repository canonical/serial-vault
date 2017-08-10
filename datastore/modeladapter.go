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
	"regexp"
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
	err := validateModelName(model.Name)
	if err != nil {
		return "", err
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
	err := validateModelName(model.Name)
	if err != nil {
		return Model{}, "", err
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

// validateModelName validates name for the model; the rule is: lowercase with no spaces
func validateModelName(name string) error {
	return validateSyntax("Name", name, validModelNameRegexp)
}
