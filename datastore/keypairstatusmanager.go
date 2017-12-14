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

const createKeypairStatusTableSQL = `
CREATE TABLE IF NOT EXISTS keypairstatus (
	id            serial primary key not null,
	authority_id  varchar(200) not null,
	key_name      varchar(200) not null,
	keypair_id    int references keypair null,
	status        varchar(20)
)
`

const createKeypairStatusSQL = "insert into keypairstatus (authority_id,key_name,status) values ($1,$2,$3) RETURNING id"

const getKeypairStatusSQL = `
select id, authority_id, key_name, keypair_id, status
from keypairstatus
where authority_id=$1 and key_name=$2`

const listKeypairStatusProgressSQL = `
select id, authority_id, key_name, keypair_id, status
from keypairstatus ks
where ks.keypair_id is null
order by authority_id, key_name
`

const listKeypairStatusProgressForUserSQL = `
select id, authority_id, key_name, keypair_id, status
from keypairstatus ks
inner join account acc on acc.authority_id=ks.authority_id
inner join useraccountlink ua on ua.account_id=acc.id
inner join userinfo u on ua.user_id=u.id
where u.username=$1 and ks.keypair_id is null
order by authority_id, key_name
`

const updateKeypairStatusSQL = `
update keypairstatus
set status=$3
where authority_id=$1 and key_name=$2`

const updateKeypairStatusWithIDSQL = `
update keypairstatus
set keypair_id=$3, status=$4
where authority_id=$1 and key_name=$2`

// Indexes
const createKeypairStatusAuthKeyIndexSQL = "CREATE UNIQUE INDEX IF NOT EXISTS auth_key_idx ON keypairstatus (authority_id, key_name)"

// KeypairStatus holds the keypair status in the local database
type KeypairStatus struct {
	ID          int    `json:"id"`
	AuthorityID string `json:"authority-id"`
	KeyName     string `json:"key-name"`
	KeypairID   int    `json:"keypair-id"`
	Status      string `json:"status"`
}

// Statuses for keypairs
const (
	KeypairStatusCreating   = "creating"
	KeypairStatusExporting  = "exporting"
	KeypairStatusEncrypting = "encrypting"
	KeypairStatusStoring    = "storing"
	KeypairStatusComplete   = "complete"
)

// CreateKeypairStatusTable creates the database table for a keypair status.
func (db *DB) CreateKeypairStatusTable() error {
	_, err := db.Exec(createKeypairStatusTableSQL)
	return err
}

// AlterKeypairStatusTable adds indexes to the table
func (db *DB) AlterKeypairStatusTable() error {
	// Create the index on the auth / key
	_, err := db.Exec(createKeypairStatusAuthKeyIndexSQL)
	return err
}

// CreateKeypairStatus adds a keypair status record to track the generation of a keypair
func (db *DB) CreateKeypairStatus(ks KeypairStatus) (int, error) {
	// Create the keypair status in the database
	var createdID int
	err := db.QueryRow(createKeypairStatusSQL, ks.AuthorityID, ks.KeyName, KeypairStatusCreating).Scan(&createdID)
	if err != nil {
		log.Printf("Error creating the keypair status: %v\n", err)
	}
	return createdID, err
}

// UpdateKeypairStatus updates the status of generating
func (db *DB) UpdateKeypairStatus(ks KeypairStatus) error {

	var err error

	if ks.KeypairID > 0 {
		_, err = db.Exec(updateKeypairStatusWithIDSQL, ks.AuthorityID, ks.KeyName, ks.KeypairID, ks.Status)
	} else {
		_, err = db.Exec(updateKeypairStatusSQL, ks.AuthorityID, ks.KeyName, ks.Status)
	}

	if err != nil {
		log.Printf("Error updating the keypair status: %v\n", err)
	}

	return err

}

// GetKeypairStatus fetches the keypair status
func (db *DB) GetKeypairStatus(authorityID, keyName string) (KeypairStatus, error) {
	var keypairID sql.NullInt64
	ks := KeypairStatus{}
	err := db.QueryRow(getKeypairStatusSQL, authorityID, keyName).Scan(&ks.ID, &ks.AuthorityID, &ks.KeyName, &keypairID, &ks.Status)
	if err != nil {
		log.Printf("Error fetching the keypair status: %v\n", err)
		return ks, err
	}

	if keypairID.Valid {
		ks.KeypairID = int(keypairID.Int64)
	}

	return ks, err
}

func (db *DB) listAllKeypairStatus() ([]KeypairStatus, error) {
	return db.listKeypairStatusFilteredByUser(anyUserFilter)
}

func (db *DB) listKeypairStatusFilteredByUser(username string) ([]KeypairStatus, error) {
	keypairs := []KeypairStatus{}
	var keypairID sql.NullInt64

	var (
		rows *sql.Rows
		err  error
	)

	if len(username) == 0 {
		rows, err = db.Query(listKeypairStatusProgressSQL)
	} else {
		rows, err = db.Query(listKeypairStatusProgressForUserSQL, username)
	}
	if err != nil {
		log.Printf("Error retrieving database keypairs: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		ks := KeypairStatus{}
		err := rows.Scan(&ks.ID, &ks.AuthorityID, &ks.KeyName, &keypairID, &ks.Status)
		if err != nil {
			return nil, err
		}

		if keypairID.Valid {
			ks.KeypairID = int(keypairID.Int64)
		}

		keypairs = append(keypairs, ks)
	}

	return keypairs, nil
}
