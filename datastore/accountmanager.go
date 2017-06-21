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
	"errors"
	"log"
	"strings"
)

const createAccountTableSQL = `
	CREATE TABLE IF NOT EXISTS account (
		id            serial primary key not null,
		authority_id  varchar(200) not null unique,
		assertion     text default ''
	)
`
const listAccountsSQL = "select id, authority_id, assertion from account order by authority_id"
const getAccountSQL = "select id, authority_id, assertion from account where authority_id=$1"
const updateAccountSQL = "update account set assertion=$2 where authority_id=$1"
const upsertAccountSQL = `
	WITH upsert AS (
		update account set authority_id=$1, assertion=$2
		where authority_id=$1
		RETURNING *
	)
	insert into account (authority_id,assertion)
	select $1, $2
	where not exists (select * from upsert)
`

// Account holds the store account assertion in the local database
type Account struct {
	ID          int
	AuthorityID string
	Assertion   string
}

// CreateAccountTable creates the database table for a account.
func (db *DB) CreateAccountTable() error {
	_, err := db.Exec(createAccountTableSQL)
	return err
}

// ListAccounts fetches the available accounts from the database.
func (db *DB) ListAccounts() ([]Account, error) {
	accounts := []Account{}

	rows, err := db.Query(listAccountsSQL)
	if err != nil {
		log.Printf("Error retrieving database accounts: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		account := Account{}
		err := rows.Scan(&account.ID, &account.AuthorityID, &account.Assertion)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// GetAccount fetches a single account from the database by the authority ID
func (db *DB) GetAccount(authorityID string) (Account, error) {
	account := Account{}

	err := db.QueryRow(getAccountSQL, authorityID).Scan(&account.ID, &account.AuthorityID, &account.Assertion)
	if err != nil {
		log.Printf("Error retrieving account: %v\n", err)
		return account, err
	}

	return account, nil
}

// UpdateAccountAssertion sets the account-key assertion of a keypair
func (db *DB) UpdateAccountAssertion(authorityID, assertion string) error {
	_, err := db.Exec(updateAccountSQL, authorityID, assertion)
	if err != nil {
		log.Printf("Error updating the database account assertion: %v\n", err)
		return err
	}

	return nil
}

// PutAccount stores an account in the database
func (db *DB) PutAccount(account Account) (string, error) {
	// Validate the data
	if strings.TrimSpace(account.AuthorityID) == "" {
		return "error-validate-account", errors.New("The Authority ID must be entered")
	}

	_, err := db.Exec(upsertAccountSQL, account.AuthorityID, account.Assertion)
	if err != nil {
		log.Printf("Error updating the database account: %v\n", err)
		return "", err
	}

	return "", nil
}
