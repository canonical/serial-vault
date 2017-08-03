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

package main

import (
	"testing"

	"github.com/CanonicalLtd/serial-vault/datastore"

	"gopkg.in/check.v1"
)

// Hook up check.v1 into the "go test" runner
func Test(t *testing.T) { check.TestingT(t) }

type manTest struct {
	Args         []string
	ErrorMessage string
}

type ManageSuite struct{}

var _ = check.Suite(&ManageSuite{})

func (s *ManageSuite) SetUpTest(c *check.C) {
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}
}

func (s *ManageSuite) TestUser(c *check.C) {
	tests := []manTest{
		manTest{
			Args:         []string{"manage", "user"},
			ErrorMessage: "Please specify one command of: add, delete, list or update"},
		manTest{
			Args:         []string{"manage", "user", "list"},
			ErrorMessage: ""},
		manTest{
			Args:         []string{"manage", "user", "add"},
			ErrorMessage: "the required flags `-n, --name' and `-r, --role' were not specified"},
		manTest{
			Args:         []string{"manage", "user", "add", "-n"},
			ErrorMessage: "expected argument for flag `-n, --name'"},
		manTest{
			Args:         []string{"manage", "user", "add", "-n", "John Smith", "-r", "invalid"},
			ErrorMessage: "Invalid value `invalid' for option `-r, --role'. Allowed values are: standard, admin or superuser"},
		manTest{
			Args:         []string{"manage", "user", "add", "-n", "John Smith", "-r", "admin"},
			ErrorMessage: "Add user expects a single 'username' argument"},
		manTest{
			Args:         []string{"manage", "user", "add", "ddan", "-n", "Desperate Dan", "-r", "admin"},
			ErrorMessage: ""},
		manTest{
			Args:         []string{"manage", "user", "add", "ddan", "-n", "Desperate Dan", "-r", "admin", "-bad"},
			ErrorMessage: "unknown flag `b'"},
		manTest{
			Args:         []string{"manage", "user", "update"},
			ErrorMessage: "Update user expects a single 'username' argument"},
		manTest{
			Args:         []string{"manage", "user", "update", "sv"},
			ErrorMessage: ""},
		manTest{
			Args:         []string{"manage", "user", "update", "sv", "-n", "Simon Vault"},
			ErrorMessage: ""},
		manTest{
			Args:         []string{"manage", "user", "delete"},
			ErrorMessage: "Delete user expects a single 'username' argument"},
		manTest{
			Args:         []string{"manage", "user", "delete", "sv"},
			ErrorMessage: ""},
	}

	for _, t := range tests {
		s.runTest(c, t.Args, t.ErrorMessage)
	}
}

func (s *ManageSuite) runTest(c *check.C, args []string, errorMessage string) {

	restore := mockArgs(args...)
	defer restore()

	err := RunMain()

	if len(errorMessage) == 0 {
		c.Check(err, check.IsNil)
	} else {
		c.Assert(err, check.ErrorMatches, errorMessage)
	}
}
