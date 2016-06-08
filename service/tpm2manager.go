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
	"io/ioutil"
	"log"
	"os/exec"
)

// TPM2InitializeKeystore initializes the TPM 2.0 module by taking ownership of the module
// and generating the key for sealing and unsealing signing keys. The context values provide
// the TPM 2.0 authentication and these are stored in the database.
func TPM2InitializeKeystore(env Env) error {
	log.Println("Initialize the TPM Keystore...")

	// Generate a unique file name to hold the primary key context
	primaryKeyContext, err := ioutil.TempFile("keystore", ".primary")
	if err != nil {
		return err
	}

	// Take ownership of the TPM 2.0 module
	cmd := exec.Command("tpm2_takeownership", "-c")
	_, err = cmd.Output()
	if err != nil {
		log.Printf("Error in TPM takeownership, %v", err)
		return err
	}

	// Create the primary key in the heirarchy
	cmd = exec.Command("tpm2_createprimary", "-A", "o", "-g", algSHA256, "-G", algRSA, "-C", primaryKeyContext.Name())
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Error in TPM createprimary, %v", err)
		log.Println(string(out[:]))
		return err
	}

	// Save the primary key context filepath in the database
	err = env.DB.PutSetting(Setting{Code: "parent", Data: primaryKeyContext.Name()})
	if err != nil {
		log.Printf("Error in saving the parent key path in settings, %v", err)
		return err
	}

	return nil
}
