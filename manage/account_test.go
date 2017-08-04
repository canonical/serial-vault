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
	"github.com/CanonicalLtd/serial-vault/account"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"gopkg.in/check.v1"
)

type AccountSuite struct{}

var _ = check.Suite(&AccountSuite{})

func (s *AccountSuite) SetUpTest(c *check.C) {
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}

	// Mock the retrieval of the assertion from the store (using a fixed assertion)
	account.FetchAssertionFromStore = account.MockFetchAssertionFromStore
}

func (s *AccountSuite) TestAccount(c *check.C) {
	tests := []manTest{
		manTest{
			Args:         []string{"manage", "account"},
			ErrorMessage: "Please specify the cache command"},
		manTest{
			Args:         []string{"manage", "account", "invalid"},
			ErrorMessage: "Unknown command `invalid'. You should use the cache command"},
		manTest{
			Args:         []string{"manage", "account", "cache"},
			ErrorMessage: ""},
	}

	for _, t := range tests {
		s.runTest(c, t.Args, t.ErrorMessage)
	}
}

func (s *AccountSuite) runTest(c *check.C, args []string, errorMessage string) {

	restore := mockArgs(args...)
	defer restore()

	err := RunMain()

	if len(errorMessage) == 0 {
		c.Check(err, check.IsNil)
	} else {
		c.Assert(err, check.ErrorMatches, errorMessage)
	}
}
