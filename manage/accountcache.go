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

	"github.com/CanonicalLtd/serial-vault/account"
	"github.com/CanonicalLtd/serial-vault/datastore"
)

// AccountCacheCommand handles the caching of account assertions from the store.
// This command would normally be run as a cron
type AccountCacheCommand struct{}

// Execute the caching of account assertions
func (cmd AccountCacheCommand) Execute(args []string) error {
	fmt.Println("Update account assertions from the Ubuntu store...")

	openDatabase()

	// Cache the account assertions from the store in the database
	account.CacheAccountAssertions(datastore.Environ)

	return nil
}
