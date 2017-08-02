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

import "log"

// UserCommand a
type UserCommand struct {
	List UserListCommand `command:"list" alias:"ls" description:"List the users"`
	Add  UserAddCommand  `command:"add" alias:"a" description:"Add a new user"`
}

// Execute a
func (cmd UserCommand) Execute(args []string) error {
	log.Println("Executing the user command")
	return nil
}
