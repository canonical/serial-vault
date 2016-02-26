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
	"bufio"
	"errors"
	"log"
	"os"
	"os/user"
	"strings"
	"sync"
)

// AuthorizedKeystore interface to manage the authorized ssh keys
type AuthorizedKeystore interface {
	List() []string
	Add(string) error
	Delete(string) error
}

// AuthorizedKeys interface to manage the authorized ssh keys
type AuthorizedKeys struct {
	path string
	mu   sync.RWMutex
}

// InitializeAuthorizedKeys initializes the path to the authorized keys file.
func InitializeAuthorizedKeys(sshKeysPath string) (AuthorizedKeystore, error) {
	// Get the user's home directory
	usr, err := user.Current()
	if err != nil {
		return &AuthorizedKeys{}, err
	}

	return &AuthorizedKeys{path: usr.HomeDir + sshKeysPath}, nil
}

// List retrieves the authorized ssh keys
func (auth *AuthorizedKeys) List() []string {
	var keys []string

	// Turn on the read lock
	auth.mu.RLock()
	defer auth.mu.RUnlock()

	// Read the authorized_keys file
	file, err := os.Open(auth.path)
	if err != nil {
		return keys
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		keys = append(keys, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading authorized_keys: %v", err)
	}

	return keys
}

func (auth *AuthorizedKeys) findKey(key string) (int, []string) {
	authKeyPosition := -1

	// Get the list of authorized keys
	keys := auth.List()
	if len(keys) == 0 {
		return authKeyPosition, keys
	}

	// Get the index of the key
	for index, authKey := range keys {
		if authKey == key {
			authKeyPosition = index
			break
		}
	}
	return authKeyPosition, keys
}

// Add saves a new authorized ssh key
func (auth *AuthorizedKeys) Add(key string) error {
	// Check the public key
	key = strings.Trim(key, " ")
	if len(key) == 0 {
		return errors.New("The public key must be entered.")
	}

	// Check if the key already exists
	authKeyPosition, _ := auth.findKey(key)
	if authKeyPosition >= 0 {
		return errors.New("The ssh public key already exists.")
	}

	// Turn on the write lock
	auth.mu.Lock()
	defer auth.mu.Unlock()

	// Open the authorized_keys file for writing, create it if it does not exist.
	file, err := os.OpenFile(auth.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(key + "\n")
	return err
}

// Delete removes an authorized ssh key
func (auth *AuthorizedKeys) Delete(key string) error {

	// Get the index of the key
	authKeyPosition, keys := auth.findKey(key)
	if authKeyPosition < 0 {
		return errors.New("The ssh public key cannot be found.")
	}

	// Remove the unwanted ssh key from the list
	keys = append(keys[:authKeyPosition], keys[authKeyPosition+1:]...)

	// Turn on the write lock
	auth.mu.Lock()
	defer auth.mu.Unlock()

	// Overwrite the contents of the authorized keys file
	file, err := os.Create(auth.path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Rewrite the contents of the file
	for _, deviceKey := range keys {
		_, err = file.WriteString(deviceKey + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}
