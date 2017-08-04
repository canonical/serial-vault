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
	"fmt"

	"github.com/CanonicalLtd/serial-vault/datastore"
)

// UserAddCommand handles adding a new user for the manage command
type UserAddCommand struct {
	Name     string `short:"n" long:"name" description:"Full name of the user" required:"yes"`
	RoleName string `short:"r" long:"role" description:"Role of the user" required:"yes" choice:"standard" choice:"admin" choice:"superuser"`
	Email    string `short:"e" long:"email" description:"Email of the user"`
}

// Execute the adding a new user
func (cmd UserAddCommand) Execute(args []string) error {

	err := checkUsernameArg(args, "Add")
	if err != nil {
		return err
	}

	// Convert the rolename to an ID
	roleID, ok := datastore.RoleID[cmd.RoleName]
	if !ok {
		return fmt.Errorf("Cannot find the role ID for role '%s'", cmd.RoleName)
	}

	// Open the database and create the user
	openDatabase()
	user := datastore.User{
		Username: args[0],
		Name:     cmd.Name,
		Role:     roleID,
		Email:    cmd.Email,
		Accounts: []datastore.Account{},
	}

	_, err = datastore.Environ.DB.CreateUser(user)
	if err != nil {
		return fmt.Errorf("Error creating the user: %v", err)
	}

	fmt.Printf("User '%s' created successfully\n", user.Username)
	return nil
}
