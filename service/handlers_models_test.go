// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2017 Canonical Ltd
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

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	check "gopkg.in/check.v1"
)

func TestModelsSuite(t *testing.T) { check.TestingT(t) }

type ModelsSuite struct{}

type ModelsSuiteTest struct {
	Data []byte
	Code int
}

var _ = check.Suite(&ModelsSuite{})

func (s *ModelsSuite) SetUpTest(c *check.C) {
	// Mock the database
	config := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
}

func (s *ModelsSuite) sendGETRequest(url string, permissions int) (*httptest.ResponseRecorder, error) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", url, nil)

	if datastore.Environ.Config.EnableUserAuth {
		// Create a JWT and add it to the request
		err := createJWTWithRole(r, permissions)
		if err != nil {
			return nil, err
		}
	}

	AdminRouter().ServeHTTP(w, r)
	return w, nil
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

	AdminRouter().ServeHTTP(w, r)
	return w, nil
}

func (s *ModelsSuite) parseModelsResponse(w *httptest.ResponseRecorder) (ModelsResponse, error) {
	// Check the JSON response
	result := ModelsResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func (s *ModelsSuite) parseModelResponse(w *httptest.ResponseRecorder) (ModelResponse, error) {
	// Check the JSON response
	result := ModelResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func (s *ModelsSuite) TestModelsHandler(c *check.C) {
	datastore.Environ.Config.EnableUserAuth = false

	w, err := s.sendGETRequest("/v1/models", 0)
	c.Assert(err, check.IsNil)
	c.Assert(w.Code, check.Equals, 200)

	result, err := s.parseModelsResponse(w)
	c.Assert(result.Success, check.Equals, true)
	c.Assert(len(result.Models), check.Equals, 6)
	c.Assert(result.Models[0].Name, check.Equals, "alder")
}

func (s *ModelsSuite) TestModelsHandlerWithPermissions(c *check.C) {
	w, err := s.sendGETRequest("/v1/models", datastore.Admin)
	c.Assert(err, check.IsNil)
	c.Assert(w.Code, check.Equals, 200)

	result, err := s.parseModelsResponse(w)
	c.Assert(result.Success, check.Equals, true)
	c.Assert(len(result.Models), check.Equals, 3)
	c.Assert(result.Models[0].Name, check.Equals, "alder")
}

func (s *ModelsSuite) TestModelsHandlerWithoutPermissions(c *check.C) {
	w, err := s.sendGETRequest("/v1/models", datastore.Invalid)
	c.Assert(err, check.IsNil)
	c.Assert(w.Code, check.Equals, 200)

	result, err := s.parseModelsResponse(w)
	c.Assert(result.Success, check.Equals, false)
	c.Assert(result.ErrorCode, check.Equals, "error-auth")
}

func (s *ModelsSuite) TestModelsHandlerWithError(c *check.C) {
	datastore.Environ.Config.EnableUserAuth = false
	datastore.Environ.DB = &datastore.ErrorMockDB{}

	w, err := s.sendGETRequest("/v1/models", datastore.Invalid)
	c.Assert(err, check.IsNil)
	c.Assert(w.Code, check.Equals, 400)

	result, err := s.parseModelsResponse(w)
	c.Assert(result.Success, check.Equals, false)
}

func (s *ModelsSuite) TestModelGetHandler(c *check.C) {
	w, err := s.sendGETRequest("/v1/models/1", datastore.Admin)
	c.Assert(err, check.IsNil)
	c.Assert(w.Code, check.Equals, 200)

	result, err := s.parseModelResponse(w)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, true)
	c.Assert(result.Model.ID, check.Equals, 1)
	c.Assert(result.Model.Name, check.Equals, "alder")
}

func (s *ModelsSuite) TestModelAssertionHeadersHandler(c *check.C) {
	d := datastore.ModelAssertion{
		ModelID: 1, KeypairID: 1,
		Series: 16, Architecture: "amd64", Revision: 1,
		Gadget: "mygadget", Kernel: "mykernel", Store: "ubuntu",
	}
	data, _ := json.Marshal(d)

	w, err := s.sendPOSTRequest("/v1/models/assertion", bytes.NewReader(data), datastore.Admin)
	c.Assert(err, check.IsNil)
	c.Assert(w.Code, check.Equals, 200)

	result, err := s.parseModelResponse(w)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, true)
}

