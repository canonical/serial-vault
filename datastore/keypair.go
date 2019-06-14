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

package datastore

import (
	"encoding/base64"
	"os/exec"

	"github.com/CanonicalLtd/serial-vault/service/log"

	"github.com/snapcore/snapd/asserts"
)

// GenerateKeypair generates a new passwordless signing-key for signing assertions
func GenerateKeypair(authorityID, passphrase, keyName string) error {
	// Create a new keypair status record to track progress
	ks := KeypairStatus{AuthorityID: authorityID, KeyName: keyName}

	base64PrivateKey, err := generateKeypair(&ks, passphrase)
	if err != nil {
		return err
	}

	publicID, sealedPrivateKey, err := importPrivateKey(&ks, base64PrivateKey)
	if err != nil {
		return err
	}

	err = storePrivateKey(&ks, publicID, sealedPrivateKey)
	if err != nil {
		return err
	}

	err = Environ.DB.DeleteKeypairStatus(ks)
	if err != nil {
		return err
	}

	// Delete the key from the local store
	manager := asserts.NewGPGKeypairManager()
	err = manager.Delete(keyName)
	if err != nil {
		log.Printf("Error removing temporary key: %v", err)
	}
	return err
}

func generateKeypair(ks *KeypairStatus, passphrase string) (string, error) {
	id, err := Environ.DB.CreateKeypairStatus(*ks)
	if err != nil {
		return "", err
	}
	ks.ID = id

	// Generate the keypair
	manager := asserts.NewGPGKeypairManager()
	err = manager.Generate(passphrase, ks.KeyName)
	if err != nil {
		log.Println("Error fetching the generated key", err)
		return "", err
	}

	// Export the ascii-armored GPG key
	ks.Status = KeypairStatusExporting
	if err = Environ.DB.UpdateKeypairStatus(*ks); err != nil {
		return "", err
	}
	out, err := exec.Command("gpg", "--homedir", "~/.snap/gnupg", "--armor", "--export-secret-key", ks.KeyName).Output()
	if err != nil {
		log.Println("Error exporting the generated key", err)
		return "", err
	}

	return base64.StdEncoding.EncodeToString(out), nil
}

func importPrivateKey(ks *KeypairStatus, base64PrivateKey string) (string, string, error) {

	// Store the signing-key in the keypair store using the asserts module
	ks.Status = KeypairStatusEncrypting
	if err := Environ.DB.UpdateKeypairStatus(*ks); err != nil {
		return "", "", err
	}
	privateKey, sealedPrivateKey, err := Environ.KeypairDB.ImportSigningKey(ks.AuthorityID, base64PrivateKey)
	if err != nil {
		log.Printf("Error storing the private key: %v", err)
		return "", "", err
	}

	return privateKey.PublicKey().ID(), sealedPrivateKey, nil
}

func storePrivateKey(ks *KeypairStatus, publicID, sealedPrivateKey string) error {
	// Store the sealed signing-key in the database
	ks.Status = KeypairStatusStoring
	if err := Environ.DB.UpdateKeypairStatus(*ks); err != nil {
		return err
	}
	keypair := Keypair{
		AuthorityID: ks.AuthorityID,
		KeyID:       publicID,
		SealedKey:   sealedPrivateKey,
		KeyName:     ks.KeyName,
	}
	_, err := Environ.DB.PutKeypair(keypair)
	if err != nil {
		log.Printf("Error storing the private key: %v", err)
		return err
	}

	return err
}

// CreateKeyName assigns a key name to an existing key
func CreateKeyName(k Keypair) error {
	kp, err := Environ.DB.GetKeypairByPublicID(k.AuthorityID, k.KeyID)
	if err != nil {
		log.Printf("Error fetching the private key: %v", err)
		return err
	}

	ks := KeypairStatus{
		AuthorityID: k.AuthorityID, KeyName: k.KeyName, KeypairID: kp.ID, Status: KeypairStatusComplete,
	}
	if ks.KeyName == "" {
		ks.KeyName = k.AuthorityID
	}
	statusID, err := Environ.DB.CreateKeypairStatus(ks)
	if err != nil {
		return err
	}
	ks.ID = statusID

	// Update the status and link to the generated keypair record
	return Environ.DB.UpdateKeypairStatus(ks)
}
