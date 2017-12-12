/*
 * Copyright (C) 2017-2018 Canonical Ltd
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

package store

import (
	"encoding/base64"
	"log"
	"os/exec"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/snapcore/snapd/asserts"
)

// GenerateKeypair generates a new passwordless signing-key for signing assertions
func GenerateKeypair(authorityID, passphrase, keyName string) error {
	// Validate the keypair name - check the key name does not exist for the brand

	// Generate the keypair
	log.Println("---Generate key...")
	manager := asserts.NewGPGKeypairManager()
	err := manager.Generate(passphrase, keyName)
	if err != nil {
		log.Println("Error fetching the generated key", err)
		return err
	}

	// Export the ascii-armored GPG key
	log.Println("---Export key...")
	out, err := exec.Command("gpg", "--homedir", "~/.snap/gnupg", "--armor", "--export-secret-key", keyName).Output()
	if err != nil {
		log.Println("Error exporting the generated key", err)
		return err
	}
	base64PrivateKey := base64.StdEncoding.EncodeToString(out)
	log.Println(string(out))

	// Store the signing-key in the keypair store using the asserts module
	log.Println("---Seal key...")
	privateKey, sealedPrivateKey, err := datastore.Environ.KeypairDB.ImportSigningKey(authorityID, base64PrivateKey)
	if err != nil {
		log.Printf("Error storing the private key: %v", err)
		return err
	}

	// Store the sealed signing-key in the database
	log.Println("---Store key...")
	keypair := datastore.Keypair{
		AuthorityID: authorityID,
		KeyID:       privateKey.PublicKey().ID(),
		SealedKey:   sealedPrivateKey,
	}
	_, err = datastore.Environ.DB.PutKeypair(keypair)
	if err != nil {
		log.Printf("Error storing the private key: %v", err)
		return err
	}

	// Delete the key from the local store
	log.Println("---Delete key...")
	err = manager.Delete(keyName)
	if err != nil {
		log.Printf("Error removing temporary key: %v", err)
	}
	return err
}
