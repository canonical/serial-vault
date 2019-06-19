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
	"os/exec"

	"github.com/CanonicalLtd/serial-vault/service/log"
)

// TPM20Command is an interface for wrapping the TPM2.0 shell commands
type TPM20Command interface {
	runCommand(command string, args ...string) error
}

type tpm20Command struct{}

func (tcmd *tpm20Command) runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Error in TPM %s, %v", command, err)
		log.Println(string(out[:]))
		return err
	}

	return nil
}
