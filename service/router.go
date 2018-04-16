// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2018 Canonical Ltd
 * License granted by Canonical Limited
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

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/account"
	"github.com/CanonicalLtd/serial-vault/service/app"
	"github.com/CanonicalLtd/serial-vault/service/assertion"
	"github.com/CanonicalLtd/serial-vault/service/core"
	"github.com/CanonicalLtd/serial-vault/service/keypair"
	"github.com/CanonicalLtd/serial-vault/service/model"
	"github.com/CanonicalLtd/serial-vault/service/pivot"
	"github.com/CanonicalLtd/serial-vault/service/sign"
	"github.com/CanonicalLtd/serial-vault/service/signinglog"
	"github.com/CanonicalLtd/serial-vault/service/store"
	"github.com/CanonicalLtd/serial-vault/service/substore"
	"github.com/CanonicalLtd/serial-vault/service/testlog"
	"github.com/CanonicalLtd/serial-vault/service/user"
	"github.com/CanonicalLtd/serial-vault/usso"
	"github.com/gorilla/mux"
)

// SigningRouter returns the application route handler for the signing service methods
func SigningRouter() *mux.Router {
	// Start the web service router
	router := mux.NewRouter()

	// API routes
	router.Handle("/v1/version", Middleware(http.HandlerFunc(core.Version))).Methods("GET")
	router.Handle("/v1/health", Middleware(http.HandlerFunc(core.Health))).Methods("GET")
	router.Handle("/v1/serial", Middleware(ErrorHandler(sign.Serial))).Methods("POST")
	router.Handle("/v1/request-id", Middleware(ErrorHandler(sign.RequestID))).Methods("POST")
	router.Handle("/v1/model", Middleware(ErrorHandler(assertion.ModelAssertion))).Methods("POST")
	router.Handle("/v1/pivot", Middleware(ErrorHandler(pivot.Model))).Methods("POST")
	router.Handle("/v1/pivotmodel", Middleware(ErrorHandler(pivot.ModelAssertion))).Methods("POST")
	router.Handle("/v1/pivotserial", Middleware(ErrorHandler(pivot.SerialAssertion))).Methods("POST")

	// Test log upload routes (only in the factory)
	if datastore.InFactory() {
		router.Handle("/testlog", Middleware(http.HandlerFunc(testlog.Index))).Methods("GET")
		router.Handle("/testlog", Middleware(http.HandlerFunc(testlog.Submit))).Methods("POST")
	}

	return router
}

