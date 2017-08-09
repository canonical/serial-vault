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
	"errors"
	"fmt"
	"strings"
	"time"
)

// MockDB holds the successful mocks for the database
type MockDB struct {
	encryptedAuthKeyHash string
}

// CreateModelTable mock for the create model table method
func (mdb *MockDB) CreateModelTable() error {
	return nil
}

// AlterModelTable mock for the alter model table method
func (mdb *MockDB) AlterModelTable() error {
	return nil
}

// CreateKeypairTable mock for the create keypair table method
func (mdb *MockDB) CreateKeypairTable() error {
	return nil
}

// AlterKeypairTable mock for the alter keypair table method
func (mdb *MockDB) AlterKeypairTable() error {
	return nil
}

// UpdateKeypairAssertion mock to update the account-key assertion of a keypair
func (mdb *MockDB) UpdateKeypairAssertion(keypairID int, assertion string) error {
	return nil
}

// CreateSettingsTable mock for the create settings table method
func (mdb *MockDB) CreateSettingsTable() error {
	return nil
}

// CreateAccountTable mock for the create account table method
func (mdb *MockDB) CreateAccountTable() error {
	return nil
}

// GetAccount mock to return a single account key
func (mdb *MockDB) GetAccount(authorityID string) (Account, error) {
	accounts, _ := mdb.ListAccounts("")

	for _, acc := range accounts {
		if acc.AuthorityID == authorityID {
			return acc, nil
		}
	}
	return Account{}, errors.New("Cannot found the account assertion")
}

// ListAccounts mock to return a list of the available accounts
func (mdb *MockDB) ListAccounts(username string) ([]Account, error) {
	var accounts []Account
	accounts = append(accounts, Account{ID: 1, AuthorityID: "System", Assertion: "assertion\n"})
	return accounts, nil
}

// PutAccount mock to update abn account assertion
func (mdb *MockDB) PutAccount(account Account) (string, error) {
	return "", nil
}

// UpdateAccountAssertion mock to update the account assertion
func (mdb *MockDB) UpdateAccountAssertion(authorityID, assertion string) error {
	return nil
}

