// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2017 Canonical Ltd
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
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/usso"
	"github.com/snapcore/snapd/asserts"
)

// DeviceAssertion defines the device identity.
type DeviceAssertion struct {
	Type         string `yaml:"type"`
	Brand        string `yaml:"brand-id"`
	Model        string `yaml:"model"`
	SerialNumber string `yaml:"serial"`
	Timestamp    string `yaml:"timestamp"`
	Revision     int    `yaml:"revision"`
	PublicKey    string `yaml:"device-key"`
}

// ModelType is the default type of a model
const ModelType = "device"

// BooleanResponse is the JSON response from an API method, indicating success or failure.
type BooleanResponse struct {
	Success      bool   `json:"success"`
	ErrorCode    string `json:"error_code"`
	ErrorSubcode string `json:"error_subcode"`
	ErrorMessage string `json:"message"`
}

func formatSignResponse(success bool, errorCode, errorSubcode, message string, assertion asserts.Assertion, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", asserts.MediaType)
	w.WriteHeader(http.StatusOK)
	encoder := asserts.NewEncoder(w)
	err := encoder.Encode(assertion)
	if err != nil {
		// Not much we can do if we're here - apart from panic!
		log.Println("Error encoding the assertion.")
		return err
	}

	return nil
}

func formatModelsResponse(success bool, errorCode, errorSubcode, message string, models []ModelSerialize, w http.ResponseWriter) error {
	response := ModelsResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, Models: models}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the models response.")
		return err
	}
	return nil
}

func formatBooleanResponse(success bool, errorCode, errorSubcode, message string, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	response := BooleanResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the boolean response.")
		return err
	}
	return nil
}

func formatModelResponse(success bool, errorCode, errorSubcode, message string, model ModelSerialize, w http.ResponseWriter) error {
	response := ModelResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, Model: model}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the model response.")
		return err
	}
	return nil
}

func formatKeypairsResponse(success bool, errorCode, errorSubcode, message string, keypairs []datastore.Keypair, w http.ResponseWriter) error {
	response := KeypairsResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, Keypairs: keypairs}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the keypairs response.")
		return err
	}
	return nil
}

func formatAccountsResponse(success bool, errorCode, errorSubcode, message string, accounts []datastore.Account, w http.ResponseWriter) error {
	response := AccountsResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, Accounts: accounts}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the accounts response.")
		return err
	}
	return nil
}

func formatSigningLogResponse(success bool, errorCode, errorSubcode, message string, logs []datastore.SigningLog, w http.ResponseWriter) error {
	response := SigningLogResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, SigningLog: logs}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the signing log response.")
		return err
	}
	return nil
}

func formatSigningLogFiltersResponse(success bool, errorCode, errorSubcode, message string, filters datastore.SigningLogFilters, w http.ResponseWriter) error {
	response := SigningLogFiltersResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, SigningLogFilters: filters}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the signing log response.")
		return err
	}
	return nil
}

func formatRequestIDResponse(success bool, message string, nonce datastore.DeviceNonce, w http.ResponseWriter) error {
	response := RequestIDResponse{Success: success, ErrorMessage: message, RequestID: nonce.Nonce}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error forming the request-id response: %v\n", err)
		return err
	}
	return nil
}

// checkAPIKey the API key header to make sure it is an allowed header
func checkAPIKey(apiKey string) error {
	if datastore.Environ.Config.APIKeys == nil || len(datastore.Environ.Config.APIKeys) == 0 {
		log.Println("No API key authorisation defined - default policy is allow")
		return nil
	}

	if _, ok := datastore.Environ.Config.APIKeysMap[apiKey]; !ok {
		return errors.New("Unauthorized API key used")
	}

	return nil
}

// logMessage logs a message in a fixed format so it can be analyzed by log handlers
// e.g. "METHOD CODE descriptive reason"
func logMessage(method, code, reason string) {
	log.Printf("%s %s %s\n", method, code, reason)
}

// define pattern for model name validation
const validModelAllowed = "^[a-z0-9](?:-?[a-z0-9])*$"

var validModelNamePattern = regexp.MustCompile(validModelAllowed)

// validateModelName validates
func validateModelName(name string) error {
	if len(name) == 0 {
		return errors.New("Name must not be empty")
	}

	if strings.ToLower(name) != name {
		return errors.New("Name must not contain uppercase characters")
	}

	if !validModelNamePattern.MatchString(name) {
		return fmt.Errorf("Name contains invalid characters, allowed %q", validModelAllowed)
	}

	return nil
}

// checkUserPermissions retrieves the user from the JWT.
// The user will be restricted by the accounts the username can access and their role.
// If user authentication is turned off, the JWT will irrelevant.
func checkUserPermissions(w http.ResponseWriter, r *http.Request) (string, error) {

	if !datastore.Environ.Config.EnableUserAuth {
		// User authentication is turned off
		return "", nil
	}

	// Check the authentication token
	token, err := JWTCheck(w, r)
	if err != nil {
		return "", err
	}

	// Get the user from the token
	return token.Claims[usso.ClaimsUsername].(string), nil
}
