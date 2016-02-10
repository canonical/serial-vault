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

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/clearsign"
)

// ClearSign the assertions using the private key and return a clear-signed buffer.
// If the private key is encrypted, the passphrase must be supplied to decrypt
// the key.
func ClearSign(assertions, armoredPrivateKey, passphrase string) ([]byte, error) {
	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(armoredPrivateKey))
	if err != nil {
		return nil, err
	}

	// Decrypt the private key, if it is encrypted
	privateKey := keyring[0].PrivateKey
	if privateKey.Encrypted {
		err = privateKey.Decrypt([]byte(passphrase))
		if err != nil {
			return nil, err
		}
	}

	// Sign the assertions and save the result in the buffer
	var buf bytes.Buffer
	plaintext, err := clearsign.Encode(&buf, privateKey, nil)
	if err != nil {
		return nil, err
	}
	_, err = plaintext.Write([]byte(assertions))
	if err != nil {
		return nil, err
	}
	err = plaintext.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
