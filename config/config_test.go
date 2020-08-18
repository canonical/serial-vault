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

package config

import "testing"

func TestReadConfig(t *testing.T) {
	settings := Settings{}
	err := ReadConfig(&settings, "../settings.yaml.example")
	if err != nil {
		t.Errorf("Error reading config file: %v", err)
	}
}

func TestReadConfigInvalidPath(t *testing.T) {
	settings := Settings{}
	err := ReadConfig(&settings, "not a good path")
	if err == nil {
		t.Error("Expected an error with an invalid config file.")
	}
}

func TestReadConfigInvalidFile(t *testing.T) {
	settings := Settings{}
	err := ReadConfig(&settings, "../README.md")
	if err == nil {
		t.Error("Expected an error with an invalid config file.")
	}
}
