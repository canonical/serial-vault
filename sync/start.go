// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
 * License granted by Canonical Limited
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

package sync

import (
	"errors"

	"github.com/CanonicalLtd/serial-vault/service/log"
)

// StartCommand starts the sync process
type StartCommand struct {
	URL      string `short:"s" long:"svurl" description:"Sync URL for the cloud serial-vault" default:"https://serial-vault-partners.canonical.com/api/"`
	Username string `short:"u" long:"user" description:"Sync username for the cloud serial-vault"`
	APIKey   string `short:"a" long:"apikey" description:"Sync API key for the cloud serial-vault"`
}

// Execute the sync for the factory
func (cmd StartCommand) Execute(args []string) error {
	withErrors := false

	if len(cmd.URL) == 0 || len(cmd.Username) == 0 || len(cmd.APIKey) == 0 {
		return errors.New("The cloud serial vault URL, username and API key must be provided")
	}

	// Open the connection to the factory database
	openDatabase()

	// Initialize the factory client
	client := NewFactoryClient(cmd.URL, cmd.Username, cmd.APIKey)

	// Sync the accounts
	log.Info("Sync the accounts from the cloud")
	err := client.Accounts()
	if err != nil {
		withErrors = true
	}

	// Sync the signing-keys

	// Sync the models

	// Sync the signing logs

	if withErrors {
		return errors.New("Sync completed with errors")
	}

	return nil
}
