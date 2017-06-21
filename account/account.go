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
	"flag"
	"log"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/snapcore/snapd/asserts"
	"github.com/snapcore/snapd/overlord/auth"
	"github.com/snapcore/snapd/store"
)

// SettingsFile is the path to the settings YAML file
var SettingsFile string

// ParseArgs checks the command line arguments
func ParseArgs() {
	flag.StringVar(&SettingsFile, "config", "./settings.yaml", "Path to the config file")
	flag.Parse()
}

// FetchAssertionFromStore retrieves an assertion from the store
var FetchAssertionFromStore = func(modelType *asserts.AssertionType, headers []string) (asserts.Assertion, error) {
	var user *auth.UserState
	var authContext auth.AuthContext
	sto := store.New(nil, authContext)

	return sto.Assertion(modelType, headers, user)
}

// CacheAccountAssertions fetches the account assertions from the store and caches them in the database
func CacheAccountAssertions(env *service.Env) {

	// Get the active signing-keys from the database
	keypairs, err := env.DB.ListKeypairs()
	if err != nil {
		log.Fatalf("Error retrieving the keypairs: %v\n", err)
	}

	// Get the account assertions from the snap store and cache them locally
	for _, k := range keypairs {
		log.Printf("-- Processing keypair - %s\n", k.KeyID)
		if !k.Active {
			// Ignore disabled keys
			log.Println("Disabled, so skipping")
			continue
		}

		// Get the account assertion from the store
		accountAssert, err := FetchAssertionFromStore(asserts.AccountType, []string{k.AuthorityID})
		if err != nil {
			log.Printf("Error fetching the account assertion from the store: %v\n", err)
			continue
		}

		_, err = env.DB.PutAccount(datastore.Account{AuthorityID: k.AuthorityID, Assertion: string(asserts.Encode(accountAssert))})
		if err != nil {
			log.Printf("Error storing the account assertion from the store: %v\n", err)
			continue
		}

		// Get the account-key assertion from the store
		accountKeyAssert, err := FetchAssertionFromStore(asserts.AccountKeyType, []string{k.KeyID})
		if err != nil {
			log.Printf("Error fetching the key assertion from the store: %v\n", err)
			continue
		}

		err = env.DB.UpdateKeypairAssertion(k.ID, string(asserts.Encode(accountKeyAssert)))
		if err != nil {
			log.Printf("Error on saving the account key assertion to the database: %v\n", err)
		}
	}

}
