// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
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

package main

import (
	"log"

	"github.com/snapcore/snapd/asserts"
	"github.com/ubuntu-core/identity-vault/account"
	"github.com/ubuntu-core/identity-vault/service"
)

func main() {
	env := service.Env{}

	// Parse the command line arguments
	account.ParseArgs()
	err := service.ReadConfig(&env.Config, account.SettingsFile)
	if err != nil {
		log.Fatalf("Error parsing the config file: %v\n", err)
	}

	// Open the connection to the local database
	env.DB = service.OpenSysDatabase(env.Config.Driver, env.Config.DataSource)

	// Get the active signing-keys from the database
	keypairs, err := env.DB.ListKeypairs()
	if err != nil {
		log.Fatalf("Error retrieving the keypairs: %v\n", err)
	}

	// Get the account assertions from the snap store and cache them locally
	for _, k := range keypairs {
		if !k.Active {
			// Ignore disabled keys
			continue
		}

		// Get the account assertion from the store
		accountAssert, err := service.FetchAssertionFromStore(asserts.AccountKeyType, []string{k.KeyID})
		if err != nil {
			log.Printf("Error fetching the assertion from the store '%s': %v\n", k.AuthorityID, err)
			continue
		}
		log.Printf("%s\n", asserts.Encode(accountAssert))

		// Store account keys in the database
	}

}
