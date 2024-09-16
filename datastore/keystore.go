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
	"github.com/CanonicalLtd/serial-vault/crypt"
	"github.com/snapcore/snapd/asserts"
)

// KeypairStoreType defines the capabilities of a keypair storage method
type KeypairStoreType struct {
	Name string
}

// Understood keypair storage types
var (
	FilesystemStore = KeypairStoreType{"filesystem"}
	DatabaseStore   = KeypairStoreType{"database"}
	TPM20Store      = KeypairStoreType{"tpm2.0"}
)

// Common error messages.
var (
	ErrorInvalidKeystoreType = errors.New("Invalid keystore type specified")
)

// KeypairStore interface to wrap the signing-key store interactions for all store types
type KeypairStore interface {
	ImportSigningKey(string, string) (asserts.PrivateKey, string, error)
	SignAssertion(*asserts.AssertionType, map[string]string, []byte, string) (asserts.Assertion, error)
	LoadKeypair(authorityID string, keyID string, base64SealedSigningKey string) error
}

// KeypairOperator interface used by some keypair stores to seal and unseal signing-keys for storage
type KeypairOperator interface {
	ImportKeypair(authorityID, keyID, base64PrivateKey string) (string, error)
	UnsealKeypair(authorityID string, keyID string, base64SealedSigningKey string) error
}

// KeypairDatabase holds the
type KeypairDatabase struct {
	KeyStoreType KeypairStoreType
	*asserts.Database
	keypairOperator KeypairOperator
}

var keypairDB KeypairDatabase

// OpenKeyStore returns the keystore as defined in the config file
func OpenKeyStore(config config.Settings) error {
	keypairDB, err := getKeyStore(config)
	if err != nil {
		return err
	}

	Environ.KeypairDB = keypairDB
	return nil
}

func getKeyStore(config config.Settings) (*KeypairDatabase, error) {
	switch config.KeyStoreType {
	case DatabaseStore.Name:
		// Prepare the memory store for the unsealed keys
		memStore := asserts.NewMemoryKeypairManager()
		db, err := asserts.OpenDatabase(&asserts.DatabaseConfig{
			KeypairManager: memStore,
		})

		dbOperator := DatabaseKeypairOperator{}

		keypairDB = KeypairDatabase{DatabaseStore, db, &dbOperator}
		return &keypairDB, err

	case TPM20Store.Name:
		// Initialize the TPM store
		tpm20 := TPM20KeypairOperator{config.KeyStorePath, config.KeyStoreSecret, &tpm20Command{}}

		// Prepare the memory store for the unsealed keys
		memStore := asserts.NewMemoryKeypairManager()
		db, err := asserts.OpenDatabase(&asserts.DatabaseConfig{
			KeypairManager: memStore,
		})

		keypairDB = KeypairDatabase{TPM20Store, db, &tpm20}
		return &keypairDB, err

	case FilesystemStore.Name:
		fsStore, err := asserts.OpenFSKeypairManager(config.KeyStorePath)
		if err != nil {
			return nil, err
		}
		db, err := asserts.OpenDatabase(&asserts.DatabaseConfig{
			KeypairManager: fsStore,
		})

		keypairDB = KeypairDatabase{FilesystemStore, db, nil}
		return &keypairDB, err

	default:
		return nil, ErrorInvalidKeystoreType
	}
}

// ImportSigningKey adds a new signing-key for an authority into the keypair store
func (kdb *KeypairDatabase) ImportSigningKey(authorityID, base64PrivateKey string) (asserts.PrivateKey, string, error) {
	privateKey, _, err := crypt.DeserializePrivateKey(base64PrivateKey)
	if err != nil {
		return nil, "", err
	}

	switch kdb.KeyStoreType.Name {
	case DatabaseStore.Name:
		fallthrough

	case TPM20Store.Name:
		// Use an internal operator to handle encryption of signing-keys for storage
		sealedPrivateKey, err := kdb.keypairOperator.ImportKeypair(authorityID, privateKey.PublicKey().ID(), base64PrivateKey)
		return privateKey, sealedPrivateKey, err

	default:
		// Keypairs are handled by the snapd library, so this is a pass-through to the core library
		return privateKey, "", kdb.ImportKey(privateKey)
	}
}

// SignAssertion signs an assertion using the signing-key from the keypair store
func (kdb *KeypairDatabase) SignAssertion(assertType *asserts.AssertionType, headers map[string]interface{}, body []byte, authorityID string, keyID string, sealedSigningKey string) (asserts.Assertion, error) {
	switch kdb.KeyStoreType.Name {

	case DatabaseStore.Name:
		fallthrough

	case TPM20Store.Name:
		// Use an internal operator to handle decryption of signing-keys from storage
		err := kdb.keypairOperator.UnsealKeypair(authorityID, keyID, sealedSigningKey)
		if err != nil {
			return nil, err
		}

		// Sign the key using the unsealed key in the memory keypair store
		return kdb.Sign(assertType, headers, body, keyID)

	default:
		// Filesystem keypairs are handled by the snapd library, so this is a pass-through to the core library
		return kdb.Sign(assertType, headers, body, keyID)
	}
}

// LoadKeypair checks if a keypair is in the memory store and (unseals and) loads it if it isn't
func (kdb *KeypairDatabase) LoadKeypair(authorityID string, keyID string, sealedSigningKey string) error {
	switch kdb.KeyStoreType.Name {
	case DatabaseStore.Name:
		fallthrough

	case TPM20Store.Name:
		// Use an internal operator to handle decryption of signing-keys from storage
		err := kdb.keypairOperator.UnsealKeypair(authorityID, keyID, sealedSigningKey)
		return err

	default:
		// Filesystem keypairs are all loaded, so this is a no-op
		return nil
	}
}