// ListAllowedModels Mock the database response for a list of models
func (mdb *MockDB) ListAllowedModels(authorization User) ([]Model, error) {

	var models []Model
	if authorization.Username == "" || authorization.Username == "sv" {
		models = append(models, Model{ID: 1, BrandID: "Vendor", Name: "alder", KeypairID: 1, AuthorityID: "System", KeyID: "61abf588e52be7a3", SealedKey: "", KeyActive: true, KeypairIDUser: 1, AuthorityIDUser: "System", KeyIDUser: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", SealedKeyUser: "", KeyActiveUser: true})
		models = append(models, Model{ID: 2, BrandID: "Vendor", Name: "ash", KeypairID: 1, AuthorityID: "System", KeyID: "61abf588e52be7a3", SealedKey: "", KeyActive: false})
		models = append(models, Model{ID: 3, BrandID: "Vendor", Name: "basswood", KeypairID: 1, AuthorityID: "System", KeyID: "61abf588e52be7a3", SealedKey: "", KeyActive: true})
	}
	if authorization.Username == "" {
		models = append(models, Model{ID: 4, BrandID: "Vendor", Name: "korina", KeypairID: 1, AuthorityID: "System", KeyID: "61abf588e52be7a3", SealedKey: "", KeyActive: true})
		models = append(models, Model{ID: 5, BrandID: "Vendor", Name: "mahogany", KeypairID: 1, AuthorityID: "System", KeyID: "61abf588e52be7a3", SealedKey: "", KeyActive: true})
		models = append(models, Model{ID: 6, BrandID: "Vendor", Name: "maple", KeypairID: 1, AuthorityID: "System", KeyID: "61abf588e52be7a3", SealedKey: "", KeyActive: true})
	}
	return models, nil
}

// FindModel mocks the database response for finding a model
func (mdb *MockDB) FindModel(brandID, modelName, apiKey string) (Model, error) {
	model := Model{ID: 1, BrandID: "System", Name: "alder", KeypairID: 1, AuthorityID: "System", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", KeyActive: true, SealedKey: ""}
	if modelName == "inactive" {
		model = Model{ID: 1, BrandID: "System", Name: "inactive", KeypairID: 1, AuthorityID: "System", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", KeyActive: false, SealedKey: ""}
	}
	if model.BrandID != brandID || model.Name != modelName {
		return model, errors.New("Cannot find a model for that brand and model")
	}
	if apiKey == "NoModelForApiKey" {
		return model, errors.New("Cannot find a model for that brand and model for the API key")
	}
	return model, nil
}

// CheckAPIKey mocks the database response to check the API key
func (mdb *MockDB) CheckAPIKey(apiKey string) bool {
	if apiKey == "InvalidAPIKey" {
		return false
	}
	return true
}

// GetAllowedModel mocks the model from the database by ID.
func (mdb *MockDB) GetAllowedModel(modelID int, authorization User) (Model, error) {

	var model Model
	found := false
	models, _ := mdb.ListAllowedModels(authorization)

	for _, mdl := range models {
		if mdl.ID == modelID {
			model = mdl
			found = true
			break
		}
	}

	if !found {
		return model, errors.New("Cannot find the model")
	}

	return model, nil
}

// UpdateAllowedModel mocks the model update.
func (mdb *MockDB) UpdateAllowedModel(model Model, authorization User) (string, error) {
	models, _ := mdb.ListAllowedModels(authorization)
	found := false

	for _, mdl := range models {
		if mdl.ID == model.ID {
			found = true
			break
		}
	}

	if !found {
		return "", errors.New("Cannot find the model")
	}
	return "", nil
}

// DeleteAllowedModel mocks the model deletion.
func (mdb *MockDB) DeleteAllowedModel(model Model, authorization User) (string, error) {
	models, _ := mdb.ListAllowedModels(authorization)
	found := false

	for _, mdl := range models {
		if mdl.ID == model.ID {
			found = true
			break
		}
	}

	if !found {
		return "", errors.New("Cannot find the model")
	}
	return "", nil
}

// CreateAllowedModel mocks creating a new model.
func (mdb *MockDB) CreateAllowedModel(model Model, authorization User) (Model, string, error) {
	model = Model{ID: 7, BrandID: "System", Name: "the-model", KeypairID: 1, AuthorityID: "system", KeyID: "61abf588e52be7a3"}

	return model, "", nil
}

// GetKeypair mocks getting a model by ID
func (mdb *MockDB) GetKeypair(keypairID int) (Keypair, error) {
	keypair := Keypair{ID: 1, AuthorityID: "system", KeyID: "61abf588e52be7a3", Active: true}
	return keypair, nil
}

// ListAllowedKeypairs mocks listing the keypairs
func (mdb *MockDB) ListAllowedKeypairs(authorization User) ([]Keypair, error) {
	var keypairs []Keypair
	if authorization.Username == "" || authorization.Username == "sv" {
		keypairs = append(keypairs, Keypair{ID: 1, AuthorityID: "system", KeyID: "61abf588e52be7a3", Active: true})
		keypairs = append(keypairs, Keypair{ID: 2, AuthorityID: "system", KeyID: "invalidone", Active: true})
	}
	if authorization.Username == "" {
		keypairs = append(keypairs, Keypair{ID: 3, AuthorityID: "systemone", KeyID: "61abf588e52be7a3", Active: true})
		keypairs = append(keypairs, Keypair{ID: 3, AuthorityID: "system", KeyID: "inactiveone", Active: false})
	}
	return keypairs, nil
}

// PutKeypair database mock
func (mdb *MockDB) PutKeypair(keypair Keypair) (string, error) {
	return "", nil
}

// UpdateAllowedKeypairActive database mock
func (mdb *MockDB) UpdateAllowedKeypairActive(keypairID int, active bool, authorization User) error {
	return nil
}

// GetSetting database mock
func (mdb *MockDB) GetSetting(code string) (Setting, error) {
	switch code {
	case "System/12345678abcdef":
		// Returning the encrypted, base64 encoded HMAC-ed auth-key: fake-hmac-ed-data
		return Setting{Code: "System/12345678abcdef", Data: "pmXt1iwvM5P947KATp24rMQFHEnAf2tUXGl1XXyfhDhf"}, nil

	case "System/abcdef12345678":
		return Setting{Code: "System/abcdef12345678", Data: mdb.encryptedAuthKeyHash}, nil

	case "do-not-find":
		return Setting{}, errors.New("Cannot find 'do-not-find'")

	default:
		return Setting{Code: code, Data: code}, nil
	}
}

// PutSetting database mock
func (mdb *MockDB) PutSetting(setting Setting) error {
	if setting.Code == "System/abcdef12345678" {
		mdb.encryptedAuthKeyHash = setting.Data
	}
	return nil
}

// CreateSigningLogTable database mock
func (mdb *MockDB) CreateSigningLogTable() error {
	return nil
}

// CheckForDuplicate database mock
func (mdb *MockDB) CheckForDuplicate(signLog *SigningLog) (bool, int, error) {
	switch signLog.SerialNumber {
	case "Aduplicate":
		return true, 3, nil
	case "AnError":
		return false, 0, errors.New("Error in check for duplicate")
	}
	return false, 0, nil
}

// CreateSigningLog database mock
func (mdb *MockDB) CreateSigningLog(signLog SigningLog) error {
	if signLog.SerialNumber == "AsigninglogError" {
		return errors.New("Error in check for create signing log entry")
	}
	return nil
}

// ListSigningLog database mock
func (mdb *MockDB) ListSigningLog(username string) ([]SigningLog, error) {
	var fromID = 11
	signingLog := []SigningLog{}

	if len(username) > 0 {
		fromID = 5
	}

	for i := 1; i < fromID; i++ {
		signingLog = append(signingLog, SigningLog{ID: i, Make: "System", Model: "Router 3400", SerialNumber: fmt.Sprintf("A%d", i), Fingerprint: fmt.Sprintf("a%d", i), Created: time.Now()})
	}
	return signingLog, nil
}

// SigningLogFilterValues database mock
func (mdb *MockDB) SigningLogFilterValues(username string) (SigningLogFilters, error) {
	return SigningLogFilters{Makes: []string{"System"}, Models: []string{"Router 3400"}}, nil
}

// CreateDeviceNonceTable database mock
func (mdb *MockDB) CreateDeviceNonceTable() error {
	return nil
}

// DeleteExpiredDeviceNonces database mock
func (mdb *MockDB) DeleteExpiredDeviceNonces() error {
	return nil
}

// CreateDeviceNonce database mock
func (mdb *MockDB) CreateDeviceNonce() (DeviceNonce, error) {
	return DeviceNonce{Nonce: "1234567890", TimeStamp: 1234567890}, nil
}

// ValidateDeviceNonce database mock
func (mdb *MockDB) ValidateDeviceNonce(nonce string) error {
	return nil
}

// CreateOpenidNonceTable database mock
func (mdb *MockDB) CreateOpenidNonceTable() error {
	return nil
}

// CreateOpenidNonce database mock
func (mdb *MockDB) CreateOpenidNonce(nonce OpenidNonce) error {
	return nil
}

// CheckUserInAccount verifies that a user has permissions to a specific account
func (mdb *MockDB) CheckUserInAccount(username, authorityID string) bool {
	return true
}

// RoleForUser fetches the user's permissions
func (mdb *MockDB) RoleForUser(username string) int {
	user, _ := mdb.GetUserByUsername(username)
	return user.Role
}

// CreateUserTable mock for creating database User table operation
func (mdb *MockDB) CreateUserTable() error {
	return nil
}

// CreateAccountUserLinkTable mock for creating database AccountUserLink table operation
func (mdb *MockDB) CreateAccountUserLinkTable() error {
	return nil
}

// AlterUserTable mock for modifications on User table operation
func (mdb *MockDB) AlterUserTable() error {
	return nil
}

// CreateUser mock for create user operation
func (mdb *MockDB) CreateUser(user User) (int, error) {
	return 740, nil
}

// ListUsers mock returning a fixed list of users
func (mdb *MockDB) ListUsers() ([]User, error) {
	var users []User
	users = append(users, User{
		ID:       1,
		Username: "user1",
		Name:     "Rigoberto Picaporte",
		Email:    "rigoberto.picaporte@ubuntu.com",
		Role:     Standard,
		Accounts: []Account{
			Account{
				ID:          1,
				AuthorityID: "authority1",
				Assertion:   "assertioncontent1",
			},
		}})
	users = append(users, User{
		ID:       2,
		Username: "user2",
		Name:     "Nancy Reagan",
		Email:    "nancy.reagan@usa.gov",
		Role:     Standard,
		Accounts: []Account{
			Account{
				ID:          2,
				AuthorityID: "authority2",
				Assertion:   "assertioncontent2",
			},
		}})
	users = append(users, User{
		ID:       3,
		Username: "sv",
		Name:     "Steven Vault",
		Email:    "sv@example.com",
		Role:     Admin,
		Accounts: []Account{
			Account{
				ID:          3,
				AuthorityID: "authority3",
				Assertion:   "assertioncontent3",
			},
			Account{
				ID:          4,
				AuthorityID: "authority4",
				Assertion:   "assertioncontent4",
			},
		}})
	users = append(users, User{
		ID:       4,
		Username: "a",
		Name:     "A",
		Email:    "a@example.com",
		Role:     Standard})
	users = append(users, User{
		ID:       5,
		Username: "root",
		Name:     "Root User",
		Email:    "the_root_user@thisdb.com",
		Role:     Superuser})
	return users, nil
}

// FindUsers mock trying to find a user in a fixed list of users
func (mdb *MockDB) FindUsers(query string) ([]User, error) {
	users, _ := mdb.ListUsers()
	returnArray := make([]User, 2)

	for _, u := range users {
		if strings.Contains(u.Username, query) || strings.Contains(u.Email, query) {
			returnArray = append(returnArray, u)
		}
	}
	return returnArray, nil
}

// GetUser mock returning the user if found by id in a fixed list of users
func (mdb *MockDB) GetUser(userID int) (User, error) {
	users, _ := mdb.ListUsers()
	for _, u := range users {
		if u.ID == userID {
			return u, nil
		}
	}
	return User{}, errors.New("Cannot find the user")
}

// GetUserByUsername mock returning the user if found by username in a fixed list of users
func (mdb *MockDB) GetUserByUsername(username string) (User, error) {
	users, _ := mdb.ListUsers()
	for _, u := range users {
		if u.Username == username {
			return u, nil
		}
	}
	return User{}, errors.New("Cannot find the user")
}

// UpdateUser mock for update user operation. Returns error if user not found in a fixed list of users
func (mdb *MockDB) UpdateUser(user User) error {
	_, err := mdb.GetUser(user.ID)
	return err
}

// DeleteUser mock for delete user operation. Returns error if user not found in a fixed list of users
func (mdb *MockDB) DeleteUser(userID int) error {
	_, err := mdb.GetUser(userID)
	return err
}

// ListUserAccounts mock returning a fixed list of accounts
func (mdb *MockDB) ListUserAccounts(username string) ([]Account, error) {
	var accounts []Account
	accounts = append(accounts, Account{ID: 1, AuthorityID: "System", Assertion: "assertion\n"})
	return accounts, nil
}

// ListNotUserAccounts mock returning a fixed list of accounts
func (mdb *MockDB) ListNotUserAccounts(username string) ([]Account, error) {
	var accounts []Account
	accounts = append(accounts, Account{ID: 2, AuthorityID: "Other", Assertion: "other assertion\n"})
	return accounts, nil
}

// ListAccountUsers mock returning a fixed list of users
func (mdb *MockDB) ListAccountUsers(authorityID string) ([]User, error) {
	return mdb.ListUsers()
}

// ErrorMockDB holds the unsuccessful mocks for the database
type ErrorMockDB struct{}

// CreateModelTable mock for the create model table method
func (mdb *ErrorMockDB) CreateModelTable() error {
	return errors.New("Error creating the model table")
}

// AlterModelTable mock for the alter model table method
func (mdb *ErrorMockDB) AlterModelTable() error {
	return nil
}

// CreateKeypairTable mock for the create keypair table method
func (mdb *ErrorMockDB) CreateKeypairTable() error {
	return nil
}

// AlterKeypairTable mock for the alter keypair table method
func (mdb *ErrorMockDB) AlterKeypairTable() error {
	return nil
}

// UpdateKeypairAssertion mock to update the account-key assertion of a keypair
func (mdb *ErrorMockDB) UpdateKeypairAssertion(keypairID int, assertion string) error {
	return errors.New("MOCK Error updating the keypair assertion")
}

// CreateSettingsTable mock for the create settings table method
func (mdb *ErrorMockDB) CreateSettingsTable() error {
	return nil
}

// CreateAccountTable mock for the create account table method
func (mdb *ErrorMockDB) CreateAccountTable() error {
	return nil
}

// GetAccount mock to return a single account key
func (mdb *ErrorMockDB) GetAccount(authorityID string) (Account, error) {

	accounts, _ := mdb.ListAccounts("")

	for _, acc := range accounts {
		if acc.AuthorityID == authorityID {
			return acc, nil
		}
	}
	return Account{}, errors.New("Cannot found the account assertion")
}

// ListAccounts mock to return a list of the available accounts
func (mdb *ErrorMockDB) ListAccounts(username string) ([]Account, error) {
	return nil, errors.New("Error getting the accounts")
}

// PutAccount mock to update abn account assertion
func (mdb *ErrorMockDB) PutAccount(account Account) (string, error) {
	return "", errors.New("MOCK error upserting the account")
}

// UpdateAccountAssertion mock to update the account assertion
func (mdb *ErrorMockDB) UpdateAccountAssertion(authorityID, assertion string) error {
	return nil
}

// ListAllowedModels ModelsList Mock the database response for a list of models
func (mdb *ErrorMockDB) ListAllowedModels(authorization User) ([]Model, error) {
	return nil, errors.New("Error getting the models")
}

// FindModel mocks the database response for finding a model, returning an invalid signing-key
func (mdb *ErrorMockDB) FindModel(brandID, modelName, apiKey string) (Model, error) {
	return Model{}, errors.New("Error finding the model")
}

// CheckAPIKey mocks the database response to check the API key
func (mdb *ErrorMockDB) CheckAPIKey(apiKey string) bool {
	return true
}

// GetAllowedModel mocks the model from the database by ID, returning an error.
func (mdb *ErrorMockDB) GetAllowedModel(modelID int, authorization User) (Model, error) {
	return Model{}, errors.New("Error retrieving the model")
}

// UpdateAllowedModel mocks the model update, returning an error.
func (mdb *ErrorMockDB) UpdateAllowedModel(model Model, authorization User) (string, error) {
	return "", errors.New("Error updating the database model")
}

// DeleteAllowedModel mocks the model deletion, returning an error.
func (mdb *ErrorMockDB) DeleteAllowedModel(model Model, authorization User) (string, error) {
	return "", errors.New("Error deleting the database model")
}

// CreateAllowedModel mocks creating a new model, returning an error.
func (mdb *ErrorMockDB) CreateAllowedModel(model Model, authorization User) (Model, string, error) {
	return Model{}, "", errors.New("Error creating the database model")
}

// GetKeypair error mock for the database
func (mdb *ErrorMockDB) GetKeypair(keypairID int) (Keypair, error) {
	keypair := Keypair{AuthorityID: "system", KeyID: "61abf588e52be7a3", Active: true}
	return keypair, errors.New("Error fetching from the database")
}

// ListAllowedKeypairs error mock for the database
func (mdb *ErrorMockDB) ListAllowedKeypairs(authorization User) ([]Keypair, error) {
	var keypairs []Keypair
	return keypairs, errors.New("MOCK Error fetching from the database")
}

// PutKeypair error mock for the database
func (mdb *ErrorMockDB) PutKeypair(keypair Keypair) (string, error) {
	return "", errors.New("Error updating the database")
}

// UpdateAllowedKeypairActive error mock for the database
func (mdb *ErrorMockDB) UpdateAllowedKeypairActive(keypairID int, active bool, authorization User) error {
	return errors.New("Error updating the database")
}

// GetSetting error mock for the database
func (mdb *ErrorMockDB) GetSetting(code string) (Setting, error) {
	return Setting{Code: code, Data: code}, nil
}

// PutSetting error mock for the database
func (mdb *ErrorMockDB) PutSetting(setting Setting) error {
	return nil
}

// CheckForDuplicate error mock for the database
func (mdb *ErrorMockDB) CheckForDuplicate(signLog *SigningLog) (bool, int, error) {
	return false, 0, nil
}

// CreateSigningLog error mock for the database
func (mdb *ErrorMockDB) CreateSigningLog(signLog SigningLog) error {
	return nil
}

// CreateSigningLogTable error mock for the database
func (mdb *ErrorMockDB) CreateSigningLogTable() error {
	return nil
}

// ListSigningLog error mock for the database
func (mdb *ErrorMockDB) ListSigningLog(username string) ([]SigningLog, error) {
	var signingLog []SigningLog
	return signingLog, errors.New("Error retrieving the signing logs")
}

// SigningLogFilterValues error mock for the database
func (mdb *ErrorMockDB) SigningLogFilterValues(username string) (SigningLogFilters, error) {
	return SigningLogFilters{}, errors.New("Error retrieving the signing log filters")
}

// CreateDeviceNonceTable error mock for the database
func (mdb *ErrorMockDB) CreateDeviceNonceTable() error {
	return nil
}

// DeleteExpiredDeviceNonces error mock for the database
func (mdb *ErrorMockDB) DeleteExpiredDeviceNonces() error {
	return nil
}

// CreateDeviceNonce error mock for the database
func (mdb *ErrorMockDB) CreateDeviceNonce() (DeviceNonce, error) {
	return DeviceNonce{}, errors.New("MOCK error generating the nonce")
}

// ValidateDeviceNonce error mock for the database
func (mdb *ErrorMockDB) ValidateDeviceNonce(nonce string) error {
	return errors.New("MOCK error validating a nonce")
}

// CreateOpenidNonceTable database mock
func (mdb *ErrorMockDB) CreateOpenidNonceTable() error {
	return nil
}

// CreateOpenidNonce database mock
func (mdb *ErrorMockDB) CreateOpenidNonce(nonce OpenidNonce) error {
	return errors.New("MOCK error generating the nonce")
}

// CheckUserInAccount verifies that a user has permissions to a specific account
func (mdb *ErrorMockDB) CheckUserInAccount(username, authorityID string) bool {
	return true
}

// RoleForUser fetches the user's permissions
func (mdb *ErrorMockDB) RoleForUser(username string) int {
	// in case username is 'root' return a valid value, as that is used for auth
	if username == "root" {
		return Superuser
	}
	return 0
}

// CreateUserTable mock for creating database User table operation
func (mdb *ErrorMockDB) CreateUserTable() error {
	return errors.New("Could not create User table")
}

// CreateAccountUserLinkTable mock for creating database AccountUserLink table operation
func (mdb *ErrorMockDB) CreateAccountUserLinkTable() error {
	return errors.New("Could not create AccountUserLink table")
}

// AlterUserTable mock for modifications on User table operation
func (mdb *ErrorMockDB) AlterUserTable() error {
	return errors.New("Could not alter User table")
}

// CreateUser error mock for create user operation
func (mdb *ErrorMockDB) CreateUser(user User) (int, error) {
	return 0, errors.New("Cannot create user")
}

// ListUsers mock returning an error for list users operation
func (mdb *ErrorMockDB) ListUsers() ([]User, error) {
	return []User{}, errors.New("Could not retrieve users list")
}

// FindUsers mock returning an error for find users operation
func (mdb *ErrorMockDB) FindUsers(query string) ([]User, error) {
	return []User{}, errors.New("Could not find any user")
}

// GetUser mock returning an error for get user operation
func (mdb *ErrorMockDB) GetUser(userID int) (User, error) {
	return User{}, errors.New("Cannot get the user")
}

// GetUserByUsername returns error for get user by username operation
func (mdb *ErrorMockDB) GetUserByUsername(username string) (User, error) {
	return User{}, errors.New("Cannot get the user")
}

// UpdateUser mock returning an error for update user operation
func (mdb *ErrorMockDB) UpdateUser(user User) error {
	return errors.New("Cannot update the user")
}

// DeleteUser mock returning an error for delete user operation
func (mdb *ErrorMockDB) DeleteUser(userID int) error {
	return errors.New("Cannot delete the user")
}

// ListUserAccounts mock returning an error for list user accounts operation
func (mdb *ErrorMockDB) ListUserAccounts(username string) ([]Account, error) {
	return []Account{}, errors.New("Could not get accounts for that user")
}

// ListNotUserAccounts mock returning an error for list non-user accounts operation
func (mdb *ErrorMockDB) ListNotUserAccounts(username string) ([]Account, error) {
	return []Account{}, errors.New("Could not get accounts not related to that user")
}

// ListAccountUsers mock returning an error for list account users operation
func (mdb *ErrorMockDB) ListAccountUsers(authorityID string) ([]User, error) {
	return []User{}, errors.New("Could not get any user for that account")
}
