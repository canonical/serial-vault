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

package keypair

import (
	"encoding/json"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/auth"
	"github.com/CanonicalLtd/serial-vault/service/log"
	"github.com/CanonicalLtd/serial-vault/service/response"
)

// SyncResponse is the response to fetch keypairs
type SyncResponse struct {
	Success  bool                    `json:"success"`
	Keypairs []datastore.SyncKeypair `json:"keypairs"`
}

// syncHandler fetches the signing-keys accessible by a user
// A encryption secret is provided and the keypairs are decrypted and re-encrypted
// using the supplied keystore secret
func syncHandler(w http.ResponseWriter, user datastore.User, apiCall bool, request SyncRequest) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.SyncUser, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, "error-auth", "", "", w)
		return
	}

	if len(request.Secret) == 0 {
		if err != nil {
			response.FormatStandardResponse(false, "error-sync-keypairs", "The keystore secret cannot be empty", "", w)
			return
		}
	}

	// Get the keypairs that the user can access (does not include the sealed key)
	keypairs, err := datastore.Environ.DB.ListAllowedKeypairs(user)
	if err != nil {
		response.FormatStandardResponse(false, "error-sync-keypairs", "", err.Error(), w)
		return
	}

	syncKeypairs := []datastore.SyncKeypair{}

	for _, k := range keypairs {
		// Get the keypair with the sealed key
		keypair, err := datastore.Environ.DB.GetKeypair(k.ID)
		if err != nil {
			response.FormatStandardResponse(false, "error-sync-keypair", "", err.Error(), w)
			return
		}

		// Decrypt and re-encrypt the keypair with the supplied keystore secret
		base64SealedSigningkey, base64AuthKeyHash, err := datastore.ReEncryptKeypair(keypair, request.Secret)
		if err != nil {
			response.FormatStandardResponse(false, "error-sync-encrypt", "", err.Error(), w)
			return
		}

		// Update the sealed key - encrypted with the new keystore secret
		keypair.SealedKey = base64SealedSigningkey

		skp := datastore.SyncKeypair{Keypair: keypair, AuthKeyHash: base64AuthKeyHash}
		syncKeypairs = append(syncKeypairs, skp)
	}

	// Return successful JSON response with the list of models
	w.WriteHeader(http.StatusOK)
	formatSyncResponse(syncKeypairs, w)
}

func formatSyncResponse(keypairs []datastore.SyncKeypair, w http.ResponseWriter) error {
	response := SyncResponse{Success: true, Keypairs: keypairs}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Info("Error forming the keypair status response.")
		return err
	}
	return nil
}
