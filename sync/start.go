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
	"time"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/log"
)

const sleepHours = 1

// StartCommand starts the sync process
type StartCommand struct {
	URL      string `short:"s" long:"svurl" description:"Sync URL for the cloud serial-vault" default:"https://serial-vault-partners.canonical.com/api/"`
	Username string `short:"u" long:"user" description:"Sync username for the cloud serial-vault"`
	APIKey   string `short:"a" long:"apikey" description:"Sync API key for the cloud serial-vault"`
	Daemon   bool   `short:"d" long:"daemon" description:"Starts the sync as a scheduled process"`
}

// Execute the sync for the factory
func (cmd StartCommand) Execute(args []string) error {
	withErrors := false
	repeat := true

	// Open the connection to the factory database
	openDatabase()

	// Check that the parameters are set in the config file or command line
	if err := cmd.verifyParameters(); err != nil {
		return err
	}

	for repeat {
		withErrors = false

		// Initialize the factory client
		client := NewFactoryClient(
			datastore.Environ.Config.SyncURL, datastore.Environ.Config.SyncUser, datastore.Environ.Config.SyncAPIKey)

		// Sync the accounts
		log.Info("Sync the accounts from the cloud")
		err := client.Accounts()
		if err != nil {
			withErrors = true
		}

		// Sync the signing-keys
		log.Info("Sync the signing-keys from the cloud")
		err = client.SigningKeys()
		if err != nil {
			withErrors = true
		}

		// Sync the models
		log.Info("Sync the models from the cloud")
		err = client.Models()
		if err != nil {
			withErrors = true
		}

		// Sync the signing logs
		log.Info("Sync the signing logs to the cloud")
		err = client.SigningLogs()
		if err != nil {
			withErrors = true
		}

		if withErrors {
			log.Error("Sync completed with errors")
		}

		if cmd.Daemon {
			// For daemon mode, wait before re-running the sync
			time.Sleep(sleepHours * time.Hour)
		} else {
			// For command mode, not need to repeat
			repeat = false
		}
	}

	if withErrors {
		return errors.New("Sync completed with errors")
	}

	return nil
}

func (cmd StartCommand) verifyParameters() error {

	// Use the sync parameters from config file first
	if len(datastore.Environ.Config.SyncURL) == 0 {
		datastore.Environ.Config.SyncURL = cmd.URL
	}
	if len(datastore.Environ.Config.SyncUser) == 0 {
		datastore.Environ.Config.SyncUser = cmd.Username
	}
	if len(datastore.Environ.Config.SyncAPIKey) == 0 {
		datastore.Environ.Config.SyncAPIKey = cmd.APIKey
	}

	if len(datastore.Environ.Config.SyncURL) == 0 || len(datastore.Environ.Config.SyncUser) == 0 || len(datastore.Environ.Config.SyncAPIKey) == 0 {
		return errors.New("The cloud serial vault URL, username and API key must be provided")
	}

	return nil
}
