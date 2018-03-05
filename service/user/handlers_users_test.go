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

package user_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/user"
	"github.com/CanonicalLtd/serial-vault/usso"
	"github.com/juju/usso/openid"
	check "gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type ServiceSuite struct{}

type UserTest struct {
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

var _ = check.Suite(&ServiceSuite{})

func (s *ServiceSuite) SetUpTest(c *check.C) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)

	datastore.Environ.Config.EnableUserAuth = true

	// Disable CSRF for tests as we do not have a secure connection
	service.MiddlewareWithCSRF = service.Middleware
}

func (s *ServiceSuite) TestUsersHandler(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	result := s.sendRequestRepliesUsersList("GET", "/v1/users", nil, c)
	c.Assert(len(result.Users), check.Equals, 5)
	c.Assert(result.Users[0].Name, check.Equals, "Rigoberto Picaporte")
	c.Assert(result.Users[1].Name, check.Equals, "Nancy Reagan")
	c.Assert(result.Users[2].Name, check.Equals, "Steven Vault")
	c.Assert(result.Users[3].Name, check.Equals, "A")
	c.Assert(result.Users[4].Name, check.Equals, "Root User")
}

func (s *ServiceSuite) TestUsersHandlerWithError(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}

	s.sendRequestRepliesUsersListError("GET", "/v1/users", nil, c)
}

func (s *ServiceSuite) TestUsersHandlerWithoutPermissions(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	s.sendRequestWithoutPermissions("GET", "/v1/users", nil, c)
}

func (s *ServiceSuite) TestGetUserHandler(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	result := s.sendRequestRepliesUser("GET", "/v1/users/4", nil, c)
	c.Assert(result.User.ID, check.Equals, 4)
	c.Assert(result.User.Username, check.Equals, "a")
	c.Assert(result.User.Name, check.Equals, "A")
	c.Assert(result.User.Email, check.Equals, "a@example.com")
	c.Assert(result.User.Role, check.Equals, datastore.Standard)
	c.Assert(len(result.User.Accounts), check.Equals, 0)
}

func (s *ServiceSuite) TestGetUserHandlerWithAccount(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	result := s.sendRequestRepliesUser("GET", "/v1/users/2", nil, c)
	c.Assert(result.User.ID, check.Equals, 2)
	c.Assert(result.User.Username, check.Equals, "user2")
	c.Assert(result.User.Name, check.Equals, "Nancy Reagan")
	c.Assert(result.User.Email, check.Equals, "nancy.reagan@usa.gov")
	c.Assert(result.User.Role, check.Equals, datastore.Standard)
	c.Assert(len(result.User.Accounts), check.Equals, 1)
	c.Assert(result.User.Accounts[0].ID, check.Equals, 2)
	c.Assert(result.User.Accounts[0].AuthorityID, check.Equals, "authority2")
	c.Assert(result.User.Accounts[0].Assertion, check.Equals, "assertioncontent2")
}

func (s *ServiceSuite) TestGetUserHandlerWithError(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}

	s.sendRequestRepliesUserError("GET", "/v1/users/2", nil, c)
}

func (s *ServiceSuite) TestGetUserHandlerWithoutPermissions(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	s.sendRequestWithoutPermissions("GET", "/v1/users/2", nil, c)
}

func (s *ServiceSuite) TestCreateUserHandler(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := user.Request{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)

	s.sendRequestRepliesUser("POST", "/v1/users", bytes.NewReader(data), c)
}

func (s *ServiceSuite) TestCreateUserHandlerWithOneAccount(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := user.Request{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
		Accounts: []string{"theauthorityid1"},
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)

	s.sendRequestRepliesUser("POST", "/v1/users", bytes.NewReader(data), c)
}

func (s *ServiceSuite) TestCreateUserHandlerWithAccounts(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := user.Request{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
		Accounts: []string{"theauthorityid1", "theauthorityid2"},
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)

	s.sendRequestRepliesUser("POST", "/v1/users", bytes.NewReader(data), c)
}

func (s *ServiceSuite) TestCreateUserHandlerWithError(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}

	user := user.Request{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
		Accounts: []string{"theauthorityid1"},
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)
	s.sendRequestRepliesUserError("POST", "/v1/users", bytes.NewReader(data), c)
}

func (s *ServiceSuite) TestCreateUserHandlerWithoutPermissions(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := user.Request{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
		Accounts: []string{"theauthorityid1"},
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)
	s.sendRequestWithoutPermissions("POST", "/v1/users", bytes.NewReader(data), c)
}

func (s *ServiceSuite) TestUpdateUserHandler(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := user.Request{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
		Accounts: []string{"theauthorityid1"},
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)

	result := s.sendRequestRepliesUser("PUT", "/v1/users/2", bytes.NewReader(data), c)
	c.Assert(result.Success, check.Equals, true)
}

func (s *ServiceSuite) TestUpdateUserHandlerWithAccount(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := user.Request{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
		Accounts: []string{"theauthorityid1"},
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)

	result := s.sendRequestRepliesUser("PUT", "/v1/users/2", bytes.NewReader(data), c)
	c.Assert(result.Success, check.Equals, true)
}

