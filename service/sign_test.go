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
	const assertions = "ABCD123456||聖誕快樂||A1234/L"

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
