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

import "gopkg.in/check.v1"

type ClientSuite struct{}

var _ = check.Suite(&ClientSuite{})

func (s *ClientSuite) SetUpTest(c *check.C) {}

func (s *ClientSuite) TestAccount(c *check.C) {
	tests := []manTest{
		manTest{
			Args:         []string{"manage", "client"},
			ErrorMessage: "the required flags `-a, --api', `-b, --brand', `-m, --model', `-s, --serial' and `-u, --url' were not specified"},
		manTest{
			Args:         []string{"manage", "client", "invalid"},
			ErrorMessage: "the required flags `-a, --api', `-b, --brand', `-m, --model', `-s, --serial' and `-u, --url' were not specified"},
	}

	for _, t := range tests {
		runTest(c, t.Args, t.ErrorMessage)
	}
}
