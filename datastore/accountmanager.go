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

const createAccountSQL = "INSERT INTO account (authority_id, assertion, resellerapi) VALUES ($1,$2,$3)"
const listAccountsSQL = "select id, authority_id, assertion, resellerapi from account order by authority_id"
const getAccountSQL = "select id, authority_id, assertion, resellerapi from account where authority_id=$1"

const getAccountByIDSQL = "select id, authority_id, assertion, resellerapi from account where id=$1"
const getUserAccountByIDSQL = `
	select a.id, a.authority_id, a.assertion, a.resellerapi 
	from account a
	inner join useraccountlink l on a.id = l.account_id
	inner join userinfo u on l.user_id = u.id
	where a.id=$1 and u.username=$2`

const updateAccountSQL = "update account set authority_id=$2, assertion=$3, resellerapi=$4 where id=$1"
const updateUserAccountSQL = `
	UPDATE account a
	SET authority_id=$3, assertion=$4, resellerapi=$5 
	INNER JOIN useraccountlink l on a.id = l.account_id
	INNER JOIN userinfo u on l.user_id = u.id
	WHERE a.id=$1 AND u.username=$2
`
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
	select a.id, a.authority_id, a.assertion, a.resellerapi 
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

// sqlite3 syntax for syncing data locally
const syncUpsertAccountSQL = `
	INSERT OR REPLACE INTO account
	(id,authority_id,assertion,resellerapi)
	VALUES ($1, $2, $3, $4)
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

// CreateAccount creates an account in the database
func (db *DB) CreateAccount(account Account) error {
	_, err := db.Exec(createAccountSQL, account.AuthorityID, account.Assertion, account.ResellerAPI)
	if err != nil {
		log.Printf("Error creating the database account: %v\n", err)
		return err
	}
	return nil
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

// getAccountByID fetches a single account from the database by the ID
func (db *DB) getAccountByID(accountID int) (Account, error) {
	account := Account{}

	err := db.QueryRow(getAccountByIDSQL, accountID).Scan(&account.ID, &account.AuthorityID, &account.Assertion, &account.ResellerAPI)
	if err != nil {
		log.Printf("Error retrieving account: %v\n", err)
		return account, err
	}

	return account, nil
}

// getUserAccountByID fetches a single account from the database by the ID
func (db *DB) getUserAccountByID(accountID int, username string) (Account, error) {
	account := Account{}

	err := db.QueryRow(getUserAccountByIDSQL, accountID, username).Scan(&account.ID, &account.AuthorityID, &account.Assertion, &account.ResellerAPI)
	if err != nil {
		log.Printf("Error retrieving account: %v\n", err)
		return account, err
	}

	return account, nil
}

// updateAccount updates an account in the database
func (db *DB) updateAccount(account Account) error {
	_, err := db.Exec(updateAccountSQL, account.ID, account.AuthorityID, account.Assertion, account.ResellerAPI)
	if err != nil {
		log.Printf("Error updating the database account: %v\n", err)
		return err
	}

	return nil
}

// updateUserAccount updates an account in the database
func (db *DB) updateUserAccount(account Account, username string) error {
	_, err := db.Exec(updateUserAccountSQL, account.ID, username, account.AuthorityID, account.Assertion, account.ResellerAPI)
	if err != nil {
		log.Printf("Error updating the database account: %v\n", err)
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

// syncAccount stores an account in the database
func (db *DB) syncAccount(account Account) error {
	_, err := db.Exec(syncUpsertAccountSQL, account.ID, account.AuthorityID, account.Assertion, account.ResellerAPI)
	if err != nil {
		log.Printf("Error updating the database account: %v\n", err)
		return err
	}

	return nil
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
