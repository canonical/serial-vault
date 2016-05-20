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

	"github.com/ubuntu-core/snappy/asserts"
)

type errorMockKeypairManager struct{}

func getMemoryKeyStore(config ConfigSettings) (*KeypairDatabase, error) {
	db, err := asserts.OpenDatabase(&asserts.DatabaseConfig{
		KeypairManager: asserts.NewMemoryKeypairManager(),
	})
	keypairDB := KeypairDatabase{FilesystemStore, db}
	return &keypairDB, err
}

func (emkdb *errorMockKeypairManager) Get(authorityID, keyID string) (asserts.PrivateKey, error) {
	return nil, errors.New("MOCK error fetching the private key")
}

func (emkdb *errorMockKeypairManager) Put(authorityID string, privKey asserts.PrivateKey) error {
	return errors.New("MOCK error saving the private key")
}

func getErrorMockKeyStore(config ConfigSettings) (*KeypairDatabase, error) {
	mockStore := new(errorMockKeypairManager)

	db, err := asserts.OpenDatabase(&asserts.DatabaseConfig{
		KeypairManager: mockStore,
	})
	keypairDB := KeypairDatabase{FilesystemStore, db}
	return &keypairDB, err
}
