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

package request

import (
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
)

// CheckUserAPI validates the user and API key
func CheckUserAPI(r *http.Request) (datastore.User, error) {
	// Get the user and API key from the header
	username := r.Header.Get("user")
	apiKey := r.Header.Get("api-key")

	// Find the user by API key
	return datastore.Environ.DB.GetUserByAPIKey(apiKey, username)
}
