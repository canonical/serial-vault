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

package pivot_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CanonicalLtd/serial-vault/service/assertion"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/pivot"
	"github.com/snapcore/snapd/asserts"
	check "gopkg.in/check.v1"
)

func TestPivotSuite(t *testing.T) { check.TestingT(t) }

type PivotSuite struct{}

type PivotTest struct {
	Method  string
	URL     string
	Data    []byte
	Code    int
	Type    string
	APIKey  string
	Success bool
}

var _ = check.Suite(&PivotSuite{})

const jsonType = "application/json; charset=UTF-8"

func parsePivotResponse(w *httptest.ResponseRecorder) (pivot.Response, error) {
	// Check the JSON response
	result := pivot.Response{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func (s *PivotSuite) SetUpTest(c *check.C) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)
}

func (s *PivotSuite) TestPivotModelHandler(c *check.C) {

	tests := []PivotTest{
		{"POST", "/v1/pivot", nil, 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivot", []byte{}, 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivot", []byte("invalid"), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivot", []byte(pivot.SerialAssert), 200, jsonType, "ValidAPIKey", true},
		{"POST", "/v1/pivot", []byte(pivot.SerialAssert), 400, jsonType, "InvalidAPIKey", false},
		{"POST", "/v1/pivot", []byte(pivot.SerialAssertInvalid), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivot", []byte(pivot.SerialAssertInvalidBrand), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivot", []byte(pivot.SerialAssertNonReseller), 400, jsonType, "ValidAPIKey", false},
	}

	for _, t := range tests {
		w := sendSigningRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.APIKey, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := parsePivotResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)
	}

}

func (s *PivotSuite) TestPivotModelSerialAssertionHandler(c *check.C) {

	tests := []PivotTest{
		{"POST", "/v1/pivotmodel", nil, 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotmodel", []byte{}, 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotmodel", []byte("invalid"), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotmodel", []byte(pivot.AssertionWrongType), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotmodel", []byte(pivot.SerialAssert), 200, asserts.MediaType, "ValidAPIKey", true},
		{"POST", "/v1/pivotmodel", []byte(pivot.SerialAssert), 400, jsonType, "InvalidAPIKey", false},
		{"POST", "/v1/pivotmodel", []byte(pivot.SerialAssertInvalid), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotmodel", []byte(pivot.SerialAssertInvalidBrand), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotmodel", []byte(pivot.SerialAssertNonReseller), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotserial", nil, 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotserial", []byte{}, 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotserial", []byte("invalid"), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotserial", []byte(pivot.SerialAssert), 200, asserts.MediaType, "ValidAPIKey", true},
		{"POST", "/v1/pivotserial", []byte(pivot.SerialAssert), 400, jsonType, "InvalidAPIKey", false},
		{"POST", "/v1/pivotserial", []byte(pivot.SerialAssertInvalid), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotserial", []byte(pivot.SerialAssertInvalidBrand), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotserial", []byte(pivot.SerialAssertNonReseller), 400, jsonType, "ValidAPIKey", false},
	}

	for _, t := range tests {
		w := sendSigningRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.APIKey, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		if t.Type == jsonType {
			result, err := parsePivotResponse(w)
			c.Assert(err, check.IsNil)
			c.Assert(result.Success, check.Equals, t.Success)
		}
	}

}

func (s *PivotSuite) TestPivotSystemUserAssertionHandler(c *check.C) {
	r := assertion.PivotSystemUserRequest{
		SystemUserRequest: assertion.SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", Since: "2017-03-24T12:34:00Z"},
		Brand:             "system", ModelName: "alder", SerialNumber: "abcd1234",
	}
	req, _ := json.Marshal(r)

	tests := []PivotTest{
		{"POST", "/v1/pivotuser", nil, 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotuser", []byte{}, 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotuser", []byte("invalid"), 400, jsonType, "ValidAPIKey", false},
		{"POST", "/v1/pivotuser", req, 200, asserts.MediaType, "ValidAPIKey", true},
	}

	for _, t := range tests {
		w := sendSigningRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.APIKey, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		if t.Type == jsonType {
			result, err := parsePivotResponse(w)
			c.Assert(err, check.IsNil)
			c.Assert(result.Success, check.Equals, t.Success)
		}
	}
}

func sendSigningRequest(method, url string, data io.Reader, apiKey string, c *check.C) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)
	r.Header.Set("api-key", apiKey)

	service.SigningRouter().ServeHTTP(w, r)

	return w
}
