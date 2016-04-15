// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2017 Canonical Ltd
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

package service

import (
	"errors"

	"github.com/ubuntu-core/snappy/asserts"
)

// keypairDatabase holds the storage of the private keys for signing. The are
// accessed by using the authority-id and key-id.
var keypairDatabase *asserts.Database

// GetKeyStore returns the keystore as defined in the config file
func GetKeyStore(config ConfigSettings) (*asserts.Database, error) {
	switch {
	case config.KeyStoreType == "filesystem":
		fsStore, err := asserts.OpenFSKeypairManager(config.KeyStorePath)
		if err != nil {
			return nil, err
		}
		db, err := asserts.OpenDatabase(&asserts.DatabaseConfig{
			KeypairManager: fsStore,
		})
		return db, err
	}
	return nil, errors.New("Invalid keystore type specified.")
}