func (s *ModelsSuite) TestModelAssertionHeadersHandlerWithoutPermissions(c *check.C) {
	d := datastore.ModelAssertion{
		ModelID: 1, KeypairID: 1,
		Series: 16, Architecture: "amd64", Revision: 1,
		Gadget: "mygadget", Kernel: "mykernel", Store: "ubuntu",
	}
	data, _ := json.Marshal(d)

	w, err := s.sendPOSTRequest("/v1/models/assertion", bytes.NewReader(data), datastore.Standard)
	c.Assert(err, check.IsNil)
	c.Assert(w.Code, check.Equals, 200)

	result, err := s.parseModelResponse(w)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, false)
	c.Assert(result.ErrorCode, check.Equals, "error-auth")
}

func (s *ModelsSuite) TestModelAssertionHeadersHandlerInvalid(c *check.C) {
	w, err := s.sendPOSTRequest("/v1/models/assertion", bytes.NewReader([]byte{}), datastore.Admin)
	c.Assert(err, check.IsNil)
	c.Assert(w.Code, check.Equals, 400)

	result, err := s.parseModelResponse(w)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, false)
}

func (s *ModelsSuite) TestModelAssertionHeadersHandlerInvalidModel(c *check.C) {
	d := datastore.ModelAssertion{
		ModelID: 999, KeypairID: 1,
		Series: 16, Architecture: "amd64", Revision: 1,
		Gadget: "mygadget", Kernel: "mykernel", Store: "ubuntu",
	}
	data, _ := json.Marshal(d)

	w, err := s.sendPOSTRequest("/v1/models/assertion", bytes.NewReader(data), datastore.Admin)
	c.Assert(err, check.IsNil)
	c.Assert(w.Code, check.Equals, 400)

	result, err := s.parseModelResponse(w)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, false)
	c.Assert(result.ErrorCode, check.Equals, "error-get-model")
}

// -------------------------------------------------------------------------

func TestModelGetHandlerWithPermissions(t *testing.T) {

	// Mock the database
	c := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	result, _ := sendRequest(t, "GET", "/v1/models/1", nil)

	if result.Model.ID != 1 {
		t.Errorf("Expected model with ID 1, got %d", result.Model.ID)
	}
	if result.Model.Name != "alder" {
		t.Errorf("Expected model name 'alder', got %s", result.Model.Name)
	}
}

func TestModelGetHandlerWithoutPermissions(t *testing.T) {

	// Mock the database
	c := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	sendRequestWithoutPermissions(t, "GET", "/v1/models/1", nil)
}

func TestModelGetHandlerWithError(t *testing.T) {

	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	sendRequestExpectError(t, "GET", "/v1/models/999999", nil)
}

func TestModelGetHandlerWithBadID(t *testing.T) {

	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	sendRequestExpectError(t, "GET", "/v1/models/999999999999999999999999999999", nil)
}

func TestModelUpdateHandler(t *testing.T) {
	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	// Update a model
	data := `
	{
		"id": 1,
		"brand-id": "System",
		"model":"the-model",
		"serial":"A1234-L",
		"device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
	}`

	result, _ := sendRequest(t, "PUT", "/v1/models/1", bytes.NewBufferString(data))

	if result.Model.ID != 1 {
		t.Errorf("Expected model with ID 1, got %d", result.Model.ID)
	}
	if result.Model.Name != "the-model" {
		t.Errorf("Expected model name 'the-model', got %s", result.Model.Name)
	}
}

func TestModelUpdateHandlerWithPermissions(t *testing.T) {
	// Mock the database
	c := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	// Update a model
	data := `
	{
		"id": 1,
		"brand-id": "System",
		"model":"the-model",
		"serial":"A1234-L",
		"device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
	}`

	result, _ := sendRequest(t, "PUT", "/v1/models/1", bytes.NewBufferString(data))

	if result.Model.ID != 1 {
		t.Errorf("Expected model with ID 1, got %d", result.Model.ID)
	}
	if result.Model.Name != "the-model" {
		t.Errorf("Expected model name 'the-model', got %s", result.Model.Name)
	}
}

