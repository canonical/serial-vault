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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/juju/usso/openid"
)

type testJWT struct {
	resp     openid.Response
	expected []string
}

func TestNewJWTToken(t *testing.T) {
	test1 := testJWT{
		resp:     openid.Response{ID: "id", Teams: []string{"teamone", "team2"}},
		expected: []string{"", "", ""},
	}
	test2 := testJWT{
		resp:     openid.Response{ID: "id", Teams: []string{"teamone", "team2"}, SReg: map[string]string{"nickname": "jwt"}},
		expected: []string{"jwt", "", ""},
	}
	test3 := testJWT{
		resp:     openid.Response{ID: "id", Teams: []string{"teamone", "team2"}, SReg: map[string]string{"nickname": "jwt", "email": "jwt@example.com", "fullname": "John W Thompson"}},
		expected: []string{"jwt", "jwt@example.com", "John W Thompson"},
	}

	for _, r := range []testJWT{test1, test2, test3} {

		jwtToken, err := NewJWTToken(&r.resp)
		if err != nil {
			t.Errorf("Error creating JWT: %v", err)
		}

		expectedToken(t, jwtToken, &r.resp, r.expected[0], r.expected[1], r.expected[2])

	}
}

func expectedToken(t *testing.T, jwtToken string, resp *openid.Response, username, email, name string) {
	token, err := VerifyJWT(jwtToken)
	if err != nil {
		t.Errorf("Error validating JWT: %v", err)
	}

	if token.Claims[ClaimsIdentity] != resp.ID {
		t.Errorf("JWT ID does not match: %v", token.Claims[ClaimsIdentity])
	}
	if token.Claims[ClaimsUsername] != username {
		t.Errorf("JWT username does not match: %v", token.Claims[ClaimsUsername])
	}
	if token.Claims[ClaimsEmail] != email {
		t.Errorf("JWT email does not match: %v", token.Claims[ClaimsEmail])
	}
	if token.Claims[ClaimsName] != name {
		t.Errorf("JWT name does not match: %v", token.Claims[ClaimsName])
	}
}

func testHandler(w http.ResponseWriter, r *http.Request) {

}

func TestAddJWTCookie(t *testing.T) {
	w := httptest.NewRecorder()
	AddJWTCookie("ThisShouldBeAJWT", w)

	// Copy the Cookie over to a new Request
	request := &http.Request{Header: http.Header{"Cookie": w.HeaderMap["Set-Cookie"]}}

	// Extract the dropped cookie from the request
	jwtToken, err := JWTExtractor(request)
	if err != nil {
		t.Errorf("Error getting the JWT cookie: %v", err)
	}
	if jwtToken != "ThisShouldBeAJWT" {
		t.Errorf("Expected 'ThisShouldBeAJWT', got '%v'", jwtToken)
	}
}
