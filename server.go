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

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
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

	var router *mux.Router
	var address string

	switch service.ServiceMode {
	case "admin":
		// Create the admin web service router
		router = service.AdminRouter(&env)
		address = ":8081"
	default:
		// Create the user web service router
		router = service.SigningRouter(&env)
		address = ":8080"
	}

	CSRF := csrf.Protect(
		[]byte(env.Config.CSRFAuthKey),
		// UNCOMMENT next line if not working in https. This is a temporal parameter, needed
		// in devmode as gorilla csrf library doesn't send csrf cookies if not set to false.
		// In production this must be removed, as it is supposed to use https, and with https
		// the cookies are sent.
		// (see https://github.com/gorilla/csrf#html-forms comments):
		//
		// csrf.Secure(false),
	)

	log.Fatal(http.ListenAndServe(address, CSRF(router)))
}
