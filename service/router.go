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

// SigningRouter returns the application route handler for the signing service methods
func SigningRouter(env *Env) *mux.Router {

	// Start the web service router
	router := mux.NewRouter()

	// API routes
	router.Handle("/v1/version", Middleware(http.HandlerFunc(VersionHandler), env)).Methods("GET")
	router.Handle("/v1/serial", Middleware(ErrorHandler(SignHandler), env)).Methods("POST")
	router.Handle("/v1/request-id", Middleware(ErrorHandler(RequestIDHandler), env)).Methods("POST")

	return router
}

// AdminRouter returns the application route handler for administrating the application
func AdminRouter(env *Env) *mux.Router {

	// Start the web service router
	router := mux.NewRouter()

	// API routes: csrf token
	router.Handle("/v1/token", Middleware(http.HandlerFunc(TokenHandler), env)).Methods("GET")

	// API routes: models admin
	router.Handle("/v1/version", Middleware(http.HandlerFunc(VersionHandler), env)).Methods("GET")
	router.Handle("/v1/models", Middleware(http.HandlerFunc(ModelsHandler), env)).Methods("GET")
	router.Handle("/v1/models", Middleware(http.HandlerFunc(ModelCreateHandler), env)).Methods("POST")
	router.Handle("/v1/models/{id:[0-9]+}", Middleware(http.HandlerFunc(ModelGetHandler), env)).Methods("GET")
	router.Handle("/v1/models/{id:[0-9]+}", Middleware(http.HandlerFunc(ModelUpdateHandler), env)).Methods("PUT")
	router.Handle("/v1/models/{id:[0-9]+}", Middleware(http.HandlerFunc(ModelDeleteHandler), env)).Methods("DELETE")

	// API routes: signing-keys
	router.Handle("/v1/keypairs", Middleware(http.HandlerFunc(KeypairListHandler), env)).Methods("GET")
	router.Handle("/v1/keypairs", Middleware(http.HandlerFunc(KeypairCreateHandler), env)).Methods("POST")
	router.Handle("/v1/keypairs/{id:[0-9]+}/disable", Middleware(http.HandlerFunc(KeypairDisableHandler), env)).Methods("POST")
	router.Handle("/v1/keypairs/{id:[0-9]+}/enable", Middleware(http.HandlerFunc(KeypairEnableHandler), env)).Methods("POST")

	// API routes: signing log
	router.Handle("/v1/signinglog", Middleware(http.HandlerFunc(SigningLogHandler), env)).Methods("GET")
	router.Handle("/v1/signinglog/filters", Middleware(http.HandlerFunc(SigningLogFiltersHandler), env)).Methods("GET")
	router.Handle("/v1/signinglog/{id:[0-9]+}", Middleware(http.HandlerFunc(SigningLogDeleteHandler), env)).Methods("DELETE")

	// API routes: account assertions
	router.Handle("/v1/accounts", Middleware(http.HandlerFunc(AccountsHandler), env)).Methods("GET")
	router.Handle("/v1/accounts", Middleware(http.HandlerFunc(AccountsUpsertHandler), env)).Methods("POST")

	// Web application routes
	path := []string{env.Config.DocRoot, "/static/"}
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir(strings.Join(path, ""))))
	router.PathPrefix("/static/").Handler(fs)
	router.PathPrefix("/models").Handler(Middleware(http.HandlerFunc(IndexHandler), env))
	router.PathPrefix("/accounts").Handler(Middleware(http.HandlerFunc(IndexHandler), env))
	router.PathPrefix("/signinglog").Handler(Middleware(http.HandlerFunc(IndexHandler), env))
	router.Handle("/", Middleware(http.HandlerFunc(IndexHandler), env)).Methods("GET")

	return router
}

// SystemUserRouter returns the application route handler for the system-user service methods
func SystemUserRouter(env *Env) *mux.Router {

	// Start the web service router
	router := mux.NewRouter()

	// API routes
	router.Handle("/v1/version", Middleware(http.HandlerFunc(VersionHandler), env)).Methods("GET")
	router.Handle("/v1/token", Middleware(http.HandlerFunc(TokenHandler), env)).Methods("GET")
	router.Handle("/v1/models", Middleware(http.HandlerFunc(ModelsHandler), env)).Methods("GET")
	router.Handle("/v1/assertions", Middleware(http.HandlerFunc(SystemUserAssertionHandler), env)).Methods("POST")

	// Web application routes
	path := []string{env.Config.DocRoot, "/static/"}
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir(strings.Join(path, ""))))
	router.PathPrefix("/static/").Handler(fs)
	router.Handle("/", Middleware(http.HandlerFunc(UserIndexHandler), env)).Methods("GET")

	return router
}
