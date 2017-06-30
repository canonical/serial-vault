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

package datastore

import (
	"errors"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/snapcore/snapd/asserts"
)

type errorMockKeypairManager struct{}

// GetMemoryKeyStore creates a mocked keystore
func GetMemoryKeyStore(config config.Settings) (*KeypairDatabase, error) {
	db, err := asserts.OpenDatabase(&asserts.DatabaseConfig{
		KeypairManager: asserts.NewMemoryKeypairManager(),
	})
	kdb := KeypairDatabase{FilesystemStore, db, nil}
	return &kdb, err
}

func (emkdb *errorMockKeypairManager) Get(keyID string) (asserts.PrivateKey, error) {
	return nil, errors.New("MOCK error fetching the private key")
}

func (emkdb *errorMockKeypairManager) Put(privKey asserts.PrivateKey) error {
	return errors.New("MOCK error saving the private key")
}

// GetErrorMockKeyStore creates a mocked keystore
func GetErrorMockKeyStore(config config.Settings) (*KeypairDatabase, error) {
	mockStore := new(errorMockKeypairManager)

	db, err := asserts.OpenDatabase(&asserts.DatabaseConfig{
		KeypairManager: mockStore,
	})
	kdb := KeypairDatabase{FilesystemStore, db, nil}
	return &kdb, err
}
