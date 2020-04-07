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
	"github.com/CanonicalLtd/serial-vault/service/metric"
	"github.com/CanonicalLtd/serial-vault/service/model"
	"github.com/CanonicalLtd/serial-vault/service/pivot"
	"github.com/CanonicalLtd/serial-vault/service/sign"
	"github.com/CanonicalLtd/serial-vault/service/signinglog"
	"github.com/CanonicalLtd/serial-vault/service/status"
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

	router.Handle("/v1/version", Middleware(http.HandlerFunc(core.Version))).Methods("GET")
	router.Handle("/v1/health", Middleware(http.HandlerFunc(core.Health))).Methods("GET")

	// API routes
	router.Handle("/v1/serial", metric.CollectAPIStats("signSerial",
		Middleware(ErrorHandler(sign.Serial)))).
		Methods("POST")
	router.Handle("/v1/request-id", metric.CollectAPIStats("signRequestID",
		Middleware(ErrorHandler(sign.RequestID)))).
		Methods("POST")
	router.Handle("/v1/model", metric.CollectAPIStats("assertionModelAssertion",
		Middleware(ErrorHandler(assertion.ModelAssertion)))).
		Methods("POST")
	router.Handle("/v1/pivot", metric.CollectAPIStats("pivotModel",
		Middleware(ErrorHandler(pivot.Model)))).
		Methods("POST")
	router.Handle("/v1/pivotmodel", metric.CollectAPIStats("pivotModelAssertion",
		Middleware(ErrorHandler(pivot.ModelAssertion)))).
		Methods("POST")
	router.Handle("/v1/pivotserial", metric.CollectAPIStats("pivotSerialAssertion",
		Middleware(ErrorHandler(pivot.SerialAssertion)))).
		Methods("POST")
	router.Handle("/v1/pivotuser", metric.CollectAPIStats("pivotSystemUserAssertion",
		Middleware(ErrorHandler(pivot.SystemUserAssertion)))).
		Methods("POST")

	// Test log upload routes (only in the factory)
	if datastore.InFactory() {
		router.Handle("/testlog", Middleware(http.HandlerFunc(testlog.Index))).Methods("GET")
		router.Handle("/testlog", Middleware(http.HandlerFunc(testlog.Submit))).Methods("POST")
	}

	// prometheus metrics endpoint
	router.Handle("/_status/metrics", metric.NewServer()).Methods("GET")
	// status endpoints
	status.AddStatusEndpoints("/_status", router)

	return router
}

