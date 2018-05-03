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

package keypair_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/crypt"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/keypair"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/CanonicalLtd/serial-vault/usso"
	"github.com/juju/usso/openid"
	"github.com/snapcore/snapd/asserts"
	check "gopkg.in/check.v1"
)

func TestKeypairSuite(t *testing.T) { check.TestingT(t) }

type KeypairSuite struct{}

var _ = check.Suite(&KeypairSuite{})

type KeypairTest struct {
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

func (s *KeypairSuite) SetUpTest(c *check.C) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)

	// Disable CSRF for tests as we do not have a secure connection
	service.MiddlewareWithCSRF = service.Middleware
}

func (s *KeypairSuite) TestListStatusHandler(c *check.C) {
	tests := []KeypairTest{
		{"GET", "/v1/keypairs", nil, 200, response.JSONHeader, 0, false, true, 4},
		{"GET", "/v1/keypairs", nil, 200, response.JSONHeader, datastore.Admin, true, true, 2},
		{"GET", "/v1/keypairs", nil, 400, response.JSONHeader, datastore.Standard, true, false, 0},
		{"GET", "/v1/keypairs", nil, 400, response.JSONHeader, 0, true, false, 0},

		{"GET", "/v1/keypairs/status/system/key1", nil, 200, response.JSONHeader, 0, false, true, 0},
		{"GET", "/v1/keypairs/status/system/key1", nil, 200, response.JSONHeader, datastore.Admin, true, true, 0},
		{"GET", "/v1/keypairs/status/system/key1", nil, 400, response.JSONHeader, datastore.Standard, true, false, 0},
		{"GET", "/v1/keypairs/status/system/key1", nil, 400, response.JSONHeader, 0, true, false, 0},

		{"GET", "/v1/keypairs/status", nil, 200, response.JSONHeader, 0, false, true, 0},
		{"GET", "/v1/keypairs/status", nil, 200, response.JSONHeader, datastore.Admin, true, true, 0},
		{"GET", "/v1/keypairs/status", nil, 400, response.JSONHeader, datastore.Standard, true, false, 0},
		{"GET", "/v1/keypairs/status", nil, 400, response.JSONHeader, 0, true, false, 0},
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
		c.Assert(len(result.Keypairs), check.Equals, t.List)

		datastore.Environ.Config.EnableUserAuth = false
	}
}

func (s *KeypairSuite) TestKeypairsErrorHandler(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}
	tests := []KeypairTest{
		{"GET", "/v1/keypairs", nil, 400, response.JSONHeader, 0, false, false, 0},
		{"GET", "/v1/keypairs", nil, 400, response.JSONHeader, datastore.Admin, true, false, 0},
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
		c.Assert(len(result.Keypairs), check.Equals, t.List)

		datastore.Environ.Config.EnableUserAuth = false
	}
}

