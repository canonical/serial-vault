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
	// Create a new keypair status record to track progress
	ks := datastore.KeypairStatus{AuthorityID: authorityID, KeyName: keyName}

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

	err = updateKeyID(&ks, publicID)
	if err != nil {
		return err
	}

	// Delete the key from the local store
	log.Println("---Delete key...")
	manager := asserts.NewGPGKeypairManager()
	err = manager.Delete(keyName)
	if err != nil {
		log.Printf("Error removing temporary key: %v", err)
	}
	return err
}

func generateKeypair(ks *datastore.KeypairStatus, passphrase string) (string, error) {
	_, err := datastore.Environ.DB.CreateKeypairStatus(*ks)
	if err != nil {
		return "", err
	}

	// Generate the keypair
	log.Println("---Generate key...")
	manager := asserts.NewGPGKeypairManager()
	err = manager.Generate(passphrase, ks.KeyName)
	if err != nil {
		log.Println("Error fetching the generated key", err)
		return "", err
	}

	// Export the ascii-armored GPG key
	log.Println("---Export key...")
	ks.Status = datastore.KeypairStatusExporting
	if err = datastore.Environ.DB.UpdateKeypairStatus(*ks); err != nil {
		return "", err
	}
	out, err := exec.Command("gpg", "--homedir", "~/.snap/gnupg", "--armor", "--export-secret-key", ks.KeyName).Output()
	if err != nil {
		log.Println("Error exporting the generated key", err)
		return "", err
	}

	return base64.StdEncoding.EncodeToString(out), nil
}

func importPrivateKey(ks *datastore.KeypairStatus, base64PrivateKey string) (string, string, error) {

	// Store the signing-key in the keypair store using the asserts module
	log.Println("---Seal key...")
	ks.Status = datastore.KeypairStatusEncrypting
	if err := datastore.Environ.DB.UpdateKeypairStatus(*ks); err != nil {
		return "", "", err
	}
	privateKey, sealedPrivateKey, err := datastore.Environ.KeypairDB.ImportSigningKey(ks.AuthorityID, base64PrivateKey)
	if err != nil {
		log.Printf("Error storing the private key: %v", err)
		return "", "", err
	}

	return privateKey.PublicKey().ID(), sealedPrivateKey, nil
}

func storePrivateKey(ks *datastore.KeypairStatus, publicID, sealedPrivateKey string) error {
	// Store the sealed signing-key in the database
	log.Println("---Store key...")
	ks.Status = datastore.KeypairStatusStoring
	if err := datastore.Environ.DB.UpdateKeypairStatus(*ks); err != nil {
		return err
	}
	keypair := datastore.Keypair{
		AuthorityID: ks.AuthorityID,
		KeyID:       publicID,
		SealedKey:   sealedPrivateKey,
	}
	_, err := datastore.Environ.DB.PutKeypair(keypair)
	if err != nil {
		log.Printf("Error storing the private key: %v", err)
		return err
	}

	return err
}

func updateKeyID(ks *datastore.KeypairStatus, keyID string) error {
	kp, err := datastore.Environ.DB.GetKeypairByPublicID(ks.AuthorityID, keyID)
	if err != nil {
		log.Printf("Error fetching the private key: %v", err)
		return err
	}

	// Update the status and link to the generated keypair record
	ks.Status = datastore.KeypairStatusComplete
	ks.KeypairID = kp.ID
	err = datastore.Environ.DB.UpdateKeypairStatus(*ks)
	return err
}
