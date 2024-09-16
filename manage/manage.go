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

package manage

import (
	"log"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
)

// Command defines the options for the serial-vault-admin command-line utility
type Command struct {
	SettingsFile string `short:"c" long:"config" description:"Path to the config file" default:"./settings.yaml"`

	Account  AccountCommand  `command:"account" alias:"a" description:"Account management"`
	Client   ClientCommand   `command:"client" alias:"c" description:"Serial-Vault Client to generate a test serial assertion request"`
	Database DatabaseCommand `command:"database" alias:"d" description:"Database schema update"`
	User     UserCommand     `command:"user" alias:"u" description:"User management"`
}

// Manage is the implementation of the command configuration for the serial-vault-admin command-line
var Manage Command

func openDatabase() {
	// Check that the database has not been set e.g. by a mock
	if datastore.Environ.DB != nil {
		return
	}

	config.ReadConfig(&datastore.Environ.Config, Manage.SettingsFile)

	// Open the connection to the database
	datastore.OpenSysDatabase(datastore.Environ.Config.Driver, datastore.Environ.Config.DataSource)

	// Opening the keypair manager to create the signing database
	err := datastore.OpenKeyStore(datastore.Environ.Config)
	if err != nil {
		log.Panicf("Error initializing the signing-key database: %v", err)
	}
}
