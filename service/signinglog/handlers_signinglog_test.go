// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018-2019 Canonical Ltd
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

package signinglog_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/signinglog"
	"github.com/CanonicalLtd/serial-vault/usso"
	"github.com/juju/usso/openid"
	check "gopkg.in/check.v1"
)

func TestSigningLogSuite(t *testing.T) { check.TestingT(t) }

type SigningLogSuite struct{}

var _ = check.Suite(&SigningLogSuite{})

type SigningLogTest struct {
	Method      string
	URL         string
	Data        []byte
	Code        int
	Type        string
	Permissions int
	EnableAuth  bool
	Success     bool
	List        int
}

func (s *SigningLogSuite) SetUpTest(c *check.C) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)

	// Disable CSRF for tests as we do not have a secure connection
	service.MiddlewareWithCSRF = service.Middleware
}

func (s *SigningLogSuite) TestSigningLogHandler(c *check.C) {
	tests := []SigningLogTest{
		{"GET", "/v1/signinglog", nil, 200, "application/json; charset=UTF-8", 0, false, true, 10},
		{"GET", "/v1/signinglog", nil, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 4},
		{"GET", "/v1/signinglog", nil, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
		{"GET", "/v1/signinglog", nil, 400, "application/json; charset=UTF-8", 0, true, false, 0},
	}

	for _, t := range tests {
		if t.EnableAuth {
			datastore.Environ.Config.EnableUserAuth = true
		}

		w := sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := parseListResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)
		c.Assert(len(result.SigningLog), check.Equals, t.List)

		datastore.Environ.Config.EnableUserAuth = false
	}
}

func (s *SigningLogSuite) TestSigningLogErrorHandler(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}
	tests := []SigningLogTest{
		{"GET", "/v1/signinglog", nil, 400, "application/json; charset=UTF-8", 0, false, false, 0},
		{"GET", "/v1/signinglog", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
	}

	for _, t := range tests {
		if t.EnableAuth {
			datastore.Environ.Config.EnableUserAuth = true
		}

		w := sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := parseListResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)
		c.Assert(len(result.SigningLog), check.Equals, t.List)

		datastore.Environ.Config.EnableUserAuth = false
	}
}

func parseListResponse(w *httptest.ResponseRecorder) (signinglog.ListResponse, error) {
	// Check the JSON response
	result := signinglog.ListResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func sendAdminRequest(method, url string, data io.Reader, permissions int, c *check.C) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	if permissions > 0 {
		// Create a JWT and add it to the request
		err := createJWTWithRole(r, permissions)
		c.Assert(err, check.IsNil)
	}

	service.AdminRouter().ServeHTTP(w, r)

	return w
}

func createJWTWithRole(r *http.Request, role int) error {
	sreg := map[string]string{"nickname": "sv", "fullname": "Steven Vault", "email": "sv@example.com"}
	resp := openid.Response{ID: "identity", Teams: []string{}, SReg: sreg}
	jwtToken, err := usso.NewJWTToken(&resp, role)
	if err != nil {
		return fmt.Errorf("Error creating a JWT: %v", err)
	}
	r.Header.Set("Authorization", "Bearer "+jwtToken)
	return nil
}
