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

package service

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

	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)
}

func (s *AssertionSuite) sendRequest(method, url string, data io.Reader, apiKey string, c *check.C) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)
	r.Header.Set("api-key", apiKey)

	SigningRouter().ServeHTTP(w, r)

	return w
}

func (s *AssertionSuite) TestAssertionHandler(c *check.C) {
	tests := []AssertionTest{
		AssertionTest{nil, 400, "application/json; charset=UTF-8", "ValidAPIKey"},
		AssertionTest{[]byte{}, 400, "application/json; charset=UTF-8", "ValidAPIKey"},
		AssertionTest{validModel(), 200, asserts.MediaType, "ValidAPIKey"},
		AssertionTest{invalidModel(), 400, "application/json; charset=UTF-8", "ValidAPIKey"},
		AssertionTest{unauthBrand(), 400, "application/json; charset=UTF-8", "ValidAPIKey"},
		AssertionTest{unknownBrand(), 400, "application/json; charset=UTF-8", "ValidAPIKey"},
	}

	for _, t := range tests {
		w := s.sendRequest("POST", "/v1/model", bytes.NewReader(t.Data), t.APIKey, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)
	}

}

func validModel() []byte {
	a := ModelAssertionRequest{
		BrandID: "system",
		Name:    "alder",
	}
	d, _ := json.Marshal(a)
	return d
}

func invalidModel() []byte {
	a := ModelAssertionRequest{
		BrandID: "system",
		Name:    "invalid",
	}
	d, _ := json.Marshal(a)
	return d
}

func unauthBrand() []byte {
	a := ModelAssertionRequest{
		BrandID: "vendor",
		Name:    "alder",
	}
	d, _ := json.Marshal(a)
	return d
}

func unknownBrand() []byte {
	a := ModelAssertionRequest{
		BrandID: "unknown",
		Name:    "alder",
	}
	d, _ := json.Marshal(a)
	return d
}
