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

package core_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/core"
	"github.com/CanonicalLtd/serial-vault/service/response"
	check "gopkg.in/check.v1"
)

func TestCoreSuite(t *testing.T) { check.TestingT(t) }

type CoreSuite struct{}

type SuiteTest struct {
	MockError bool
	Method    string
	URL       string
	Data      []byte
	Code      int
	Type      string
	Result    string
	Success   bool
}

var _ = check.Suite(&CoreSuite{})

func (s *CoreSuite) SetUpTest(c *check.C) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)
}

func sendRequest(method, url string, data io.Reader, c *check.C) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	service.SigningRouter().ServeHTTP(w, r)

	return w
}

func sendAdminRequest(method, url string, data io.Reader, c *check.C) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	service.AdminRouter().ServeHTTP(w, r)

	return w
}

func (s *CoreSuite) TestVersionHandler(c *check.C) {
	tests := []SuiteTest{
		{false, "GET", "/v1/version", nil, 200, response.JSONHeader, "", true},
	}

	for _, t := range tests {
		if t.MockError {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}

		w := sendRequest(t.Method, t.URL, bytes.NewReader(t.Data), c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		datastore.Environ.DB = &datastore.MockDB{}
	}
}

func (s *CoreSuite) TestTokenHandler(c *check.C) {
	tests := []SuiteTest{
		{false, "GET", "/v1/authtoken", nil, 200, response.JSONHeader, "", true},
		{false, "GET", "/v1/token", nil, 200, response.JSONHeader, "", true},
	}

	for _, t := range tests {
		if t.MockError {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}

		w := sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		datastore.Environ.DB = &datastore.MockDB{}
	}
}

func (s *CoreSuite) TestHealthHandler(c *check.C) {
	tests := []SuiteTest{
		{false, "GET", "/v1/health", nil, 200, response.JSONHeader, "healthy", true},
		{true, "GET", "/v1/health", nil, 400, response.JSONHeader, "", false},
	}

	for _, t := range tests {
		if t.MockError {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}

		w := sendRequest(t.Method, t.URL, bytes.NewReader(t.Data), c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result := core.HealthResponse{}
		err := json.NewDecoder(w.Body).Decode(&result)
		c.Assert(err, check.IsNil)
		if t.Success {
			c.Assert(result.Database, check.Equals, t.Result)
		}

		datastore.Environ.DB = &datastore.MockDB{}
	}
}
