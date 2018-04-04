// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
 * License granted by Canonical Limited
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

package main

import (
	"fmt"
	"os"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/sync"
	"github.com/jessevdk/go-flags"
)

func main() {
	datastore.Environ = &datastore.Env{}

	err := run()
	if err != nil {
		os.Exit(1)
	}

}

func run() error {
	// Parse the command line arguments and execute the command
	parser := flags.NewParser(&sync.Sync, flags.HelpFlag)
	_, err := parser.Parse()

	if err != nil {
		if e, ok := err.(*flags.Error); ok {
			if e.Type == flags.ErrHelp || e.Type == flags.ErrCommandRequired {
				parser.WriteHelp(os.Stdout)
				return nil
			}
		}
		fmt.Println(err)
	}

	return err
}
