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

// KeypairStoreType defines the capabilities of a keypair storage method
type KeypairStoreType struct {
	Name    string
	CanSign bool
}

// Understood keypair storage types
var (
	FilesystemStore = KeypairStoreType{"filesystem", false}
	TPM20Store      = KeypairStoreType{"tpm2.0", true}
)

// Common error messages.
var (
	ErrorInvalidKeystoreType = errors.New("Invalid keystore type specified")
)

// KeypairStore interface to wrap the signing-key store interactions for all store types
type KeypairStore interface {
	ImportSigningKey(string, asserts.PrivateKey) error
	SignAssertion(*asserts.AssertionType, map[string]string, []byte, string) (asserts.Assertion, error)
}

// KeypairDatabase holds the
type KeypairDatabase struct {
	KeyStoreType KeypairStoreType
	*asserts.Database
}

var keypairDB KeypairDatabase

// GetKeyStore returns the keystore as defined in the config file
func GetKeyStore(config ConfigSettings) (*KeypairDatabase, error) {
	switch {
	case config.KeyStoreType == TPM20Store.Name:
		return nil, nil

	case config.KeyStoreType == FilesystemStore.Name:
		fsStore, err := asserts.OpenFSKeypairManager(config.KeyStorePath)
		if err != nil {
			return nil, err
		}
		db, err := asserts.OpenDatabase(&asserts.DatabaseConfig{
			KeypairManager: fsStore,
		})

		keypairDB = KeypairDatabase{FilesystemStore, db}
		return &keypairDB, err

	default:
		return nil, ErrorInvalidKeystoreType
	}
}

// ImportSigningKey adds a new signing-key for an authority into the keypair store
func (kdb *KeypairDatabase) ImportSigningKey(authorityID string, privateKey asserts.PrivateKey) error {

	switch {
	case kdb.KeyStoreType.Name == TPM20Store.Name:
		return nil

	default:
		// Keypairs are handled by the ubuntu-core library, so this is a pass-through to the core library
		return kdb.ImportKey(authorityID, privateKey)
	}
}

// SignAssertion signs an assertion using the signing-key from the keypair store
func (kdb *KeypairDatabase) SignAssertion(assertType *asserts.AssertionType, headers map[string]string, body []byte, keyID string) (asserts.Assertion, error) {

	switch {
	case kdb.KeyStoreType.Name == TPM20Store.Name:
		return nil, nil

	default:
		// Filesystem keypairs are handled by the ubuntu-core library, so this is a pass-through to the core library
		return kdb.Sign(assertType, headers, body, keyID)
	}
}