// AdminRouter returns the application route handler for administrating the application
func AdminRouter() *mux.Router {
	// Start the web service router
	router := mux.NewRouter()

	router.Handle("/v1/version", Middleware(http.HandlerFunc(core.Version))).Methods("GET")
	router.Handle("/v1/health", Middleware(http.HandlerFunc(core.Health))).Methods("GET")

	// API routes: csrf token and auth token
	router.Handle("/v1/token", MiddlewareWithCSRF(http.HandlerFunc(core.Token))).Methods("GET")
	router.Handle("/v1/authtoken", MiddlewareWithCSRF(http.HandlerFunc(core.Token))).Methods("GET")

	// API routes: models admin
	router.Handle("/v1/models", MiddlewareWithCSRF(http.HandlerFunc(model.List))).Methods("GET")
	router.Handle("/v1/models/assertion", MiddlewareWithCSRF(http.HandlerFunc(model.AssertionHeaders))).Methods("POST")
	router.Handle("/v1/models", MiddlewareWithCSRF(http.HandlerFunc(model.Create))).Methods("POST")
	router.Handle("/v1/models/{id:[0-9]+}", MiddlewareWithCSRF(http.HandlerFunc(model.Get))).Methods("GET")
	router.Handle("/v1/models/{id:[0-9]+}", MiddlewareWithCSRF(http.HandlerFunc(model.Update))).Methods("PUT")
	router.Handle("/v1/models/{id:[0-9]+}", MiddlewareWithCSRF(http.HandlerFunc(model.Delete))).Methods("DELETE")

	// API routes: signing-keys
	router.Handle("/v1/keypairs", MiddlewareWithCSRF(http.HandlerFunc(keypair.List))).Methods("GET")
	router.Handle("/v1/keypairs", MiddlewareWithCSRF(http.HandlerFunc(keypair.Create))).Methods("POST")
	router.Handle("/v1/keypairs/{id:[0-9]+}/disable", MiddlewareWithCSRF(http.HandlerFunc(keypair.Disable))).Methods("POST")
	router.Handle("/v1/keypairs/{id:[0-9]+}/enable", MiddlewareWithCSRF(http.HandlerFunc(keypair.Enable))).Methods("POST")
	router.Handle("/v1/keypairs/assertion", MiddlewareWithCSRF(http.HandlerFunc(keypair.Assertion))).Methods("POST")

	router.Handle("/v1/keypairs/generate", MiddlewareWithCSRF(http.HandlerFunc(keypair.Generate))).Methods("POST")
	router.Handle("/v1/keypairs/status/{authorityID}/{keyName}", MiddlewareWithCSRF(http.HandlerFunc(keypair.Status))).Methods("GET")
	router.Handle("/v1/keypairs/status", MiddlewareWithCSRF(http.HandlerFunc(keypair.Progress))).Methods("GET")
	router.Handle("/v1/keypairs/register", MiddlewareWithCSRF(http.HandlerFunc(store.KeyRegister))).Methods("POST")

	// API routes: signing log
	router.Handle("/v1/signinglog", MiddlewareWithCSRF(http.HandlerFunc(signinglog.List))).Methods("GET")
	router.Handle("/v1/signinglog/filters", MiddlewareWithCSRF(http.HandlerFunc(signinglog.ListFilters))).Methods("GET")

	// API routes: account assertions
	router.Handle("/v1/accounts", MiddlewareWithCSRF(http.HandlerFunc(account.List))).Methods("GET")
	router.Handle("/v1/accounts", MiddlewareWithCSRF(http.HandlerFunc(account.Create))).Methods("POST")
	router.Handle("/v1/accounts/{id:[0-9]+}", MiddlewareWithCSRF(http.HandlerFunc(account.Update))).Methods("PUT")
	router.Handle("/v1/accounts/{id:[0-9]+}", MiddlewareWithCSRF(http.HandlerFunc(account.Get))).Methods("GET")
	router.Handle("/v1/accounts/upload", MiddlewareWithCSRF(http.HandlerFunc(account.Upload))).Methods("POST")
	router.Handle("/v1/accounts/{id:[0-9]+}/stores", MiddlewareWithCSRF(http.HandlerFunc(substore.List))).Methods("GET")
	router.Handle("/v1/accounts/stores/{id:[0-9]+}", MiddlewareWithCSRF(http.HandlerFunc(substore.Update))).Methods("PUT")
	router.Handle("/v1/accounts/stores/{id:[0-9]+}", MiddlewareWithCSRF(http.HandlerFunc(substore.Delete))).Methods("DELETE")
	router.Handle("/v1/accounts/stores", MiddlewareWithCSRF(http.HandlerFunc(substore.Create))).Methods("POST")

	// API routes: system-user assertion
	router.Handle("/v1/assertions", MiddlewareWithCSRF(http.HandlerFunc(app.SystemUserAssertion))).Methods("POST")

	// API routes: users management
	router.Handle("/v1/users", MiddlewareWithCSRF(http.HandlerFunc(user.List))).Methods("GET")
	router.Handle("/v1/users", MiddlewareWithCSRF(http.HandlerFunc(user.Create))).Methods("POST")
	router.Handle("/v1/users/{id:[0-9]+}", MiddlewareWithCSRF(http.HandlerFunc(user.Get))).Methods("GET")
	router.Handle("/v1/users/{id:[0-9]+}", MiddlewareWithCSRF(http.HandlerFunc(user.Update))).Methods("PUT")
	router.Handle("/v1/users/{id:[0-9]+}", MiddlewareWithCSRF(http.HandlerFunc(user.Delete))).Methods("DELETE")
	router.Handle("/v1/users/{id:[0-9]+}/otheraccounts", MiddlewareWithCSRF(http.HandlerFunc(user.GetOtherAccounts))).Methods("GET")

	// OpenID routes: using Ubuntu SSO
	router.Handle("/login", MiddlewareWithCSRF(http.HandlerFunc(usso.LoginHandler)))
	router.Handle("/logout", MiddlewareWithCSRF(http.HandlerFunc(usso.LogoutHandler)))

	// Web application routes
	path := []string{datastore.Environ.Config.DocRoot, "/static/"}
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir(strings.Join(path, ""))))
	router.PathPrefix("/static/").Handler(fs)
	router.PathPrefix("/signing-keys").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.PathPrefix("/models").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.PathPrefix("/keypairs").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.PathPrefix("/accounts").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.PathPrefix("/signinglog").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.PathPrefix("/systemuser").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.PathPrefix("/users").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.PathPrefix("/notfound").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.Handle("/", MiddlewareWithCSRF(http.HandlerFunc(app.Index))).Methods("GET")

	// Admin API routes
	router.Handle("/api/signinglog", Middleware(http.HandlerFunc(signinglog.APIList))).Methods("GET")
	router.Handle("/api/keypairs", Middleware(http.HandlerFunc(keypair.APIList))).Methods("GET")
	router.Handle("/api/accounts/{id:[0-9]+}/stores", Middleware(http.HandlerFunc(substore.APIList))).Methods("GET")
	router.Handle("/api/accounts/stores/{id:[0-9]+}", Middleware(http.HandlerFunc(substore.APIUpdate))).Methods("PUT")
	router.Handle("/api/accounts/stores/{id:[0-9]+}", Middleware(http.HandlerFunc(substore.APIDelete))).Methods("DELETE")
	router.Handle("/api/accounts/stores", Middleware(http.HandlerFunc(substore.APICreate))).Methods("POST")

	// Sync API routes
	router.Handle("/api/accounts", Middleware(http.HandlerFunc(account.APIList))).Methods("GET")
	router.Handle("/api/keypairs/sync", Middleware(http.HandlerFunc(keypair.APISyncKeypairs))).Methods("POST")
	router.Handle("/api/models", Middleware(http.HandlerFunc(model.APIList))).Methods("GET")
	router.Handle("/api/signinglog", Middleware(http.HandlerFunc(signinglog.APISyncLog))).Methods("POST")
	router.Handle("/api/testlog", Middleware(http.HandlerFunc(testlog.APIListLog))).Methods("GET")
	router.Handle("/api/testlog", Middleware(http.HandlerFunc(testlog.APISyncLog))).Methods("POST")
	router.Handle("/api/testlog/{id:[0-9]+}", Middleware(http.HandlerFunc(testlog.APISyncUpdateLog))).Methods("PUT")

	return router
}
