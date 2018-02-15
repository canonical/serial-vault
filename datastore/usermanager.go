// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
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
)

const createUserTableSQL = `
	CREATE TABLE IF NOT EXISTS userinfo (
		id               serial primary key not null,
		username         varchar(200) not null unique,
		name             varchar(200),
		email            varchar(255) not null,
		userrole         int not null,
		api_key          varchar(200) not null
	)
`

const createAccountUserLinkTableSQL = `
	CREATE TABLE IF NOT EXISTS useraccountlink (
		user_id          int references userinfo not null,
		account_id     	 int references account not null
	)
`

const listUsersSQL = "select id, username, name, email, userrole, api_key from userinfo order by username"
const getUserSQL = "select id, username, name, email, userrole, api_key from userinfo where id=$1"
const getUserByUsernameSQL = "select id, username, name, email, userrole, api_key from userinfo where username=$1"
const findUsersSQL = "select id, username, name, email, userrole, api_key from userinfo where username like '%$1%' or name like '%$1%'"
const createUserSQL = "insert into userinfo (username, name, email, userrole, api_key) values ($1,$2,$3,$4,$5) RETURNING id"
const updateUserSQL = "update userinfo set username=$1, name=$2, email=$3, userrole=$4, api_key=$6 where id=$5"
const deleteUserSQL = "delete from userinfo where id=$1"

const listAccountUsersSQL = `
	select id, username, name, email, userrole, api_key
	from userinfo u
	inner join useraccountlink l on u.id = l.user_id
	inner join account a on l.account_id = a.id
	where a.authority_id=$1
`

const findAccountUserSQL = `
	select count(*) 
	from userinfo u
	inner join useraccountlink l on u.id = l.user_id
	inner join account a on l.account_id = a.id
	where u.username=$1 and a.authority_id=$2
`

const deleteUserAccountsSQL = "delete from useraccountlink where user_id=$1"
const linkAccountToUserSQL = "insert into useraccountlink (user_id, account_id) values ($1,$2)"

const alterUserRemoveOpenIDIdentity = "alter table userinfo drop column if exists openid_identity"

// Add the API key field to the models table (nullable)
const alterUserAPIKey = "alter table userinfo add column api_key varchar(200) default ''"

// Make the API key not-nullable
const alterUserAPIKeyNotNullable = `alter table userinfo
	alter column api_key set not null,
	alter column api_key drop default
`

// Available user roles:
//
// * Invalid:	default value set in case there is no authentication previous process for this user and thus not got a valid role.
// * Standard:	role for regular users. This is the less privileged role
// * Admin:		role for admin users, including standard role permissions but not superuser ones
// * Superuser:	role for users having all the permissions
const (
	Invalid   = iota       // 0
	Standard  = 100 * iota // 100
	Admin                  // 200
	Superuser              // 300
)

// RoleName holds the names for each of the roles
var RoleName = map[int]string{0: "", 100: "standard", 200: "admin", 300: "superuser"}

// RoleID holds the ID for each of the named roles
var RoleID = map[string]int{"": 0, "standard": 100, "admin": 200, "superuser": 300}

// User holds user personal, authentication and authorization info
type User struct {
	ID       int
	Username string
	Name     string
	Email    string
	APIKey   string
	Role     int
	Accounts []Account
}

// CreateUserTable creates User table in database
func (db *DB) CreateUserTable() error {
	_, err := db.Exec(createUserTableSQL)
	return err
}

// CreateAccountUserLinkTable creates table to link User and Account tables in a m-m relationship
func (db *DB) CreateAccountUserLinkTable() error {
	// Add and populate the API key field (ignore error as it may already be there)
	db.addUserAPIKeyField()

	_, err := db.Exec(createAccountUserLinkTableSQL)
	return err
}

// AlterUserTable includes all user table definition modifications
func (db *DB) AlterUserTable() error {
	_, err := db.Exec(alterUserRemoveOpenIDIdentity)
	return err
}

// addUserAPIKeyField adds and defaults the API key field to the user table
func (db *DB) addUserAPIKeyField() error {

	// Add the API key field to the user table
	_, err := db.Exec(alterUserAPIKey)
	if err != nil {
		// Field already exists so skip
		return nil
	}

	// Default the API key for any records where it is empty
	users, err := db.ListUsers()
	if err != nil {
		return err
	}
	for _, user := range users {
		if len(user.APIKey) > 0 {
			continue
		}

		// Generate an random API key and update the record
		apiKey, err := generateAPIKey()
		if err != nil {
			log.Printf("Could not generate random string for the API key")
			return errors.New("Error generating random string for the API key")
		}

		// Update the API key on the model
		user.APIKey = apiKey
		db.updateUser(user)
	}

	// Add the constraints to the API key field
	_, err = db.Exec(alterUserAPIKeyNotNullable)
	if err != nil {
		return err
	}

	return nil
}