// AdminRouter returns the application route handler for administrating the application
func AdminRouter() *mux.Router {
	// Start the web service router
	router := mux.NewRouter()

	router.Handle("/v1/version", Middleware(http.HandlerFunc(core.Version))).Methods("GET")
	router.Handle("/v1/health", Middleware(http.HandlerFunc(core.Health))).Methods("GET")

	// API routes: csrf token and auth token
	router.Handle("/v1/token", metric.CollectAPIStats("coreToken",
		MiddlewareWithCSRF(http.HandlerFunc(core.Token)))).
		Methods("GET")
	router.Handle("/v1/authtoken", metric.CollectAPIStats("coreToken",
		MiddlewareWithCSRF(http.HandlerFunc(core.Token)))).
		Methods("GET")

	// API routes: models admin
	router.Handle("/v1/models", metric.CollectAPIStats("modelList",
		MiddlewareWithCSRF(http.HandlerFunc(model.List)))).
		Methods("GET")
	router.Handle("/v1/models/assertion", metric.CollectAPIStats("modelAssertionHeaders",
		MiddlewareWithCSRF(http.HandlerFunc(model.AssertionHeaders)))).
		Methods("POST")
	router.Handle("/v1/models", metric.CollectAPIStats("modelCreate",
		MiddlewareWithCSRF(http.HandlerFunc(model.Create)))).
		Methods("POST")
	router.Handle("/v1/models/{id:[0-9]+}", metric.CollectAPIStats("modelGet",
		MiddlewareWithCSRF(http.HandlerFunc(model.Get)))).
		Methods("GET")
	router.Handle("/v1/models/{id:[0-9]+}", metric.CollectAPIStats("modelUpdate",
		MiddlewareWithCSRF(http.HandlerFunc(model.Update)))).
		Methods("PUT")
	router.Handle("/v1/models/{id:[0-9]+}", metric.CollectAPIStats("modelDelete",
		MiddlewareWithCSRF(http.HandlerFunc(model.Delete)))).
		Methods("DELETE")

	// API routes: signing-keys
	router.Handle("/v1/keypairs", metric.CollectAPIStats("keypairList",
		MiddlewareWithCSRF(http.HandlerFunc(keypair.List)))).
		Methods("GET")
	router.Handle("/v1/keypairs", metric.CollectAPIStats("keypairCreate",
		MiddlewareWithCSRF(http.HandlerFunc(keypair.Create)))).
		Methods("POST")
	router.Handle("/v1/keypairs/{id:[0-9]+}", metric.CollectAPIStats("keypairGet",
		MiddlewareWithCSRF(http.HandlerFunc(keypair.Get)))).
		Methods("GET")
	router.Handle("/v1/keypairs/{id:[0-9]+}", metric.CollectAPIStats("keypairUpdate",
		MiddlewareWithCSRF(http.HandlerFunc(keypair.Update)))).
		Methods("PUT")
	router.Handle("/v1/keypairs/{id:[0-9]+}/disable", metric.CollectAPIStats("keypairDisable",
		MiddlewareWithCSRF(http.HandlerFunc(keypair.Disable)))).
		Methods("POST")
	router.Handle("/v1/keypairs/{id:[0-9]+}/enable", metric.CollectAPIStats("keypairEnable",
		MiddlewareWithCSRF(http.HandlerFunc(keypair.Enable)))).
		Methods("POST")
	router.Handle("/v1/keypairs/assertion", metric.CollectAPIStats("keypairAssertion",
		MiddlewareWithCSRF(http.HandlerFunc(keypair.Assertion)))).
		Methods("POST")

	router.Handle("/v1/keypairs/generate", metric.CollectAPIStats("keypairGenerate",
		MiddlewareWithCSRF(http.HandlerFunc(keypair.Generate)))).
		Methods("POST")
	router.Handle("/v1/keypairs/status/{authorityID}/{keyName}", metric.CollectAPIStats("keypairStatus",
		MiddlewareWithCSRF(http.HandlerFunc(keypair.Status)))).
		Methods("GET")
	router.Handle("/v1/keypairs/status", metric.CollectAPIStats("keypairProgress",
		MiddlewareWithCSRF(http.HandlerFunc(keypair.Progress)))).
		Methods("GET")
	router.Handle("/v1/keypairs/register", metric.CollectAPIStats("storeKeyRegister",
		MiddlewareWithCSRF(http.HandlerFunc(store.KeyRegister)))).
		Methods("POST")

	// API routes: signing log
	// TODO: GET /v1/signinglog is not really used in the frontend and could be removed
	router.Handle("/v1/signinglog", metric.CollectAPIStats("signinglogList",
		MiddlewareWithCSRF(http.HandlerFunc(signinglog.List)))).
		Methods("GET")
	router.Handle("/v1/signinglog/account/{authorityID}", metric.CollectAPIStats("signinglogListForAccount",
		MiddlewareWithCSRF(http.HandlerFunc(signinglog.ListForAccount)))).
		Methods("GET")
	router.Handle("/v1/signinglog/account/{authorityID}/filters", metric.CollectAPIStats("signinglogListFilters",
		MiddlewareWithCSRF(http.HandlerFunc(signinglog.ListFilters)))).
		Methods("GET")

	// API routes: account assertions
	router.Handle("/v1/accounts", metric.CollectAPIStats("accountList",
		MiddlewareWithCSRF(http.HandlerFunc(account.List)))).
		Methods("GET")
	router.Handle("/v1/accounts", metric.CollectAPIStats("accountCreate",
		MiddlewareWithCSRF(http.HandlerFunc(account.Create)))).
		Methods("POST")
	router.Handle("/v1/accounts/{id:[0-9]+}", metric.CollectAPIStats("accountUpdate",
		MiddlewareWithCSRF(http.HandlerFunc(account.Update)))).
		Methods("PUT")
	router.Handle("/v1/accounts/{id:[0-9]+}", metric.CollectAPIStats("accountGet",
		MiddlewareWithCSRF(http.HandlerFunc(account.Get)))).
		Methods("GET")
	router.Handle("/v1/accounts/upload", metric.CollectAPIStats("accountUpload",
		MiddlewareWithCSRF(http.HandlerFunc(account.Upload)))).
		Methods("POST")
	router.Handle("/v1/accounts/{id:[0-9]+}/stores", metric.CollectAPIStats("substoreList",
		MiddlewareWithCSRF(http.HandlerFunc(substore.List)))).
		Methods("GET")
	router.Handle("/v1/accounts/stores/{id:[0-9]+}", metric.CollectAPIStats("substoreUpdate",
		MiddlewareWithCSRF(http.HandlerFunc(substore.Update)))).
		Methods("PUT")
	router.Handle("/v1/accounts/stores/{id:[0-9]+}", metric.CollectAPIStats("substoreDelete",
		MiddlewareWithCSRF(http.HandlerFunc(substore.Delete)))).
		Methods("DELETE")
	router.Handle("/v1/accounts/stores", metric.CollectAPIStats("substoreCreate",
		MiddlewareWithCSRF(http.HandlerFunc(substore.Create)))).
		Methods("POST")

	// API routes: system-user assertion
	router.Handle("/v1/assertions", metric.CollectAPIStats("assertionSystemUserAssertion",
		MiddlewareWithCSRF(http.HandlerFunc(assertion.SystemUserAssertion)))).
		Methods("POST")

	// API routes: users management
	router.Handle("/v1/users", metric.CollectAPIStats("userList",
		MiddlewareWithCSRF(http.HandlerFunc(user.List)))).
		Methods("GET")
	router.Handle("/v1/users", metric.CollectAPIStats("userCreate",
		MiddlewareWithCSRF(http.HandlerFunc(user.Create)))).
		Methods("POST")
	router.Handle("/v1/users/{id:[0-9]+}", metric.CollectAPIStats("userGet",
		MiddlewareWithCSRF(http.HandlerFunc(user.Get)))).
		Methods("GET")
	router.Handle("/v1/users/{id:[0-9]+}", metric.CollectAPIStats("userUpdate",
		MiddlewareWithCSRF(http.HandlerFunc(user.Update)))).
		Methods("PUT")
	router.Handle("/v1/users/{id:[0-9]+}", metric.CollectAPIStats("userDelete",
		MiddlewareWithCSRF(http.HandlerFunc(user.Delete)))).
		Methods("DELETE")
	router.Handle("/v1/users/{id:[0-9]+}/otheraccounts", metric.CollectAPIStats("userGetOtherAccounts",
		MiddlewareWithCSRF(http.HandlerFunc(user.GetOtherAccounts)))).
		Methods("GET")

	// OpenID routes: using Ubuntu SSO
	router.Handle("/login", metric.CollectAPIStats("ussoLoginHandler",
		MiddlewareWithCSRF(http.HandlerFunc(usso.LoginHandler))))
	router.Handle("/logout", metric.CollectAPIStats("ussoLogoutHandler",
		MiddlewareWithCSRF(http.HandlerFunc(usso.LogoutHandler))))

	// Web application routes
	path := []string{datastore.Environ.Config.DocRoot, "/static/"}
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir(strings.Join(path, ""))))
	router.PathPrefix("/static/").Handler(fs)
	router.PathPrefix("/signing-keys").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.PathPrefix("/models").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.PathPrefix("/keypairs").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.PathPrefix("/accounts").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.PathPrefix("/signinglog").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.PathPrefix("/substores").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.PathPrefix("/systemuser").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.PathPrefix("/users").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.PathPrefix("/notfound").Handler(MiddlewareWithCSRF(http.HandlerFunc(app.Index)))
	router.Handle("/", MiddlewareWithCSRF(http.HandlerFunc(app.Index))).Methods("GET")

	// Admin API routes
	router.Handle("/api/signinglog", metric.CollectAPIStats("signinglogAPIList",
		Middleware(http.HandlerFunc(signinglog.APIList)))).
		Methods("GET")
	router.Handle("/api/keypairs", metric.CollectAPIStats("keypairAPIList",
		Middleware(http.HandlerFunc(keypair.APIList)))).
		Methods("GET")
	router.Handle("/api/accounts/{id:[0-9]+}/stores", metric.CollectAPIStats("substoreAPIList",
		Middleware(http.HandlerFunc(substore.APIList)))).
		Methods("GET")
	router.Handle("/api/accounts/stores/{id:[0-9]+}", metric.CollectAPIStats("substoreAPIUpdate",
		Middleware(http.HandlerFunc(substore.APIUpdate)))).
		Methods("PUT")
	router.Handle("/api/accounts/stores/{id:[0-9]+}", metric.CollectAPIStats("substoreAPIDelete",
		Middleware(http.HandlerFunc(substore.APIDelete)))).
		Methods("DELETE")
	router.Handle("/api/accounts/stores", metric.CollectAPIStats("substoreAPICreate",
		Middleware(http.HandlerFunc(substore.APICreate)))).
		Methods("POST")
	router.Handle("/api/accounts/stores/{modelID:[0-9]+}/{serial}", metric.CollectAPIStats("substoreAPIGet",
		Middleware(http.HandlerFunc(substore.APIGet)))).
		Methods("GET")
	router.Handle("/api/assertions/checkserial", metric.CollectAPIStats("assertionAPIValidateSerial",
		Middleware(http.HandlerFunc(assertion.APIValidateSerial)))).
		Methods("POST")
	router.Handle("/api/assertions", metric.CollectAPIStats("assertionAPISystemUser",
		Middleware(http.HandlerFunc(assertion.APISystemUser)))).
		Methods("POST")
	router.Handle("/api/models/{id:[0-9]+}", metric.CollectAPIStats("modelAPIGet",
		Middleware(http.HandlerFunc(model.APIGet)))).
		Methods("GET")
	router.Handle("/api/models/{id:[0-9]+}", metric.CollectAPIStats("modelAPIUpdate",
		Middleware(http.HandlerFunc(model.APIUpdate)))).
		Methods("PUT")
	router.Handle("/api/models/{id:[0-9]+}", metric.CollectAPIStats("modelAPIDelete",
		Middleware(http.HandlerFunc(model.APIDelete)))).
		Methods("DELETE")
	router.Handle("/api/models", metric.CollectAPIStats("modelAPICreate",
		Middleware(http.HandlerFunc(model.APICreate)))).
		Methods("POST")
	router.Handle("/api/models/assertion", metric.CollectAPIStats("modelAPIAssertionHeaders",
		Middleware(http.HandlerFunc(model.APIAssertionHeaders)))).
		Methods("POST")

	// Sync API routes
	router.Handle("/api/accounts", metric.CollectAPIStats("accountAPIList",
		Middleware(http.HandlerFunc(account.APIList)))).
		Methods("GET")
	router.Handle("/api/keypairs/sync", metric.CollectAPIStats("keypairAPISyncKeypairs)",
		Middleware(http.HandlerFunc(keypair.APISyncKeypairs)))).
		Methods("POST")
	router.Handle("/api/models", metric.CollectAPIStats("modelAPIList",
		Middleware(http.HandlerFunc(model.APIList)))).
		Methods("GET")
	router.Handle("/api/signinglog", metric.CollectAPIStats("signinglogAPISyncLog",
		Middleware(http.HandlerFunc(signinglog.APISyncLog)))).
		Methods("POST")
	router.Handle("/api/testlog", metric.CollectAPIStats("testlogAPIListLog",
		Middleware(http.HandlerFunc(testlog.APIListLog)))).
		Methods("GET")
	router.Handle("/api/testlog", metric.CollectAPIStats("testlogAPISyncLog",
		Middleware(http.HandlerFunc(testlog.APISyncLog)))).
		Methods("POST")
	router.Handle("/api/testlog/{id:[0-9]+}", metric.CollectAPIStats("testlogAPISyncUpdateLog",
		Middleware(http.HandlerFunc(testlog.APISyncUpdateLog)))).
		Methods("PUT")

	// prometheus metrics endpoint
	router.Handle("/_status/metrics", metric.NewServer()).Methods("GET")
	// status endpoints
	status.AddStatusEndpoints("/_status", router)

	return router
}
