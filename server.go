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

	// Start the web service router
	router := mux.NewRouter()

	// API routes
	router.Handle("/1.0/version", service.Middleware(http.HandlerFunc(service.VersionHandler), &env)).Methods("GET")
	router.Handle("/1.0/models", service.Middleware(http.HandlerFunc(service.ModelsHandler), &env)).Methods("GET")
	router.Handle("/1.0/sign", service.Middleware(http.HandlerFunc(service.SignHandler), &env)).Methods("POST")

	// Web application routes
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	router.PathPrefix("/static/").Handler(fs)
	router.PathPrefix("/").Handler(service.Middleware(http.HandlerFunc(service.IndexHandler), &env))

	log.Fatal(http.ListenAndServe(":8080", router))
}
