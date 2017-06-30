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
	"encoding/base64"
	"io/ioutil"
	"testing"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/crypt"
	"github.com/snapcore/snapd/asserts"
)

func getTPMKeyStore() (*KeypairDatabase, error) {
	// Set up the environment variables
	config := config.Settings{KeyStoreType: "tpm2.0", KeyStoreSecret: "this needs to be 32 bytes long!!"}
	Environ = &Env{Config: config}

	return getKeyStore(config)
}

func getTPMKeyStoreWithMockCommand() *KeypairDatabase {
	// Set up the environment variables
	config := config.Settings{KeyStorePath: "../keystore", KeyStoreType: "tpm2.0", KeyStoreSecret: "this needs to be 32 bytes long!!"}
	Environ = &Env{Config: config, DB: &MockDB{}}

	tpm20 := TPM20KeypairOperator{config.KeyStorePath, config.KeyStoreSecret, &mockTPM20Command{}}

	// Prepare the memory store for the unsealed keys
	memStore := asserts.NewMemoryKeypairManager()
	db, _ := asserts.OpenDatabase(&asserts.DatabaseConfig{
		KeypairManager: memStore,
	})

	keypairDB = KeypairDatabase{TPM20Store, db, &tpm20}
	return &keypairDB
}

func TestTPMGetKeyStore(t *testing.T) {
	keystore, err := getTPMKeyStore()
	if err != nil {
		t.Error("Error setting up the TPM keystore")
	}
	if keystore == nil {
		t.Error("Nil keystore returned")
	}
}

func TestTPMEncryptDecrypt(t *testing.T) {

	plainText := "fake-hmac-ed-data"

	cipherText, err := crypt.EncryptKey(plainText, "this needs to be 32 bytes long!!")
	if err != nil {
		t.Errorf("Error encrypting text: %v", err)
	}
	if string(cipherText[:]) == plainText {
		t.Error("Invalid encryption")
	}

	plainTextAgain, err := crypt.DecryptKey(cipherText, "this needs to be 32 bytes long!!")
	if err != nil {
		t.Errorf("Error decrypting text: %v", err)
	}
	if string(plainTextAgain[:]) != plainText {
		t.Error("Invalid decryption")
	}
}

func TestGenerateAuthKey(t *testing.T) {

	authKey := crypt.GenerateAuthKey("Hello", "World")
	if authKey != "Hello/World" {
		t.Errorf("Error generating the auth-key: %v", authKey)
	}
}

func TestTPMCreateKey(t *testing.T) {
	// Set up the environment variables
	config := config.Settings{KeyStorePath: "../keystore", KeyStoreType: "tpm2.0", KeyStoreSecret: "this needs to be 32 bytes long!!"}
	Environ = &Env{Config: config, DB: &MockDB{}}

	tpm20 := TPM20KeypairOperator{config.KeyStorePath, config.KeyStoreSecret, &mockTPM20Command{}}

	err := tpm20.createKey("primaryKeyContextPath", algKeyedHash, "test", "do-not-find")
	if err != nil {
		t.Errorf("Error creating the TPM key: %v", err)
	}
}

func TestTPMGenerateEncryptionKey(t *testing.T) {
	// Set up the environment variables
	config := config.Settings{KeyStorePath: "../keystore", KeyStoreType: "tpm2.0", KeyStoreSecret: "this needs to be 32 bytes long!!"}
	Environ = &Env{Config: config, DB: &MockDB{}}

	tpm20 := TPM20KeypairOperator{config.KeyStorePath, config.KeyStoreSecret, &mockTPM20Command{}}

	hashedAuthKey, err := tpm20.generateEncryptionKey("System", "12345678abcdef")
	if err != nil {
		t.Errorf("Error encrypting the TPM auth-key: %v", err)
	}
	if hashedAuthKey != "fake-hmac-ed-data" {
		t.Errorf("Error encrypting the auth-key: %v", hashedAuthKey)
	}
}

func TestTPMImportKeyUnsealKey(t *testing.T) {
	keypairDB := getTPMKeyStoreWithMockCommand()

	signingKey, err := ioutil.ReadFile("../keystore/TestKey.asc")
	if err != nil {
		t.Errorf("Error reading the signing-key file: %v", err)
	}
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	sealedSigningKey, err := keypairDB.keypairOperator.ImportKeypair("System", "12345678abcdef", encodedSigningKey)
	if err != nil {
		t.Errorf("Error encrypting the signing-key: %v", err)
	}
	if sealedSigningKey == encodedSigningKey {
		t.Error("The sealed and unsealed signing-keys are the same")
	}

	err = keypairDB.keypairOperator.UnsealKeypair("System", "12345678abcdef", sealedSigningKey)
	if err != nil {
		t.Errorf("Error decrypting the signing-key: %v", err)
	}
}
