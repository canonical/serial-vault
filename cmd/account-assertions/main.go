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

	"github.com/CanonicalLtd/serial-vault/account"
	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
)

func main() {
	datastore.Environ = &datastore.Env{}

	// Parse the command line arguments
	account.ParseArgs()
	err := config.ReadConfig(&datastore.Environ.Config, account.SettingsFile)
	if err != nil {
		log.Fatalf("Error parsing the config file: %v\n", err)
	}

	// Open the connection to the local database
	datastore.OpenSysDatabase(datastore.Environ.Config.Driver, datastore.Environ.Config.DataSource)

	// Cache the account assertions from the store in the database
	account.CacheAccountAssertions(datastore.Environ)
}