func (s *KeypairSuite) TestCreateGenerateEnableDisableHandler(c *check.C) {
	// Mock the database and the keystore
	config := config.Settings{KeyStoreType: "memory", EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	// Generate the keypair to upload
	signingKey, err := ioutil.ReadFile("../../keystore/TestKey.asc")
	c.Assert(err, check.IsNil)
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)
	k := keypair.WithPrivateKey{PrivateKey: string(encodedSigningKey), AuthorityID: "system", KeyName: "serial-key"}
	data, _ := json.Marshal(k)

	// Generate a bad keypair to upload
	signingKey, err = ioutil.ReadFile("../../README.md")
	c.Assert(err, check.IsNil)
	encodedSigningKey = base64.StdEncoding.EncodeToString(signingKey)
	k = keypair.WithPrivateKey{PrivateKey: string(encodedSigningKey), AuthorityID: "system"}
	dataBad, _ := json.Marshal(k)

	kp := datastore.Keypair{ID: 1, AuthorityID: "system", KeyName: "serial-key"}
	keypair, _ := json.Marshal(kp)

	tests := []KeypairTest{
		{"GET", "/v1/keypairs/1", nil, 200, response.JSONHeader, 0, false, true, 0},
		{"GET", "/v1/keypairs/1", nil, 200, response.JSONHeader, datastore.Admin, true, true, 0},
		{"GET", "/v1/keypairs/1", nil, 400, response.JSONHeader, datastore.Standard, true, false, 0},
		{"GET", "/v1/keypairs/1", nil, 400, response.JSONHeader, datastore.Admin, true, false, 1},
		{"GET", "/v1/keypairs/9999999999999999999999999", nil, 400, response.JSONHeader, datastore.Admin, true, false, 0},

		{"POST", "/v1/keypairs", data, 200, response.JSONHeader, 0, false, true, 0},
		{"POST", "/v1/keypairs", data, 200, response.JSONHeader, datastore.Admin, true, true, 0},
		{"POST", "/v1/keypairs", []byte(""), 400, response.JSONHeader, datastore.Admin, true, false, 0},
		{"POST", "/v1/keypairs", []byte("bad"), 400, response.JSONHeader, datastore.Admin, true, false, 0},
		{"POST", "/v1/keypairs", []byte("{}"), 400, response.JSONHeader, datastore.Admin, true, false, 0},
		{"POST", "/v1/keypairs", dataBad, 400, response.JSONHeader, datastore.Admin, true, false, 0},
		{"POST", "/v1/keypairs", data, 400, response.JSONHeader, datastore.Standard, true, false, 0},
		{"POST", "/v1/keypairs", data, 400, response.JSONHeader, 0, true, false, 0},

		{"PUT", "/v1/keypairs/1", keypair, 200, response.JSONHeader, 0, false, true, 0},
		{"PUT", "/v1/keypairs/1", keypair, 200, response.JSONHeader, datastore.Admin, true, true, 0},
		{"PUT", "/v1/keypairs/1", []byte(""), 400, response.JSONHeader, datastore.Admin, true, false, 0},
		{"PUT", "/v1/keypairs/1", []byte("bad"), 400, response.JSONHeader, datastore.Admin, true, false, 0},
		{"PUT", "/v1/keypairs/1", []byte("{}"), 400, response.JSONHeader, datastore.Admin, true, false, 0},
		{"PUT", "/v1/keypairs/1", dataBad, 400, response.JSONHeader, datastore.Admin, true, false, 0},
		{"PUT", "/v1/keypairs/1", keypair, 400, response.JSONHeader, datastore.Standard, true, false, 0},
		{"PUT", "/v1/keypairs/1", keypair, 400, response.JSONHeader, 0, true, false, 0},

		{"POST", "/v1/keypairs/generate", data, 202, response.JSONHeader, 0, false, true, 0},
		{"POST", "/v1/keypairs/generate", data, 202, response.JSONHeader, datastore.Admin, true, true, 0},
		{"POST", "/v1/keypairs/generate", []byte(""), 400, response.JSONHeader, datastore.Admin, true, false, 0},
		{"POST", "/v1/keypairs/generate", []byte("bad"), 400, response.JSONHeader, datastore.Admin, true, false, 0},
		{"POST", "/v1/keypairs/generate", []byte("{}"), 400, response.JSONHeader, datastore.Admin, true, false, 0},
		{"POST", "/v1/keypairs/generate", data, 400, response.JSONHeader, datastore.Standard, true, false, 0},
		{"POST", "/v1/keypairs/generate", data, 400, response.JSONHeader, 0, true, false, 0},

		{"POST", "/v1/keypairs/1/disable", []byte(""), 200, response.JSONHeader, 0, false, true, 0},
		{"POST", "/v1/keypairs/1/disable", []byte(""), 200, response.JSONHeader, datastore.Admin, true, true, 0},
		{"POST", "/v1/keypairs/1/disable", []byte(""), 400, response.JSONHeader, datastore.Standard, true, false, 0},
		{"POST", "/v1/keypairs/1/disable", []byte(""), 400, response.JSONHeader, datastore.Admin, true, false, 1},
		{"POST", "/v1/keypairs/9999999999999999999999999/disable", []byte(""), 400, response.JSONHeader, datastore.Admin, true, false, 0},

		{"POST", "/v1/keypairs/1/enable", []byte(""), 200, response.JSONHeader, 0, false, true, 0},
		{"POST", "/v1/keypairs/1/enable", []byte(""), 200, response.JSONHeader, datastore.Admin, true, true, 0},
		{"POST", "/v1/keypairs/1/enable", []byte(""), 400, response.JSONHeader, datastore.Standard, true, false, 0},
		{"POST", "/v1/keypairs/1/enable", []byte(""), 400, response.JSONHeader, datastore.Admin, true, false, 1},
		{"POST", "/v1/keypairs/9999999999999999999999999/enable", []byte(""), 400, response.JSONHeader, datastore.Admin, true, false, 0},
	}
	for _, t := range tests {
		datastore.Environ.KeypairDB, _ = datastore.GetMemoryKeyStore(config)
		if t.List > 0 {
			// Use the error database mock
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}
		datastore.Environ.Config.EnableUserAuth = t.EnableAuth

		w := sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		c.Log(string(t.Data))
		c.Log(w.Body)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := response.ParseStandardResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)

		datastore.Environ.Config.EnableUserAuth = false
		datastore.Environ.DB = &datastore.MockDB{}
	}
}

func (s *KeypairSuite) TestCreateKeyStoreError(c *check.C) {
	// Mock the database and the keystore
	config := config.Settings{KeyStoreType: "memory", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	signingKey, err := ioutil.ReadFile("../../keystore/TestKey.asc")
	c.Assert(err, check.IsNil)
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	k := keypair.WithPrivateKey{PrivateKey: string(encodedSigningKey), AuthorityID: "System"}
	data, _ := json.Marshal(k)

	tests := []KeypairTest{
		{"POST", "/v1/keypairs", data, 400, response.JSONHeader, datastore.Admin, true, false, 0},
		{"POST", "/v1/keypairs", data, 400, response.JSONHeader, datastore.Admin, true, false, 1},
	}
	for _, t := range tests {
		datastore.Environ.KeypairDB, _ = datastore.GetErrorMockKeyStore(config)
		if t.List > 0 {
			// Use the error database mock
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}

		datastore.Environ.Config.EnableUserAuth = t.EnableAuth

		w := sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := response.ParseStandardResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)

		datastore.Environ.Config.EnableUserAuth = false
		datastore.Environ.DB = &datastore.MockDB{}
	}
}

