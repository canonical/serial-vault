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

// TestList checks the list method using the test authorized_keys file.
func TestList(t *testing.T) {
	auth := &AuthorizedKeys{path: "../test/authorized_keys"}
	keys := auth.List()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got: %d", len(keys))
	}
}

func TestListInvalidFile(t *testing.T) {
	auth := &AuthorizedKeys{path: "does not exist"}
	keys := auth.List()
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got: %d", len(keys))
	}
}

func addSSHKeys(t *testing.T, auth *AuthorizedKeys) []string {
	// Add a test ssh key (should create the file)
	err := auth.Add("test ssh key")
	if err != nil {
		t.Errorf("Expected success, got: %v", err)
	}

	err = auth.Add("another ssh key")
	if err != nil {
		t.Errorf("Expected success, got: %v", err)
	}

	err = auth.Add("yet another ssh key")
	if err != nil {
		t.Errorf("Expected success, got: %v", err)
	}

	keys := auth.List()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got: %d", len(keys))
	}

	return keys
}

// TestAdd adds an ssh public key to the authorized_keys file.
func TestAdd(t *testing.T) {
	auth := &AuthorizedKeys{path: "../test/authorized_keys_new"}

	// Add some a ssh keys and check they were added successfully
	addSSHKeys(t, auth)

	// Delete the created file
	os.Remove(auth.path)
}

func TestAddInvalidKey(t *testing.T) {
	auth := &AuthorizedKeys{path: "../test/authorized_keys_new"}

	err := auth.Add("     ")
	if err == nil {
		t.Error("Expected failure, got success.")
	}

	// Delete the created file
	os.Remove(auth.path)
}

func TestAddDuplicateKey(t *testing.T) {
	auth := &AuthorizedKeys{path: "../test/authorized_keys_new"}

	// Add some a ssh keys and check they were added successfully
	addSSHKeys(t, auth)

	// Add a duplicate key
	err := auth.Add("  yet another ssh key  ")
	if err == nil {
		t.Error("Expected failure, got success.")
	}

	// Delete the created file
	os.Remove(auth.path)
}

// TestDelete adds some ssh public keys to the authorized_keys file and
// then deletes them.
func TestDelete(t *testing.T) {
	auth := &AuthorizedKeys{path: "../test/authorized_keys_new"}

	// Add some a ssh keys and check they were added successfully
	keys := addSSHKeys(t, auth)

	// Delete one of the keys
	err := auth.Delete(keys[1])
	if err != nil {
		t.Errorf("Expected success, got: %v", err)
	}

	keysAfterDelete := auth.List()
	if len(keysAfterDelete) != len(keys)-1 {
		t.Errorf("Expected %d keys, got: %d", len(keys)-1, len(keysAfterDelete))
	}

	// Delete the created file
	os.Remove(auth.path)
}

func TestDeleteKeyNotFound(t *testing.T) {
	auth := &AuthorizedKeys{path: "../test/authorized_keys_new"}

	// Add some a ssh keys and check they were added successfully
	keys := addSSHKeys(t, auth)

	// Delete one of the keys
	err := auth.Delete("key that does not exist")
	if err == nil {
		t.Error("Expected error, got success.")
	}

	keysAfterDelete := auth.List()
	if len(keysAfterDelete) != len(keys) {
		t.Errorf("Expected %d keys, got: %d", len(keys), len(keysAfterDelete))
	}

	// Delete the created file
	os.Remove(auth.path)
}

func TestDeleteFileNotFound(t *testing.T) {
	auth := &AuthorizedKeys{path: "../test/authorized_keys_new"}

	// Delete one of the keys
	err := auth.Delete("key that does not exist")
	if err == nil {
		t.Error("Expected error, got success.")
	}

	keysAfterDelete := auth.List()
	if len(keysAfterDelete) != 0 {
		t.Errorf("Expected 0 keys, got: %d", len(keysAfterDelete))
	}
}
