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

package substore_test

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
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/CanonicalLtd/serial-vault/service/substore"
	"github.com/CanonicalLtd/serial-vault/usso"
	"github.com/juju/usso/openid"
	check "gopkg.in/check.v1"
)

func TestSubstoreSuite(t *testing.T) { check.TestingT(t) }

type SubstoreSuite struct{}

type SubstoreTest struct {
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

var _ = check.Suite(&SubstoreSuite{})

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

func sendSigningRequest(method, url string, data io.Reader, apiKey string, c *check.C) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)
	r.Header.Set("api-key", apiKey)

	service.SigningRouter().ServeHTTP(w, r)

	return w
}

func (s *SubstoreSuite) SetUpTest(c *check.C) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)

	// Disable CSRF for tests as we do not have a secure connection
	service.MiddlewareWithCSRF = service.Middleware
}

func parseInstanceResponse(w *httptest.ResponseRecorder) (substore.InstanceResponse, error) {
	// Check the JSON response
	result := substore.InstanceResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func parseListResponse(w *httptest.ResponseRecorder) (substore.ListResponse, error) {
	// Check the JSON response
	result := substore.ListResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func (s *SubstoreSuite) TestSubstoresHandler(c *check.C) {
	tests := []SubstoreTest{
		{"GET", "/v1/accounts/1/stores", nil, 200, "application/json; charset=UTF-8", 0, false, true, 2},
		{"GET", "/v1/accounts/1/stores", nil, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 2},
		{"GET", "/v1/accounts/1/stores", nil, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
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
		c.Assert(len(result.Substores), check.Equals, t.List)

		datastore.Environ.Config.EnableUserAuth = false
	}
}

func (s *SubstoreSuite) TestSubstoresCreateUpdateDeleteHandler(c *check.C) {
	substoreNew := datastore.Substore{AccountID: 1, FromModelID: 1, Store: "mybrand", SerialNumber: "a11112222", ModelName: "alder-mybrand"}
	ssn, _ := json.Marshal(substoreNew)

	substore := datastore.Substore{ID: 1, AccountID: 1, FromModelID: 1, Store: "mybrand", SerialNumber: "a11112222", ModelName: "alder-mybrand"}
	ss, _ := json.Marshal(substore)

	tests := []SubstoreTest{
		{"POST", "/v1/accounts/stores", ssn, 200, "application/json; charset=UTF-8", 0, false, true, 0},
		{"POST", "/v1/accounts/stores", ssn, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 0},
		{"POST", "/v1/accounts/stores", ssn, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
		{"POST", "/v1/accounts/stores", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{"PUT", "/v1/accounts/stores/1", ss, 200, "application/json; charset=UTF-8", 0, false, true, 0},
		{"PUT", "/v1/accounts/stores/1", ss, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 0},
		{"PUT", "/v1/accounts/stores/99", ss, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{"PUT", "/v1/accounts/stores/1", ss, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
		{"PUT", "/v1/accounts/stores/1", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{"DELETE", "/v1/accounts/stores/1", nil, 200, "application/json; charset=UTF-8", 0, false, true, 0},
		{"DELETE", "/v1/accounts/stores/1", nil, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 0},
		{"DELETE", "/v1/accounts/stores/1", nil, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
	}

	for _, t := range tests {
		if t.EnableAuth {
			datastore.Environ.Config.EnableUserAuth = true
		}

		w := sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)

		result, err := response.ParseStandardResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)

		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		datastore.Environ.Config.EnableUserAuth = false
	}
}

func (s *SubstoreSuite) TestSubstoresCreateHandlerReturnSubstore(c *check.C) {
	substoreNew := datastore.Substore{AccountID: 1, FromModelID: 1, Store: "mybrand", SerialNumber: "a11112222", ModelName: "alder-mybrand"}
	ssn, _ := json.Marshal(substoreNew)

	w := sendAdminRequest("POST", "/v1/accounts/stores", bytes.NewReader(ssn), 0, c)

	result, err := parseInstanceResponse(w)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, true)

	c.Assert(w.Code, check.Equals, 200)
	c.Assert(w.Header().Get("Content-Type"), check.Equals, "application/json; charset=UTF-8")
	// Substore is returned in the response, ID was set
	c.Assert(result.Substore.ID > 0, check.Equals, true)
	c.Assert(result.Substore.FromModelID, check.Equals, substoreNew.FromModelID)
	c.Assert(result.Substore.SerialNumber, check.Equals, substoreNew.SerialNumber)
}

func (s *SubstoreSuite) TestSubstoresErrorHandler(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}

	tests := []SubstoreTest{
		{"GET", "/v1/accounts/1/stores", nil, 400, "application/json; charset=UTF-8", 0, false, false, 0},
		{"GET", "/v1/accounts/1/stores", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{"GET", "/v1/accounts/1/stores", nil, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
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
		c.Assert(len(result.Substores), check.Equals, t.List)

		datastore.Environ.Config.EnableUserAuth = false
	}
}

func (s *SubstoreSuite) TestSubstoresUpdateErrorHandler(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}

	substoreNew := datastore.Substore{AccountID: 1, FromModelID: 1, Store: "mybrand", SerialNumber: "a11112222", ModelName: "alder-mybrand"}
	ssn, _ := json.Marshal(substoreNew)

	substore := datastore.Substore{ID: 1, AccountID: 1, FromModelID: 1, Store: "mybrand", SerialNumber: "a11112222", ModelName: "alder-mybrand"}
	ss, _ := json.Marshal(substore)

	tests := []SubstoreTest{
		{"POST", "/v1/accounts/stores", ssn, 400, "application/json; charset=UTF-8", 0, false, false, 0},
		{"POST", "/v1/accounts/stores", ssn, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{"POST", "/v1/accounts/stores", ssn, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
		{"POST", "/v1/accounts/stores", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{"PUT", "/v1/accounts/stores/1", ss, 400, "application/json; charset=UTF-8", 0, false, false, 0},
		{"PUT", "/v1/accounts/stores/1", ss, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{"PUT", "/v1/accounts/stores/1", ss, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
		{"PUT", "/v1/accounts/stores/1", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{"DELETE", "/v1/accounts/stores/1", nil, 400, "application/json; charset=UTF-8", 0, false, false, 0},
		{"DELETE", "/v1/accounts/stores/1", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{"DELETE", "/v1/accounts/stores/1", nil, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
	}

	for _, t := range tests {
		if t.EnableAuth {
			datastore.Environ.Config.EnableUserAuth = true
		}

		w := sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)

		result, err := response.ParseStandardResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)

		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		datastore.Environ.Config.EnableUserAuth = false
	}
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
