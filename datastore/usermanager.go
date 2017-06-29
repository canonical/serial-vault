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
	CREATE TABLE IF NOT EXISTS user (
		id               serial primary key not null,
		username         citext not null unique,
		openid_token     text not null,
		name             varchar(200),
		email            citext not null,
		role             int not null
	)
`

const createAccountUserLinkTableSQL = `
	CREATE TABLE IF NOT EXISTS useraccountlink (
		user_id          int references user not null,
		account_id     	 int references account not null
	)
`

const listUsersSQL = "select id, username, openid_token, name, email, role from user order by username"
const getUserSQL = "select id, username, openid_token, name, email, role from user where username=$1"
const findUsersSQL = "select id, username, openid_token, name, email, role from user where username like '%$1%' or name like '%$1%"
const createUserSQL = "insert into user (username, openid_token, name, email, role) values ($1,$2,$3,$4,$5)"
const updateUserSQL = "update user set username=$1, openid_token=$2, name=$3, email=$4, role=$5 where username=$6"
const deleteUserSQL = "delete from username where username=$1"

const listAccountUsersSQL = `
	select id, username, openid_token, name, email, role
	from user u
	inner join accountuserlink l on u.id = l.user_id
	inner join accounts a on l.account_id = a.id
	where a.authority_id=$1
`

// User holds user personal, authentication and authorization info
type User struct {
	ID          int
	Username    string
	OpenIDToken string
	Name        string
	Email       string
	Role        int
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
		err := rows.Scan(&user.ID, &user.Username, &user.OpenIDToken, &user.Name, &user.Email, &user.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUser fetches a single user from the database by the username
func (db *DB) GetUser(username string) (User, error) {
	user := User{}

	err := db.QueryRow(getUserSQL, username).Scan(&user.ID, &user.Username, &user.OpenIDToken, &user.Name, &user.Email, &user.Role)
	if err != nil {
		log.Printf("Error retrieving user %v: %v\n", username, err)
		return user, err
	}

	return user, nil
}

// CreateUser adds a new record to User database table
func (db *DB) CreateUser(user User) error {
	_, err := db.Exec(createUserSQL, user.Username, user.OpenIDToken, user.Name, user.Email, user.Role)
	if err != nil {
		log.Printf("Error creating user %v: %v\n", user.Username, err)
		return err
	}

	return nil
}

// UpdateUser sets user new values for an existing record.
func (db *DB) UpdateUser(username string, user User) error {
	_, err := db.Exec(updateUserSQL, user.Username, user.OpenIDToken, user.Name, user.Email, user.Role, username)
	if err != nil {
		log.Printf("Error updating database user %v: %v\n", username, err)
		return err
	}

	return nil
}

// DeleteUser deletes user matching username param
func (db *DB) DeleteUser(username string) error {
	_, err := db.Exec(deleteUserSQL, username)
	if err != nil {
		log.Printf("Error deleting database user %v: %v\n", username, err)
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
		err := rows.Scan(&user.ID, &user.Username, &user.OpenIDToken, &user.Name, &user.Email, &user.Role)
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
		err := rows.Scan(&user.ID, &user.Username, &user.OpenIDToken, &user.Name, &user.Email, &user.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
