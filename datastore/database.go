// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2018 Canonical Ltd
 * License granted by Canonical Limited
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

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/snapcore/snapd/asserts"
)

const anyUserFilter = ""

// SigningLogParams holds extra parameters for the SigningLog search
type SigningLogParams struct {
	Limit        uint64 // 0 means no LIMIT here
	Offset       uint64
	Filter       []string
	Serialnumber string
}

// Datastore interface for the database logic
type Datastore interface {
	ListAllowedModels(authorization User) ([]Model, error)
	FindModel(brandID, modelName, apiKey string) (Model, error)
	GetAllowedModel(modelID int, authorization User) (Model, error)
	UpdateAllowedModel(model Model, authorization User) (string, error)
	DeleteAllowedModel(model Model, authorization User) (string, error)
	CreateAllowedModel(model Model, authorization User) (Model, string, error)
	CreateModelTable() error
	AlterModelTable() error
	CheckAPIKey(apiKey string) bool
	CheckModelExists(brandID, name string) bool

	CreateModelAssertTable() error
	AlterModelAssertTable() error
	CreateModelAssert(m ModelAssertion) (int, error)
	UpdateModelAssert(m ModelAssertion) error
	GetModelAssert(modelID int) (ModelAssertion, error)
	UpsertModelAssert(m ModelAssertion) error

	CreateSignedModelAssertTable() error
	UpsertSignedModelAssert(modelID int, revision int, assertion asserts.Assertion) error
	GetSignedModelAssert(modelID int) (asserts.Assertion, error)

	ListAllowedKeypairs(authorization User) ([]Keypair, error)
	GetKeypair(keypairID int) (Keypair, error)
	GetKeypairByPublicID(authorityID, keyID string) (Keypair, error)
	GetKeypairByName(authorityID, keyName string) (Keypair, error)
	PutKeypair(keypair Keypair) (string, error)
	UpdateAllowedKeypairActive(keypairID int, active bool, authorization User) error
	UpdateKeypairAssertion(keypair Keypair, authorization User) (string, error)
	CreateKeypairTable() error
	AlterKeypairTable() error
	CheckKeypairKeynameExists(authorityID, name string) bool

	CreateSettingsTable() error
	PutSetting(setting Setting) error
	GetSetting(code string) (Setting, error)

	CreateSigningLogTable() error
	CheckForDuplicate(signLog *SigningLog) (bool, int, error)
	CreateSigningLog(signLog SigningLog) error
	ListAllowedSigningLog(authorization User) ([]SigningLog, error)
	ListAllowedSigningLogForAccount(authorization User, authorityID string, params *SigningLogParams) ([]SigningLog, error)
	AllowedSigningLogFilterValues(authorization User, authorityID string) (SigningLogFilters, error)

	CreateDeviceNonceTable() error
	DeleteExpiredDeviceNonces() error
	CreateDeviceNonce() (DeviceNonce, error)
	ValidateDeviceNonce(nonce string) error

	CreateAccountTable() error
	AlterAccountTable() error
	ListAllowedAccounts(authorization User) ([]Account, error)
	GetAllowedAccount(authorityID string, authorization User) (Account, error)
	GetAccount(authorityID string) (Account, error)
	GetAccountByID(accountID int, authorization User) (Account, error)
	CreateAccount(account Account) error
	UpdateAccount(account Account, authorization User) error
	PutAccount(account Account, authorization User) (string, error)

	CreateOpenidNonceTable() error
	CreateOpenidNonce(nonce OpenidNonce) error

	CreateUser(user User) (int, error)
	ListUsers() ([]User, error)
	FindUsers(query string) ([]User, error)
	GetUser(userID int) (User, error)
	GetUserByUsername(username string) (User, error)
	GetUserByAPIKey(apiKey, username string) (User, error)
	UpdateUser(user User) error
	DeleteUser(userID int) error
	CreateUserTable() error
	CreateAccountUserLinkTable() error
	CheckUserInAccount(username, authorityID string) bool
	AlterUserTable() error

	ListUserAccounts(username string) ([]Account, error)
	ListNotUserAccounts(username string) ([]Account, error)
	ListAccountUsers(authorityID string) ([]User, error)

	CreateKeypairStatusTable() error
	AlterKeypairStatusTable() error
	CreateKeypairStatus(ks KeypairStatus) (int, error)
	UpdateKeypairStatus(ks KeypairStatus) error
	DeleteKeypairStatus(ks KeypairStatus) error
	GetKeypairStatus(authorityID, keyName string) (KeypairStatus, error)
	ListAllowedKeypairStatus(authorization User) ([]KeypairStatus, error)

	CreateSubstoreTable() error
	CreateAllowedSubstore(store Substore, authorization User) (Substore, error)
	ListSubstores(accountID int, authorization User) ([]Substore, error)
	UpdateAllowedSubstore(store Substore, authorization User) error
	DeleteAllowedSubstore(storeID int, authorization User) (string, error)
	GetAllowedSubstore(fromModelID int, serialNumber string, authorization User) (Substore, error)
	GetSubstore(fromModelID int, serialNumber string) (Substore, error)
	GetSubstoreModel(brand, model, serialNumber string) (Substore, error)

	CreateTestLogTable() error
	CreateTestLog(testLog TestLog) error
	ListAllowedTestLog(authorization User) ([]TestLog, error)

	HealthCheck() error

	SyncAccount(account Account) error
	SyncKeypair(keypair SyncKeypair) error
	SyncModel(m Model) error
	CheckForMatching(signLog SigningLog) (bool, error)
	CreateSigningLogSync(signLog SigningLog) error
	SyncSigningLog() ([]SigningLog, error)
	SyncUpdateSigningLog(id int) error
	SyncListTestLogs() ([]TestLog, error)
	SyncDeleteTestLog(ID int) error
	UpdateAllowedTestLog(ID int, authorization User) error
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

// OpenSysDatabase return an open database connection
func OpenSysDatabase(driver, dataSource string) {
	// Open the database connection
	if driver == "sqlite3" {
		openSQLiteDatabase(driver, dataSource)
	} else {
		openPostgreSQLDatabase(driver, dataSource)
	}
}

func (db *DB) transaction(txFunc func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	err = txFunc(tx)
	return err
}

// InFactory checks if we are running in the factory (with a sqlite database)
func InFactory() bool {
	if Environ.Config.Driver == "sqlite3" {
		return true
	}
	return false
}
