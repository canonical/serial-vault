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

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/snapcore/snapd/asserts"
)

func TestCacheAccountAssertions(t *testing.T) {
	// Mock the database
	mockDB := datastore.MockDB{}
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	datastore.Environ = &datastore.Env{DB: &mockDB, Config: config}
	datastore.OpenKeyStore(config)

	// Mock the retrieval of the assertion from the store (using a fixed assertion)
	FetchAssertionFromStore = MockFetchAssertionFromStore

	CacheAccountAssertions(datastore.Environ)
}

func TestCacheAccountAssertionsFetchError(t *testing.T) {
	// Mock the database
	mockDB := datastore.MockDB{}
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	datastore.Environ = &datastore.Env{DB: &mockDB, Config: config}
	datastore.OpenKeyStore(config)

	// Mock the retrieval of the assertion from the store (using a fixed assertion)
	FetchAssertionFromStore = mockErrorFetchAssertionFromStore

	CacheAccountAssertions(datastore.Environ)
}

// Mock the retrieval of the assertion from the store (with an error)
func mockErrorFetchAssertionFromStore(modelType *asserts.AssertionType, headers []string) (asserts.Assertion, error) {
	if headers[0] == "systemone" || headers[0] == "invalidone" {
		return nil, errors.New("Error retrieving the account assertion from the store")
	}
	return MockFetchAssertionFromStore(modelType, headers)
}
