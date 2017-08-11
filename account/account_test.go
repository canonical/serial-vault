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

package account

import (
	"errors"
	"testing"

	"gopkg.in/check.v1"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/snapcore/snapd/asserts"
)

type AccountSuite struct{}

var _ = check.Suite(&AccountSuite{})

func Test(t *testing.T) { check.TestingT(t) }

func (s *AccountSuite) SetUpTest(c *check.C) {
	//datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}

	// Mock the database
	mockDB := datastore.MockDB{}
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	datastore.Environ = &datastore.Env{DB: &mockDB, Config: config}
	datastore.OpenKeyStore(config)
}

func (s *AccountSuite) TestAccountAssertionsFromKeypairs(c *check.C) {
	tests := []bool{false, true}

	for _, t := range tests {
		if t {
			// Mock the retrieval of the assertion from the store (using a fixed assertion)
			FetchAssertionFromStore = MockFetchAssertionFromStore
		} else {
			// Mock error in retrieval of the assertion from the store (using a fixed assertion)
			FetchAssertionFromStore = mockErrorFetchAssertionFromStore
		}

		CacheAccountAssertions(datastore.Environ)
	}
}

func (s *AccountSuite) TestAccountAssertions(c *check.C) {
	tests := []bool{false, true}

	for _, t := range tests {
		if t {
			// Mock the retrieval of the assertion from the store (using a fixed assertion)
			FetchAssertionFromStore = MockFetchAssertionFromStore
		} else {
			// Mock error in retrieval of the assertion from the store (using a fixed assertion)
			FetchAssertionFromStore = mockErrorFetchAssertionFromStore
		}

		CacheAccounts(datastore.Environ)
	}
}

// Mock the retrieval of the assertion from the store (with an error)
func mockErrorFetchAssertionFromStore(modelType *asserts.AssertionType, headers []string) (asserts.Assertion, error) {

	if modelType == asserts.AccountType || headers[0] == "systemone" || headers[0] == "invalidone" {
		return nil, errors.New("Error retrieving the account assertion from the store")
	}
	return MockFetchAssertionFromStore(modelType, headers)
}
