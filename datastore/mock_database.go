// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2018 Canonical Ltd
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

	"github.com/snapcore/snapd/asserts"
)

// MockDB holds the successful mocks for the database
type MockDB struct {
	encryptedAuthKeyHash string
}

// CreateSignedModelAssertTable mock for the create signed model table method
func (mdb *MockDB) CreateSignedModelAssertTable() error {
	return nil
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
func (mdb *MockDB) UpdateKeypairAssertion(keypair Keypair, authorization User) (string, error) {
	return "", nil
}

// CreateSettingsTable mock for the create settings table method
func (mdb *MockDB) CreateSettingsTable() error {
	return nil
}

// CreateAccountTable mock for the create account table method
func (mdb *MockDB) CreateAccountTable() error {
	return nil
}

// AlterAccountTable mock for the create account table method
func (mdb *MockDB) AlterAccountTable() error {
	return nil
}

// CreateAccount mock to create an account record
func (mdb *MockDB) CreateAccount(account Account) error {
	return nil
}

// GetAccount mock to return a single account key
func (mdb *MockDB) GetAccount(authorityID string) (Account, error) {
	accounts, _ := mdb.ListAllowedAccounts(User{})

	for _, acc := range accounts {
		if acc.AuthorityID == authorityID {
			return acc, nil
		}
	}
	return Account{}, errors.New("Cannot found the account assertion")
}

// GetAccountByID mock to return a single account key
func (mdb *MockDB) GetAccountByID(ID int, user User) (Account, error) {
	accounts, _ := mdb.ListAllowedAccounts(user)

	for _, acc := range accounts {
		if acc.ID == ID {
			return acc, nil
		}
	}
	return Account{}, errors.New("Cannot found the account assertion")
}

// GetAllowedAccount mock to fetch account
func (mdb *MockDB) GetAllowedAccount(authorityID string, authorization User) (Account, error) {
	return mdb.GetAccount(authorityID)
}

// ListAllowedAccounts mock to return a list of the available accounts
func (mdb *MockDB) ListAllowedAccounts(authorization User) ([]Account, error) {
	var accounts []Account
	accounts = append(accounts, Account{ID: 1, AuthorityID: "system", Assertion: "assertion\n", ResellerAPI: true})
	accounts = append(accounts, Account{ID: 2, AuthorityID: "vendor", Assertion: "assertion\n", ResellerAPI: false})
	accounts = append(accounts, Account{ID: 3, AuthorityID: "generic", Assertion: "assertion\n", ResellerAPI: true})
	return accounts, nil
}

// PutAccount mock to update abn account assertion
func (mdb *MockDB) PutAccount(account Account, authorization User) (string, error) {
	return "", nil
}

// UpdateAccountAssertion mock to update the account assertion
func (mdb *MockDB) UpdateAccountAssertion(authorityID, assertion string, resellerAPI bool) error {
	return nil
}

// UpdateAccount mock to update the account
func (mdb *MockDB) UpdateAccount(account Account, authorization User) error {
	return nil
}

// ListAllowedModels Mock the database response for a list of models
func (mdb *MockDB) ListAllowedModels(authorization User) ([]Model, error) {

	var models []Model
	if authorization.Username == "" || authorization.Username == "sv" || authorization.Username == "sync" {
		models = append(models, Model{ID: 1, BrandID: "system", Name: "alder", KeypairID: 1, AuthorityID: "system", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", SealedKey: "", KeyActive: true, KeypairIDUser: 1, AuthorityIDUser: "system", KeyIDUser: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", SealedKeyUser: "", KeyActiveUser: true})
		models = append(models, Model{ID: 2, BrandID: "system", Name: "ash", KeypairID: 1, AuthorityID: "system", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", SealedKey: "", KeyActive: false})
		models = append(models, Model{ID: 3, BrandID: "system", Name: "basswood", KeypairID: 1, AuthorityID: "system", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", SealedKey: "", KeyActive: true})
	}
	if authorization.Username == "" {
		models = append(models, Model{ID: 4, BrandID: "system", Name: "korina", KeypairID: 1, AuthorityID: "system", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", SealedKey: "", KeyActive: true})
		models = append(models, Model{ID: 5, BrandID: "system", Name: "mahogany", KeypairID: 1, AuthorityID: "system", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", SealedKey: "", KeyActive: true})
		models = append(models, Model{ID: 6, BrandID: "system", Name: "maple", KeypairID: 1, AuthorityID: "system", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", SealedKey: "", KeyActive: true})
	}
	return models, nil
}

// SyncAccount mock to update the account
func (mdb *MockDB) SyncAccount(account Account) error {
	return nil
}

// FindModel mocks the database response for finding a model
func (mdb *MockDB) FindModel(brandID, modelName, apiKey string) (Model, error) {
	model := Model{ID: 1, BrandID: "system", Name: "alder", KeypairID: 1, AuthorityID: "system", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", KeyActive: true, SealedKey: ""}
	if modelName == "ash" {
		model = Model{ID: 2, BrandID: "system", Name: "ash", KeypairID: 1, AuthorityID: "system", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", KeyActive: true, SealedKey: ""}
	}
	if modelName == "alder-mybrand" {
		model = Model{ID: 1, BrandID: "mybrand", Name: "alder-mybrand", KeypairID: 1, AuthorityID: "mybrand", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", KeyActive: true, SealedKey: ""}
	}
	if modelName == "alder-mybrand-2" {
		model = Model{ID: 2, BrandID: "mybrand", Name: "alder-mybrand-2", KeypairID: 1, AuthorityID: "mybrand", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", KeyActive: true, SealedKey: ""}
	}
	if modelName == "generic-classic" {
		model = Model{ID: 1, BrandID: "generic", Name: "generic-classic", KeypairID: 1, AuthorityID: "generic", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", KeyActive: true, SealedKey: ""}
	}
	if modelName == "inactive" {
		model = Model{ID: 1, BrandID: "system", Name: "inactive", KeypairID: 1, AuthorityID: "system", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", KeyActive: false, SealedKey: ""}
	}
	if model.BrandID != brandID || model.Name != modelName || modelName == "invalid" {
		return model, errors.New("Cannot find a model for that brand and model")
	}
	if apiKey == "NoModelForApiKey" {
		return model, errors.New("Cannot find a model for that brand and model for the API key")
	}
	return model, nil
}

// CheckModelExists mocks the database response for finding a model
func (mdb *MockDB) CheckModelExists(brandID, modelName string) bool {
	model := Model{ID: 1, BrandID: "system", Name: "alder", KeypairID: 1, AuthorityID: "system", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", KeyActive: true, SealedKey: ""}
	if model.BrandID != brandID || model.Name != modelName || modelName == "invalid" {
		return false
	}
	return true
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

	if model.ID == 1 && model.Name == "ash" {
		return "error-model-exists", errors.New("A device with the same Brand and Model already exists")
	}

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

func keypairSystem() Keypair {
	return Keypair{ID: 1, AuthorityID: "system", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", Active: true}
}

// CreateAllowedModel mocks creating a new model.
func (mdb *MockDB) CreateAllowedModel(model Model, authorization User) (Model, string, error) {
	model = Model{ID: 7, BrandID: model.BrandID, Name: model.Name, KeypairID: model.KeypairID, AuthorityID: "system", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO"}

	return model, "", nil
}

// SyncModel mocks creating a new model
func (mdb *MockDB) SyncModel(model Model) error {
	return nil
}

// GetKeypair mocks getting a keypair by ID
func (mdb *MockDB) GetKeypair(keypairID int) (Keypair, error) {
	keypair := keypairSystem()
	return keypair, nil
}

// GetKeypairByPublicID mocks getting a keypair by key ID
func (mdb *MockDB) GetKeypairByPublicID(auth, keyID string) (Keypair, error) {
	keypair := keypairSystem()
	return keypair, nil
}

// GetKeypairByName mocks getting a keypair by name
func (mdb *MockDB) GetKeypairByName(authorityID, keyName string) (Keypair, error) {
	keypair := keypairSystem()
	return keypair, nil
}

// ListAllowedKeypairs mocks listing the keypairs
func (mdb *MockDB) ListAllowedKeypairs(authorization User) ([]Keypair, error) {
	var keypairs []Keypair
	if authorization.Username == "" || authorization.Username == "sv" || authorization.Username == "sync" {
		keypairs = append(keypairs, keypairSystem())
		keypairs = append(keypairs, Keypair{ID: 2, AuthorityID: "system", KeyID: "invalidone", Active: true})
	}
	if authorization.Username == "" {
		keypairs = append(keypairs, Keypair{ID: 3, AuthorityID: "systemone", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", Active: true})
		keypairs = append(keypairs, Keypair{ID: 3, AuthorityID: "system", KeyID: "inactiveone", Active: false})
	}
	return keypairs, nil
}

// PutKeypair database mock
func (mdb *MockDB) PutKeypair(keypair Keypair) (string, error) {
	return "", nil
}

// CheckKeypairKeynameExists mock checking for keypair by name
func (mdb *MockDB) CheckKeypairKeynameExists(authorityID, name string) bool {
	if name == "invalid" {
		return true
	}
	return false
}

// SyncKeypair database mock
func (mdb *MockDB) SyncKeypair(keypair SyncKeypair) error {
	return nil
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

	case "system/UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO":
		return Setting{Code: "system/UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", Data: mdb.encryptedAuthKeyHash}, nil

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

// CreateTestLogTable error mock for the database
func (mdb *MockDB) CreateTestLogTable() error {
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

// CheckForMatching database mock
func (mdb *MockDB) CheckForMatching(signLog SigningLog) (bool, error) {
	switch signLog.SerialNumber {
	case "Aduplicate":
		return true, nil
	case "AnError":
		return false, errors.New("Error in check for duplicate")
	}
	return false, nil
}

// CreateSigningLog database mock
func (mdb *MockDB) CreateSigningLog(signLog SigningLog) error {
	if signLog.SerialNumber == "AsigninglogError" {
		return errors.New("Error in check for create signing log entry")
	}
	return nil
}

// CreateSigningLogSync database mock
func (mdb *MockDB) CreateSigningLogSync(signLog SigningLog) error {
	if signLog.SerialNumber == "AsigninglogError" {
		return errors.New("Error in check for create signing log entry")
	}
	return nil
}

// ListAllowedSigningLog database mock
func (mdb *MockDB) ListAllowedSigningLog(authorization User) ([]SigningLog, error) {
	var fromID = 11
	signingLog := []SigningLog{}

	if len(authorization.Username) > 0 {
		fromID = 5
	}

	for i := 1; i < fromID; i++ {
		signingLog = append(signingLog, SigningLog{ID: i, Make: "System", Model: "Router 3400", SerialNumber: fmt.Sprintf("A%d", i), Fingerprint: fmt.Sprintf("a%d", i), Created: time.Now()})
	}
	return signingLog, nil
}

// ListAllowedSigningLogForAccount database mock
func (mdb *MockDB) ListAllowedSigningLogForAccount(authorization User, authorityID string, params *SigningLogParams) ([]SigningLog, error) {
	return mdb.ListAllowedSigningLog(authorization)
}

// SyncSigningLog database mock
func (mdb *MockDB) SyncSigningLog() ([]SigningLog, error) {
	signingLog := []SigningLog{}
	for i := 1; i < 5; i++ {
		signingLog = append(signingLog, SigningLog{ID: i, Make: "system", Model: "alder", SerialNumber: fmt.Sprintf("A%d", i), Fingerprint: fmt.Sprintf("a%d", i), Created: time.Now()})
	}
	return signingLog, nil
}

// SyncUpdateSigningLog database mock
func (mdb *MockDB) SyncUpdateSigningLog(id int) error {
	return nil
}

// AllowedSigningLogFilterValues database mock
func (mdb *MockDB) AllowedSigningLogFilterValues(authorization User, authorityID string) (SigningLogFilters, error) {
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

// CreateUserTable §
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
			{
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
			{
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
			{
				ID:          3,
				AuthorityID: "authority3",
				Assertion:   "assertioncontent3",
			},
			{
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
	users = append(users, User{
		ID:       6,
		Username: "sync",
		Name:     "Sync Vault",
		Email:    "sync@example.com",
		Role:     SyncUser,
		Accounts: []Account{
			{
				ID:          3,
				AuthorityID: "authority3",
				Assertion:   "assertioncontent3",
			},
			{
				ID:          4,
				AuthorityID: "authority4",
				Assertion:   "assertioncontent4",
			},
		}})

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

// GetUserByAPIKey mock returning the user if found by username in a fixed list of users
func (mdb *MockDB) GetUserByAPIKey(apiKey, username string) (User, error) {
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

// CreateKeypairStatusTable mock the database creation
func (mdb *MockDB) CreateKeypairStatusTable() error {
	return nil
}

// AlterKeypairStatusTable mock the database creation
func (mdb *MockDB) AlterKeypairStatusTable() error {
	return nil
}

// CreateKeypairStatus mocks the creation of a keypair status record
func (mdb *MockDB) CreateKeypairStatus(ks KeypairStatus) (int, error) {
	return 4, nil
}

// UpdateKeypairStatus mocks the update of a keypair status record
func (mdb *MockDB) UpdateKeypairStatus(ks KeypairStatus) error {
	return nil
}

// ListAllowedKeypairStatus lists the keypair statuses
func (mdb *MockDB) ListAllowedKeypairStatus(authorization User) ([]KeypairStatus, error) {
	ks := []KeypairStatus{}
	ks = append(ks, KeypairStatus{ID: 1, AuthorityID: "system", KeyName: "key1", Status: KeypairStatusCreating})
	ks = append(ks, KeypairStatus{ID: 2, AuthorityID: "system", KeyName: "key2", Status: KeypairStatusEncrypting})
	ks = append(ks, KeypairStatus{ID: 3, AuthorityID: "system", KeyName: "key3", Status: KeypairStatusExporting})
	return ks, nil
}

// GetKeypairStatus fetches a single keypair status record
func (mdb *MockDB) GetKeypairStatus(authorityID, keyName string) (KeypairStatus, error) {

	kss, _ := mdb.ListAllowedKeypairStatus(User{})
	for _, ks := range kss {
		if ks.AuthorityID == authorityID && ks.KeyName == keyName {
			return ks, nil
		}
	}
	return KeypairStatus{}, errors.New("Cannot find the keypair status")
}

// DeleteKeypairStatus removes a keypair status
func (mdb *MockDB) DeleteKeypairStatus(ks KeypairStatus) error {
	return nil
}

// CreateModelAssertTable mock for creating database model assertion headers table
func (mdb *MockDB) CreateModelAssertTable() error {
	return nil
}

// CreateModelAssert mock for creating model assertion record
func (mdb *MockDB) CreateModelAssert(m ModelAssertion) (int, error) {
	return 1, nil
}

// AlterModelAssertTable mock for the alter model assertion table method
func (mdb *MockDB) AlterModelAssertTable() error {
	return nil
}

// UpdateModelAssert mock for updating model assertion record
func (mdb *MockDB) UpdateModelAssert(m ModelAssertion) error {
	return nil
}

// UpsertSignedModelAssert mock for updating signe model assertion record
func (mdb *MockDB) UpsertSignedModelAssert(modelID int, revision int, assertion asserts.Assertion) error {
	return nil
}

// GetSignedModelAssert mock for getting signe model assertion record
func (mdb *MockDB) GetSignedModelAssert(modelID int) (asserts.Assertion, error) {
	headers := map[string]interface{}{
		"type":              "model",
		"authority-id":      "canonical",
		"brand-id":          "canonical",
		"series":            "16",
		"model":             "test",
		"store":             "ubuntu",
		"architecture":      "amd64",
		"gadget":            "pc",
		"kernel":            "pc-kernel",
		"validation":        "certified",
		"timestamp":         "2016-04-01T00:00:00.0Z",
		"sign-key-sha3-384": "-CvQKAwRQ5h3Ffn10FILJoEZUXOv6km9FwA80-Rcj-f-6jadQ89VRswHNiEB9Lxk",
	}
	signature := []byte(`AcLDXAQAAQoABgUCV7UYzwAKCRDUpVvql9g3IK7uH/4udqNOurx5WYVknzXdwekp0ovHCQJ0iBPw
TSFxEVr9faZSzb7eqJ1WicHsShf97PYS3ClRYAiluFsjRA8Y03kkSVJHjC+sIwGFubsnkmgflt6D
WEmYIl0UBmeaEDS8uY4Xvp9NsLTzNEj2kvzy/52gKaTc1ZSl5RDL9ppMav+0V9iBYpiDPBWH2rJ+
aDSD8Rkyygm0UscfAKyDKH4lrvZ0WkYyi1YVNPrjQ/AtBySh6Q4iJ3LifzKa9woIyAuJET/4/FPY
oirqHAfuvNod36yNQIyNqEc20AvTvZNH0PSsg4rq3DLjIPzv5KbJO9lhsasNJK1OdL6x8Yqrdsbk
ldZp4qkzfjV7VOMQKaadfcZPRaVVeJWOBnBiaukzkhoNlQi1sdCdkBB/AJHZF8QXw6c7vPDcfnCV
1lW7ddQ2p8IsJbT6LzpJu3GW/P4xhNgCjtCJ1AJm9a9RqLwQYgdLZwwDa9iCRtqTbRXBlfy3apps
1VjbQ3h5iCd0hNfwDBnGVm1rhLKHCD1DUdNE43oN2ZlE7XGyh0HFV6vKlpqoW3eoXCIxWu+HBY96
+LSl/jQgCkb0nxYyzEYK4Reb31D0mYw1Nji5W+MIF5E09+DYZoOT0UvR05YMwMEOeSdI/hLWg/5P
k+GDK+/KopMmpd4D1+jjtF7ZvqDpmAV98jJGB2F88RyVb4gcjmFFyTi4Kv6vzz/oLpbm0qrizC0W
HLGDN/ymGA5sHzEgEx7U540vz/q9VX60FKqL2YZr/DcyY9GKX5kCG4sNqIIHbcJneZ4frM99oVDu
7Jv+DIx/Di6D1ULXol2XjxbbJLKHFtHksR97ceaFvcZwTogC61IYUBJCvvMoqdXAWMhEXCr0QfQ5
Xbi31XW2d4/lF/zWlAkRnGTzufIXFni7+nEuOK0SQEzO3/WaRedK1SGOOtTDjB8/3OJeW96AUYK5
oTIynkYkEyHWMNCXALg+WQW6L4/YO7aUjZ97zOWIugd7Xy63aT3r/EHafqaY2nacOhLfkeKZ830b
o/ezjoZQAxbh6ce7JnXRgE9ELxjdAhBTpGjmmmN2sYrJ7zP9bOgly0BnEPXGSQfFA+NNNw1FADx1
MUY8q9DBjmVtgqY+1KGTV5X8KvQCBMODZIf/XJPHdCRAHxMd8COypcwgL2vDIIXpOFbi1J/B0GF+
eklxk9wzBA8AecBMCwCzIRHDNpD1oa2we38bVFrOug6e/VId1k1jYFJjiLyLCDmV8IMYwEllHSXp
LQAdm3xZ7t4WnxYC8YSCk9mXf3CZg59SpmnV5Q5Z6A5Pl7Nc3sj7hcsMBZEsOMPzNC9dPsBnZvjs
WpPUffJzEdhHBFhvYMuD4Vqj6ejUv9l3oTrjQWVC`)
	return asserts.Assemble(headers, nil, nil, signature)
}

// GetModelAssert mock for updating model assertion record
func (mdb *MockDB) GetModelAssert(modelID int) (ModelAssertion, error) {
	if modelID == 2 {
		return ModelAssertion{
			ID:            1,
			ModelID:       1,
			KeypairID:     1,
			Series:        16,
			Classic:       "TRue",
			Base:          "core18",
			DisplayName:   "Classic Device",
			Architecture:  "amd64",
			Revision:      1,
			Gadget:        "pc",
			Kernel:        "pc-kernel",
			Store:         "ubuntu",
			Created:       time.Now().UTC(),
			Modified:      time.Now().UTC(),
			RequiredSnaps: "snapweb,juju",
		}, nil
	}

	return ModelAssertion{
		ID:            1,
		ModelID:       1,
		KeypairID:     1,
		Series:        16,
		Base:          "core18",
		Architecture:  "amd64",
		Revision:      1,
		Gadget:        "pc",
		Kernel:        "pc-kernel",
		Store:         "ubuntu",
		Created:       time.Now().UTC(),
		Modified:      time.Now().UTC(),
		RequiredSnaps: "snapweb,juju",
	}, nil
}

// UpsertModelAssert mock for creating or updating model assertion record
func (mdb *MockDB) UpsertModelAssert(m ModelAssertion) error {
	return nil
}

// CreateSubstoreTable mock for the create substore table method
func (mdb *MockDB) CreateSubstoreTable() error {
	return nil
}

// CreateAllowedSubstore mock to create a substore record
func (mdb *MockDB) CreateAllowedSubstore(store Substore, authorization User) (Substore, error) {
	substore := Substore{ID: 1, AccountID: store.AccountID, FromModelID: store.FromModelID, Store: store.Store, SerialNumber: store.SerialNumber, ModelName: store.ModelName}

	return substore, nil
}

// ListSubstores mock to list substore records
func (mdb *MockDB) ListSubstores(accountID int, authorization User) ([]Substore, error) {
	fromModel, _ := mdb.GetAllowedModel(1, authorization)

	substores := []Substore{
		{ID: 1, AccountID: 1, FromModelID: fromModel.ID, FromModel: fromModel, Store: "mybrand", SerialNumber: "abc1234", ModelName: "alder-mybrand"},
		{ID: 2, AccountID: 1, FromModelID: fromModel.ID, FromModel: fromModel, Store: "mybrand", SerialNumber: "abc5678", ModelName: "alder-mybrand"},
	}

	return substores, nil
}

// UpdateAllowedSubstore mock to update a substore record
func (mdb *MockDB) UpdateAllowedSubstore(store Substore, authorization User) error {
	return nil
}

// DeleteAllowedSubstore mock to update a substore record
func (mdb *MockDB) DeleteAllowedSubstore(storeID int, authorization User) (string, error) {
	return "", nil
}

// GetSubstore mock to get a substore record
func (mdb *MockDB) GetSubstore(fromModelID int, serialNumber string) (Substore, error) {
	if serialNumber == "XXX" {
		return Substore{}, fmt.Errorf("error retrieving database substore for %s, from model %d", serialNumber, fromModelID)
	}

	fromModel := Model{ID: 1, BrandID: "generic", Name: "generic-classic", KeypairID: 1, AuthorityID: "generic", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", KeyActive: true, KeyActiveUser: true, AuthorityIDUser: "generic", KeyIDUser: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO"}

	if serialNumber == "abc1234X" {
		return Substore{ID: 2, AccountID: 1, FromModelID: 1, FromModel: fromModel, Store: "mybrand", SerialNumber: "abc1234X", ModelName: "alder-mybrand-2"}, nil
	}

	return Substore{ID: 1, AccountID: 1, FromModelID: 1, FromModel: fromModel, Store: "mybrand", SerialNumber: "abc1234", ModelName: "alder-mybrand"}, nil
}

// GetSubstoreModel mock to get a substore record
func (mdb *MockDB) GetSubstoreModel(brand, model, serialNumber string) (Substore, error) {
	if model == "invalid" {
		return Substore{}, errors.New("MOCK invalid model")
	}
	return mdb.GetSubstore(1, serialNumber)
}

// GetAllowedSubstore mock to get a substore record
func (mdb *MockDB) GetAllowedSubstore(fromModelID int, serialNumber string, authorization User) (Substore, error) {
	return mdb.GetSubstore(fromModelID, serialNumber)
}

// CreateTestLog mock to create a test log
func (mdb *MockDB) CreateTestLog(testLog TestLog) error {
	return nil
}

// ListAllowedTestLog database mock
func (mdb *MockDB) ListAllowedTestLog(authorization User) ([]TestLog, error) {
	logs := []TestLog{
		{ID: 1, Brand: "system", Model: "alder", Filename: "test1.xml"},
		{ID: 2, Brand: "system", Model: "alder", Filename: "test2.xml"},
	}
	return logs, nil
}

// SyncListTestLogs database mock
func (mdb *MockDB) SyncListTestLogs() ([]TestLog, error) {
	logs := []TestLog{
		{ID: 1, Brand: "system", Model: "alder", Filename: "test1.xml"},
		{ID: 2, Brand: "system", Model: "alder", Filename: "test2.xml"},
	}
	return logs, nil
}

// SyncDeleteTestLog database mock
func (mdb *MockDB) SyncDeleteTestLog(ID int) error {
	if ID > 2 {
		return errors.New("MOCK error deleting the test log")
	}
	return nil
}

// UpdateAllowedTestLog database mock
func (mdb *MockDB) UpdateAllowedTestLog(ID int, authorization User) error {
	if authorization.Role >= SyncUser {
		return nil
	}
	return errors.New("MOCK no permissions to update the test log")
}

// HealthCheck mock for a healthy datastore
func (mdb *MockDB) HealthCheck() error {
	return nil
}

// -----------------------------------------------------------------------------

// ErrorMockDB holds the unsuccessful mocks for the database
type ErrorMockDB struct{}

// UpsertSignedModelAssert mock an error
func (mdb *ErrorMockDB) UpsertSignedModelAssert(modelID int, revision int, assertion asserts.Assertion) error {
	return errors.New("error")
}

// GetSignedModelAssert mock an error
func (mdb *ErrorMockDB) GetSignedModelAssert(modelID int) (asserts.Assertion, error) {
	return nil, errors.New("error")
}

// CreateSignedModelAssertTable mock for the create model table method
func (mdb *ErrorMockDB) CreateSignedModelAssertTable() error {
	return errors.New("Error creating the signed model assert table")
}

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
func (mdb *ErrorMockDB) UpdateKeypairAssertion(keypair Keypair, authorization User) (string, error) {
	return "invalid-assertion", errors.New("MOCK Error updating the keypair assertion")
}

// CreateSettingsTable mock for the create settings table method
func (mdb *ErrorMockDB) CreateSettingsTable() error {
	return nil
}

// CreateAccountTable mock for the create account table method
func (mdb *ErrorMockDB) CreateAccountTable() error {
	return nil
}

// AlterAccountTable mock for the create account table method
func (mdb *ErrorMockDB) AlterAccountTable() error {
	return nil
}

// CreateAccount mock to create an account record
func (mdb *ErrorMockDB) CreateAccount(account Account) error {
	return errors.New("MOCK creating the account")
}

// GetAccount mock to return a single account key
func (mdb *ErrorMockDB) GetAccount(authorityID string) (Account, error) {

	accounts, _ := mdb.ListAllowedAccounts(User{})

	for _, acc := range accounts {
		if acc.AuthorityID == authorityID {
			return acc, nil
		}
	}
	return Account{}, errors.New("Cannot found the account assertion")
}

// GetAccountByID mock to return a single account key
func (mdb *ErrorMockDB) GetAccountByID(ID int, user User) (Account, error) {

	accounts, _ := mdb.ListAllowedAccounts(user)

	for _, acc := range accounts {
		if acc.ID == ID {
			return acc, nil
		}
	}
	return Account{}, errors.New("Cannot found the account assertion")
}

// GetAllowedAccount mock to fetch account
func (mdb *ErrorMockDB) GetAllowedAccount(authorityID string, authorization User) (Account, error) {
	return Account{}, errors.New("MOCK error getting the account")
}

// ListAllowedAccounts mock to return a list of the available accounts
func (mdb *ErrorMockDB) ListAllowedAccounts(authorization User) ([]Account, error) {
	return nil, errors.New("Error getting the accounts")
}

// PutAccount mock to update an account assertion
func (mdb *ErrorMockDB) PutAccount(account Account, authorization User) (string, error) {
	return "", errors.New("MOCK error upserting the account")
}

// SyncAccount mock to update the account
func (mdb *ErrorMockDB) SyncAccount(account Account) error {
	return errors.New("MOCK error syncing the account")
}

// UpdateAccountAssertion mock to update the account assertion
func (mdb *ErrorMockDB) UpdateAccountAssertion(authorityID, assertion string, resellerAPI bool) error {
	return nil
}

// UpdateAccount mock to update the account
func (mdb *ErrorMockDB) UpdateAccount(account Account, authorization User) error {
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

// CheckModelExists mocks the database response for finding a model
func (mdb *ErrorMockDB) CheckModelExists(brandID, modelName string) bool {
	return false
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

// SyncModel mocks creating a new model, returning an error.
func (mdb *ErrorMockDB) SyncModel(model Model) error {
	return errors.New("Error creating the database model")
}

// GetKeypair error mock for the database
func (mdb *ErrorMockDB) GetKeypair(keypairID int) (Keypair, error) {
	keypair := Keypair{AuthorityID: "system", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", Active: true}
	return keypair, errors.New("Error fetching from the database")
}

// GetKeypairByPublicID error mock for the database
func (mdb *ErrorMockDB) GetKeypairByPublicID(auth, keyID string) (Keypair, error) {
	keypair := Keypair{AuthorityID: "system", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", Active: true}
	return keypair, errors.New("Error fetching from the database")
}

// GetKeypairByName mocks getting a keypair by name
func (mdb *ErrorMockDB) GetKeypairByName(authorityID, keyName string) (Keypair, error) {
	keypair := Keypair{ID: 1, AuthorityID: "system", KeyID: "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO", Active: true}
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

// CheckKeypairKeynameExists mock checking for keypair by name
func (mdb *ErrorMockDB) CheckKeypairKeynameExists(authorityID, name string) bool {
	if name == "invalid" {
		return true
	}
	return false
}

// SyncKeypair error mock for the database
func (mdb *ErrorMockDB) SyncKeypair(keypair SyncKeypair) error {
	return errors.New("Error updating the database")
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

// CheckForMatching error mock for the database
func (mdb *ErrorMockDB) CheckForMatching(signLog SigningLog) (bool, error) {
	return false, nil
}

// CreateSigningLog error mock for the database
func (mdb *ErrorMockDB) CreateSigningLog(signLog SigningLog) error {
	return nil
}

// CreateSigningLogSync error mock for the database
func (mdb *ErrorMockDB) CreateSigningLogSync(signLog SigningLog) error {
	return nil
}

// CreateSigningLogTable error mock for the database
func (mdb *ErrorMockDB) CreateSigningLogTable() error {
	return nil
}

// CreateTestLogTable error mock for the database
func (mdb *ErrorMockDB) CreateTestLogTable() error {
	return nil
}

// ListAllowedSigningLog error mock for the database
func (mdb *ErrorMockDB) ListAllowedSigningLog(authorization User) ([]SigningLog, error) {
	var signingLog []SigningLog
	return signingLog, errors.New("Error retrieving the signing logs")
}

// ListAllowedSigningLogForAccount database mock
func (mdb *ErrorMockDB) ListAllowedSigningLogForAccount(authorization User, authorityID string, params *SigningLogParams) ([]SigningLog, error) {
	return mdb.ListAllowedSigningLog(authorization)
}

// SyncSigningLog error mock for the database
func (mdb *ErrorMockDB) SyncSigningLog() ([]SigningLog, error) {
	var signingLog []SigningLog
	return signingLog, errors.New("Error retrieving the signing logs")
}

// SyncUpdateSigningLog database mock
func (mdb *ErrorMockDB) SyncUpdateSigningLog(id int) error {
	return errors.New("Error updating the signing log")
}

// AllowedSigningLogFilterValues error mock for the database
func (mdb *ErrorMockDB) AllowedSigningLogFilterValues(authorization User, authorityID string) (SigningLogFilters, error) {
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

// GetUserByAPIKey returns error for get user by username operation
func (mdb *ErrorMockDB) GetUserByAPIKey(apiKey, username string) (User, error) {
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

// CreateKeypairStatusTable mock the database creation with error
func (mdb *ErrorMockDB) CreateKeypairStatusTable() error {
	return errors.New("Could not create the Keypair Status table")
}

// AlterKeypairStatusTable mock the database creation
func (mdb *ErrorMockDB) AlterKeypairStatusTable() error {
	return errors.New("Could not update the Keypair Status table")
}

// CreateKeypairStatus mocks the creation of a keypair status record
func (mdb *ErrorMockDB) CreateKeypairStatus(ks KeypairStatus) (int, error) {
	return 0, errors.New("Cannot create keypair status record")
}

// UpdateKeypairStatus mocks the update of a keypair status record
func (mdb *ErrorMockDB) UpdateKeypairStatus(ks KeypairStatus) error {
	return errors.New("Cannot update keypair status record")
}

// ListAllowedKeypairStatus lists the keypair statuses
func (mdb *ErrorMockDB) ListAllowedKeypairStatus(authorization User) ([]KeypairStatus, error) {
	return []KeypairStatus{}, errors.New("Cannot update keypair status record")
}

// GetKeypairStatus fetches a single keypair status record
func (mdb *ErrorMockDB) GetKeypairStatus(authorityID, keyName string) (KeypairStatus, error) {
	return KeypairStatus{}, errors.New("Cannot find the keypair status")
}

// DeleteKeypairStatus removes a keypair status
func (mdb *ErrorMockDB) DeleteKeypairStatus(ks KeypairStatus) error {
	return errors.New("Cannot delete the keypair status")
}

// CreateModelAssertTable mock for creating database model assertion headers table
func (mdb *ErrorMockDB) CreateModelAssertTable() error {
	return errors.New("Cannot create the model assertion table")
}

// CreateModelAssert mock for creating model assertion record
func (mdb *ErrorMockDB) CreateModelAssert(m ModelAssertion) (int, error) {
	return 0, errors.New("Cannot create the model assertion record")
}

// AlterModelAssertTable mock for the alter model assertion table method
func (mdb *ErrorMockDB) AlterModelAssertTable() error {
	return errors.New("Cannot amend the model assertion table")
}

// UpdateModelAssert mock for updating model assertion record
func (mdb *ErrorMockDB) UpdateModelAssert(m ModelAssertion) error {
	return errors.New("Cannot update the model assertion record")
}

// GetModelAssert mock for updating model assertion record
func (mdb *ErrorMockDB) GetModelAssert(modelID int) (ModelAssertion, error) {
	return ModelAssertion{}, errors.New("Cannot find the model assertion record")
}

// UpsertModelAssert mock for creating or updating model assertion record
func (mdb *ErrorMockDB) UpsertModelAssert(m ModelAssertion) error {
	return errors.New("Cannot upsert the model assertion record")
}

// CreateSubstoreTable mock for the create substore table method
func (mdb *ErrorMockDB) CreateSubstoreTable() error {
	return nil
}

// CreateAllowedSubstore mock to create a substore record
func (mdb *ErrorMockDB) CreateAllowedSubstore(store Substore, authorization User) (Substore, error) {
	return store, errors.New("Cannot create the sub-store model")
}

// ListSubstores mock to list substore records
func (mdb *ErrorMockDB) ListSubstores(accountID int, authorization User) ([]Substore, error) {
	fromModel, _ := mdb.GetAllowedModel(1, authorization)

	substores := []Substore{
		{ID: 1, FromModelID: fromModel.ID, FromModel: fromModel, Store: "mybrand", SerialNumber: "abc1234", ModelName: "alder-mybrand"},
		{ID: 1, FromModelID: fromModel.ID, FromModel: fromModel, Store: "mybrand", SerialNumber: "abc5678", ModelName: "alder-mybrand"},
	}

	return substores, errors.New("Cannot list the sub-stores")
}

// UpdateAllowedSubstore mock to update a substore record
func (mdb *ErrorMockDB) UpdateAllowedSubstore(store Substore, authorization User) error {
	return errors.New("Cannot update the sub-store model")
}

// DeleteAllowedSubstore mock to update a substore record
func (mdb *ErrorMockDB) DeleteAllowedSubstore(storeID int, authorization User) (string, error) {
	return "", errors.New("Cannot delete the sub-store model")
}

// GetSubstore mock to get a substore record
func (mdb *ErrorMockDB) GetSubstore(fromModelID int, serialNumber string) (Substore, error) {
	return Substore{}, errors.New("Cannot get the sub-store model")
}

// GetSubstoreModel mock to get a substore record
func (mdb *ErrorMockDB) GetSubstoreModel(brand, model, serialNumber string) (Substore, error) {
	return Substore{}, errors.New("Cannot get the sub-store model")
}

// GetAllowedSubstore mock to get a substore record
func (mdb *ErrorMockDB) GetAllowedSubstore(fromModelID int, serialNumber string, authorization User) (Substore, error) {
	return mdb.GetSubstore(fromModelID, serialNumber)
}

// CreateTestLog mock to create a test log
func (mdb *ErrorMockDB) CreateTestLog(testLog TestLog) error {
	return errors.New("MOCK Cannot create the test log")
}

// ListAllowedTestLog database mock
func (mdb *ErrorMockDB) ListAllowedTestLog(authorization User) ([]TestLog, error) {
	return nil, errors.New("MOCK Cannot fetch the test logs")
}

// SyncListTestLogs database mock
func (mdb *ErrorMockDB) SyncListTestLogs() ([]TestLog, error) {
	return nil, errors.New("MOCK Cannot fetch the test logs")
}

// SyncDeleteTestLog database mock
func (mdb *ErrorMockDB) SyncDeleteTestLog(ID int) error {
	return errors.New("MOCK error deleting the test log")
}

// UpdateAllowedTestLog database mock
func (mdb *ErrorMockDB) UpdateAllowedTestLog(ID int, authorization User) error {
	return errors.New("MOCK error updating the test log")
}

// HealthCheck mock to simulate failed HealthCheck
func (mdb *ErrorMockDB) HealthCheck() error {
	return errors.New("Health check failed")
}
