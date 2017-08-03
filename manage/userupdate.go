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
	"errors"
	"fmt"

	"github.com/CanonicalLtd/serial-vault/datastore"
)

// UserUpdateCommand handles updating a user for the manage command
type UserUpdateCommand struct {
	Name           string `short:"n" long:"name" description:"Full name of the user"`
	RoleName       string `short:"r" long:"role" description:"Role of the user" choice:"standard" choice:"admin" choice:"superuser"`
	Email          string `short:"e" long:"email" description:"Email of the user"`
	OpenIDIdentity string `short:"i" long:"identity" description:"OpenID Identity of the user"`
}

// Execute the user update
func (cmd UserUpdateCommand) Execute(args []string) error {
	if len(args) != 1 {
		return errors.New("Update user expects a single 'username' argument")
	}

	// Convert the rolename to an ID
	roleID, ok := datastore.RoleID[cmd.RoleName]
	if !ok {
		return fmt.Errorf("Cannot find the role ID for role '%s'", cmd.RoleName)
	}

	// Open the database and get the user from the database
	openDatabase()
	user, err := datastore.Environ.DB.GetUserByUsername(args[0])
	if err != nil {
		return fmt.Errorf("Error finding the user '%s'", args[0])
	}

	// Only update the fields that have been supplied
	if len(cmd.Name) > 0 {
		user.Name = cmd.Name
	}
	if roleID > 0 {
		user.Role = roleID
	}
	if len(cmd.Email) > 0 {
		user.Email = cmd.Email
	}
	if len(cmd.OpenIDIdentity) > 0 {
		user.OpenIDIdentity = cmd.OpenIDIdentity
	}

	err = datastore.Environ.DB.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("Error updating the user: %v", err)
	}

	fmt.Printf("User '%s' updated successfully\n", user.Username)
	return nil
}
