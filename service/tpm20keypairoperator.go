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
	"log"
	"os"

	"github.com/CanonicalLtd/serial-vault/datastore"
)

const (
	algSHA256    = "0x000B"
	algSHA512    = "0x000D"
	algRSA       = "0x0001"
	algKeyedHash = "0x0008"
	algSymCipher = "0x0025"

	handleHash = "0x81010002"
	handleSym  = "0x81010003"
)

// TPM20KeypairOperator is the operator that handles interactions with the TPM2.0 device and signing-keys
type TPM20KeypairOperator struct {
	path       string
	secret     string
	tpmCommand TPM20Command
}

// ImportKeypair adds a new signing-key to the TPM2.0 store.
// The main TPM2.0 operations:
//  * Use the auth/key-id as the key
//  * Create an KeyedHash key for context
//  * Use TPM to HMAC the auth-key (using KeyedHash context)
//  * Use AES symmetric encryption to encrypt the signing-key file (using Go)
//  * Encrypt the auth-key and store in the database (using Go)
func (tpmStore *TPM20KeypairOperator) ImportKeypair(authorityID, keyID, base64PrivateKey string) (string, error) {

	// Get the parent context from the database settings table
	setting, err := Environ.DB.GetSetting("parent")
	if err != nil {
		return "", nil
	}

	// Create a KeyedHash key to create the context for the hash key
	err = tpmStore.createKey(setting.Data, algKeyedHash, "hash", handleHash)
	if err != nil {
		return "", err
	}

	// Generate an HMAC hash of the signing-key details
	authKeyHash, err := tpmStore.generateEncryptionKey(authorityID, keyID)
	if err != nil {
		return "", err
	}

	// Use the HMAC-ed auth-key as the key to encrypt the signing-key
	sealedSigningKey, err := encryptKey(base64PrivateKey, authKeyHash)

	// base64 encode the sealed signing-key for storage
	base64SealedSigningkey := base64.StdEncoding.EncodeToString(sealedSigningKey)

	return base64SealedSigningkey, err
}

// UnsealKeypair unseals a TPM-sealed signing-key and stores it in the memory store
//  * Decrypt the auth-key
//  * Decrypt the signing key
//  * Load into memory store
func (tpmStore *TPM20KeypairOperator) UnsealKeypair(authorityID string, keyID string, base64SealedSigningKey string) error {

	return unsealKeypair(authorityID, keyID, base64SealedSigningKey)
}

// generateEncryptionKey takes the authority and the key details and uses the TPM 2.0 module to create a HMAC hash of the data.
// This hash is used as the key for symmetric encryption of the signing-key.
func (tpmStore *TPM20KeypairOperator) generateEncryptionKey(authorityID, keyID string) (string, error) {
	// Generate a file with the plain-text base of the symmetric encryption key
	keyText := generateAuthKey(authorityID, keyID)
	tmpfile, err := ioutil.TempFile("", "tmp")
	if err != nil {
		return "", err
	}
	os.Remove(tmpfile.Name())
	err = ioutil.WriteFile(tmpfile.Name(), []byte(keyText), 0600)
	if err != nil {
		return "", err
	}

	// Create a file to hold the symmetric encryption key that will be used
	hashKey, err := ioutil.TempFile("", "tmp")
	if err != nil {
		return "", err
	}
	os.Remove(hashKey.Name())

	// Use the TPM module to hash the plain-text
	err = tpmStore.tpmCommand.runCommand("tpm2_hmac", "-k", handleHash, "-g", algSHA256, "-I", tmpfile.Name(), "-o", hashKey.Name())
	if err != nil {
		return "", err
	}

	// Read the HMAC-ed data
	encryptionKey, err := ioutil.ReadFile(hashKey.Name())
	if err != nil {
		return "", err
	}

	// Encrypt and store the auth-key hash
	encryptedAuthKeyHash, err := encryptKey(string(encryptionKey[:]), tpmStore.secret)
	if err != nil {
		return "", err
	}

	// Encrypt the HMAC-ed auth-key for storage
	base64AuthKeyHash := base64.StdEncoding.EncodeToString([]byte(encryptedAuthKeyHash))
	Environ.DB.PutSetting(datastore.Setting{Code: generateAuthKey(authorityID, keyID), Data: base64AuthKeyHash})

	// Remove the temporary files
	os.Remove(tmpfile.Name())
	os.Remove(hashKey.Name())

	return string(encryptionKey[:]), nil
}

// createKey uses the TPM 2.0 module to generate a key and load it into the TPM 2.0 module.
func (tpmStore *TPM20KeypairOperator) createKey(primaryKeyContextPath, algorithm, prefix, handle string) error {

	// Check if we've already created a key for this operation
	_, err := Environ.DB.GetSetting(handle)
	if err == nil {
		// Already created a key, so let's use it
		log.Printf("Using the existing key for '%s'", prefix)
		return nil
	}

	// Generate a unique file name to hold the key context
	keyContext, err := ioutil.TempFile(tpmStore.path, ".context")
	if err != nil {
		return err
	}

	// Generate a unique file name to hold the public and private key and name file
	publicKey, err := ioutil.TempFile(tpmStore.path, ".pub")
	if err != nil {
		return err
	}
	privateKey, err := ioutil.TempFile(tpmStore.path, ".prv")
	if err != nil {
		return err
	}
	nameFile, err := ioutil.TempFile(tpmStore.path, ".name")
	if err != nil {
		return err
	}

	// Remove the temporary files as the TPM2.0 tools will create them
	os.Remove(keyContext.Name())
	os.Remove(publicKey.Name())
	os.Remove(privateKey.Name())
	os.Remove(nameFile.Name())

	// Create the key in the heirarchy
	err = tpmStore.tpmCommand.runCommand("tpm2_create", "-g", algSHA256, "-G", algorithm, "-c", primaryKeyContextPath, "-o", publicKey.Name(), "-O", privateKey.Name())
	if err != nil {
		return err
	}

	// Load the key in the heirarchy
	err = tpmStore.tpmCommand.runCommand("tpm2_load", "-c", primaryKeyContextPath, "-u", publicKey.Name(), "-r", privateKey.Name(), "-n", nameFile.Name(), "-C", keyContext.Name())
	if err != nil {
		return err
	}

	// Move the key to non-volatile storage, so it will survive a power cycle
	err = tpmStore.tpmCommand.runCommand("tpm2_evictcontrol", "-A", "o", "-c", keyContext.Name(), "-S", handle)
	if err != nil {
		return err
	}

	// Store the handle so we know that it has been created
	Environ.DB.PutSetting(datastore.Setting{Code: handle, Data: handle})

	// Clean up the created files
	os.Remove(keyContext.Name())
	os.Remove(publicKey.Name())
	os.Remove(privateKey.Name())
	os.Remove(nameFile.Name())

	return nil
}
