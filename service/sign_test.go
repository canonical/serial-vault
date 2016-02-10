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

func TestClearSignFile(t *testing.T) {
	const assertions = `
  {
	  "brand-id": "System",
    "model":"聖誕快樂",
    "serial":"A1234/L",
		"revision": 2,
    "device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
  }`

	// Get the test private key
	key, err := getPrivateKey(TestPrivateKeyPath)
	if err != nil {
		t.Errorf("Error reading the private key file: %v", err)
	}

	response, err := ClearSign(assertions, string(key), "")
	if err != nil {
		t.Errorf("Error signing the assertions text: %v", err)
	}
	if len(response) == 0 {
		t.Errorf("Empty signed data returned.")
	}
}

func TestClearSignInvalidFile(t *testing.T) {
	const assertions = `
  {
	  "brand-id": "System",
    "model":"聖誕快樂",
    "serial":"A1234/L",
		"revision": 2,
    "device-key":"ssh-rsa NNhqloxPyIYXiTP+3JTPWV/mNoBar2geWIf"
  }`

	// Get an invalid private key file
	_, err := getPrivateKey("../README.md")
	if err != nil {
		t.Error("Expected an error using an invalid private key file.")
	}
}
