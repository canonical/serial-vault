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
	"encoding/json"
	"log"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/utils"
)

// ListResponse is the JSON response from the API Signing Log method
type ListResponse struct {
	Success      bool                   `json:"success"`
	ErrorCode    string                 `json:"error_code"`
	ErrorSubcode string                 `json:"error_subcode"`
	ErrorMessage string                 `json:"message"`
	SigningLog   []datastore.SigningLog `json:"logs"`
}

// listHandler is the API method to fetch the log records from signing
func listHandler(w http.ResponseWriter, user datastore.User) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := utils.CheckUserPermissions(user, datastore.Admin)
	if err != nil {
		utils.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	logs, err := datastore.Environ.DB.ListAllowedSigningLog(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatListResponse(false, "error-fetch-signinglog", "", err.Error(), nil, w)
		return
	}

	// Return successful JSON response with the list of models
	w.WriteHeader(http.StatusOK)
	formatListResponse(true, "", "", "", logs, w)
}

func formatListResponse(success bool, errorCode, errorSubcode, message string, logs []datastore.SigningLog, w http.ResponseWriter) error {
	response := ListResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, SigningLog: logs}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the signing log response.")
		return err
	}
	return nil
}
