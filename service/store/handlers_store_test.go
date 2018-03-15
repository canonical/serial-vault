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

package store_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/store"
	"github.com/CanonicalLtd/serial-vault/usso"
	"github.com/juju/usso/openid"
	check "gopkg.in/check.v1"
)

func TestStoreSuite(t *testing.T) { check.TestingT(t) }

type StoreSuite struct{}

type StoreSuiteTest struct {
	Data         []byte
	Code         int
	MockError    bool
	MockRegError bool
	Permissions  int
	EnableAuth   bool
	SkipJWT      bool
}

var _ = check.Suite(&StoreSuite{})

func (s *StoreSuite) SetUpTest(c *check.C) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)

	// Mocks the submission of the key to the store
	store.RegisterKey = mockRegisterKey

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

func (s *StoreSuite) TestStoreHandler(c *check.C) {
	tests := []StoreSuiteTest{
		{nil, 400, false, false, 0, false, false},
		{validKeyRegister(), 200, false, false, 0, false, false},
		{validKeyRegister(), 400, true, false, 0, false, false},
		{nil, 400, false, false, datastore.Admin, true, false},
		{validKeyRegister(), 200, false, false, datastore.Admin, true, false},
		{validKeyRegister(), 400, false, false, datastore.Admin, true, true},
		{validKeyRegister(), 400, false, true, datastore.Admin, true, false},
		{validKeyRegister(), 400, false, false, datastore.Standard, true, false},
	}

	for _, t := range tests {
		if t.MockError {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}
		if t.MockRegError {
			store.RegisterKey = mockRegisterKeyError
		}
		datastore.Environ.Config.EnableUserAuth = t.EnableAuth

		w := sendAdminRequest("POST", "/v1/keypairs/register", bytes.NewReader(t.Data), t.Permissions, t.SkipJWT, c)
		c.Assert(w.Code, check.Equals, t.Code)

		datastore.Environ.DB = &datastore.MockDB{}
		store.RegisterKey = mockRegisterKey
		datastore.Environ.Config.EnableUserAuth = false
	}
}

func validKeyRegister() []byte {
	a := store.KeyRegister{
		Auth:        store.Auth{Email: "john@example.com", Password: "password", OTP: ""},
		AuthorityID: "system",
		KeyName:     "UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO",
	}
	d, _ := json.Marshal(a)
	return d
}

func mockRegisterKey(keyAuth store.KeyRegister, keypair datastore.Keypair) error {
	return nil
}

func mockRegisterKeyError(keyAuth store.KeyRegister, keypair datastore.Keypair) error {
	return errors.New("MOCK error on register key")
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
