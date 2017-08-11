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

package datastore

import (
	"testing"

	check "gopkg.in/check.v1"
)

func TestAccountAdapter(t *testing.T) { check.TestingT(t) }

type accountAdapterSuite struct{}

var _ = check.Suite(&accountAdapterSuite{})

func (as *accountAdapterSuite) TestValidAuthorityID(c *check.C) {
	authorityID := "JADNF9478NA84MAPD8"
	err := validateAuthorityID(authorityID)
	c.Assert(err, check.IsNil)
}

func (as *accountAdapterSuite) TestAuthorityIDEmpty(c *check.C) {
	authorityID := ""
	err := validateAuthorityID(authorityID)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "Authority ID must not be empty")
}

func (as *accountAdapterSuite) TestAuthorityIDTrailingSpace(c *check.C) {
	authorityID := " "
	err := validateAuthorityID(authorityID)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "Authority ID must not be empty")
}
