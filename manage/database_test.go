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

package manage

import (
	"log"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"

	"gopkg.in/check.v1"
)

type databaseSuite struct{}

var _ = check.Suite(&databaseSuite{})

func (s *databaseSuite) SetUpTest(c *check.C) {

	mockDB := datastore.MockDB{}
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	datastore.Environ = &datastore.Env{DB: &mockDB, Config: config}
	datastore.OpenKeyStore(config)
}

func (s *databaseSuite) TestDoTable(c *check.C) {
	m1 := func() error {
		log.Println("Successful execution 1")
		return nil
	}

	m2 := func() error {
		log.Println("Successful execution 2")
		return nil
	}

	exec([]operation{
		{m1, create, "the table 1", false},
		{m2, update, "the table 2", false},
		{m1, create, "the table 1", true},
		{m2, update, "the table 2", true},
	})
}

func (s *databaseSuite) TestDatabase(c *check.C) {
	tests := []manTest{
		{
			Args:         []string{"serial-vault-admin", "database"},
			ErrorMessage: ""},
	}

	for _, t := range tests {
		runTest(c, t.Args, t.ErrorMessage)
	}
}
