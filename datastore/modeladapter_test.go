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
	"testing"
)

func TestModelName(t *testing.T) {
	name := "my-model-01"
	err := validateModelName(name)
	if err != nil {
		t.Errorf("Not a valid model name: %v", err)
	}
}

func TestModelNameEmpty(t *testing.T) {
	name := ""
	err := validateModelName(name)
	if err == nil {
		t.Error("Expected name not to be valid, but it is")
	}
	if err.Error() != "Model name must not be empty" {
		t.Error("Error happening is not the one searched for")
	}
}

func TestModelNameUpperChar(t *testing.T) {
	name := "my-Model-01"
	err := validateModelName(name)
	if err == nil {
		t.Error("Expected name not to be valid, but it is")
	}
	if err.Error() != "Model name must not contain uppercase characters" {
		t.Error("Error happening is not the one searched for")
	}
}