func TestModelUpdateHandlerWithoutPermissions(t *testing.T) {
	// Mock the database
	c := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	// Update a model
	data := `
	{
		"id": 1,
		"brand-id": "System",
		"model":"the-model",
		"serial":"A1234-L",
		"device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
	}`

	sendRequestWithoutPermissions(t, "PUT", "/v1/models/1", bytes.NewBufferString(data))
}

func TestModelUpdateHandlerWithPermissionsNotFound(t *testing.T) {
	// Mock the database
	c := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	// Update a model
	data := `
	{
		"id": 5,
		"brand-id": "System",
		"model":"the-model",
		"serial":"A1234-L",
		"device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
	}`

	result, _ := sendRequestExpectError(t, "PUT", "/v1/models/5", bytes.NewBufferString(data))

	if result.Model.ID != 5 {
		t.Errorf("Expected model with ID 5, got %d", result.Model.ID)
	}
	if result.Model.Name != "the-model" {
		t.Errorf("Expected model name 'the-model', got %s", result.Model.Name)
	}
}

func TestModelUpdateHandlerWithErrors(t *testing.T) {
	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	// Update a model
	data := `{}`

	sendRequestExpectError(t, "PUT", "/v1/models/1", bytes.NewBufferString(data))
}

func TestModelUpdateHandlerWithNilData(t *testing.T) {
	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	sendRequestExpectError(t, "PUT", "/v1/models/1", nil)
}

func TestModelUpdateHandlerWithEmptyData(t *testing.T) {
	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	sendRequestExpectError(t, "PUT", "/v1/models/1", bytes.NewBufferString(""))
}

func TestModelUpdateHandlerWithBadData(t *testing.T) {
	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	sendRequestExpectError(t, "PUT", "/v1/models/1", bytes.NewBufferString("bad"))
}

func TestModelUpdateHandlerWithBadID(t *testing.T) {
	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	sendRequestExpectError(t, "PUT", "/v1/models/999999999999999999999999999999", bytes.NewBufferString("bad"))
}

func TestModelDeleteHandler(t *testing.T) {
	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	// Delete a model
	data := "{}"
	sendRequest(t, "DELETE", "/v1/models/1", bytes.NewBufferString(data))
}

func TestModelDeleteHandlerWithPermissions(t *testing.T) {
	// Mock the database
	c := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	// Delete a model
	data := "{}"
	sendRequest(t, "DELETE", "/v1/models/1", bytes.NewBufferString(data))
}

func TestModelDeleteHandlerWithoutPermissions(t *testing.T) {
	// Mock the database
	c := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	// Delete a model
	data := "{}"
	sendRequestWithoutPermissions(t, "DELETE", "/v1/models/1", bytes.NewBufferString(data))
}

func TestModelDeleteHandlerWithPermissionsNotFound(t *testing.T) {
	// Mock the database
	c := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: c}

	// Delete a model
	data := "{}"
	sendRequestExpectError(t, "DELETE", "/v1/models/5", bytes.NewBufferString(data))
}

func TestModelDeleteHandlerWithErrors(t *testing.T) {
	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	// Delete a model
	data := `{}`

	sendRequestExpectError(t, "DELETE", "/v1/models/1", bytes.NewBufferString(data))
}

func TestModelDeleteHandlerWithBadID(t *testing.T) {
	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	sendRequestExpectError(t, "DELETE", "/v1/models/999999999999999999999999999999", bytes.NewBufferString("bad"))
}

func TestModelCreateHandler(t *testing.T) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	// Define a model linked with the signing-key as JSON
	model := ModelSerialize{BrandID: "System", Name: "the-model", KeypairID: 1}
	data, _ := json.Marshal(model)

	result, _ := sendRequest(t, "POST", "/v1/models", bytes.NewReader(data))
	if result.Model.ID != 7 {
		t.Errorf("Expected model with ID 7, got %d", result.Model.ID)
	}
	if result.Model.Name != "the-model" {
		t.Errorf("Expected model name 'the-model', got %s", result.Model.Name)
	}
}

