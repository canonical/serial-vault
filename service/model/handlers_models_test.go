// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2018 Canonical Ltd
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

package model_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/model"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/CanonicalLtd/serial-vault/usso"
	"github.com/juju/usso/openid"
	check "gopkg.in/check.v1"
)

func TestModelsSuite(t *testing.T) { check.TestingT(t) }

type ModelsSuite struct{}

type SuiteTest struct {
	MockError   bool
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

var _ = check.Suite(&ModelsSuite{})

func (s *ModelsSuite) SetUpTest(c *check.C) {
	// Mock the database
	config := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.Environ.KeypairDB, _ = datastore.GetErrorMockKeyStore(config)

	// Disable CSRF for tests as we do not have a secure connection
	service.MiddlewareWithCSRF = service.Middleware
}

func sendAdminRequest(method, url string, data io.Reader, permissions int, c *check.C) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	if datastore.Environ.Config.EnableUserAuth {
		// Create a JWT and add it to the request
		err := createJWTWithRole(r, permissions)
		c.Assert(err, check.IsNil)
	}

	service.AdminRouter().ServeHTTP(w, r)

	return w
}

func (s *ModelsSuite) sendPOSTRequest(url string, data io.Reader, permissions int) (*httptest.ResponseRecorder, error) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", url, data)

	if datastore.Environ.Config.EnableUserAuth {
		// Create a JWT and add it to the request
		err := createJWTWithRole(r, permissions)
		if err != nil {
			return nil, err
		}
	}

	service.AdminRouter().ServeHTTP(w, r)
	return w, nil
}

