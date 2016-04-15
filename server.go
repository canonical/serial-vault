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
	"net/http"

	"github.com/ubuntu-core/identity-vault/service"
)

const sshKeysPath = "/.ssh/authorized_keys"

func main() {
	env := service.Env{}
	// Parse the command line arguments
	service.ParseArgs()
	err := service.ReadConfig(&env.Config)
	if err != nil {
		log.Fatalf("Error parsing the config file: %v", err)
	}

	// Initialize the authorized keys manager
	env.AuthorizedKeys, err = service.InitializeAuthorizedKeys(sshKeysPath)
	if err != nil {
		log.Fatalf("Error initializing the Authorized Keys manager: %v", err)
	}

	// Open the connection to the local database
	env.DB = service.OpenSysDatabase(env.Config.Driver, env.Config.DataSource)

	// Opening the keypair manager to create the signing database
	env.KeypairDB, err = service.GetKeyStore(env.Config)
	if err != nil {
		log.Fatalf("Error initializing the signing-key database: %v", err)
	}

	// Start the web service router
	router := service.Router(&env)

	log.Fatal(http.ListenAndServe(":8080", router))
}
