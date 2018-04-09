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

package signinglog

import (
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/auth"
	"github.com/CanonicalLtd/serial-vault/service/response"
)

func syncLogHandler(w http.ResponseWriter, user datastore.User, apiCall bool, signLog datastore.SigningLog) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.SyncUser, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	// Create the the signing-log if it does not exist
	exists, err := datastore.Environ.DB.CheckForMatching(signLog)
	if err != nil {
		response.FormatStandardResponse(false, "error-signinglog-match", "", err.Error(), w)
		return
	}

	if !exists {
		// The signing log has not been sync-ed, so create it (keep the same create timestamp)
		err = datastore.Environ.DB.CreateSigningLogSync(signLog)
		if err != nil {
			response.FormatStandardResponse(false, "error-signinglog-create", "", err.Error(), w)
			return
		}
	}

	// Return successful JSON response
	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", "", w)
}
