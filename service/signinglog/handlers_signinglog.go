// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018-2019 Canonical Ltd
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

	"github.com/CanonicalLtd/serial-vault/service/utils"
)

// ListHandler is the API method to fetch the log records from signing
func ListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := utils.GetUserFromJWT(w, r)
	if err != nil {
		utils.FormatStandardResponse(false, "error-auth", "", err.Error(), w)
		return
	}

	listHandler(w, authUser)
}
