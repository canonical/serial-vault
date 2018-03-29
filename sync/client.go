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
	"errors"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/account"
	"github.com/CanonicalLtd/serial-vault/service/log"
)

// Client is the sync interface for the serial vault
type Client interface {
	Accounts() error
}

// FactoryClient is the implementation of the factory sync for the serial vault
type FactoryClient struct {
	URL      string
	Username string
	APIKey   string
}

// NewFactoryClient creates a factory client to sync data with the cloud serial-vault
func NewFactoryClient(url, username, apiKey string) *FactoryClient {
	return &FactoryClient{
		URL: url, Username: username, APIKey: apiKey,
	}
}

// SendRequest sends the request to the serial vault
var SendRequest = func(method, url, endpoint, username, apikey string, data []byte) (*http.Response, error) {
	r, _ := http.NewRequest(method, url, bytes.NewReader(data))
	r.Header.Set("user", username)
	r.Header.Set("api-key", apikey)

	client := http.Client{}
	return client.Do(r)
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

// Accounts synchronizes the account details to the factory instance
func (c *FactoryClient) Accounts() error {
	// Fetch the accounts from the serial-vault
	result, err := FetchAccounts(c.URL, c.Username, c.APIKey)
	if err != nil {
		log.Errorf("Error parsing accounts: %v", err)
		return err
	}
	if !result.Success {
		log.Errorf("Error fetching accounts: %s", result.ErrorMessage)
		return errors.New(result.ErrorMessage)
	}

	// Update the factory database with the accounts
	for _, a := range result.Accounts {
		_, err = datastore.Environ.DB.SyncAccount(a)
		if err != nil {
			log.Errorf("Error updating accounts: %v", err)
			return err
		}
	}

	return nil
}

func parseAccountResponse(w *http.Response) (account.ListResponse, error) {
	// Check the JSON response
	result := account.ListResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}
