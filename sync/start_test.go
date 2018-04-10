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
	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
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
	sync.FetchSigningKeys = mockFetchSigningKeys
	datastore.ReEncryptKeypair = mockReEncryptKeypair
	sync.FetchModels = mockFetchModels
	sync.SendSigningLog = mockSendSigningLog
}

func (s *startSuite) TestStart(c *check.C) {
	tests := []suiteTest{
		{
			Args:         []string{"factory", "sync"},
			ErrorMessage: "The cloud serial vault URL, username and API key must be provided"},
		{
			Args:         []string{"factory", "sync", "--user=sync", "--apikey=ValidAPIKey"},
			ErrorMessage: ""},
		{
			Args:         []string{"factory", "sync", "--user=sync", "--apikey=ValidAPIKey"},
			ErrorMessage: "Sync completed with errors",
			MockErrorDB:  true},
		{
			Args:         []string{"factory", "sync", "--user=sync", "--apikey=ValidAPIKey"},
			ErrorMessage: "Sync completed with errors",
			MockFail:     true},
	}

	for _, t := range tests {
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

		runTest(c, t.Args, t.ErrorMessage)

		datastore.Environ.DB = &datastore.MockDB{}
		sync.FetchAccounts = mockFetchAccounts
		sync.FetchSigningKeys = mockFetchSigningKeys
		sync.FetchModels = mockFetchModels
		sync.SendSigningLog = mockSendSigningLog
	}
}