// ListUsers returns current available users in database
func (db *DB) ListUsers() ([]User, error) {
	rows, err := db.Query(listUsersSQL)
	if err != nil {
		log.Printf("Error retrieving database users: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	return db.rowsToUsers(rows)
}

// FindUsers returns array of users matching query string in username or name
func (db *DB) FindUsers(query string) ([]User, error) {
	rows, err := db.Query(findUsersSQL, query)
	if err != nil {
		log.Printf("Error searching for database users: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	return db.rowsToUsers(rows)
}

// GetUser fetches a single user from database
func (db *DB) GetUser(userID int) (User, error) {
	row := db.QueryRow(getUserSQL, userID)
	user, err := db.rowToUser(row)
	if err != nil {
		log.Printf("Error retrieving user %v: %v\n", userID, err)
	}
	return user, err
}

// GetUserByUsername fetches a single user from database
func (db *DB) GetUserByUsername(username string) (User, error) {
	row := db.QueryRow(getUserByUsernameSQL, username)
	user, err := db.rowToUser(row)
	if err != nil {
		log.Printf("Error retrieving user %v: %v\n", username, err)
	}
	return user, err
}

// createUser adds a new record to User database table, Returns new record identifier if success
func (db *DB) createUser(user User) (int, error) {

	createdUserID := -1

	err := db.transaction(func(tx *sql.Tx) error {

		err := tx.QueryRow(createUserSQL, user.Username, user.Name, user.Email, user.Role, user.APIKey).Scan(&createdUserID)
		if err != nil {
			log.Printf("Error creating user %v: %v\n", user.Username, err)
			return err
		}

		err = db.putUserAccounts(createdUserID, user.Accounts, tx)
		if err != nil {
			log.Printf("Error creating user %v: %v\n", user.Username, err)
			return err
		}

		return nil
	})

	return createdUserID, err
}

// updateUser sets user new values for an existing record. Also updates useraccount link. All that in a transaction
func (db *DB) updateUser(user User) error {

	return db.transaction(func(tx *sql.Tx) error {

		_, err := tx.Exec(updateUserSQL, user.Username, user.Name, user.Email, user.Role, user.ID, user.APIKey)
		if err != nil {
			log.Printf("Error updating database user %v: %v\n", user.ID, err)
			return err
		}

		err = db.putUserAccounts(user.ID, user.Accounts, tx)
		if err != nil {
			log.Printf("Error creating user %v: %v\n", user.Username, err)
			return err
		}

		return nil
	})
}

// DeleteUser deletes a user
func (db *DB) DeleteUser(userID int) error {

	return db.transaction(func(tx *sql.Tx) error {

		_, err := tx.Exec(deleteUserSQL, userID)
		if err != nil {
			log.Printf("Error deleting database user %v: %v\n", userID, err)
			return err
		}

		_, err = tx.Exec(deleteUserAccountsSQL, userID)
		if err != nil {
			log.Printf("Error deleting user accounts: %v", err)
			return err
		}

		return nil
	})
}

// ListAccountUsers returns list of User related with certain account
func (db *DB) ListAccountUsers(authorityID string) ([]User, error) {
	users := []User{}

	rows, err := db.Query(listAccountUsersSQL, authorityID)
	if err != nil {
		log.Printf("Error retrieving database users of certain account: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user := User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Name, &user.Email, &user.Role, &user.APIKey)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// CheckUserInAccount verifies that a user has permissions to a specific account
func (db *DB) CheckUserInAccount(username, authorityID string) bool {
	if username == "" {
		return true
	}

	var count int

	row := db.QueryRow(findAccountUserSQL, username, authorityID)
	err := row.Scan(&count)
	if err != nil {
		log.Printf("Error retrieving database account of certain user: %v\n", err)
		return false
	}

	return count > 0
}

func (db *DB) putUserAccounts(userID int, accounts []Account, tx *sql.Tx) error {
	// first, delete previous registers if any
	_, err := tx.Exec(deleteUserAccountsSQL, userID)
	if err != nil {
		log.Printf("Could not delete user accounts: %v", err)
		return err
	}

	// link received data
	for _, account := range accounts {

		// if account id is not a valid identifier, fetch Account from DB using autorithyID field
		if account.ID == 0 {
			account, err = db.GetAccount(account.AuthorityID)
			if err != nil {
				log.Printf("Invalid account: %v", err)
				return err
			}
		}

		_, err := tx.Exec(linkAccountToUserSQL, userID, account.ID)
		if err != nil {
			log.Printf("Could not complete linking user to account transaction: %v", err)
			return err
		}
	}

	return nil
}

func (db *DB) rowToUser(row *sql.Row) (User, error) {
	user := User{}
	err := row.Scan(&user.ID, &user.Username, &user.Name, &user.Email, &user.Role, &user.APIKey)
	if err != nil {
		return User{}, err
	}

	// Get related accounts and fill related User field
	user.Accounts, err = db.listAccountsFilteredByUser(user.Username)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) rowsToUser(rows *sql.Rows) (User, error) {
	user := User{}
	err := rows.Scan(&user.ID, &user.Username, &user.Name, &user.Email, &user.Role, &user.APIKey)
	if err != nil {
		return User{}, err
	}

	// Get related accounts and fill related User field
	user.Accounts, err = db.listAccountsFilteredByUser(user.Username)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) rowsToUsers(rows *sql.Rows) ([]User, error) {
	users := []User{}

	for rows.Next() {
		user, err := db.rowsToUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
