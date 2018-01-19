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

const createAccountTableSQL = `
	CREATE TABLE IF NOT EXISTS account (
		id            serial primary key not null,
		authority_id  varchar(200) not null unique,
		assertion     text default '',
		resellerapi   bool default false
	)
`
const listAccountsSQL = "select id, authority_id, assertion, resellerapi from account order by authority_id"
const getAccountSQL = "select id, authority_id, assertion, resellerapi from account where authority_id=$1"
const updateAccountSQL = "update account set assertion=$2, resellerapi=$3 where authority_id=$1"
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

const listUserAccountsSQL = `
	select a.id, authority_id, assertion, resellerapi 
	from account a
	inner join useraccountlink l on a.id = l.account_id
	inner join userinfo u on l.user_id = u.id
	where u.username=$1
`

const listNotUserAccountsSQL = `
	select id, authority_id, assertion, resellerapi 
	from account
	where id not in (
		select a.id 
		from account a
		inner join useraccountlink l on a.id = l.account_id
		inner join userinfo u on l.user_id = u.id
		where u.username=$1
	)
`

// Add the reseller API field to indicate whether the reseller functions are available for an account
const alterAccountResellerAPI = "alter table account add column resellerapi bool default false"

// Account holds the store account assertion in the local database
type Account struct {
	ID          int
	AuthorityID string
	Assertion   string
	ResellerAPI bool
}

// CreateAccountTable creates the database table for an account.
func (db *DB) CreateAccountTable() error {
	_, err := db.Exec(createAccountTableSQL)
	return err
}

// AlterAccountTable modifies the database table for an account.
func (db *DB) AlterAccountTable() error {
	db.Exec(alterAccountResellerAPI)
	return nil
}

func (db *DB) listAllAccounts() ([]Account, error) {
	return db.listAccountsFilteredByUser(anyUserFilter)
}

func (db *DB) listAccountsFilteredByUser(username string) ([]Account, error) {

	var (
		rows *sql.Rows
		err  error
	)

	if len(username) == 0 {
		rows, err = db.Query(listAccountsSQL)
	} else {
		rows, err = db.Query(listUserAccountsSQL, username)
	}
	if err != nil {
		log.Printf("Error retrieving database accounts: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	return rowsToAccounts(rows)
}

// GetAccount fetches a single account from the database by the authority ID
func (db *DB) GetAccount(authorityID string) (Account, error) {
	account := Account{}

	err := db.QueryRow(getAccountSQL, authorityID).Scan(&account.ID, &account.AuthorityID, &account.Assertion, &account.ResellerAPI)
	if err != nil {
		log.Printf("Error retrieving account: %v\n", err)
		return account, err
	}

	return account, nil
}

// UpdateAccountAssertion sets the account-key assertion of a keypair
func (db *DB) UpdateAccountAssertion(authorityID, assertion string, resellerAPI bool) error {
	_, err := db.Exec(updateAccountSQL, authorityID, assertion, resellerAPI)
	if err != nil {
		log.Printf("Error updating the database account assertion: %v\n", err)
		return err
	}

	return nil
}

// putAccount stores an account in the database
func (db *DB) putAccount(account Account) (string, error) {
	_, err := db.Exec(upsertAccountSQL, account.AuthorityID, account.Assertion)
	if err != nil {
		log.Printf("Error updating the database account: %v\n", err)
		return "", err
	}

	return "", nil
}

// ListUserAccounts returns a list of Account objects related with certain user
func (db *DB) ListUserAccounts(username string) ([]Account, error) {
	rows, err := db.Query(listUserAccountsSQL, username)
	if err != nil {
		log.Printf("Error retrieving database accounts of certain user: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	return rowsToAccounts(rows)
}

// ListNotUserAccounts returns a list of Account objects that are not related with certain user
func (db *DB) ListNotUserAccounts(username string) ([]Account, error) {
	rows, err := db.Query(listNotUserAccountsSQL, username)
	if err != nil {
		log.Printf("Error retrieving database accounts not belonging to certain user: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	return rowsToAccounts(rows)
}

func rowsToAccounts(rows *sql.Rows) ([]Account, error) {
	accounts := []Account{}

	for rows.Next() {
		account := Account{}
		err := rows.Scan(&account.ID, &account.AuthorityID, &account.Assertion, &account.ResellerAPI)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// BuildAccountsFromAuthorityIDs from a list of strings representing authority ids, build related datastore.Account objects
func BuildAccountsFromAuthorityIDs(authorityIDs []string) []Account {
	var accounts []Account
	for _, authorityID := range authorityIDs {
		accounts = append(accounts, BuildAccountFromAuthorityID(authorityID))
	}
	return accounts
}

// BuildAccountFromAuthorityID from a string representing authority id, build related datastore.Account object
func BuildAccountFromAuthorityID(authorityID string) Account {
	return Account{AuthorityID: authorityID}
}
