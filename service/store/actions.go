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

package store

import (
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/auth"
	"github.com/CanonicalLtd/serial-vault/service/log"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/CanonicalLtd/serial-vault/store"
)

func keyRegisterHandler(w http.ResponseWriter, user datastore.User, apiCall bool, keyAuth store.KeyRegister) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := auth.CheckUserPermissions(user, datastore.Admin, apiCall)
	if err != nil {
		response.FormatStandardResponse(false, response.ErrorAuth.Code, "", err.Error(), w)
		return
	}

	// Check that the user has permissions to this authority-id
	if !datastore.Environ.DB.CheckUserInAccount(user.Username, keyAuth.AuthorityID) {
		response.FormatStandardResponse(false, response.ErrorAuth.Code, "", response.ErrorAuth.Message, w)
		return
	}

	keypair, err := datastore.Environ.DB.GetKeypairByName(keyAuth.AuthorityID, keyAuth.KeyName)
	if err != nil {
		log.Message("KEYPAIR", response.ErrorFetchKeypair.Code, err.Error())
		response.FormatStandardResponse(false, response.ErrorFetchKeypair.Code, "", "Cannot find the signing key", w)
		return
	}

	// Register the account key with the store
	err = store.RegisterKey(keyAuth, keypair)
	if err != nil {
		log.Message("KEYPAIR", response.ErrorStoreKeypair.Code, err.Error())
		response.FormatStandardResponse(false, response.ErrorStoreKeypair.Code, "", err.Error(), w)
		return
	}

	// Return successful JSON response
	w.WriteHeader(http.StatusOK)
	response.FormatStandardResponse(true, "", "", "", w)
}
