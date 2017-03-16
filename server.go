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

	"github.com/gorilla/csrf"
	"github.com/ubuntu-core/identity-vault/service"
)

func main() {
	env := service.Env{}
	// Parse the command line arguments
	service.ParseArgs()
	err := service.ReadConfig(&env.Config)
	if err != nil {
		log.Fatalf("Error parsing the config file: %v", err)
	}

	// Open the connection to the local database
	env.DB = service.OpenSysDatabase(env.Config.Driver, env.Config.DataSource)

	// Opening the keypair manager to create the signing database
	env.KeypairDB, err = service.GetKeyStore(env.Config)
	if err != nil {
		log.Fatalf("Error initializing the signing-key database: %v", err)
	}

	var handler http.Handler
	var address string

	switch service.ServiceMode {
	case "admin":
		// configure request forgery protection
		csrfSecure := true
		csrfSecureEnv := os.Getenv("CSRF_SECURE")
		if csrfSecureEnv == "disable" {
			csrfSecure = false
		}

		CSRF := csrf.Protect(
			[]byte(env.Config.CSRFAuthKey),
			csrf.Secure(csrfSecure),
		)

		// Create the admin web service router
		handler = CSRF(service.AdminRouter(&env))
		address = ":8081"
	case "system-user":
		// configure request forgery protection
		csrfSecure := true
		csrfSecureEnv := os.Getenv("CSRF_SECURE")
		if csrfSecureEnv == "disable" {
			csrfSecure = false
		}

		CSRF := csrf.Protect(
			[]byte(env.Config.CSRFAuthKey),
			csrf.Secure(csrfSecure),
		)

		// Create the admin web service router
		handler = CSRF(service.SystemUserRouter(&env))
		address = ":8082"
	default:
		// Create the user web service router
		handler = service.SigningRouter(&env)
		address = ":8080"
	}

	log.Fatal(http.ListenAndServe(address, handler))
}
