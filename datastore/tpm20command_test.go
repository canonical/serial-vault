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
	"testing"

	"github.com/CanonicalLtd/serial-vault/service/log"
)

type mockTPM20Command struct{}

func TestRunCommand(t *testing.T) {
	command := tpm20Command{}
	err := command.runCommand("ls", "-l")
	if err != nil {
		t.Errorf("Error running shell command: %v", err)
	}
}

func TestRunCommandBadCommand(t *testing.T) {
	command := tpm20Command{}

	err := command.runCommand("thisreallyshouldnotwork", "-l")
	if err == nil {
		t.Error("Expected error, got success")
	}
}

func (tcmd *mockTPM20Command) runCommand(command string, args ...string) error {
	log.Printf("  Mock command: %s\n", command)
	if command == "tpm2_hmac" {
		filename := args[len(args)-1]
		err := ioutil.WriteFile(filename, []byte("fake-hmac-ed-data"), 0600)
		return err
	}
	return nil
}
