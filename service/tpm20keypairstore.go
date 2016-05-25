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
	"io"

	"github.com/google/go-tpm/tpm"
	"github.com/snapcore/snapd/asserts"
)

// TPM20KeypairStore is the storage container for signing-keys in the TPM2.0 device
type TPM20KeypairStore struct {
	path string
	rw   io.ReadWriter
}

// OpenTPMStore opens access to the TPM2.0 device
func OpenTPMStore(path string) (io.ReadWriter, error) {
	// Use the TPM library to open the store
	rw, err := tpm.OpenTPM(path)
	return rw, err
}

// TPM20ImportKey adds a new signing-key to the TPM2.0 store
func (tpmStore *TPM20KeypairStore) TPM20ImportKey(authorityID string, privateKey asserts.PrivateKey) error {
	// Params: heirarchy, public, private, flag

	// Output the public and private keys to temporary files

	// Use tpm2_tools to load the external keypair

	return errors.New("Not implemented")
}
