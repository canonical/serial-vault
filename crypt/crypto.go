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

package crypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"github.com/CanonicalLtd/serial-vault/service/log"

	"github.com/snapcore/snapd/asserts"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

// GenerateAuthKey generates an key from the signing key details
func GenerateAuthKey(authorityID, keyID string) string {
	return strings.Join([]string{authorityID, "/", keyID}, "")
}

// CreateSecret generates a secret that can be used for encryption
func CreateSecret(length int) (string, error) {
	rb := make([]byte, length)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(rb), nil
}

// EncryptKey uses symmetric encryption to encrypt the data for storage
func EncryptKey(plainTextKey, keyText string) ([]byte, error) {
	// The AES key needs to be 16 or 32 bytes i.e. AES-128 or AES-256
	aesKey := padRight(keyText, "x", 32)

	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		log.Printf("Error creating the cipher block: %v", err)
		return nil, err
	}

	// The IV needs to be unique, but not secure. Including it at the start of the plaintext
	ciphertext := make([]byte, aes.BlockSize+len(plainTextKey))
	iv := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		log.Printf("Error creating the IV for the cipher: %v", err)
		return nil, err
	}

	// Use CFB mode for the encryption
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plainTextKey))

	return ciphertext, nil
}

// DecryptKey handles the decryption of a sealed signing key
func DecryptKey(sealedKey []byte, keyText string) ([]byte, error) {
	aesKey := padRight(keyText, "x", 32)

	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		log.Printf("Error creating the cipher block: %v", err)
		return nil, err
	}

	if len(sealedKey) < aes.BlockSize {
		return nil, errors.New("Cipher text too short")
	}

	iv := sealedKey[:aes.BlockSize]
	sealedKey = sealedKey[aes.BlockSize:]

	// Use CFB mode for the decryption
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(sealedKey, sealedKey)

	return sealedKey, nil
}

// DeserializePrivateKey decodes a base64 encoded private key file and converts
// it to a private key that can be used for storage in the keypair store
func DeserializePrivateKey(base64PrivateKey string) (asserts.PrivateKey, string, error) {
	// The private-key is base64 encoded, so we need to decode it
	decodedPrivateKey, err := base64.StdEncoding.DecodeString(base64PrivateKey)
	if err != nil {
		return nil, "error-decode-key", err
	}

	return privateKeyToAssertsKey(decodedPrivateKey)
}

func privateKeyToAssertsKey(key []byte) (asserts.PrivateKey, string, error) {
	const errorInvalidKey = "invalid-keypair"

	// Validate the signing-key
	block, err := armor.Decode(bytes.NewReader(key))
	if err != nil {
		return nil, errorInvalidKey, err
	}

	pkt, err := packet.Read(block.Body)
	if err != nil {
		return nil, errorInvalidKey, err
	}

	privk, ok := pkt.(*packet.PrivateKey)
	if !ok {
		return nil, errorInvalidKey, errors.New("Not a private key")
	}
	if _, ok := privk.PrivateKey.(*rsa.PrivateKey); !ok {
		return nil, errorInvalidKey, errors.New("Not an RSA private key")
	}
	return asserts.RSAPrivateKey(privk.PrivateKey.(*rsa.PrivateKey)), "", nil
}

// padRight truncates a string to a specific length, padding with a named
// character for shorter strings.
func padRight(str, pad string, length int) string {
	for {
		str += pad
		if len(str) > length {
			return str[0:length]
		}
	}
}
