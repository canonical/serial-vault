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

const (
	create = "Create"
	update = "Update"
)

type operation struct {
	method func() error
	action string
	table  string
}

func execOne(method func() error, action, tableName string) {
	err := method()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("%sd the '%s' table.", action, tableName)
	}
}

func exec(operations []operation) {
	for _, op := range operations {
		execOne(op.method, op.action, op.table)
	}
}

func main() {
	datastore.Environ = &datastore.Env{}
	// Parse the command line arguments
	config.ParseArgs()
	config.ReadConfig(&datastore.Environ.Config, config.SettingsFile)

	// Open the connection to the local database
	datastore.OpenSysDatabase(datastore.Environ.Config.Driver, datastore.Environ.Config.DataSource)

	// Execute all create and alter table operations
	operations := []operation{
		// Create the keypair table, if it does not exist
		{datastore.Environ.DB.CreateKeypairTable, create, "keypair"},

		// Create the model table, if it does not exist
		{datastore.Environ.DB.CreateModelTable, create, "model"},

		// Create the keypair table, if it does not exist
		{datastore.Environ.DB.CreateSettingsTable, create, "settings"},

		// Create the signinglog table, if it does not exist
		{datastore.Environ.DB.CreateSigningLogTable, create, "signinglog"},

		// Create the nonce table, if it does not exist
		{datastore.Environ.DB.CreateDeviceNonceTable, create, "nonce"},

		// Create the account table, if it does not exist
		{datastore.Environ.DB.CreateAccountTable, create, "account"},

		// Update the model table, adding the new user-keypair field
		{datastore.Environ.DB.AlterModelTable, update, "model"},

		// Update the keypair table, adding the new fields
		{datastore.Environ.DB.AlterKeypairTable, update, "keypair"},

		// Create the OpenID nonce table, if it does not exist
		{datastore.Environ.DB.CreateOpenidNonceTable, create, "openid nonce"},

		// Create the User table, if it does not exist
		{datastore.Environ.DB.CreateUserTable, create, "userinfo"},

		// Create the AccountUserLink table, if it does not exist
		{datastore.Environ.DB.CreateAccountUserLinkTable, create, "account-user link"},

		// Update the User table, removing not needed openid_identity field
		{datastore.Environ.DB.AlterUserTable, update, "userinfo"},
	}

	exec(operations)

	// Create the test key (if the filesystem store is used)
	if datastore.Environ.Config.KeyStoreType == "filesystem" {
		// Create the test key as it is in the default filesystem keystore
		datastore.Environ.DB.PutKeypair(datastore.Keypair{AuthorityID: "System", KeyID: "61abf588e52be7a3"})
	}

	// Initalize the TPM store, authenticating with the TPM 2.0 module
	if datastore.Environ.Config.KeyStoreType == datastore.TPM20Store.Name {
		log.Println("Initialize the TPM2.0 store")
		err := datastore.TPM2InitializeKeystore(nil)
		if err != nil {
			log.Fatal(err)
		} else {
			log.Println("Initialized TPM 2.0 module.")
		}
	}

}
