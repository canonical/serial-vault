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

package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/gorilla/mux"
)

// SubstoresResponse is the JSON response from the API Sub-Stores method
type SubstoresResponse struct {
	Success      bool                 `json:"success"`
	ErrorCode    string               `json:"error_code"`
	ErrorSubcode string               `json:"error_subcode"`
	ErrorMessage string               `json:"message"`
	Substores    []datastore.Substore `json:"substores"`
}

// SubstoreDeleteHandler is the API method to delete a sub-store model.
func SubstoreDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	authUser, err := checkIsAdminAndGetUserFromJWT(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-auth", "", "", w)
		return
	}

	// Get the sub-store primary key
	vars := mux.Vars(r)
	storeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", vars["id"])
		formatBooleanResponse(false, "error-invalid-store", "", errorMessage, w)
		return
	}

	// Update the database
	errorSubcode, err := datastore.Environ.DB.DeleteAllowedSubstore(storeID, authUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatBooleanResponse(false, "error-deleting-store", errorSubcode, err.Error(), w)
		return
	}

	w.WriteHeader(http.StatusOK)
	formatBooleanResponse(true, "", "", "", w)
}
