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

package sync

import (
	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/log"
)

// Command defines the options for the serial-vault-admin command-line utility
type Command struct {
	SettingsFile string          `short:"c" long:"config" description:"Path to the config file" default:"./settings.yaml"`
	Start        StartCommand    `command:"sync" alias:"s" description:"Start the factory sync process"`
	Database     DatabaseCommand `command:"database" alias:"d" description:"Database schema update"`
}

// Sync is the implementation of the command configuration for the serial-vault-admin command-line
var Sync Command

func openDatabase() {
	// Check that the database has not been set e.g. by a mock
	if datastore.Environ.DB != nil {
		return
	}

	log.Infof("Open the settings file: %s", Sync.SettingsFile)
	config.ReadConfig(&datastore.Environ.Config, Sync.SettingsFile)

	// Open the connection to the database
	datastore.OpenSysDatabase(datastore.Environ.Config.Driver, datastore.Environ.Config.DataSource)
}