func (s *KeypairSuite) TestAssertionHandler(c *check.C) {
	// Create the account key assertion
	assertAcc, err := generateAccountAssertion(asserts.AccountKeyType, "alder", "maple-inc")
	c.Assert(err, check.IsNil)

	// Encode the assertion and create the request
	encodedAssert := base64.StdEncoding.EncodeToString([]byte(assertAcc))
	data, err := json.Marshal(keypair.AssertionRequest{ID: 1, Assertion: encodedAssert})
	c.Assert(err, check.IsNil)

	dataBad1, err := json.Marshal(keypair.AssertionRequest{ID: 0, Assertion: "InvalidData"})
	c.Assert(err, check.IsNil)
	dataBad2, err := json.Marshal(keypair.AssertionRequest{ID: 1, Assertion: "InvalidData"})
	c.Assert(err, check.IsNil)

	// Wrong type
	assertAcc, err = generateAccountAssertion(asserts.AccountType, "alder", "maple-inc")
	c.Assert(err, check.IsNil)
	encodedAssert = base64.StdEncoding.EncodeToString([]byte(assertAcc))
	dataBad3, err := json.Marshal(keypair.AssertionRequest{ID: 1, Assertion: encodedAssert})
	c.Assert(err, check.IsNil)

	tests := []KeypairTest{
		{"POST", "/v1/keypairs/assertion", data, 200, response.JSONHeader, 0, false, true, 0},
		{"POST", "/v1/keypairs/assertion", data, 200, response.JSONHeader, datastore.Admin, true, true, 0},
		{"POST", "/v1/keypairs/assertion", data, 400, response.JSONHeader, datastore.Standard, true, false, 0},
		{"POST", "/v1/keypairs/assertion", data, 400, response.JSONHeader, 0, true, false, 0},
		{"POST", "/v1/keypairs/assertion", []byte(""), 400, response.JSONHeader, 0, true, false, 0},
		{"POST", "/v1/keypairs/assertion", []byte("bad"), 400, response.JSONHeader, 0, true, false, 0},
		{"POST", "/v1/keypairs/assertion", dataBad1, 400, response.JSONHeader, datastore.Admin, true, false, 0},
		{"POST", "/v1/keypairs/assertion", dataBad2, 400, response.JSONHeader, datastore.Admin, true, false, 0},
		{"POST", "/v1/keypairs/assertion", dataBad3, 400, response.JSONHeader, datastore.Admin, true, false, 0},
		{"POST", "/v1/keypairs/assertion", data, 400, response.JSONHeader, datastore.Admin, true, false, 1},
	}
	for _, t := range tests {
		if t.List > 0 {
			// Use the error database mock
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}
		datastore.Environ.Config.EnableUserAuth = t.EnableAuth

		w := sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := response.ParseStandardResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)

		datastore.Environ.Config.EnableUserAuth = false
		datastore.Environ.DB = &datastore.MockDB{}
	}
}

func parseListResponse(w *httptest.ResponseRecorder) (keypair.ListResponse, error) {
	// Check the JSON response
	result := keypair.ListResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
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

func generateAccountAssertion(assertType *asserts.AssertionType, accountID, username string) (string, error) {
	privateKey, _ := generatePrivateKey()

	headers := map[string]interface{}{
		"authority-id": "system",
		"account-id":   accountID,
		"username":     username,
		"display-name": username,
		"mail":         "test@example.com",
		"revision":     "1",
		"timestamp":    "2016-01-02T15:04:05Z",
		"validation":   "unproven",
	}

	var body []byte

	switch assertType {
	case asserts.AccountType:
		headers["sign-key-sha3-384"] = privateKey.PublicKey().ID()
		body = nil
	case asserts.AccountKeyType:
		headers["public-key-sha3-384"] = privateKey.PublicKey().ID()
		headers["since"] = "2016-01-02T15:04:05Z"
		body, _ = asserts.EncodePublicKey(privateKey.PublicKey())
	}

	accAssert, err := datastore.Environ.KeypairDB.Sign(assertType, headers, body, "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO")
	if err != nil {
		return "", err
	}

	assertAcc := asserts.Encode(accAssert)
	return string(assertAcc), nil
}

func generatePrivateKey() (asserts.PrivateKey, error) {
	signingKey, err := ioutil.ReadFile("../../keystore/TestDeviceKey.asc")
	if err != nil {
		return nil, err
	}
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	privateKey, _, err := crypt.DeserializePrivateKey(encodedSigningKey)
	return privateKey, err
}
