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
	"strings"
)

const createKeypairTableSQL = `
	CREATE TABLE IF NOT EXISTS keypair (
		id            serial primary key not null,
		authority_id  varchar(200) not null,
		key_id        varchar(200) not null,
		active        boolean default true,
		sealed_key    text,
		assertion     text default ''
	)
`
const listKeypairsSQL = "select id, authority_id, key_id, active, assertion from keypair order by authority_id, key_id"
const listKeypairsForUserSQL = `
	select k.id, k.authority_id, k.key_id, k.active, k.assertion 
	from keypair k
	inner join account acc on acc.authority_id=k.authority_id
	inner join useraccountlink ua on ua.account_id=acc.id
	inner join userinfo u on ua.user_id=u.id
	where u.username=$1
	order by authority_id, key_id`
const getKeypairSQL = "select id, authority_id, key_id, active, sealed_key, assertion from keypair where id=$1"
const getKeypairByPublicIDSQL = "select id, authority_id, key_id, active, sealed_key, assertion from keypair where authority_id=$1 and key_id=$2"
const toggleKeypairSQL = "update keypair set active=$2 where id=$1"
const toggleKeypairForUserSQL = `
	update keypair k
	set active=$2
	from account acc 
	inner join useraccountlink ua on ua.account_id=acc.id
	inner join userinfo u on ua.user_id=u.id
	where k.id=$1 and u.username=$3 and acc.authority_id=k.authority_id`
const upsertKeypairSQL = `
	WITH upsert AS (
		update keypair set authority_id=$1, key_id=$2, sealed_key=$3, assertion=$4
		where authority_id=$1 and key_id=$2
		RETURNING *
	)
	insert into keypair (authority_id,key_id,sealed_key,assertion)
	select $1, $2, $3, $4
	where not exists (select * from upsert)
`
const updateKeypairSQL = "update keypair set assertion=$2 where id=$1"

// Add the assertion field to store the assertion for the account key to the table
const alterKeypairAddAssertion = "alter table keypair add column assertion text default ''"

// Keypair holds the keypair reference details in the local database
type Keypair struct {
	ID          int
	AuthorityID string
	KeyID       string
	Active      bool
	SealedKey   string
	Assertion   string
}

// CreateKeypairTable creates the database table for a keypair.
func (db *DB) CreateKeypairTable() error {
	_, err := db.Exec(createKeypairTableSQL)
	return err
}

// AlterKeypairTable adds extra fields to an existing keypair database table
func (db *DB) AlterKeypairTable() error {
	db.Exec(alterKeypairAddAssertion)
	// Ignore errors as the field may already be added
	return nil
}

func (db *DB) listAllKeypairs() ([]Keypair, error) {
	return db.listKeypairsFilteredByUser(anyUserFilter)
}

func (db *DB) listKeypairsFilteredByUser(username string) ([]Keypair, error) {
	var keypairs []Keypair

	var (
		rows *sql.Rows
		err  error
	)

	if len(username) == 0 {
		rows, err = db.Query(listKeypairsSQL)
	} else {
		rows, err = db.Query(listKeypairsForUserSQL, username)
	}
	if err != nil {
		log.Printf("Error retrieving database keypairs: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		keypair := Keypair{}
		err := rows.Scan(&keypair.ID, &keypair.AuthorityID, &keypair.KeyID, &keypair.Active, &keypair.Assertion)
		if err != nil {
			return nil, err
		}
		keypairs = append(keypairs, keypair)
	}

	return keypairs, nil
}

// GetKeypair fetches a single keypair from the database by ID
func (db *DB) GetKeypair(keypairID int) (Keypair, error) {
	keypair := Keypair{}

	err := db.QueryRow(getKeypairSQL, keypairID).Scan(&keypair.ID, &keypair.AuthorityID, &keypair.KeyID, &keypair.Active, &keypair.SealedKey, &keypair.Assertion)
	if err != nil {
		log.Printf("Error retrieving keypair by ID: %v\n", err)
		return keypair, err
	}

	return keypair, nil
}

// GetKeypairByPublicID fetches a single keypair from the database by ID
func (db *DB) GetKeypairByPublicID(authorityID, keyID string) (Keypair, error) {
	keypair := Keypair{}

	err := db.QueryRow(getKeypairByPublicIDSQL, authorityID, keyID).Scan(&keypair.ID, &keypair.AuthorityID, &keypair.KeyID, &keypair.Active, &keypair.SealedKey, &keypair.Assertion)
	if err != nil {
		log.Printf("Error retrieving keypair by ID: %v\n", err)
		return keypair, err
	}

	return keypair, nil
}

// PutKeypair stores a keypair in the database
func (db *DB) PutKeypair(keypair Keypair) (string, error) {
	// Validate the data
	if strings.TrimSpace(keypair.AuthorityID) == "" || strings.TrimSpace(keypair.KeyID) == "" {
		return "error-validate-keypair", errors.New("The Authority ID and the Key ID must be entered")
	}

	_, err := db.Exec(upsertKeypairSQL, keypair.AuthorityID, keypair.KeyID, keypair.SealedKey, keypair.Assertion)
	if err != nil {
		log.Printf("Error updating the database keypair: %v\n", err)
		return "", err
	}

	return "", nil
}

func (db *DB) updateKeypairActive(keypairID int, active bool) error {
	return db.updateKeypairActiveFilteredByUser(keypairID, active, anyUserFilter)
}

func (db *DB) updateKeypairActiveFilteredByUser(keypairID int, active bool, username string) error {
	var err error

	if len(username) == 0 {
		_, err = db.Exec(toggleKeypairSQL, keypairID, active)
	} else {
		_, err = db.Exec(toggleKeypairForUserSQL, keypairID, active, username)
	}
	if err != nil {
		log.Printf("Error updating the database keypair: %v\n", err)
		return err
	}

	return nil
}

// updateKeypairAssertion sets the account-key assertion of a keypair
func (db *DB) updateKeypairAssertion(keypairID int, assertion string) error {
	_, err := db.Exec(updateKeypairSQL, keypairID, assertion)
	if err != nil {
		log.Printf("Error updating the database keypair assertion: %v\n", err)
		return err
	}

	return nil
}
