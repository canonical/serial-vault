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

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccountsHandler(t *testing.T) {

	// Mock the database
	Environ = &Env{DB: &MockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/v1/accounts", nil)
	http.HandlerFunc(AccountsHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := AccountsResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the accounts response: %v", err)
	}
	if len(result.Accounts) != 1 {
		t.Errorf("Expected 1 accounts, got %d", len(result.Accounts))
	}
	// if result.Models[0].Name != "alder" {
	// 	t.Errorf("Expected model name 'alder', got %s", result.Models[0].Name)
	// }
}

func TestAccountsHandlerError(t *testing.T) {

	// Mock the database
	Environ = &Env{DB: &ErrorMockDB{}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/v1/accounts", nil)
	http.HandlerFunc(AccountsHandler).ServeHTTP(w, r)

	// Check the JSON response
	result := AccountsResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the accounts response: %v", err)
	}
	if result.Success {
		t.Error("Expected error, got success response")
	}
}
