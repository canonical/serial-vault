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

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/store"
	check "gopkg.in/check.v1"
)

func TestStoreSuite(t *testing.T) { check.TestingT(t) }

type StoreSuite struct{}

type StoreSuiteTest struct {
	Data []byte
	Code int
}

var _ = check.Suite(&StoreSuite{})

func (s *StoreSuite) SetUpSuite(c *check.C) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)

	// Mocks the submission of the key to the store
	store.RegisterKey = mockRegisterKey
}

func (s *StoreSuite) sendRequest(method, url string, data io.Reader, c *check.C) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	AdminRouter().ServeHTTP(w, r)

	return w
}

func (s *StoreSuite) TestStoreHandler(c *check.C) {
	tests := []StoreSuiteTest{
		StoreSuiteTest{nil, 400},
		StoreSuiteTest{validKeyRegister(), 200},
	}

	for _, t := range tests {
		w := s.sendRequest("POST", "/v1/keypairs/register", bytes.NewReader(t.Data), c)
		c.Assert(w.Code, check.Equals, t.Code)
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
