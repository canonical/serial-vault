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

package datastore

import (
	"io/ioutil"
	"log"
)

// TPM2InitializeKeystore initializes the TPM 2.0 module by taking ownership of the module
// and generating the key for sealing and unsealing signing keys. The context values provide
// the TPM 2.0 authentication and these are stored in the database.
// Main TPM 2.0 operations:
//  * takeownership
//  * createprimary
func TPM2InitializeKeystore(command TPM20Command) error {
	log.Println("Initialize the TPM Keystore...")

	// Generate a unique file name to hold the primary key context
	primaryKeyContext, err := ioutil.TempFile(Environ.Config.KeyStorePath, ".primary")
	if err != nil {
		return err
	}

	if command == nil {
		command = &tpm20Command{}
	}

	// Take ownership of the TPM 2.0 module
	err = command.runCommand("tpm2_takeownership", "-c")
	if err != nil {
		log.Printf("Error in TPM takeownership, %v", err)
		return err
	}

	// Create the primary key in the heirarchy
	err = command.runCommand("tpm2_createprimary", "-A", "o", "-g", algSHA256, "-G", algRSA, "-C", primaryKeyContext.Name())
	if err != nil {
		log.Printf("Error in TPM createprimary, %v", err)
		return err
	}

	// Save the primary key context filepath in the database
	err = Environ.DB.PutSetting(Setting{Code: "parent", Data: primaryKeyContext.Name()})
	if err != nil {
		log.Printf("Error in saving the parent key path in settings, %v", err)
		return err
	}

	return nil
}
