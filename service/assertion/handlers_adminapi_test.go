// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
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

package assertion_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/pivot"
	"github.com/CanonicalLtd/serial-vault/service/response"
	check "gopkg.in/check.v1"
)

type SuiteTest struct {
	Method      string
	URL         string
	Data        []byte
	Code        int
	Type        string
	Permissions int
	EnableAuth  bool
	Success     bool
	SkipJWT     bool
	MockError   bool
}

func (s *AssertionSuite) TestAPISystemUserHandler(c *check.C) {
	tests := []SuiteTest{
		{"POST", "/api/assertions", []byte(generateSystemUserRequest()), 400, "application/json; charset=UTF-8", 0, false, false, false, false},
		{"POST", "/api/assertions", []byte(generateSystemUserRequest()), 200, "application/json; charset=UTF-8", datastore.SyncUser, true, true, false, false},
		{"POST", "/api/assertions", []byte(generateSystemUserRequest()), 200, "application/json; charset=UTF-8", datastore.Standard, true, true, false, false},
		{"POST", "/api/assertions", []byte(generateSystemUserRequestInactiveModel()), 400, "application/json; charset=UTF-8", datastore.Standard, true, false, false, false},
		{"POST", "/api/assertions", []byte(generateSystemUserRequest()), 400, "application/json; charset=UTF-8", 0, true, false, false, false},
	}

	for _, t := range tests {
		if t.EnableAuth {
			datastore.Environ.Config.EnableUserAuth = true
		}

		w := sendAdminAPIRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		datastore.Environ.Config.EnableUserAuth = false
	}
}

func (s *AssertionSuite) TestAPIValidSerialHandler(c *check.C) {
	tests := []SuiteTest{
		{"POST", "/api/assertions/checkserial", nil, 400, response.JSONHeader, 0, false, false, false, false},
		{"POST", "/api/assertions/checkserial", nil, 400, response.JSONHeader, datastore.Admin, false, false, false, false},
		{"POST", "/api/assertions/checkserial", []byte{}, 400, response.JSONHeader, datastore.Admin, false, false, false, false},
		{"POST", "/api/assertions/checkserial", []byte("invalid"), 400, response.JSONHeader, datastore.Admin, false, false, false, false},
		{"POST", "/api/assertions/checkserial", []byte(pivot.AssertionWrongType), 400, response.JSONHeader, datastore.Admin, true, false, false, false},
		{"POST", "/api/assertions/checkserial", []byte(pivot.SerialAssert), 200, response.JSONHeader, datastore.Admin, true, true, false, false},
		{"POST", "/api/assertions/checkserial", []byte(pivot.SerialAssert), 400, response.JSONHeader, datastore.Standard, true, false, false, false},
		{"POST", "/api/assertions/checkserial", []byte(pivot.SerialAssertInvalid), 400, response.JSONHeader, datastore.Admin, true, false, false, false},
		{"POST", "/api/assertions/checkserial", []byte(pivot.SerialAssertInvalidBrand), 400, response.JSONHeader, datastore.Admin, true, false, false, false},
	}

	for _, t := range tests {
		w := sendAdminAPIRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := response.ParseStandardResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)
	}

}

func sendAdminAPIRequest(method, url string, data io.Reader, permissions int, c *check.C) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	switch permissions {
	case datastore.SyncUser:
		r.Header.Set("user", "sync")
		r.Header.Set("api-key", "ValidAPIKey")
	case datastore.Standard:
		r.Header.Set("user", "user1")
		r.Header.Set("api-key", "ValidAPIKey")
	case datastore.Admin:
		r.Header.Set("user", "sv")
		r.Header.Set("api-key", "ValidAPIKey")
	default:
		break
	}

	service.AdminRouter().ServeHTTP(w, r)

	return w
}
