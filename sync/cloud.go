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

package sync

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/account"
	"github.com/CanonicalLtd/serial-vault/service/keypair"
	"github.com/CanonicalLtd/serial-vault/service/log"
	"github.com/CanonicalLtd/serial-vault/service/model"
	"github.com/CanonicalLtd/serial-vault/service/response"
)

var hclient http.Client

// SendRequest sends the request to the serial vault
var SendRequest = func(method, url, endpoint, username, apikey string, data []byte) (*http.Response, error) {
	log.Infof("Call the cloud %s", url+endpoint)
	r, _ := http.NewRequest(method, url+endpoint, bytes.NewReader(data))
	r.Header.Set("user", username)
	r.Header.Set("api-key", apikey)

	return hclient.Do(r)
}

// FetchAccounts fetches the accounts from the cloud serial vault
var FetchAccounts = func(url, username, apikey string) (account.ListResponse, error) {
	w, err := SendRequest("GET", url, "accounts", username, apikey, nil)
	if err != nil {
		log.Errorf("Error fetching accounts: %v", err)
		return account.ListResponse{}, err
	}

	// Parse the response from the accounts
	return parseAccountResponse(w)
}

// FetchSigningKeys fetches the signing-keys from the cloud serial vault
// Send our keystore secret to the cloud and get back the keys encrypted using our secret
var FetchSigningKeys = func(url, username, apikey string, data []byte) (keypair.SyncResponse, error) {
	w, err := SendRequest("POST", url, "keypairs/sync", username, apikey, data)
	if err != nil {
		log.Errorf("Error fetching accounts: %v", err)
		return keypair.SyncResponse{}, err
	}

	// Parse the response from the signing-key request
	return parseSigningKeyResponse(w)
}

// FetchModels fetches the models from the cloud serial vault
var FetchModels = func(url, username, apikey string) (model.ListResponse, error) {
	w, err := SendRequest("GET", url, "models", username, apikey, nil)
	if err != nil {
		log.Errorf("Error fetching models: %v", err)
		return model.ListResponse{}, err
	}

	// Parse the response from the cloud
	return parseModelResponse(w)
}

// SendSigningLog sends a signing log to the cloud serial vault
var SendSigningLog = func(url, username, apikey string, signLog datastore.SigningLog) (bool, error) {

	data, err := json.Marshal(signLog)
	if err != nil {
		log.Errorf("Error marshalling signing log: %v", err)
		return false, err
	}

	w, err := SendRequest("POST", url, "signinglog", username, apikey, data)
	if err != nil {
		log.Errorf("Error syncing signing log: %v", err)
		return false, err
	}

	// Parse the response from the cloud
	result, err := parseStandardResponse(w)
	if err != nil {
		log.Errorf("Error parsing signing log: %v", err)
		return false, err
	}
	if !result.Success {
		log.Errorf("Error syncing signing log: %v", result.ErrorMessage)
		return false, err
	}

	return result.Success, nil
}

func parseAccountResponse(w *http.Response) (account.ListResponse, error) {
	// Check the JSON response
	result := account.ListResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func parseSigningKeyResponse(w *http.Response) (keypair.SyncResponse, error) {
	// Check the JSON response
	result := keypair.SyncResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func parseModelResponse(w *http.Response) (model.ListResponse, error) {
	// Check the JSON response
	result := model.ListResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}

func parseStandardResponse(w *http.Response) (response.StandardResponse, error) {
	// Check the JSON response
	result := response.StandardResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}
