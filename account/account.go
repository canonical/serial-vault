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
	"fmt"
	"log"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/snapcore/snapd/asserts"
	"github.com/snapcore/snapd/overlord/auth"
	"github.com/snapcore/snapd/store"
)

// FetchAssertionFromStore retrieves an assertion from the store
var FetchAssertionFromStore = func(modelType *asserts.AssertionType, headers []string) (asserts.Assertion, error) {
	var user *auth.UserState
	var authContext auth.AuthContext
	sto := store.New(nil, authContext)

	return sto.Assertion(modelType, headers, user)
}

// CacheAccountAssertions fetches the account/account-key assertions from the store and caches them in the database
// (reads through the keypairs and refreshes the account/account-key assertions)
func CacheAccountAssertions(env *datastore.Env) {

	// Get the active signing-keys from the database. This operation is not filtered by authorization
	keypairs, err := env.DB.ListAllowedKeypairs(datastore.User{})
	if err != nil {
		log.Fatalf("Error retrieving the keypairs: %v\n", err)
	}

	// Get the account/account-key assertions from the snap store and cache them locally
	for _, k := range keypairs {
		fmt.Printf("Processing keypair - %s\n", k.KeyID)
		if !k.Active {
			// Ignore disabled keys
			fmt.Printf("Keypair %s disabled, so skipping\n", k.KeyID)
			continue
		}

		// Get the account assertion from the store
		accountAssert, err := FetchAssertionFromStore(asserts.AccountType, []string{k.AuthorityID})
		if err != nil {
			fmt.Printf("Error fetching the account assertion from the store: %v\n", err)
			continue
		}

		account := datastore.Account{
			AuthorityID: k.AuthorityID,
			Assertion:   string(asserts.Encode(accountAssert)),
		}

		_, err = env.DB.PutAccount(account, datastore.User{})
		if err != nil {
			fmt.Printf("Error storing the account assertion from the store: %v\n", err)
			continue
		}

		// Get the account-key assertion from the store
		accountKeyAssert, err := FetchAssertionFromStore(asserts.AccountKeyType, []string{k.KeyID})
		if err != nil {
			fmt.Printf("Error fetching the key assertion from the store: %v\n", err)
			continue
		}

		keypair := datastore.Keypair{
			ID:          k.ID,
			AuthorityID: k.AuthorityID,
			KeyID:       k.KeyID,
			Assertion:   string(asserts.Encode(accountKeyAssert)),
		}

		errorCode, err := env.DB.UpdateKeypairAssertion(keypair, datastore.User{})
		if err != nil {
			fmt.Printf("Error on saving the account key assertion to the database: %v - %v\n", errorCode, err)
		}
	}
}

// CacheAccounts fetches the account assertions from the store and caches them in the database
// (reads through the accounts and refreshes the account assertions)
func CacheAccounts(env *datastore.Env) {

	// Get the accounts from the database. This operation is not filtered by authorization
	accounts, err := env.DB.ListAllowedAccounts(datastore.User{})
	if err != nil {
		log.Fatalf("Error retrieving the keypairs: %v\n", err)
	}

	// Get the account assertions from the snap store and cache them locally
	for _, acc := range accounts {
		fmt.Printf("Processing account - %s\n", acc.AuthorityID)

		// Get the account assertion from the store
		accountAssert, err := FetchAssertionFromStore(asserts.AccountType, []string{acc.AuthorityID})
		if err != nil {
			fmt.Printf("Error fetching the account assertion from the store: %v\n", err)
			continue
		}

		account := datastore.Account{
			AuthorityID: acc.AuthorityID,
			Assertion:   string(asserts.Encode(accountAssert)),
		}

		_, err = env.DB.PutAccount(account, datastore.User{})
		if err != nil {
			fmt.Printf("Error storing the account assertion from the store: %v\n", err)
			continue
		}

	}
}
