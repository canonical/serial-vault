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

package random

import "testing"

func TestRandomGeneration(t *testing.T) {
	n := 10
	// search for random strings enough smalls as to see if they are random
	tokens := make(map[string]string)
	for i := 0; i < n; i++ {
		// generate minimum amount of random data to verify it is enough random
		token, err := GenerateRandomString(10)
		if err != nil {
			t.Errorf("Error generating random string: %v", err)
		}
		tokens[token] = token
	}

	// Check that we have n different tokens stored in the map
	if len(tokens) < n {
		t.Error("Generated random numbers are not unique")
	}
}
