// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
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
	"log"
)

const createSubstoreTableSQL = `
	CREATE TABLE IF NOT EXISTS substore (
		id               serial primary key not null,
		account_id       int references account not null,
		from_model_id    int references model not null,
		to_model_id      int references model not null,
		store            varchar(200) not null,
		serial_number    varchar(200) not null
	)
`

const createSubstoreSQL = `
	INSERT INTO substore 
	(account_id, from_model_id, to_model_id, store, serial_number) 
	VALUES ($1,$2,$3,$4,$5)`

const listSubstoreSQL = `
	SELECT id, account_id, from_model_id, to_model_id, store, serial_number 
	FROM substore 
	WHERE account_id=$1`

const listUserSubstoreSQL = `
	SELECT s.id, account_id, from_model_id, to_model_id, store, serial_number 
	FROM substore s
	INNER JOIN useraccountlink l ON s.account_id = l.account_id
	INNER JOIN userinfo u ON l.user_id = u.id
	WHERE s.account_id=$1 AND u.username=$2
`
const updateSubstoreSQL = `
	UPDATE substore 
	SET account_id=$2, from_model_id=$3, to_model_id=$4, store=$5, serial_number=$6 
	WHERE id=$1`
const updateSubstoreForUserSQL = `
	UPDATE substore s 
	SET account_id=$2, from_model_id=$3, to_model_id=$4, store=$5, serial_number=$6 
	FROM useraccountlink ua ON ua.account_id=s.account_id
	INNER JOIN userinfo u ON ua.user_id=u.id
	WHERE s.id=$1 AND u.username=$7`

// Substore holds the substore details for an account in the local database
type Substore struct {
	ID           int    `json:"id"`
	AccountID    int    `json:"accountID"`
	FromModelID  int    `json:"fromModelID"`
	ToModelID    int    `json:"toModelID"`
	FromModel    Model  `json:"fromModel"`
	ToModel      Model  `json:"toModel"`
	Store        string `json:"store"`
	SerialNumber string `json:"serialnumber"`
}

// CreateSubstoreTable creates the database table for a sub-store
func (db *DB) CreateSubstoreTable() error {
	_, err := db.Exec(createSubstoreTableSQL)
	return err
}

// createSubstore creates an sub-store in the database
func (db *DB) createSubstore(store Substore) error {
	_, err := db.Exec(createSubstoreSQL, store.AccountID, store.FromModelID, store.ToModelID, store.Store, store.SerialNumber)
	if err != nil {
		log.Printf("Error creating the database sub-store: %v\n", err)
		return err
	}
	return nil
}

// ListSubstores returns a list of sub-stores
func (db *DB) listSubstores(accountID int) ([]Substore, error) {
	rows, err := db.Query(listSubstoreSQL, accountID)
	if err != nil {
		log.Printf("Error retrieving sub-stores: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	return db.rowsToSubstores(rows)
}

// listSubstoresFilteredByUser returns a list of sub-stores
func (db *DB) listSubstoresFilteredByUser(accountID int, username string) ([]Substore, error) {
	rows, err := db.Query(listUserSubstoreSQL, accountID, username)
	if err != nil {
		log.Printf("Error retrieving sub-stores of a user: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	return db.rowsToSubstores(rows)
}

func (db *DB) rowsToSubstores(rows *sql.Rows) ([]Substore, error) {
	stores := []Substore{}

	for rows.Next() {
		store := Substore{}
		err := rows.Scan(&store.ID, &store.AccountID, &store.FromModelID, &store.ToModelID, &store.Store, &store.SerialNumber)
		if err != nil {
			return nil, err
		}

		store.FromModel, err = db.getModel(store.FromModelID)
		if err != nil {
			log.Printf("Error retrieving database model: %v\n", err)
			return stores, err
		}

		store.ToModel, err = db.getModel(store.ToModelID)
		if err != nil {
			log.Printf("Error retrieving database model: %v\n", err)
			return stores, err
		}

		stores = append(stores, store)
	}

	return stores, nil
}

func (db *DB) updateSubstore(store Substore) (string, error) {
	return db.updateSubstoreFilteredByUser(store, anyUserFilter)
}

func (db *DB) updateSubstoreFilteredByUser(store Substore, username string) (string, error) {
	var err error

	if len(username) == 0 {
		_, err = db.Exec(updateSubstoreSQL, store.ID, store.AccountID, store.FromModelID, store.ToModelID, store.Store, store.SerialNumber)
	} else {
		_, err = db.Exec(updateSubstoreForUserSQL, store.ID, store.AccountID, store.FromModelID, store.ToModelID, store.Store, store.SerialNumber, username)
	}
	if err != nil {
		log.Printf("Error updating the database sub-store: %v\n", err)
		return "", err
	}

	return "", nil
}
