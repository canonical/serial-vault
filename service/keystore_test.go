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
	"os"
	"testing"
)

func TestGetKeyStoreFilesystem(t *testing.T) {
	// Set up the environment variables
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore"}
	Environ = &Env{Config: config}

	result := GetKeyStore()
	if result == nil {
		t.Error("Error setting up the filesystem keystore")
	}

}

func TestGetKeyStoreInvalid(t *testing.T) {
	// Set up the environment variables
	config := ConfigSettings{KeyStoreType: "invalid", KeyStorePath: "../keystore"}
	Environ = &Env{Config: config}

	result := GetKeyStore()
	if result != nil {
		t.Errorf("Expected nil keystore, but got a response: %v", result)
	}
}

func TestPutGetKeyStore(t *testing.T) {
	// Set up the environment variables
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore"}
	Environ = &Env{Config: config}

	data := []byte("Test Data")
	model := Model{ID: 999999}
	keystore := GetKeyStore()

	// Save the data to the keystore
	path, err := keystore.Put(data, model)
	if err != nil {
		t.Errorf("Error saving signing-key to keystore: %v", err)
	}

	// Check that we can read the file
	fetchedData, err := keystore.Get(model)
	if err != nil {
		t.Errorf("Error retrieving the signing-key from the keystore: %v", err)
	}
	if string(data) != string(fetchedData) {
		t.Errorf("Error in the fetched signing-key: %s", string(fetchedData))
	}

	// Clean up - remove the keystore file
	os.Remove(path)
}

func TestGetKeyStoreError(t *testing.T) {
	// Set up the environment variables
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore"}
	Environ = &Env{Config: config}

	model := Model{ID: 999999}
	keystore := GetKeyStore()

	// Attempt to the signing-key for an invalid model
	_, err := keystore.Get(model)
	if err == nil {
		t.Error("Expected error, got success")
	}
}

func TestPutKeyStoreError(t *testing.T) {
	// Set up the environment variables
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "INVALID PATH"}
	Environ = &Env{Config: config}

	data := []byte("Test Data")
	model := Model{ID: 999999}
	keystore := GetKeyStore()

	// Save the data to the invalid keystore path
	_, err := keystore.Put(data, model)
	if err == nil {
		t.Error("Expected error, got success", err)
	}
}
