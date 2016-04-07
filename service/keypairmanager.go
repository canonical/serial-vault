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
	"errors"
	"log"
	"strings"
)

const createKeypairTableSQL = `
	CREATE TABLE IF NOT EXISTS keypair (
		id            serial primary key not null,
		authority_id  varchar(200) not null,
		key_id        varchar(200) not null,
		active        boolean default true
	)
`
const listKeypairsSQL = "select id, authority_id, key_id from keypair order by authority_id, key_id"
const getKeypairSQL = "select id, authority_id, key_id from keypair where id=$1"
const toggleKeypairSQL = "update keypair set active=$2 where id=$1"
const upsertKeypairSQL = `
	WITH upsert AS (
		update keypair set authority_id=$1, key_id=$2 RETURNING *
	)
	insert into keypair (authority_id,key_id)
	select $1, $2
	where not exists (select * from upsert)
`

// Keypair holds the keypair reference details in the local database
type Keypair struct {
	ID          int
	AuthorityID string
	KeyID       string
	Active      bool
}

// CreateKeypairTable creates the database table for a keypair.
func (db *DB) CreateKeypairTable() error {
	_, err := db.Exec(createKeypairTableSQL)
	return err
}

// ListKeypairs fetches the available keypairs from the database.
func (db *DB) ListKeypairs() ([]Keypair, error) {
	var keypairs []Keypair

	rows, err := db.Query(listKeypairsSQL)
	if err != nil {
		log.Printf("Error retrieving database keypairs: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		keypair := Keypair{}
		err := rows.Scan(&keypair.ID, &keypair.AuthorityID, &keypair.KeyID)
		if err != nil {
			return nil, err
		}
		keypairs = append(keypairs, keypair)
	}

	return keypairs, nil
}

// GetKeypair fetches a single keypair from the database by ID
func (db *DB) GetKeypair(keypairID int) (*Keypair, error) {
	keypair := Keypair{}

	err := db.QueryRow(getKeypairSQL, keypairID).Scan(&keypair.ID, &keypair.AuthorityID, &keypair.KeyID, &keypair.Active)
	if err != nil {
		log.Printf("Error retrieving keypair by ID: %v\n", err)
		return nil, err
	}

	return &keypair, nil
}

// PutKeypair stores a keypair in the database
func (db *DB) PutKeypair(keypair Keypair) (string, error) {
	// Validate the data
	if strings.TrimSpace(keypair.AuthorityID) == "" || strings.TrimSpace(keypair.KeyID) == "" {
		return "error-validate-keypair", errors.New("The Authority ID and the Key ID must be entered.")
	}

	_, err := db.Exec(upsertKeypairSQL, keypair.AuthorityID, keypair.KeyID)
	if err != nil {
		log.Printf("Error updating the database keypair: %v\n", err)
		return "", err
	}

	return "", nil
}

// UpdateKeypairActive sets the active state of a keypair
func (db *DB) UpdateKeypairActive(keypairID int, active bool) error {
	_, err := db.Exec(toggleKeypairSQL, keypairID, active)
	if err != nil {
		log.Printf("Error updating the database keypair: %v\n", err)
		return err
	}

	return nil
}
