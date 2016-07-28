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
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/snapcore/snapd/asserts"
	"golang.org/x/crypto/openpgp/packet"
)

// decodePublicKey replicates a private method in snapcore asserts to convert the device-key header
// into a usable PublicKey format.
func decodePublicKey(pubKey []byte) (asserts.PublicKey, error) {
	pkt, err := decodeOpenpgp(pubKey, "public key")
	if err != nil {
		return nil, err
	}
	pubk, ok := pkt.(*packet.PublicKey)
	if !ok {
		return nil, fmt.Errorf("expected public key, got instead: %T", pkt)
	}
	return asserts.OpenPGPPublicKey(pubk), nil
}

func generateAuthKey(authorityID, keyID string) string {
	return strings.Join([]string{authorityID, "/", keyID}, "")
}

func createSecret(length int) (string, error) {
	rb := make([]byte, length)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(rb), nil
}

// encryptKey uses symmetric encryption to encrypt the data for storage
func encryptKey(plainTextKey, keyText string) ([]byte, error) {
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

func decryptKey(sealedKey []byte, keyText string) ([]byte, error) {
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

func decodeOpenpgp(formatAndBase64 []byte, kind string) (packet.Packet, error) {
	if len(formatAndBase64) == 0 {
		return nil, fmt.Errorf("empty %s", kind)
	}
	format, data, err := splitFormatAndBase64Decode(formatAndBase64)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", kind, err)
	}
	if format != "openpgp" {
		return nil, fmt.Errorf("unsupported %s format: %q", kind, format)
	}
	pkt, err := packet.Read(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("could not decode %s data: %v", kind, err)
	}
	return pkt, nil
}

func splitFormatAndBase64Decode(formatAndBase64 []byte) (string, []byte, error) {
	parts := bytes.SplitN(formatAndBase64, []byte(" "), 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("expected format and base64 data separated by space")
	}
	buf := make([]byte, base64.StdEncoding.DecodedLen(len(parts[1])))
	n, err := base64.StdEncoding.Decode(buf, parts[1])
	if err != nil {
		return "", nil, fmt.Errorf("could not decode base64 data: %v", err)
	}
	return string(parts[0]), buf[:n], nil
}
