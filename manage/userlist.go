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
	"os"
	"text/tabwriter"

	"github.com/CanonicalLtd/serial-vault/datastore"
)

// UserListCommand handles the list of users for the manage command
type UserListCommand struct{}

// Execute the list of users
func (cmd UserListCommand) Execute(args []string) error {

	// Get the list of users from the database
	openDatabase()
	users, err := datastore.Environ.DB.ListUsers()
	if err != nil {
		fmt.Printf("Error list the users: %v\n", err)
	}

	// Create a tabwriter to format the output
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 5, 0, 4, ' ', 0)

	// Print the headers
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Username\tName\tRole\tEmail")

	// Print the user list
	for _, u := range users {
		role, _ := datastore.RoleName[u.Role]

		s := fmt.Sprintf("%s\t%s\t%s\t%s", u.Username, u.Name, role, u.Email)
		fmt.Fprintln(w, s)
	}
	fmt.Fprintln(w, "")
	w.Flush()

	return nil
}
