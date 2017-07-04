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
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"fmt"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/juju/usso"
	"github.com/juju/usso/openid"
)

func TestLoginHandlerUSSORedirect(t *testing.T) {

	// Mock the database
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/login", nil)
	http.HandlerFunc(LoginHandler).ServeHTTP(w, r)

	if w.Code != http.StatusFound {
		t.Errorf("Expected HTTP status '302', got: %v", w.Code)
	}

	u, err := url.Parse(w.Header().Get("Location"))
	if err != nil {
		t.Errorf("Error Parsing the redirect URL: %v", u)
	}

	// Check that the redirect is to the Ubuntu SSO service
	url := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	if url != usso.ProductionUbuntuSSOServer.LoginURL() {
		t.Errorf("Unexpected redirect URL: %v", url)
	}
}

func TestLoginHandlerReturn(t *testing.T) {
	// Response parameters from OpenID login
	const url = "/login?openid.ns=http://specs.openid.net/auth/2.0&openid.mode=id_res&openid.op_endpoint=https://login.ubuntu.com/%2Bopenid&openid.claimed_id=https://login.ubuntu.com/%2Bid/AAAAAA&openid.identity=https://login.ubuntu.com/%2Bid/AAAAAA&openid.return_to=http://return.to&openid.response_nonce=2005-05-15T17:11:51ZUNIQUE&openid.assoc_handle=1&openid.signed=op_endpoint,return_to,response_nonce,assoc_handle,claimed_id,identity,sreg.email,sreg.fullname&openid.sig=AAAA&openid.ns.sreg=http://openid.net/extensions/sreg/1.1&openid.sreg.email=a@example.org&openid.sreg.fullname=A&openid.sreg.nickname=a"

	// Mock the database and OpenID verification
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}
	verify = verifySuccess

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", url, nil)
	http.HandlerFunc(LoginHandler).ServeHTTP(w, r)

	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected HTTP status '307', got: %v", w.Code)
	}

	// Copy the headers over to a new Request
	request := &http.Request{Header: w.Header()}

	// Extract the cookie from the request
	jwtToken, err := JWTExtractor(request)
	if err != nil {
		t.Errorf("Error getting the JWT cookie: %v", err)
	}

	// Check the JWT details
	response, _ := verifySuccess(url)

	// Get User Role
	user, err := datastore.Environ.DB.GetUser(response.SReg["nickname"])

	if err != nil {
		t.Errorf("Could not get datastore user %v: %v\n", response.SReg["nickname"], err)
	}

	expectedToken(t, jwtToken, response, response.SReg["nickname"], response.SReg["email"], response.SReg["fullname"], user.Role)
}

func TestLoginHandlerReturnFail(t *testing.T) {
	// Response parameters from OpenID login
	const url = "/login?openid.ns=http://specs.openid.net/auth/2.0&openid.mode=id_res&openid.op_endpoint=https://login.ubuntu.com/%2Bopenid&openid.claimed_id=https://login.ubuntu.com/%2Bid/AAAAAA&openid.identity=https://login.ubuntu.com/%2Bid/AAAAAA&openid.return_to=http://return.to&openid.response_nonce=2005-05-15T17:11:51ZUNIQUE&openid.assoc_handle=1&openid.signed=op_endpoint,return_to,response_nonce,assoc_handle,claimed_id,identity,sreg.email,sreg.fullname&openid.sig=AAAA&openid.ns.sreg=http://openid.net/extensions/sreg/1.1&openid.sreg.email=a@example.org&openid.sreg.fullname=A&openid.sreg.nickname=a"

	// Mock the database and OpenID verification
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}
	verify = verifyFail

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", url, nil)
	http.HandlerFunc(LoginHandler).ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected HTTP status '400', got: %v", w.Code)
	}
}

func verifySuccess(requestURL string) (*openid.Response, error) {
	params := make(map[string]string)

	tokens := strings.Split(requestURL, "&")
	for _, t := range tokens {
		tks := strings.Split(t, "=")

		if len(tks) == 2 {
			params[strings.TrimSpace(tks[0])] = tks[1]
		}
	}

	r := openid.Response{
		ID:    params["openid.sig"],
		Teams: []string{"ce-web-logs"},
		SReg: map[string]string{
			"nickname": params["openid.sreg.nickname"],
			"fullname": params["openid.sreg.fullname"],
			"email":    params["openid.sreg.email"],
		},
	}

	return &r, nil
}

func verifyFail(requestURL string) (*openid.Response, error) {
	return nil, errors.New("MOCK error from OpenID verification")
}
