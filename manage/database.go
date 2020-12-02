// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
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

package manage

import (
	"fmt"

	"github.com/CanonicalLtd/serial-vault/service/log"

	"github.com/CanonicalLtd/serial-vault/datastore"
)

const (
	create = "Create"
	update = "Update"
)

type operation struct {
	method     func() error
	action     string
	table      string
	skipSqlite bool
}

func execOne(method func() error, action, tableName string) {
	err := method()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("%sd the '%s' table.\n", action, tableName)
	}
}

func exec(operations []operation) {
	for _, op := range operations {
		if op.skipSqlite && datastore.InFactory() {
			continue
		}
		execOne(op.method, op.action, op.table)
	}
}

// DatabaseCommand is the main command for database management
type DatabaseCommand struct{}

// Execute the database schema updates
func (cmd DatabaseCommand) Execute(args []string) error {
	fmt.Println("Update the database schema...")

	openDatabase()

	UpdateDatabase()

	return nil
}

// UpdateDatabase updates the database schema
func UpdateDatabase() {

	// Execute all create and alter table operations
	operations := []operation{
		// Create the keypair table, if it does not exist
		{datastore.Environ.DB.CreateKeypairTable, create, "keypair", false},

		// Create the model table, if it does not exist
		{datastore.Environ.DB.CreateModelTable, create, "model", false},

		// Create the keypair table, if it does not exist
		{datastore.Environ.DB.CreateSettingsTable, create, "settings", false},

		// Create the signinglog table, if it does not exist
		{datastore.Environ.DB.CreateSigningLogTable, create, "signinglog", false},

		// Create the nonce table, if it does not exist
		{datastore.Environ.DB.CreateDeviceNonceTable, create, "nonce", false},

		// Create the account table, if it does not exist
		{datastore.Environ.DB.CreateAccountTable, create, "account", false},
		{datastore.Environ.DB.AlterAccountTable, update, "account", false},

		// Update the model table, adding the new user-keypair field
		{datastore.Environ.DB.AlterModelTable, update, "model", false},

		// Update the keypair table, adding the new fields
		{datastore.Environ.DB.AlterKeypairTable, update, "keypair", false},

		// Create the OpenID nonce table, if it does not exist
		{datastore.Environ.DB.CreateOpenidNonceTable, create, "openid nonce", false},

		// Create the User table, if it does not exist
		{datastore.Environ.DB.CreateUserTable, create, "userinfo", false},

		// Create the AccountUserLink table, if it does not exist
		{datastore.Environ.DB.CreateAccountUserLinkTable, create, "account-user link", true},

		// Update the User table, removing not needed openid_identity field
		{datastore.Environ.DB.AlterUserTable, update, "userinfo", true},

		// Create the Keypair Status table, if it does not exist, and add indexes
		{datastore.Environ.DB.CreateKeypairStatusTable, create, "keypair status", false},
		{datastore.Environ.DB.AlterKeypairStatusTable, update, "keypair status", false},

		// Create the Model Assertion table, if it does not exist
		{datastore.Environ.DB.CreateModelAssertTable, create, "model assertion", false},
		{datastore.Environ.DB.AlterModelAssertTable, update, "model assertion", false},

		{datastore.Environ.DB.CreateSignedModelAssertTable, create, "signed model assertion", false},

		// Create the Sub-store table, if it does not exist
		{datastore.Environ.DB.CreateSubstoreTable, create, "sub-store", false},

		// Create the testlog table, if it does not exist
		{datastore.Environ.DB.CreateTestLogTable, create, "testlog", false},
	}

	exec(operations)

	// Create the test key (if the filesystem store is used)
	if datastore.Environ.Config.KeyStoreType == "filesystem" {
		// Create the test key as it is in the default filesystem keystore
		datastore.Environ.DB.PutKeypair(datastore.Keypair{AuthorityID: "System", KeyID: "61abf588e52be7a3"})
	}

	// Initialize the TPM store, authenticating with the TPM 2.0 module
	if datastore.Environ.Config.KeyStoreType == datastore.TPM20Store.Name {
		fmt.Println("Initialize the TPM2.0 store")
		err := datastore.TPM2InitializeKeystore(nil)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("Initialized TPM 2.0 module.")
		}
	}

}
