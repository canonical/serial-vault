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
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/account"
	"github.com/CanonicalLtd/serial-vault/sync"
	check "gopkg.in/check.v1"
)

type startSuite struct{}

var _ = check.Suite(&startSuite{})

func (s *startSuite) SetUpTest(c *check.C) {
	mockDB := datastore.MockDB{}
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	datastore.Environ = &datastore.Env{DB: &mockDB, Config: config}
	datastore.OpenKeyStore(config)

	sync.FetchAccounts = mockFetchAccounts
}

func (s *startSuite) TestStart(c *check.C) {
	tests := []suiteTest{
		{
			Args:         []string{"factory-sync", "start"},
			ErrorMessage: "The cloud serial vault URL, username and API key must be provided",
			MockError:    false},

		{
			Args:         []string{"factory-sync", "start", "--user=sync", "--apikey=ValidAPIKey"},
			ErrorMessage: "",
			MockError:    false},
		{
			Args:         []string{"factory-sync", "start", "--user=sync", "--apikey=ValidAPIKey"},
			ErrorMessage: "Sync completed with errors",
			MockError:    true},
	}

	for _, t := range tests {
		if t.MockError {
			sync.FetchAccounts = mockFetchAccountsError
		}

		runTest(c, t.Args, t.ErrorMessage)

		sync.FetchAccounts = mockFetchAccounts
	}
}

func mockFetchAccounts(url, username, apikey string) (account.ListResponse, error) {
	w := sendSyncAPIRequest("GET", "/api/accounts", nil)
	return parseListResponse(w)
}

func mockFetchAccountsError(url, username, apikey string) (account.ListResponse, error) {
	return account.ListResponse{}, errors.New("MOCK error fetching accounts")
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
