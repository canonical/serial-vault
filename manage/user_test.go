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
	"github.com/CanonicalLtd/serial-vault/datastore"
	"gopkg.in/check.v1"
)

type UserSuite struct{}

var _ = check.Suite(&UserSuite{})

func (s *UserSuite) SetUpTest(c *check.C) {
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}}
}

func (s *UserSuite) TestUser(c *check.C) {
	tests := []manTest{
		manTest{
			Args:         []string{"serial-vault-admin", "user"},
			ErrorMessage: "Please specify one command of: add, delete, list or update"},
		manTest{
			Args:         []string{"serial-vault-admin", "user", "list"},
			ErrorMessage: ""},
		manTest{
			Args:         []string{"serial-vault-admin", "user", "add"},
			ErrorMessage: "the required flags `-n, --name' and `-r, --role' were not specified"},
		manTest{
			Args:         []string{"serial-vault-admin", "user", "add", "-n"},
			ErrorMessage: "expected argument for flag `-n, --name'"},
		manTest{
			Args:         []string{"serial-vault-admin", "user", "add", "-n", "John Smith", "-r", "invalid"},
			ErrorMessage: "Invalid value `invalid' for option `-r, --role'. Allowed values are: standard, admin or superuser"},
		manTest{
			Args:         []string{"serial-vault-admin", "user", "add", "-n", "John Smith", "-r", "admin"},
			ErrorMessage: "Add user expects a 'username' argument"},
		manTest{
			Args:         []string{"serial-vault-admin", "user", "add", "ddan", "-n", "Desperate Dan", "-r", "admin"},
			ErrorMessage: ""},
		manTest{
			Args:         []string{"serial-vault-admin", "user", "add", "ddan", "-n", "Desperate Dan", "-r", "admin", "-bad"},
			ErrorMessage: "unknown flag `b'"},
		manTest{
			Args:         []string{"serial-vault-admin", "user", "update"},
			ErrorMessage: "Update user expects a 'username' argument"},
		manTest{
			Args:         []string{"serial-vault-admin", "user", "update", "john", "smith"},
			ErrorMessage: "Update user expects a single 'username' argument"},
		manTest{
			Args:         []string{"serial-vault-admin", "user", "update", "sv"},
			ErrorMessage: "No changes requested. Please supply user details to change"},
		manTest{
			Args:         []string{"serial-vault-admin", "user", "update", "sv", "-n", "Simon Vault"},
			ErrorMessage: ""},
		manTest{
			Args:         []string{"serial-vault-admin", "user", "update", "sv", "-u", "svault"},
			ErrorMessage: ""},
		manTest{
			Args:         []string{"serial-vault-admin", "user", "delete"},
			ErrorMessage: "Delete user expects a 'username' argument"},
		manTest{
			Args:         []string{"serial-vault-admin", "user", "delete", "john", "smith"},
			ErrorMessage: "Delete user expects a single 'username' argument"},
		manTest{
			Args:         []string{"serial-vault-admin", "user", "delete", "sv"},
			ErrorMessage: ""},
	}

	for _, t := range tests {
		runTest(c, t.Args, t.ErrorMessage)
	}
}
