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
	"log"

	"github.com/CanonicalLtd/serial-vault/config"
	_ "github.com/lib/pq" // postgresql driver
)

// Datastore interface for the database logic
type Datastore interface {
	ListModels(username string) ([]Model, error)
	FindModel(brandID, modelName string) (Model, error)
	GetModel(modelID int, username string) (Model, error)
	UpdateModel(model Model, username string) (string, error)
	DeleteModel(model Model, username string) (string, error)
	CreateModel(model Model, username string) (Model, string, error)
	CreateModelTable() error
	AlterModelTable() error

	ListKeypairs(username string) ([]Keypair, error)
	GetKeypair(keypairID int) (Keypair, error)
	PutKeypair(keypair Keypair) (string, error)
	UpdateKeypairActive(keypairID int, active bool, username string) error
	UpdateKeypairAssertion(keypairID int, assertion string) error
	CreateKeypairTable() error
	AlterKeypairTable() error

	CreateSettingsTable() error
	PutSetting(setting Setting) error
	GetSetting(code string) (Setting, error)

	CreateSigningLogTable() error
	CheckForDuplicate(signLog *SigningLog) (bool, int, error)
	CreateSigningLog(signLog SigningLog) error
	ListSigningLog(username string) ([]SigningLog, error)
	SigningLogFilterValues(username string) (SigningLogFilters, error)

	CreateDeviceNonceTable() error
	DeleteExpiredDeviceNonces() error
	CreateDeviceNonce() (DeviceNonce, error)
	ValidateDeviceNonce(nonce string) error

	CreateAccountTable() error
	ListAccounts(username string) ([]Account, error)
	GetAccount(authorityID string) (Account, error)
	UpdateAccountAssertion(authorityID, assertion string) error
	PutAccount(account Account) (string, error)

	CreateOpenidNonceTable() error
	CreateOpenidNonce(nonce OpenidNonce) error

	CreateUser(user User) error
	ListUsers() ([]User, error)
	FindUsers(query string) ([]User, error)
	GetUser(username string) (User, error)
	UpdateUser(username string, user User) error
	DeleteUser(username string) error
	CreateUserTable() error
	CreateAccountUserLinkTable() error
	CheckUserInAccount(username, authorityID string) bool
	RoleForUser(username string) int

	ListUserAccounts(username string) ([]Account, error)
	ListAccountUsers(authorityID string) ([]User, error)
}

// DB local database interface with our custom methods.
type DB struct {
	*sql.DB
}

// Env Environment struct that holds the config and data store details.
type Env struct {
	Config    config.Settings
	DB        Datastore
	KeypairDB *KeypairDatabase
}

// Environ contains the parsed config file settings.
var Environ *Env

// OpenidNonceStore contains the database nonce store for Openid
var OpenidNonceStore PgNonceStore

// OpenSysDatabase Return an open database connection
func OpenSysDatabase(driver, dataSource string) {
	// Open the database connection
	db, err := sql.Open(driver, dataSource)
	if err != nil {
		log.Fatalf("Error opening the database: %v\n", err)
	}

	// Check that we have a valid database connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error accessing the database: %v\n", err)
	} else {
		log.Println("Database opened successfully.")
	}

	Environ.DB = &DB{db}
	OpenidNonceStore.DB = &DB{db}
}
