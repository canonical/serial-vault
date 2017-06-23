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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"log"

	"github.com/CanonicalLtd/serial-vault/utils"
)

// DatabaseKeypairOperator is the storage container for signing-keys in the database
type DatabaseKeypairOperator struct{}

// ImportKeypair adds a new signing-key to the database key store.
// The main operations:
//  * Use the auth/key-id as the key
//  * Use HMAC the auth-key
//  * Use AES symmetric encryption to encrypt the signing-key file
//  * Encrypt the auth-key and store in the database
func (dbStore *DatabaseKeypairOperator) ImportKeypair(authorityID, keyID, base64PrivateKey string) (string, error) {
	// Generate an HMAC hash of the auth-key
	authKeyHash, err := dbStore.generateEncryptionKey(authorityID, keyID)

	// Use the HMAC-ed auth-key as the key to encrypt the signing-key
	sealedSigningKey, err := utils.EncryptKey(base64PrivateKey, authKeyHash)

	// base64 encode the sealed signing-key for storage
	base64SealedSigningkey := base64.StdEncoding.EncodeToString(sealedSigningKey)

	return base64SealedSigningkey, err
}

func (dbStore *DatabaseKeypairOperator) generateEncryptionKey(authorityID, keyID string) (string, error) {
	keyText := utils.GenerateAuthKey(authorityID, keyID)
	secretText, err := utils.CreateSecret(32)
	if err != nil {
		return "", err
	}
	h := hmac.New(sha256.New, []byte(secretText))
	h.Write([]byte(keyText))
	encryptionKey := string(h.Sum(nil)[:])

	// Encrypt and store the auth-key hash
	encryptedAuthKeyHash, err := utils.EncryptKey(string(encryptionKey[:]), Environ.Config.KeyStoreSecret)
	if err != nil {
		return "", err
	}

	// Encrypt the HMAC-ed auth-key for storage
	base64AuthKeyHash := base64.StdEncoding.EncodeToString([]byte(encryptedAuthKeyHash))
	Environ.DB.PutSetting(Setting{Code: utils.GenerateAuthKey(authorityID, keyID), Data: base64AuthKeyHash})

	return string(encryptionKey[:]), nil
}

// UnsealKeypair unseals a database-stored signing-key and stores it in the memory store
//  * Decrypt the auth-key
//  * Decrypt the signing key
//  * Load into memory store
func (dbStore *DatabaseKeypairOperator) UnsealKeypair(authorityID string, keyID string, base64SealedSigningKey string) error {
	return unsealKeypair(authorityID, keyID, base64SealedSigningKey)
}

func unsealKeypair(authorityID string, keyID string, base64SealedSigningKey string) error {

	// Check if we have already unsealed the key into the memory store
	_, err := keypairDB.PublicKey(keyID)

	if err != nil {
		// The key has not been unsealed and stored in the memory store

		// Decode and decrypt the auth-key
		authKeySetting, err := Environ.DB.GetSetting(utils.GenerateAuthKey(authorityID, keyID))
		if err != nil {
			log.Println("Cannot find the auth-key for the signing-key")
			return err
		}

		// Decode the auth-key from storage
		encryptedAuthKey, err := base64.StdEncoding.DecodeString(authKeySetting.Data)
		if err != nil {
			log.Println("Could not decode the auth-key for the signing-key")
			return err
		}

		// Decrypt the decoded auth-key
		authKey, err := utils.DecryptKey(encryptedAuthKey, Environ.Config.KeyStoreSecret)
		if err != nil {
			log.Println("Could not decrypt the auth-key for the signing-key")
			return err
		}

		// Decode and decrypt the signing-key
		sealedSigningKey, err := base64.StdEncoding.DecodeString(base64SealedSigningKey)
		if err != nil {
			log.Println("Could not decode the signing-key")
			return err
		}
		base64SigningKey, err := utils.DecryptKey(sealedSigningKey, string(authKey[:]))
		if err != nil {
			log.Println("Could not decrypt the signing-key")
			return err
		}

		// Convert the byte array to an asserts key
		privateKey, errorCode, err := utils.DeserializePrivateKey(string(base64SigningKey[:]))
		if err != nil {
			log.Printf("Error generating the asserts private-key: %v", errorCode)
			return err
		}

		// Add the private-key to the memory keypair store
		err = keypairDB.ImportKey(privateKey)
		if err != nil {
			log.Println("Error importing the private-key to memory store")
			return err
		}

	}

	return nil
}
