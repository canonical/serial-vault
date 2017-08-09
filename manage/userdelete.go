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

// UserDeleteCommand handles user delete for the serial-vault-admin command
type UserDeleteCommand struct{}

// Execute user deletion
func (cmd UserDeleteCommand) Execute(args []string) error {
	err := checkUsernameArg(args, "Delete")
	if err != nil {
		return err
	}

	// Open the database and get the user from the database
	openDatabase()
	user, err := datastore.Environ.DB.GetUserByUsername(args[0])
	if err != nil {
		return fmt.Errorf("Error finding the user '%s'", args[0])
	}

	err = datastore.Environ.DB.DeleteUser(user.ID)
	if err != nil {
		return fmt.Errorf("Error deleting the user: %v", err)
	}

	fmt.Printf("User '%s' deleted successfully\n", user.Username)
	return nil
}