func (s *ServiceSuite) TestUpdateUserHandlerWithAccounts(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := user.Request{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
		Accounts: []string{"theauthorityid1", "theauthorityid2"},
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)

	result := s.sendRequestRepliesUser("PUT", "/v1/users/2", bytes.NewReader(data), c)
	c.Assert(result.Success, check.Equals, true)
}

func (s *ServiceSuite) TestUpdateUserHandlerWithError(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}

	user := user.Request{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)
	s.sendRequestRepliesUserError("PUT", "/v1/users/2", bytes.NewReader(data), c)
}

func (s *ServiceSuite) TestUpdateUserHandlerWithoutPermissions(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := user.Request{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)
	s.sendRequestWithoutPermissions("PUT", "/v1/users/2", bytes.NewReader(data), c)
}

func (s *ServiceSuite) TestDeleteUserHandler(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	s.sendRequestRepliesUser("DELETE", "/v1/users/2", nil, c)
}

func (s *ServiceSuite) TestDeleteUserHandlerWithError(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}

	s.sendRequestRepliesUserError("DELETE", "/v1/users/2", nil, c)
}

func (s *ServiceSuite) TestDeleteUserHandlerWithoutPermissions(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	s.sendRequestWithoutPermissions("DELETE", "/v1/users/2", nil, c)
}

func (s *ServiceSuite) createSuperuserJWT(r *http.Request, c *check.C) {
	sreg := map[string]string{"nickname": "root", "fullname": "Root User", "email": "the_root_user@thisdb.com"}
	resp := openid.Response{ID: "identity", Teams: []string{}, SReg: sreg}
	jwtToken, err := usso.NewJWTToken(&resp, datastore.Superuser)
	c.Assert(err, check.IsNil)

	r.Header.Set("Authorization", "Bearer "+jwtToken)
}

func (s *ServiceSuite) sendRequestRepliesUser(method, url string, data io.Reader, c *check.C) user.GetResponse {
	body := s.sendRequest(method, url, data, c)

	result := user.GetResponse{}
	err := json.NewDecoder(body).Decode(&result)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, true)

	return result
}

func (s *ServiceSuite) sendRequestRepliesUsersList(method, url string, data io.Reader, c *check.C) user.ListResponse {
	body := s.sendRequest(method, url, data, c)

	result := user.ListResponse{}
	err := json.NewDecoder(body).Decode(&result)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, true)

	return result
}

func (s *ServiceSuite) sendRequest(method, url string, data io.Reader, c *check.C) *bytes.Buffer {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	s.createSuperuserJWT(r, c)

	service.AdminRouter().ServeHTTP(w, r)

	return w.Body
}

func (s *ServiceSuite) sendRequestRepliesUserError(method, url string, data io.Reader, c *check.C) {
	body := s.sendRequest(method, url, data, c)
	result := user.GetResponse{}
	err := json.NewDecoder(body).Decode(&result)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, false)
}

func (s *ServiceSuite) sendRequestRepliesUsersListError(method, url string, data io.Reader, c *check.C) {
	body := s.sendRequest(method, url, data, c)
	result := user.ListResponse{}
	err := json.NewDecoder(body).Decode(&result)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, false)
}

func (s *ServiceSuite) sendRequestWithoutPermissions(method, url string, data io.Reader, c *check.C) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)
	service.AdminRouter().ServeHTTP(w, r)

	result := user.GetResponse{}

	log.Println(string(w.Body.Bytes()))

	err := json.NewDecoder(w.Body).Decode(&result)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, false)
	c.Assert(result.ErrorCode, check.Equals, "error-auth")
}

func (s *ServiceSuite) TestOtherAccountsHandler(c *check.C) {
	tests := []UserTest{
		{"GET", "/v1/users/1/otheraccounts", nil, 200, "application/json; charset=UTF-8", datastore.Superuser, true, true, 1},
		{"GET", "/v1/users/1/otheraccounts", nil, 400, "application/json; charset=UTF-8", datastore.Admin, true, false, 0},
		{"GET", "/v1/users/1/otheraccounts", nil, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
	}

	for _, t := range tests {
		//if !t.EnableAuth {
		datastore.Environ.Config.EnableUserAuth = t.EnableAuth
		//}

		w := sendAdminRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		c.Log("---", t.Permissions)
		c.Log("---", t.EnableAuth, datastore.Environ.Config.EnableUserAuth)
		c.Log("---", string(w.Body.Bytes()))
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := parseAccountsResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)
		c.Assert(len(result.Accounts), check.Equals, t.List)

		datastore.Environ.Config.EnableUserAuth = !t.EnableAuth
	}
}

func parseAccountsResponse(w *httptest.ResponseRecorder) (user.AccountsResponse, error) {
	// Check the JSON response
	result := user.AccountsResponse{}
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
	sreg := map[string]string{"nickname": "root", "fullname": "Root User", "email": "the_root_user@thisdb.com"}
	resp := openid.Response{ID: "identity", Teams: []string{}, SReg: sreg}
	jwtToken, err := usso.NewJWTToken(&resp, role)
	if err != nil {
		return fmt.Errorf("Error creating a JWT: %v", err)
	}
	r.Header.Set("Authorization", "Bearer "+jwtToken)
	return nil
}
