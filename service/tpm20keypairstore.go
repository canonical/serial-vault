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

/* Process

createdb
[x] takeownership
[x] createprimary

importkey
[x] Use the auth/key-id as the key
[x] Create an KeyedHash key for context
[x] Use TPM to HMAC the data (using KeyedHash context)
[x] Create an TPM assymetric key (using RSA context)
[ ] Use AES symmetric encryption to encrypt the signing-key file (using TPM)
[ ] Encrypt the auth-key and store that locally


sign
[ ] Decrypt the key using the assymetric key
[x] Decrypt the signing key
[x] Load into memory store

*/

package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/snapcore/snapd/asserts"
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

// TPM20KeypairStore is the storage container for signing-keys in the TPM2.0 device
type TPM20KeypairStore struct {
	path string
	rw   io.ReadWriteCloser
}

// getAuth creates a hash from the keypair authority
func getAuth(authorityID string) [20]byte {
	auth := sha1.Sum([]byte(authorityID))
	return auth
}

// OpenTPMStore opens access to the TPM2.0 device
func OpenTPMStore(path string) (io.ReadWriteCloser, error) {
	// TODO: Check if we still need this method
	// Use the TPM library to open the store
	// rw, err := tpm.OpenTPM(path)
	// return rw, err
	return nil, nil
}

// TPM20ImportKey adds a new signing-key to the TPM2.0 store
func (tpmStore *TPM20KeypairStore) TPM20ImportKey(authorityID, keyID, base64PrivateKey string) (string, error) {

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
	sealedSigningKey, err := tpmStore.encryptSigningKey(base64PrivateKey, authKeyHash)

	// base64 encode the sealed signing-key for storage
	base64SealedSigningkey := base64.StdEncoding.EncodeToString(sealedSigningKey)

	// Encrypt the HMAC-ed auth-key for storage
	base64AuthKeyHash := base64.StdEncoding.EncodeToString([]byte(authKeyHash))
	Environ.DB.PutSetting(Setting{Code: tpmStore.generateAuthKey(authorityID, keyID), Data: base64AuthKeyHash})

	// Create an assymetric key to create the context for symmetric encryption
	err = tpmStore.createKey(setting.Data, algSymCipher, "sym", handleSym)
	if err != nil {
		return "", err
	}

	// TODO: encrypt and store the auth-key hash
	// TODO: clean up temporary files

	return base64SealedSigningkey, err
}

// TPM20UnsealKey unseals a TPM-sealed signing-key and stores it in the memory store
func (tpmStore *TPM20KeypairStore) TPM20UnsealKey(assertType *asserts.AssertionType, headers map[string]string, body []byte, authorityID string, keyID string, base64SealedSigningKey string) error {

	// Check if we have already unsealed the key into the memory store
	_, err := keypairDB.PublicKey(authorityID, keyID)

	if err != nil {
		// The key has not been unsealed and stored in the memory store

		// Decode and decrypt the auth-key
		authKeySetting, err := Environ.DB.GetSetting(tpmStore.generateAuthKey(authorityID, keyID))
		if err != nil {
			log.Println("Cannot find the auth-key for the signing-key")
			return err
		}
		// TODO: need to also decrypt the decoded auth-key
		authKey, err := base64.StdEncoding.DecodeString(authKeySetting.Data)
		if err != nil {
			log.Println("Could not decode the auth-key for the signing-key")
			return err
		}

		// Decode and decrypt the signing-key
		sealedSigningKey, err := base64.StdEncoding.DecodeString(base64SealedSigningKey)
		if err != nil {
			log.Println("Could not decode the signing-key")
			return err
		}
		base64SigningKey, err := tpmStore.decryptSigningKey(sealedSigningKey, string(authKey[:]))
		if err != nil {
			log.Println("Could not decrypt the signing-key")
			return err
		}

		// Convert the byte array to an asserts key
		privateKey, errorCode, err := deserializePrivateKey(string(base64SigningKey[:]))
		if err != nil {
			log.Printf("Error generating the asserts private-key: %v", errorCode)
			return err
		}

		// Add the private-key to the memory keypair store
		err = keypairDB.ImportKey(authorityID, privateKey)
		if err != nil {
			log.Println("Error importing the private-key to memory store")
			return err
		}

	}

	return nil
}

func (tpmStore *TPM20KeypairStore) generateAuthKey(authorityID, keyID string) string {
	return strings.Join([]string{authorityID, "/", keyID}, "")
}

