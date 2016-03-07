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
	"fmt"
	"io/ioutil"
	"log"
)

// KeyStore interface to save and retrieve a signing-key
type KeyStore interface {
	Put(data []byte, model Model) (string, error)
	Get(model Model) ([]byte, error)
}

// GetKeyStore returns the keystore as defined in the config file
func GetKeyStore() KeyStore {
	switch {
	case Environ.Config.KeyStoreType == "filesystem":
		return FilesystemKeyStore{Path: Environ.Config.KeyStorePath}
	}
	return nil
}

// FilesystemKeyStore that stores signing-keys in the filesystem
type FilesystemKeyStore struct {
	Path string
}

func (fs FilesystemKeyStore) fullPath(model Model) string {
	// Format the path name for the keystore file
	return fmt.Sprintf("%s/model%d", fs.Path, model.ID)
}

// Put creates a new signing-key in the keystore for a specific model/revision
func (fs FilesystemKeyStore) Put(data []byte, model Model) (string, error) {
	fullPath := fs.fullPath(model)
	err := ioutil.WriteFile(fullPath, data, 0600)
	if err != nil {
		log.Printf("Error saving key to keystore: %s", fullPath)
	}

	return fullPath, err
}

// Get fetches the signing-key for a specific model/revision from the keystore
func (fs FilesystemKeyStore) Get(model Model) ([]byte, error) {
	fullPath := fs.fullPath(model)
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		log.Printf("Error retrieving key to keystore: %s", fullPath)
	}
	return data, err
}
