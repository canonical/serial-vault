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

import "testing"

func TestTPM2InitializeKeystore(t *testing.T) {
	// Set up the environment variables
	config := ConfigSettings{KeyStorePath: "../keystore", KeyStoreType: "tpm2.0", KeyStoreSecret: "this needs to be 32 bytes long!!"}
	env := Env{Config: config, DB: &MockDB{}}

	err := TPM2InitializeKeystore(env, &mockTPM20Command{})
	if err != nil {
		t.Errorf("Error initializing the TPM keystore: %v", err)
	}
}