// generateEncryptionKey takes the authority and the key details and uses the TPM 2.0 module to create a HMAC hash of the data.
// This hash is used as the key for symmetric encryption of the signing-key.
func (tpmStore *TPM20KeypairStore) generateEncryptionKey(authorityID, keyID string) (string, error) {
	// Generate a file with the plain-text base of the symmetric encryption key
	keyText := tpmStore.generateAuthKey(authorityID, keyID)
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
	cmd := exec.Command("tpm2_hmac", "-k", handleHash, "-g", algSHA256, "-I", tmpfile.Name(), "-o", hashKey.Name())
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Error in TPM hmac, %v", err)
		log.Println(string(out[:]))
		return "", err
	}

	// Read the HMAC-ed data
	encryptionKey, err := ioutil.ReadFile(hashKey.Name())
	if err != nil {
		return "", err
	}

	// Remove the temporary files
	os.Remove(tmpfile.Name())
	os.Remove(hashKey.Name())

	return string(encryptionKey[:]), nil
}

// createKey uses the TPM 2.0 module to generate a key and load it into the TPM 2.0 module.
func (tpmStore *TPM20KeypairStore) createKey(primaryKeyContextPath, algorithm, prefix, handle string) error {

	// Check if we've already created a key for this operation
	_, err := Environ.DB.GetSetting(strings.Join([]string{prefix, "context"}, "_"))
	if err == nil {
		// Already created a key, so let's use it
		log.Printf("Using the existing key for '%s'", prefix)
		return nil
	}

	// Generate a unique file name to hold the key context
	keyContext, err := ioutil.TempFile("keystore", ".context")
	if err != nil {
		return err
	}

	// Generate a unique file name to hold the public and private key and name file
	publicKey, err := ioutil.TempFile("keystore", ".pub")
	if err != nil {
		return err
	}
	privateKey, err := ioutil.TempFile("keystore", ".prv")
	if err != nil {
		return err
	}
	nameFile, err := ioutil.TempFile("keystore", ".name")
	if err != nil {
		return err
	}

	// Remove the temporary files as the TPM2.0 tools will create them
	os.Remove(keyContext.Name())
	os.Remove(publicKey.Name())
	os.Remove(privateKey.Name())
	os.Remove(nameFile.Name())

	// Create the key in the heirarchy
	cmd := exec.Command("tpm2_create", "-g", algSHA256, "-G", algorithm, "-c", primaryKeyContextPath, "-o", publicKey.Name(), "-O", privateKey.Name())
	stdout, err := cmd.Output()
	if err != nil {
		log.Printf("Error in TPM create, %v", err)
		log.Println(string(stdout[:]))
		return err
	}

	// Load the key in the heirarchy
	cmd = exec.Command("tpm2_load", "-c", primaryKeyContextPath, "-u", publicKey.Name(), "-r", privateKey.Name(), "-n", nameFile.Name(), "-C", keyContext.Name())
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Error in TPM load, %v", err)
		log.Println(string(out[:]))
		return err
	}

	// Move the key to non-volatile storage, so it will survive a power cycle
	cmd = exec.Command("tpm2_evictcontrol", "-A", "o", "-c", keyContext.Name(), "-S", handle)
	stdout, err = cmd.Output()
	log.Println(string(stdout[:]))
	if err != nil {
		log.Printf("Error in TPM create, %v", err)
		log.Println(string(stdout[:]))
		return err
	}

	// Clean up the created files
	os.Remove(keyContext.Name())
	os.Remove(publicKey.Name())
	os.Remove(privateKey.Name())
	os.Remove(nameFile.Name())

	return nil
}

// encryptSigningKey uses the symmetric encryption to encrypt the signing-key for storage
func (tpmStore *TPM20KeypairStore) encryptSigningKey(base64PrivateKey, keyText string) ([]byte, error) {
	// The AES key needs to be 16 or 32 bytes i.e. AES-128 or AES-256
	aesKey := padRight(keyText, "x", 32)

	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		log.Printf("Error creating the cipher block: %v", err)
		return nil, err
	}

	// The IV needs to be unique, but not secure. Including it at the start of the plaintext
	ciphertext := make([]byte, aes.BlockSize+len(base64PrivateKey))
	iv := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		log.Printf("Error creating the IV for the cipher: %v", err)
		return nil, err
	}

	// Use CFB mode for the encryption
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(base64PrivateKey))

	return ciphertext, nil
}

func (tpmStore *TPM20KeypairStore) decryptSigningKey(sealedSigningKey []byte, authKey string) ([]byte, error) {
	aesKey := padRight(authKey, "x", 32)

	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		log.Printf("Error creating the cipher block: %v", err)
		return nil, err
	}

	if len(sealedSigningKey) < aes.BlockSize {
		return nil, errors.New("Cipher text too short")
	}

	iv := sealedSigningKey[:aes.BlockSize]
	sealedSigningKey = sealedSigningKey[aes.BlockSize:]

	// Use CFB mode for the decryption
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(sealedSigningKey, sealedSigningKey)

	return sealedSigningKey, nil
}
