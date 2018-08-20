// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
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
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CanonicalLtd/serial-vault/account"
	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/assertion"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/snapcore/snapd/asserts"
	check "gopkg.in/check.v1"
)

func TestAssertionSuite(t *testing.T) { check.TestingT(t) }

type AssertionSuite struct{}

type AssertionTest struct {
	Data   []byte
	Code   int
	Type   string
	APIKey string
}

var _ = check.Suite(&AssertionSuite{})

func (s *AssertionSuite) SetUpTest(c *check.C) {
	// Mock the store
	account.FetchAssertionFromStore = account.MockFetchAssertionFromStore
	service.MiddlewareWithCSRF = service.Middleware

	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)
}

func (s *AssertionSuite) sendRequest(method, url string, data io.Reader, apiKey string, c *check.C) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)
	r.Header.Set("api-key", apiKey)

	service.SigningRouter().ServeHTTP(w, r)

	return w
}

func (s *AssertionSuite) TestAssertionHandler(c *check.C) {
	tests := []AssertionTest{
		{nil, 400, response.JSONHeader, "ValidAPIKey"},
		{[]byte{}, 400, response.JSONHeader, "ValidAPIKey"},
		{validModel(), 200, asserts.MediaType, "ValidAPIKey"},
		{validModel(), 400, response.JSONHeader, "InvalidAPIKey"},
		{classicModel(), 200, asserts.MediaType, "ValidAPIKey"},
		{classicModel(), 400, response.JSONHeader, "InvalidAPIKey"},
		{invalidModel(), 400, response.JSONHeader, "ValidAPIKey"},
		{unauthBrand(), 400, response.JSONHeader, "ValidAPIKey"},
		{unknownBrand(), 400, response.JSONHeader, "ValidAPIKey"},
	}

	for _, t := range tests {
		w := s.sendRequest("POST", "/v1/model", bytes.NewReader(t.Data), t.APIKey, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)
	}

}

func (s *AssertionSuite) TestAssertionErrorHandler(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}
	// Mock the store with an error
	account.FetchAssertionFromStore = account.MockFetchAssertionFromStoreError

	tests := []AssertionTest{
		{validModel(), 400, response.JSONHeader, "ValidAPIKey"},
	}

	for _, t := range tests {
		w := s.sendRequest("POST", "/v1/model", bytes.NewReader(t.Data), t.APIKey, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)
	}

}

func validModel() []byte {
	a := assertion.ModelAssertionRequest{
		BrandID: "system",
		Name:    "alder",
	}
	d, _ := json.Marshal(a)
	return d
}

func invalidModel() []byte {
	a := assertion.ModelAssertionRequest{
		BrandID: "system",
		Name:    "invalid",
	}
	d, _ := json.Marshal(a)
	return d
}

func unauthBrand() []byte {
	a := assertion.ModelAssertionRequest{
		BrandID: "vendor",
		Name:    "alder",
	}
	d, _ := json.Marshal(a)
	return d
}

func unknownBrand() []byte {
	a := assertion.ModelAssertionRequest{
		BrandID: "unknown",
		Name:    "alder",
	}
	d, _ := json.Marshal(a)
	return d
}

func classicModel() []byte {
	a := assertion.ModelAssertionRequest{
		BrandID: "system",
		Name:    "ash",
	}
	d, _ := json.Marshal(a)
	return d
}
