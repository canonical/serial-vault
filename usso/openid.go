// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
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

package usso

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/juju/usso"
	"github.com/juju/usso/openid"
)

var (
	teams    = "" // e.g. ce-web-logs,canonical
	required = "email,fullname,nickname"
	optional = ""
)

var client = openid.NewClient(usso.ProductionUbuntuSSOServer, &datastore.OpenidNonceStore, nil)

// verify is used to perform the OpenID verification of the login
// response. This is declared as a variable so it can be overridden for
// testing.
var verify = client.Verify

func replyHTTPError(w http.ResponseWriter, returnCode int, err error) {
	w.Header().Set("ContentType", "text/html")
	w.WriteHeader(returnCode)
	errorTemplate.Execute(w, err)
}

// LoginHandler processes the login for Ubuntu SSO
func LoginHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	url := *r.URL

	// Set the return URL: from the OpenID login with the full domain name
	url.Scheme = datastore.Environ.Config.URLScheme
	url.Host = datastore.Environ.Config.URLHost

	if r.Form.Get("openid.ns") == "" {
		req := openid.Request{
			ReturnTo:     url.String(),
			Teams:        strings.FieldsFunc(teams, isComma),
			SRegRequired: strings.FieldsFunc(required, isComma),
			SRegOptional: strings.FieldsFunc(optional, isComma),
		}
		url := client.RedirectURL(&req)
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	resp, err := verify(url.String())
	if err != nil {
		// A mangled OpenID response is suspicious, so leave a nasty response
		replyHTTPError(w, http.StatusBadRequest, err)
		return
	}

	// Check we have the mandatory fields in the OpenID response
	username := r.Form.Get("openid.sreg.nickname")
	fullname := r.Form.Get("openid.sreg.fullname")
	if len(username) == 0 || len(fullname) == 0 {
		log.Println("Some params are missing from the OpenID response")
		http.Redirect(w, r, "/notfound", http.StatusTemporaryRedirect)
		return
	}

	User, err := datastore.Environ.DB.GetUser(username)
	if err != nil {
		// Cannot find the user, so redirect to the login page
		log.Printf("Error retrieving user from datastore: %v\n", err)
		http.Redirect(w, r, "/notfound", http.StatusTemporaryRedirect)
		return
	}

	// verify role value is valid
	if User.Role != datastore.Standard && User.Role != datastore.Admin && User.Role != datastore.Superuser {
		log.Printf("Role obtained from database for user %v has not a valid value: %v\n", username, User.Role)
		http.Redirect(w, r, "/notfound", http.StatusTemporaryRedirect)
		return
	}

	// Build the JWT
	jwtToken, err := NewJWTToken(resp, User.Role)
	if err != nil {
		// Unexpected that this should occur, so leave the detailed response
		log.Printf("Error creating the JWT: %v", err)
		replyHTTPError(w, http.StatusBadRequest, err)
		return
	}

	// Set a cookie with the JWT
	AddJWTCookie(jwtToken, w)

	// Redirect to the homepage with the JWT
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func isComma(c rune) bool {
	return c == ','
}

var errorTemplate = template.Must(template.New("failure").Parse(`<html>
<head><title>Login Error</title></head>
<body>{{.}}</body>
</html>
`))

// LogoutHandler logs the user out by removing the cookie and the JWT authorization header
func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	// Remove the authorization header with contains the bearer token
	w.Header().Del("Authorization")

	// Create a new invalid token with an unauthorized user
	jwtToken, err := createJWT("INVALID", "Not Logged-In", "", "", 0, 0)
	if err != nil {
		log.Println("Error logging out:", err.Error())
	}

	// Update the cookie with the invalid token and expired date
	c, err := r.Cookie(JWTCookie)
	if err != nil {
		log.Println("Error logging out:", err.Error())
	}
	c.Value = jwtToken
	c.Expires = time.Now().AddDate(0, 0, -1)

	// Set the bearer token and the cookie
	http.SetCookie(w, c)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
