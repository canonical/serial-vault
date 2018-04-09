// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
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

package sync_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/account"
	"github.com/CanonicalLtd/serial-vault/service/keypair"
	"github.com/CanonicalLtd/serial-vault/service/model"
	"github.com/CanonicalLtd/serial-vault/sync"
	check "gopkg.in/check.v1"
)

func (s *startSuite) TestStartUnit(c *check.C) {
	tests := []suiteTest{
		{
			Args:         []string{"account"},
			ErrorMessage: ""},
		{
			Args:         []string{"account"},
			ErrorMessage: "MOCK error fetching accounts",
			MockErrorDB:  true},
		{
			Args:         []string{"account"},
			ErrorMessage: "MOCK fail fetching accounts",
			MockFail:     true},
		{
			Args:         []string{"signingkey"},
			ErrorMessage: ""},
		{
			Args:         []string{"signingkey"},
			ErrorMessage: "MOCK error fetching signing keys",
			MockErrorDB:  true},
		{
			Args:         []string{"signingkey"},
			ErrorMessage: "Error fetching signing keys",
			MockFail:     true},
		{
			Args:         []string{"model"},
			ErrorMessage: ""},
		{
			Args:         []string{"model"},
			ErrorMessage: "MOCK error fetching models",
			MockErrorDB:  true},
		{
			Args:         []string{"model"},
			ErrorMessage: "MOCK fail fetching models",
			MockFail:     true},
		{
			Args:         []string{"signinglog"},
			ErrorMessage: ""},
		{
			Args:         []string{"signinglog"},
			ErrorMessage: "Error retrieving the signing logs",
			MockErrorDB:  true},
	}

	for _, t := range tests {
		var err error

		if t.MockErrorDB {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
			sync.FetchAccounts = mockFetchAccountsError
			sync.FetchSigningKeys = mockFetchSigningKeysError
			sync.FetchModels = mockFetchModelsError
			sync.SendSigningLog = mockSendSigningLogError
		}
		if t.MockFail {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
			sync.FetchAccounts = mockFetchAccountsFail
			sync.FetchSigningKeys = mockFetchSigningKeysFail
			sync.FetchModels = mockFetchModelsFail
			sync.SendSigningLog = mockSendSigningLogError
		}
		if !t.MockErrorDB && !t.MockFail {
			// This ensures that we treat the keypairs as new
			sync.GetKeypairByPublicID = mockGetKeypairByPublicID
		}

		client := sync.NewFactoryClient("/api/", "sync", "ValidAPIKey")

		switch t.Args[0] {
		case "account":
			err = client.Accounts()
		case "signingkey":
			err = client.SigningKeys()
		case "model":
			err = client.Models()
		case "signinglog":
			err = client.SigningLogs()
		}

		if len(t.ErrorMessage) == 0 {
			c.Assert(err, check.IsNil)
		} else {
			c.Assert(err, check.NotNil)
			c.Assert(err.Error(), check.Equals, t.ErrorMessage)
		}

		datastore.Environ.DB = &datastore.MockDB{}
		sync.FetchAccounts = mockFetchAccounts
		sync.FetchSigningKeys = mockFetchSigningKeys
		sync.FetchModels = mockFetchModels
		sync.SendSigningLog = mockSendSigningLog
	}

}

func mockFetchAccounts(url, username, apikey string) (account.ListResponse, error) {
	w := sendSyncAPIRequest("GET", "/api/accounts", nil)
	return parseListResponse(w)
}

func mockFetchAccountsError(url, username, apikey string) (account.ListResponse, error) {
	return account.ListResponse{}, errors.New("MOCK error fetching accounts")
}

func mockFetchAccountsFail(url, username, apikey string) (account.ListResponse, error) {
	return account.ListResponse{Success: false, ErrorMessage: "MOCK fail fetching accounts"}, nil
}

func mockFetchSigningKeys(url, username, apikey string, data []byte) (keypair.SyncResponse, error) {
	w := sendSyncAPIRequest("POST", "/api/keypairs/sync", bytes.NewReader(data))
	return parseKeysResponse(w)
}

func mockFetchSigningKeysError(url, username, apikey string, data []byte) (keypair.SyncResponse, error) {
	return keypair.SyncResponse{}, errors.New("MOCK error fetching signing keys")
}

func mockFetchSigningKeysFail(url, username, apikey string, data []byte) (keypair.SyncResponse, error) {
	return keypair.SyncResponse{Success: false}, nil
}

func mockFetchModels(url, username, apikey string) (model.ListResponse, error) {
	w := sendSyncAPIRequest("GET", "/api/models", nil)
	return parseModelResponse(w)
}

func mockFetchModelsError(url, username, apikey string) (model.ListResponse, error) {
	return model.ListResponse{}, errors.New("MOCK error fetching models")
}

func mockFetchModelsFail(url, username, apikey string) (model.ListResponse, error) {
	return model.ListResponse{Success: false, ErrorMessage: "MOCK fail fetching models"}, nil
}

func mockSendSigningLog(url, username, apikey string, signLog datastore.SigningLog) error {
	return nil
}

func mockSendSigningLogError(url, username, apikey string, signLog datastore.SigningLog) error {
	return errors.New("MOCK error syncing signing log")
}

func sendSyncAPIRequest(method, url string, data io.Reader) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	r.Header.Set("user", "sync")
	r.Header.Set("api-key", "ValidAPIKey")

	service.AdminRouter().ServeHTTP(w, r)

	return w
}

func parseListResponse(w *httptest.ResponseRecorder) (account.ListResponse, error) {
	// Check the JSON response
	result := account.ListResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func parseKeysResponse(w *httptest.ResponseRecorder) (keypair.SyncResponse, error) {
	// Check the JSON response
	result := keypair.SyncResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func parseModelResponse(w *httptest.ResponseRecorder) (model.ListResponse, error) {
	// Check the JSON response
	result := model.ListResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func mockReEncryptKeypair(keypair datastore.Keypair, newSecret string) (string, string, error) {
	return "Base64SealedKey", "Base64SAuthKey", nil
}

// mockGetKeypairByPublicID error mock for the database
func mockGetKeypairByPublicID(auth, keyID string) (datastore.Keypair, error) {
	return datastore.Keypair{}, errors.New("MOCK Error fetching from the database")
}
