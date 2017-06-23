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

package main

import (
	"log"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
)

func main() {
	datastore.Environ = &datastore.Env{}
	// Parse the command line arguments
	config.ParseArgs()
	config.ReadConfig(&datastore.Environ.Config, config.SettingsFile)

	// Open the connection to the local database
	datastore.OpenSysDatabase(datastore.Environ.Config.Driver, datastore.Environ.Config.DataSource)

	// Create the keypair table, if it does not exist
	err := datastore.Environ.DB.CreateKeypairTable()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Created the 'keypair' table.")

		// Create the test key (if the filesystem store is used)
		if datastore.Environ.Config.KeyStoreType == "filesystem" {
			// Create the test key as it is in the default filesystem keystore
			datastore.Environ.DB.PutKeypair(datastore.Keypair{AuthorityID: "System", KeyID: "61abf588e52be7a3"})
		}
	}

	// Create the model table, if it does not exist
	err = datastore.Environ.DB.CreateModelTable()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Created the 'model' table.")
	}

	// Create the keypair table, if it does not exist
	err = datastore.Environ.DB.CreateSettingsTable()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Created the 'settings' table.")
	}

	// Create the signinglog table, if it does not exist
	err = datastore.Environ.DB.CreateSigningLogTable()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Created the 'signinglog' table.")
	}

	// Create the nonce table, if it does not exist
	err = datastore.Environ.DB.CreateDeviceNonceTable()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Created the 'nonce' table.")
	}

	// Create the account table, if it does not exist
	err = datastore.Environ.DB.CreateAccountTable()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Created the 'account' table.")
	}

	// Update the model table, adding the new user-keypair field
	err = datastore.Environ.DB.AlterModelTable()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Updated the 'model' table.")
	}

	// Update the keypair table, adding the new fields
	err = datastore.Environ.DB.AlterKeypairTable()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Updated the 'keypair' table.")
	}

	// Initalize the TPM store, authenticating with the TPM 2.0 module
	if datastore.Environ.Config.KeyStoreType == datastore.TPM20Store.Name {
		log.Println("Initialize the TPM2.0 store")
		err = datastore.TPM2InitializeKeystore(nil)
		if err != nil {
			log.Fatal(err)
		} else {
			log.Println("Initialized TPM 2.0 module.")
		}
	}
}
