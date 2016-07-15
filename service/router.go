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

package service

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// SigningRouter returns the application route handler for the user methods
func SigningRouter(env *Env) *mux.Router {

	// Start the web service router
	router := mux.NewRouter()

	// API routes
	router.Handle("/1.0/version", Middleware(http.HandlerFunc(VersionHandler), env)).Methods("GET")
	router.Handle("/1.0/sign", Middleware(http.HandlerFunc(SignHandler), env)).Methods("POST")

	return router
}

// AdminRouter returns the application route handler for administrating the application
func AdminRouter(env *Env) *mux.Router {

	// Start the web service router
	router := mux.NewRouter()

	// API routes: models admin
	router.Handle("/1.0/version", Middleware(http.HandlerFunc(VersionHandler), env)).Methods("GET")
	router.Handle("/1.0/models", Middleware(http.HandlerFunc(ModelsHandler), env)).Methods("GET")
	router.Handle("/1.0/models", Middleware(http.HandlerFunc(ModelCreateHandler), env)).Methods("POST")
	router.Handle("/1.0/models/{id:[0-9]+}", Middleware(http.HandlerFunc(ModelGetHandler), env)).Methods("GET")
	router.Handle("/1.0/models/{id:[0-9]+}", Middleware(http.HandlerFunc(ModelUpdateHandler), env)).Methods("PUT")
	router.Handle("/1.0/models/{id:[0-9]+}", Middleware(http.HandlerFunc(ModelDeleteHandler), env)).Methods("DELETE")

	// API routes: signing-keys
	router.Handle("/1.0/keypairs", Middleware(http.HandlerFunc(KeypairListHandler), env)).Methods("GET")
	router.Handle("/1.0/keypairs", Middleware(http.HandlerFunc(KeypairCreateHandler), env)).Methods("POST")
	router.Handle("/1.0/keypairs/{id:[0-9]+}/disable", Middleware(http.HandlerFunc(KeypairDisableHandler), env)).Methods("POST")
	router.Handle("/1.0/keypairs/{id:[0-9]+}/enable", Middleware(http.HandlerFunc(KeypairEnableHandler), env)).Methods("POST")

	// API routes: signing log
	router.Handle("/1.0/signinglog", Middleware(http.HandlerFunc(SigningLogHandler), env)).Methods("GET")
	router.Handle("/1.0/signinglog/{id:[0-9]+}", Middleware(http.HandlerFunc(SigningLogDeleteHandler), env)).Methods("DELETE")

	// Web application routes
	path := []string{env.Config.DocRoot, "/static/"}
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir(strings.Join(path, ""))))
	router.PathPrefix("/static/").Handler(fs)
	router.PathPrefix("/models").Handler(Middleware(http.HandlerFunc(IndexHandler), env))
	router.PathPrefix("/signinglog").Handler(Middleware(http.HandlerFunc(IndexHandler), env))
	router.Handle("/", Middleware(http.HandlerFunc(IndexHandler), env)).Methods("GET")

	return router
}
