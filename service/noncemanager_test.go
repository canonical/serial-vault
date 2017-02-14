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

func TestNonceGeneration(t *testing.T) {
	// Generate some nonces
	var nonces [10]DeviceNonce
	for i := 0; i < 10; i++ {
		nonce, err := generateNonce()
		if err != nil {
			t.Error("Error generating nonce")
		}
		nonces[i] = nonce
	}

	// Check that the nonces look unique and valid
	for i := 1; i < 10; i++ {
		thisOne := nonces[i]
		lastOne := nonces[i-1]

		if thisOne.Nonce == lastOne.Nonce {
			t.Error("Generated nonces are not unique")
		}
		if lastOne.TimeStamp > thisOne.TimeStamp {
			t.Error("Nonce based on invalid timestamp")
		}
	}
}
