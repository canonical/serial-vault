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

package service

import "net/http"

// AccountsResponse is the JSON response from the API Accounts method
type AccountsResponse struct {
	Success      bool      `json:"success"`
	ErrorCode    string    `json:"error_code"`
	ErrorSubcode string    `json:"error_subcode"`
	ErrorMessage string    `json:"message"`
	Accounts     []Account `json:"accounts"`
}

// AccountsHandler is the API method to list the account assertions
func AccountsHandler(w http.ResponseWriter, r *http.Request) {

	accounts, err := Environ.DB.ListAccounts()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		formatAccountsResponse(false, "error-accounts-json", "", err.Error(), nil, w)
		return
	}

	// Format the model for output and return JSON response
	w.WriteHeader(http.StatusOK)
	formatAccountsResponse(true, "", "", "", accounts, w)
}