func parseListResponse(w *httptest.ResponseRecorder) (model.ListResponse, error) {
	// Check the JSON response
	result := model.ListResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func parseInstanceResponse(w *httptest.ResponseRecorder) (model.InstanceResponse, error) {
	// Check the JSON response
	result := model.InstanceResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func (s *ModelsSuite) TestListHandler(c *check.C) {
	tests := []SuiteTest{
		{false, "GET", "/v1/models", nil, 200, "application/json; charset=UTF-8", 0, false, true, 6},
		{false, "GET", "/v1/models", nil, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 3},
		{false, "GET", "/v1/models", nil, 400, "application/json; charset=UTF-8", datastore.Invalid, true, false, 0},
		{true, "GET", "/v1/models", nil, 400, "application/json; charset=UTF-8", datastore.Invalid, false, false, 0},
	}

	for _, t := range tests {
		datastore.Environ.Config.EnableUserAuth = t.EnableAuth
		if t.MockError {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}

		w := sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := parseListResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)
		c.Assert(len(result.Models), check.Equals, t.List)
		if t.List > 0 {
			c.Assert(result.Models[0].Name, check.Equals, "alder")
		}

		datastore.Environ.Config.EnableUserAuth = true
		if t.MockError {
			datastore.Environ.DB = &datastore.MockDB{}
		}
	}
}

func (s *ModelsSuite) TestGetHandler(c *check.C) {
	tests := []SuiteTest{
		{false, "GET", "/v1/models/1", nil, 200, "application/json; charset=UTF-8", 0, false, true, 0},
		{false, "GET", "/v1/models/1", nil, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 0},
		{false, "GET", "/v1/models/1", nil, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
		{false, "GET", "/v1/models/999999", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "GET", "/v1/models/999999999999999999999999999999", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{true, "GET", "/v1/models/1", nil, 400, "application/json; charset=UTF-8", 0, false, false, 0},

		// Admin API tests
		{false, "GET", "/api/models/1", nil, 400, "application/json; charset=UTF-8", 0, false, false, 0},
		{false, "GET", "/api/models/1", nil, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 0},
		{false, "GET", "/api/models/1", nil, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
		{false, "GET", "/api/models/999999", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "GET", "/api/models/999999999999999999999999999999", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{true, "GET", "/api/models/1", nil, 400, "application/json; charset=UTF-8", 0, false, false, 0},
	}

	for _, t := range tests {
		datastore.Environ.Config.EnableUserAuth = t.EnableAuth
		if t.MockError {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}

		var w *httptest.ResponseRecorder
		if strings.Contains(t.URL, "api") {
			w = sendAdminAPIRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		} else {
			w = sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		}
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := parseInstanceResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)
		if t.Success {
			c.Assert(result.Model.ID, check.Equals, 1)
			c.Assert(result.Model.Name, check.Equals, "alder")
		}

		datastore.Environ.Config.EnableUserAuth = true
		if t.MockError {
			datastore.Environ.DB = &datastore.MockDB{}
		}
	}
}

func (s *ModelsSuite) TestUpdateDeleteHandler(c *check.C) {
	data := `
	{
		"id": 1,
		"brand-id": "System",
		"model":"the-model",
		"serial":"A1234-L",
		"device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
	}`
	dataNotFound := `
	{
		"id": 5,
		"brand-id": "System",
		"model":"the-model",
		"serial":"A1234-L",
		"device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
	}`
	dataExists := `
	{
		"id": 1,
		"brand-id": "system",
		"model":"ash",
		"serial":"A1234-L",
		"device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
	}`

	// Define a model linked with the signing-key as JSON
	model := datastore.Model{BrandID: "System", Name: "the-model", KeypairID: 1}
	newData, _ := json.Marshal(model)

	tests := []SuiteTest{
		{false, "PUT", "/v1/models/1", []byte(data), 200, "application/json; charset=UTF-8", 0, false, true, 0},
		{false, "PUT", "/v1/models/1", []byte(data), 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 0},
		{false, "PUT", "/v1/models/1", []byte(dataExists), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "PUT", "/v1/models/1", []byte(dataNotFound), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "PUT", "/v1/models/1", []byte(data), 400, "application/json; charset=UTF-8", datastore.Invalid, true, false, 0},
		{false, "PUT", "/v1/models/1", []byte(data), 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
		{false, "PUT", "/v1/models/999999999999999999999999999999", []byte(data), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "PUT", "/v1/models/5", []byte(dataNotFound), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "PUT", "/v1/models/5", []byte(""), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "PUT", "/v1/models/5", []byte("bad"), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{true, "PUT", "/v1/models/1", []byte(data), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{true, "PUT", "/v1/models/1", nil, 400, "application/json; charset=UTF-8", 0, false, false, 0},

		{false, "DELETE", "/v1/models/1", nil, 200, "application/json; charset=UTF-8", 0, false, true, 0},
		{false, "DELETE", "/v1/models/1", nil, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 0},
		{false, "DELETE", "/v1/models/1", nil, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
		{false, "DELETE", "/v1/models/5", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "DELETE", "/v1/models/999999999999999999999999999999", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{true, "DELETE", "/v1/models/1", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},

		{false, "POST", "/v1/models", []byte(newData), 200, "application/json; charset=UTF-8", 0, false, true, 0},
		{false, "POST", "/v1/models", []byte(newData), 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 0},
		{false, "POST", "/v1/models", []byte(""), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "POST", "/v1/models", []byte("bad"), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{true, "POST", "/v1/models", []byte(newData), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},

		// Admin API
		{false, "PUT", "/api/models/1", []byte(data), 400, "application/json; charset=UTF-8", 0, false, false, 0},
		{false, "PUT", "/api/models/1", []byte(data), 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 0},
		{false, "PUT", "/api/models/1", []byte(dataNotFound), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "PUT", "/api/models/1", []byte(data), 400, "application/json; charset=UTF-8", datastore.Invalid, true, false, 0},
		{false, "PUT", "/api/models/1", []byte(data), 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
		{false, "PUT", "/api/models/999999999999999999999999999999", []byte(data), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "PUT", "/api/models/5", []byte(dataNotFound), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "PUT", "/api/models/5", []byte(""), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "PUT", "/api/models/5", []byte("bad"), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{true, "PUT", "/api/models/1", []byte(data), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{true, "PUT", "/api/models/1", nil, 400, "application/json; charset=UTF-8", 0, false, false, 0},

		{false, "DELETE", "/api/models/1", nil, 400, "application/json; charset=UTF-8", 0, false, false, 0},
		{false, "DELETE", "/api/models/1", nil, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 0},
		{false, "DELETE", "/api/models/1", nil, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
		{false, "DELETE", "/api/models/5", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "DELETE", "/api/models/999999999999999999999999999999", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{true, "DELETE", "/api/models/1", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},

		{false, "POST", "/api/models", []byte(newData), 400, "application/json; charset=UTF-8", 0, false, false, 0},
		{false, "POST", "/api/models", []byte(newData), 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 0},
		{false, "POST", "/api/models", []byte(""), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "POST", "/api/models", []byte("bad"), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{true, "POST", "/api/models", []byte(newData), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
	}

	for _, t := range tests {
		datastore.Environ.Config.EnableUserAuth = t.EnableAuth
		if t.MockError {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}

		var w *httptest.ResponseRecorder
		if strings.Contains(t.URL, "api") {
			w = sendAdminAPIRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		} else {
			w = sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		}
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := response.ParseStandardResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)

		datastore.Environ.Config.EnableUserAuth = true
		if t.MockError {
			datastore.Environ.DB = &datastore.MockDB{}
		}
	}
}

func (s *ModelsSuite) TestCreateHandlerReturnModel(c *check.C) {
	model := datastore.Model{BrandID: "System", Name: "the-model", KeypairID: 1}
	newData, _ := json.Marshal(model)

	datastore.Environ.Config.EnableUserAuth = false
	w := sendAdminRequest("POST", "/v1/models", bytes.NewReader(newData), 0, c)
	c.Assert(w.Code, check.Equals, 200)
	c.Assert(w.Header().Get("Content-Type"), check.Equals, "application/json; charset=UTF-8")

	result, err := parseInstanceResponse(w)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, true)
	// return model from DB, ID is set
	c.Assert(result.Model.ID > 0, check.Equals, true)
	c.Assert(result.Model.BrandID, check.Equals, model.BrandID)
	c.Assert(result.Model.Name, check.Equals, model.Name)
}

func (s *ModelsSuite) TestAssertionHandler(c *check.C) {
	d := datastore.ModelAssertion{
		ModelID: 1, KeypairID: 1,
		Series: 16, Architecture: "amd64", Revision: 1,
		Gadget: "mygadget", Kernel: "mykernel", Store: "ubuntu",
	}
	data, _ := json.Marshal(d)

	d = datastore.ModelAssertion{
		ModelID: 999, KeypairID: 1,
		Series: 16, Architecture: "amd64", Revision: 1,
		Gadget: "mygadget", Kernel: "mykernel", Store: "ubuntu",
	}
	dataInvalid, _ := json.Marshal(d)

	tests := []SuiteTest{
		{false, "POST", "/v1/models/assertion", []byte(data), 200, "application/json; charset=UTF-8", 0, false, true, 0},
		{false, "POST", "/v1/models/assertion", []byte(data), 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 0},
		{false, "POST", "/v1/models/assertion", []byte(data), 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
		{false, "POST", "/v1/models/assertion", []byte(""), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "POST", "/v1/models/assertion", []byte(dataInvalid), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "POST", "/v1/models/assertion", []byte("bad"), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{true, "POST", "/v1/models/assertion", []byte(data), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{true, "POST", "/v1/models/assertion", []byte(data), 400, "application/json; charset=UTF-8", 0, false, false, 0},

		// Admin API
		{false, "POST", "/api/models/assertion", []byte(data), 400, "application/json; charset=UTF-8", 0, false, false, 0},
		{false, "POST", "/api/models/assertion", []byte(data), 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 0},
		{false, "POST", "/api/models/assertion", []byte(data), 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
		{false, "POST", "/api/models/assertion", []byte(""), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "POST", "/api/models/assertion", []byte(dataInvalid), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{false, "POST", "/api/models/assertion", []byte("bad"), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{true, "POST", "/api/models/assertion", []byte(data), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{true, "POST", "/api/models/assertion", []byte(data), 400, "application/json; charset=UTF-8", 0, false, false, 0},
	}

	// Mock the database and the keystore
	config := config.Settings{KeyStoreType: "memory", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	for _, t := range tests {
		datastore.Environ.KeypairDB, _ = datastore.TestMemoryKeyStore(config)
		datastore.Environ.Config.EnableUserAuth = t.EnableAuth
		if t.MockError {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}

		var w *httptest.ResponseRecorder
		if strings.Contains(t.URL, "api") {
			w = sendAdminAPIRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		} else {
			w = sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		}
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := response.ParseStandardResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)

		datastore.Environ.Config.EnableUserAuth = true
		if t.MockError {
			datastore.Environ.DB = &datastore.MockDB{}
		}
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
