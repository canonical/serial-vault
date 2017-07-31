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
	"log"
)

const createUserTableSQL = `
	CREATE TABLE IF NOT EXISTS userinfo (
		id               serial primary key not null,
		username         varchar(200) not null unique,
		openid_identity  text not null,
		name             varchar(200),
		email            varchar(255) not null,
		userrole         int not null
	)
`

const createAccountUserLinkTableSQL = `
	CREATE TABLE IF NOT EXISTS useraccountlink (
		user_id          int references userinfo not null,
		account_id     	 int references account not null
	)
`

const listUsersSQL = "select id, username, openid_identity, name, email, userrole from userinfo order by username"
const getUserSQL = "select id, username, openid_identity, name, email, userrole from userinfo where id=$1"
const findUsersSQL = "select id, username, openid_identity, name, email, userrole from userinfo where username like '%$1%' or name like '%$1%'"
const createUserSQL = "insert into userinfo (username, openid_identity, name, email, userrole) values ($1,$2,$3,$4,$5) RETURNING id"
const updateUserSQL = "update userinfo set username=$1, openid_identity=$2, name=$3, email=$4, userrole=$5 where id=$6"
const deleteUserSQL = "delete from userinfo where id=$1"

const listAccountUsersSQL = `
	select id, username, openid_identity, name, email, userrole
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

// Available user roles:
//
// * Standard:	role for regular users. This is the less privileged role
// * Admin:		role for admin users, including standard role permissions but not superuser ones
// * Superuser:	role for users having all the permissions
const (
	_         = iota
	Standard  = 100 * iota // 100
	Admin                  // 200
	Superuser              // 300
)

// User holds user personal, authentication and authorization info
type User struct {
	ID             int
	Username       string
	OpenIDIdentity string
	Name           string
	Email          string
	Role           int
}

// CreateUserTable creates User table in database
func (db *DB) CreateUserTable() error {
	_, err := db.Exec(createUserTableSQL)
	return err
}

// CreateAccountUserLinkTable creates table to link User and Account tables in a m-m relationship
func (db *DB) CreateAccountUserLinkTable() error {
	_, err := db.Exec(createAccountUserLinkTableSQL)
	return err
}

// ListUsers returns current available users in database
func (db *DB) ListUsers() ([]User, error) {
	rows, err := db.Query(listUsersSQL)
	if err != nil {
		log.Printf("Error retrieving database users: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	return rowsToUsers(rows)
}

// FindUsers returns array of users matching query string in username or name
func (db *DB) FindUsers(query string) ([]User, error) {
	users := []User{}

	rows, err := db.Query(findUsersSQL, query)
	if err != nil {
		log.Printf("Error searching for database users: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user := User{}
		err := rows.Scan(&user.ID, &user.Username, &user.OpenIDIdentity, &user.Name, &user.Email, &user.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUser fetches a single user from database
func (db *DB) GetUser(userID int) (User, error) {
	user := User{}

	err := db.QueryRow(getUserSQL, userID).Scan(&user.ID, &user.Username, &user.OpenIDIdentity, &user.Name, &user.Email, &user.Role)
	if err != nil {
		log.Printf("Error retrieving user %v: %v\n", userID, err)
		return user, err
	}

	return user, nil
}

// GetUserByUsername fetches a single user from database
func (db *DB) GetUserByUsername(username string) (User, error) {
	user := User{}

	err := db.QueryRow(getUserSQL, username).Scan(&user.ID, &user.Username, &user.OpenIDIdentity, &user.Name, &user.Email, &user.Role)
	if err != nil {
		log.Printf("Error retrieving user %v: %v\n", username, err)
		return user, err
	}

	return user, nil
}

// CreateUser adds a new record to User database table, Returns new record identifier if success
func (db *DB) CreateUser(user User) (int, error) {
	var createdUserID int
	err := db.QueryRow(createUserSQL, user.Username, user.OpenIDIdentity, user.Name, user.Email, user.Role).Scan(&createdUserID)
	if err != nil {
		log.Printf("Error creating user %v: %v\n", user.Username, err)
		return 0, err
	}
	return createdUserID, nil
}

// UpdateUser sets user new values for an existing record.
func (db *DB) UpdateUser(user User) error {
	_, err := db.Exec(updateUserSQL, user.Username, user.OpenIDIdentity, user.Name, user.Email, user.Role, user.ID)
	if err != nil {
		log.Printf("Error updating database user %v: %v\n", user.ID, err)
		return err
	}

	return nil
}

// DeleteUser deletes a user
func (db *DB) DeleteUser(userID int) error {
	_, err := db.Exec(deleteUserSQL, userID)
	if err != nil {
		log.Printf("Error deleting database user %v: %v\n", userID, err)
		return err
	}

	return nil
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
		err := rows.Scan(&user.ID, &user.Username, &user.OpenIDIdentity, &user.Name, &user.Email, &user.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func rowsToUsers(rows *sql.Rows) ([]User, error) {
	users := []User{}

	for rows.Next() {
		user := User{}
		err := rows.Scan(&user.ID, &user.Username, &user.OpenIDIdentity, &user.Name, &user.Email, &user.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// RoleForUser checks the user role against the database
// If user authentication is turned off, the role defaults to Admin
func (db *DB) RoleForUser(username string) int {
	if username == "" {
		return Admin
	}

	user, err := db.GetUserByUsername(username)
	if err != nil {
		return 0
	}
	return user.Role
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
