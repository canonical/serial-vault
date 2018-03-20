// -*- Mode: Go; indent-tabs-mode: t -*-

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

package manage

import (
	"os"
	"testing"

	"github.com/jessevdk/go-flags"

	"gopkg.in/check.v1"
)

// Hook up check.v1 into the "go test" runner
func Test(t *testing.T) { check.TestingT(t) }

type manTest struct {
	Args         []string
	ErrorMessage string
}

func mockArgs(args ...string) (restore func()) {
	old := os.Args
	os.Args = args
	return func() { os.Args = old }
}

func RunMain() error {
	// Parse the command line arguments and execute the command
	parser := flags.NewParser(&Manage, flags.HelpFlag)
	_, err := parser.Parse()
	return err
}

func runTest(c *check.C, args []string, errorMessage string) {
	restore := mockArgs(args...)
	defer restore()

	err := RunMain()

	if len(errorMessage) == 0 {
		c.Check(err, check.IsNil)
	} else {
		c.Assert(err, check.ErrorMatches, errorMessage)
	}
}