func TestModelCreateHandlerWithPermissions(t *testing.T) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	// Define a model linked with the signing-key as JSON
	model := ModelSerialize{BrandID: "System", Name: "the-model", KeypairID: 1}
	data, _ := json.Marshal(model)

	result, _ := sendRequest(t, "POST", "/v1/models", bytes.NewReader(data))
	if result.Model.ID != 7 {
		t.Errorf("Expected model with ID 7, got %d", result.Model.ID)
	}
	if result.Model.Name != "the-model" {
		t.Errorf("Expected model name 'the-model', got %s", result.Model.Name)
	}
}

func TestModelCreateHandlerWithoutPermissions(t *testing.T) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	// Define a model linked with the signing-key as JSON
	model := ModelSerialize{BrandID: "System", Name: "the-model", KeypairID: 1}
	data, _ := json.Marshal(model)

	sendRequestWithoutPermissions(t, "POST", "/v1/models", bytes.NewReader(data))
}

func TestModelCreateHandlerWithError(t *testing.T) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	// Define a model linked with the signing-key as JSON
	model := ModelSerialize{BrandID: "System", Name: "the-model", KeypairID: 1}
	data, _ := json.Marshal(model)

	sendRequestExpectError(t, "POST", "/v1/models", bytes.NewReader(data))
}

func TestModelCreateHandlerWithBase64Error(t *testing.T) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	// Define a model linked with the signing-key as JSON
	model := ModelSerialize{BrandID: "System", Name: "the-model", KeypairID: 1}
	data, _ := json.Marshal(model)

	sendRequestExpectError(t, "POST", "/v1/models", bytes.NewReader(data))
}

func TestModelCreateHandlerWithNilData(t *testing.T) {
	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	sendRequestExpectError(t, "POST", "/v1/models", nil)
}

func TestModelCreateHandlerWithEmptyData(t *testing.T) {
	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	sendRequestExpectError(t, "POST", "/v1/models", bytes.NewBufferString(""))
}

func TestModelCreateHandlerWithBadData(t *testing.T) {
	// Mock the database
	config := config.Settings{JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.ErrorMockDB{}, Config: config}

	sendRequestExpectError(t, "POST", "/v1/models", bytes.NewBufferString("bad"))
}

func sendRequest(t *testing.T, method, url string, data io.Reader) (ModelResponse, error) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	// Create a JWT and add it to the request
	err := createJWTWithRole(r, datastore.Admin)
	if err != nil {
		t.Errorf("Error creating a JWT: %v", err)
	}

	AdminRouter().ServeHTTP(w, r)

	// Check the JSON response
	result := ModelResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the model response: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.ErrorMessage)
	}

	return result, err
}

func sendRequestWithoutPermissions(t *testing.T, method, url string, data io.Reader) (ModelResponse, error) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	// Create a JWT and add it to the request
	err := createJWTWithRole(r, datastore.Standard)
	if err != nil {
		t.Errorf("Error creating a JWT: %v", err)
	}

	AdminRouter().ServeHTTP(w, r)

	// Check the JSON response
	result := ModelResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the model response: %v", err)
	}
	if result.Success {
		t.Error("Expected error, got success response")
	}
	if result.ErrorCode != "error-auth" {
		t.Error("Expected error-auth code")
	}

	return result, err
}

func sendRequestExpectError(t *testing.T, method, url string, data io.Reader) (ModelResponse, error) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	// Create a JWT and add it to the request
	err := createJWTWithRole(r, datastore.Admin)
	if err != nil {
		t.Errorf("Error creating a JWT: %v", err)
	}

	AdminRouter().ServeHTTP(w, r)

	// Check the JSON response
	result := ModelResponse{}
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the model response: %v", err)
	}
	if result.Success {
		t.Error("Expected error, got success")
	}

	return result, err
}
