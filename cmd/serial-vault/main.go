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

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	svlog "github.com/CanonicalLtd/serial-vault/service/log"
	logging "github.com/op/go-logging"
)

func init() {
	svlog.InitLogger(logging.INFO)
}

func main() {
	datastore.Environ = &datastore.Env{}
	// Parse the command line arguments
	config.ParseArgs()
	err := config.ReadConfig(&datastore.Environ.Config, config.SettingsFile)
	if err != nil {
		svlog.Fatalf("Error parsing the config file: %v", err)
	}

	// Open the connection to the local database
	datastore.OpenSysDatabase(datastore.Environ.Config.Driver, datastore.Environ.Config.DataSource)

	// Opening the keypair manager to create the signing database
	err = datastore.OpenKeyStore(datastore.Environ.Config)
	if err != nil {
		svlog.Fatalf("Error initializing the signing-key database: %v", err)
	}

	var handler http.Handler
	var port string

	switch config.ServiceMode {
	case "admin":
		// Create the admin web service router
		handler = service.AdminRouter()
		port = datastore.Environ.Config.PortAdmin
		if port == "" {
			port = "8081"
		}
	default:
		// Create the user web service router
		handler = service.SigningRouter()
		port = datastore.Environ.Config.PortSignin
		if port == "" {
			port = "8080"
		}
	}

	svlog.Infof("Starting service on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
