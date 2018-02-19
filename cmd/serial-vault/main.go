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
	"os"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/gorilla/csrf"
)

func main() {
	datastore.Environ = &datastore.Env{}
	// Parse the command line arguments
	config.ParseArgs()
	err := config.ReadConfig(&datastore.Environ.Config, config.SettingsFile)
	if err != nil {
		log.Fatalf("Error parsing the config file: %v", err)
	}

	// Open the connection to the local database
	datastore.OpenSysDatabase(datastore.Environ.Config.Driver, datastore.Environ.Config.DataSource)

	// Opening the keypair manager to create the signing database
	err = datastore.OpenKeyStore(datastore.Environ.Config)
	if err != nil {
		log.Fatalf("Error initializing the signing-key database: %v", err)
	}

	var handler http.Handler
	var address string

	switch config.ServiceMode {
	case "admin":
		// Create the admin web service router
		handler = service.AdminRouter()
		address = ":8081"
	case "system-user":
		// configure request forgery protection
		csrfSecure := true
		csrfSecureEnv := os.Getenv("CSRF_SECURE")
		if csrfSecureEnv == "disable" {
			log.Println("Disable secure flag")
			csrfSecure = false
		}

		CSRF := csrf.Protect(
			[]byte(datastore.Environ.Config.CSRFAuthKey),
			csrf.Secure(csrfSecure),
			csrf.HttpOnly(csrfSecure),
			csrf.CookieName("XSRF-TOKEN"),
		)

		// Create the admin web service router
		if csrfSecure {
			handler = CSRF(service.SystemUserRouter())
		} else {
			// Allow cross-origin access for local development
			handler = service.CORSMiddleware()(CSRF(service.SystemUserRouter()))
		}
		address = ":8082"
	default:
		// Create the user web service router
		handler = service.SigningRouter()
		address = ":8080"
	}

	log.Printf("Starting service on port %s", address)
	log.Fatal(http.ListenAndServe(address, handler))
}
