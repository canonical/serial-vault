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
	"encoding/base64"
	"io/ioutil"
	"testing"

	"github.com/CanonicalLtd/serial-vault/datastore"
)

func getDatabaseKeyStore() (*KeypairDatabase, error) {
	// Set up the environment variables
	config := ConfigSettings{KeyStoreType: "database", KeyStoreSecret: "this needs to be something secure"}
	Environ = &Env{Config: config, DB: &datastore.MockDB{}}

	return GetKeyStore(config)
}

func TestDatabaseGetKeyStore(t *testing.T) {
	keystore, err := getDatabaseKeyStore()
	if err != nil {
		t.Error("Error setting up the database keystore")
	}
	if keystore == nil {
		t.Error("Nil keystore returned")
	}
}

func TestDatabaseGenerateEncryptionKey(t *testing.T) {
	// Set up the environment variables
	config := ConfigSettings{KeyStoreType: "database", KeyStoreSecret: "this needs to be something secure"}
	Environ = &Env{Config: config, DB: &datastore.MockDB{}}

	dbOperator := DatabaseKeypairOperator{}

	hashedAuthKey, err := dbOperator.generateEncryptionKey("System", "12345678abcdef")
	if err != nil {
		t.Errorf("Error encrypting the auth-key: %v", err)
	}
	if hashedAuthKey == "System/12345678abcdef" {
		t.Errorf("Error encrypting the auth-key: %v", hashedAuthKey)
	}
}

func TestImportUnsealKeypair(t *testing.T) {
	keypairDB, _ := getDatabaseKeyStore()

	signingKey, err := ioutil.ReadFile("../keystore/TestKey.asc")
	if err != nil {
		t.Errorf("Error reading the signing-key file: %v", err)
	}
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	sealedSigningKey, err := keypairDB.keypairOperator.ImportKeypair("System", "abcdef12345678", encodedSigningKey)
	if err != nil {
		t.Errorf("Error encrypting the signing-key: %v", err)
	}
	if sealedSigningKey == encodedSigningKey {
		t.Error("The sealed and unsealed signing-keys are the same")
	}

	err = keypairDB.keypairOperator.UnsealKeypair("System", "abcdef12345678", sealedSigningKey)
	if err != nil {
		t.Errorf("Error decrypting the signing-key: %v", err)
	}
}

func TestImportKey(t *testing.T) {
	keypairDB, _ := getDatabaseKeyStore()

	signingKey, err := ioutil.ReadFile("../keystore/TestKey.asc")
	if err != nil {
		t.Errorf("Error reading the signing-key file: %v", err)
	}
	base64PrivateKey := base64.StdEncoding.EncodeToString(signingKey)

	_, _, err = keypairDB.ImportSigningKey("System", base64PrivateKey)
	if err != nil {
		t.Errorf("Error importing the signing-key file: %v", err)
	}

}
