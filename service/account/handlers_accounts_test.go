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

package account_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"bytes"

	"encoding/base64"

	acc "github.com/CanonicalLtd/serial-vault/account"
	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/crypt"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/account"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/CanonicalLtd/serial-vault/usso"
	"github.com/juju/usso/openid"
	"github.com/snapcore/snapd/asserts"
	check "gopkg.in/check.v1"
)

func TestAccountSuite(t *testing.T) { check.TestingT(t) }

type AccountSuite struct{}

type AccountTest struct {
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
	Accounts    int
}

var _ = check.Suite(&AccountSuite{})

func (s *AccountSuite) SetUpTest(c *check.C) {
	// Mock the store
	acc.FetchAssertionFromStore = acc.MockFetchAssertionFromStore

	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)

	// Disable CSRF for tests as we do not have a secure connection
	service.MiddlewareWithCSRF = service.Middleware
}

func sendAdminRequest(method, url string, data io.Reader, permissions int, skipJWT bool, c *check.C) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	if datastore.Environ.Config.EnableUserAuth && !skipJWT {
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

func parseListResponse(w *httptest.ResponseRecorder) (account.ListResponse, error) {
	// Check the JSON response
	result := account.ListResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
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

func (s *AccountSuite) TestAccountsHandler(c *check.C) {

	tests := []AccountTest{
		{"GET", "/v1/accounts", nil, 200, "application/json; charset=UTF-8", 0, false, true, false, false, 3},
		{"GET", "/v1/accounts", nil, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, false, false, 3},
		{"GET", "/v1/accounts", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, true, false, 0},
		{"GET", "/v1/accounts", nil, 400, "application/json; charset=UTF-8", 0, true, false, false, false, 0},
	}

	for _, t := range tests {
		if t.EnableAuth {
			datastore.Environ.Config.EnableUserAuth = true
		}

		w := sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, t.SkipJWT, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := parseListResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)
		c.Assert(len(result.Accounts), check.Equals, t.Accounts)

		datastore.Environ.Config.EnableUserAuth = false
	}
}

func (s *AccountSuite) TestCreateGetUpdateAccountHandlers(c *check.C) {

	account := datastore.Account{ID: 2, AuthorityID: "vendor", ResellerAPI: false}
	acc, _ := json.Marshal(account)

	tests := []AccountTest{
		{"POST", "/v1/accounts", nil, 400, "application/json; charset=UTF-8", 0, false, false, false, false, 0},
		{"POST", "/v1/accounts", acc, 200, "application/json; charset=UTF-8", 0, false, true, false, false, 0},
		{"POST", "/v1/accounts", acc, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, false, false, 1},
		{"POST", "/v1/accounts", acc, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, true, false, 1},
		{"POST", "/v1/accounts", acc, 400, "application/json; charset=UTF-8", 0, true, false, false, false, 0},

		{"GET", "/v1/accounts/99999", nil, 400, "application/json; charset=UTF-8", 0, false, false, false, false, 0},
		{"GET", "/v1/accounts/1", nil, 200, "application/json; charset=UTF-8", 0, false, true, false, false, 0},
		{"GET", "/v1/accounts/1", nil, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, false, false, 0},
		{"GET", "/v1/accounts/1", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, true, false, 0},
		{"GET", "/v1/accounts/1", nil, 400, "application/json; charset=UTF-8", 0, true, false, false, false, 0},

		{"PUT", "/v1/accounts/1", nil, 400, "application/json; charset=UTF-8", 0, false, false, false, false, 0},
		{"PUT", "/v1/accounts/1", acc, 200, "application/json; charset=UTF-8", 0, false, true, false, false, 0},
		{"PUT", "/v1/accounts/1", acc, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, false, false, 0},
		{"PUT", "/v1/accounts/1", acc, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, true, false, 0},
		{"PUT", "/v1/accounts/1", acc, 400, "application/json; charset=UTF-8", 0, true, false, false, false, 0},
	}

	for _, t := range tests {
		if t.EnableAuth {
			datastore.Environ.Config.EnableUserAuth = true
		}

		w := sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, t.SkipJWT, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := response.ParseStandardResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)

		datastore.Environ.Config.EnableUserAuth = false
	}
}

func (s *AccountSuite) TestAccountsHandlerError(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}

	w := sendAdminRequest("GET", "/v1/accounts", bytes.NewReader(nil), datastore.Admin, false, c)
	c.Assert(w.Code, check.Equals, 400)

	result, err := response.ParseStandardResponse(w)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, false)
}

func (s *AccountSuite) TestAccountsUploadHandler(c *check.C) {

	// Create the account assertion
	assertAcc, err := generateAccountAssertion(asserts.AccountType, "alder", "maple-inc")
	c.Assert(err, check.IsNil)

	// Encode the assertion and create the request
	encodedAssert := base64.StdEncoding.EncodeToString([]byte(assertAcc))
	request, err := json.Marshal(account.AssertionRequest{Assertion: encodedAssert})
	c.Assert(err, check.IsNil)

	invalidRequest1, err := json.Marshal(account.AssertionRequest{Assertion: "InvalidData"})
	c.Assert(err, check.IsNil)

	encodedAssert = base64.StdEncoding.EncodeToString([]byte("InvalidData"))
	invalidRequest2, err := json.Marshal(account.AssertionRequest{Assertion: encodedAssert})
	c.Assert(err, check.IsNil)

	// Encode the assertion and create the request (account-key instead of an account assertion)
	assertion, err := generateAccountAssertion(asserts.AccountKeyType, "alder", "maple-inc")
	c.Assert(err, check.IsNil)
	encodedAssert = base64.StdEncoding.EncodeToString([]byte(assertion))
	invalidRequest3, err := json.Marshal(account.AssertionRequest{Assertion: encodedAssert})
	c.Assert(err, check.IsNil)

	tests := []AccountTest{
		{"POST", "/v1/accounts/upload", request, 200, "application/json; charset=UTF-8", 0, false, true, false, false, 0},
		{"POST", "/v1/accounts/upload", request, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, false, false, 0},
		{"POST", "/v1/accounts/upload", request, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, true, false, 0},
		{"POST", "/v1/accounts/upload", request, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, true, false, 0},
		{"POST", "/v1/accounts/upload", request, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, false, false, 0},
		{"POST", "/v1/accounts/upload", []byte("InvalidData"), 400, "application/json; charset=UTF-8", datastore.Admin, true, false, false, false, 0},
		{"POST", "/v1/accounts/upload", invalidRequest1, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, false, false, 0},
		{"POST", "/v1/accounts/upload", invalidRequest2, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, false, false, 0},
		{"POST", "/v1/accounts/upload", invalidRequest3, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, false, false, 0},
		{"POST", "/v1/accounts/upload", request, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, false, true, 0},
	}

	for _, t := range tests {
		if t.EnableAuth {
			datastore.Environ.Config.EnableUserAuth = true
		}
		if t.MockError {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}

		w := sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, t.SkipJWT, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := response.ParseStandardResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)

		datastore.Environ.Config.EnableUserAuth = false
		datastore.Environ.DB = &datastore.MockDB{}
	}

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
