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
	"github.com/dgrijalva/jwt-go"
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

func formatUserResponse(success bool, errorCode, errorSubcode, message string, user datastore.User, w http.ResponseWriter) error {
	response := UserResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, User: user}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the user response.")
		return err
	}
	return nil
}

func formatUsersResponse(success bool, errorCode, errorSubcode, message string, users []datastore.User, w http.ResponseWriter) error {
	response := UsersResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, Users: users}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the users response.")
		return err
	}
	return nil
}

// checkAPIKey the API key header to make sure it is an allowed header
func checkAPIKey(apiKey string) error {
	if len(apiKey) == 0 {
		return errors.New("Blank API key used")
	}

	if ok := datastore.Environ.DB.CheckAPIKey(apiKey); !ok {
		return errors.New("Unauthorized API key used")
	}

	return nil
}

// logMessage logs a message in a fixed format so it can be analyzed by log handlers
// e.g. "METHOD CODE descriptive reason"
func logMessage(method, code, reason string) {
	log.Printf("%s %s %s\n", method, code, reason)
}

// define patterns for validation
const validModelAllowed = "^[a-z0-9](?:-?[a-z0-9])*$"
const validUsernameAllowed = validModelAllowed

var validModelNamePattern = regexp.MustCompile(validModelAllowed)
var validUsernamePattern = regexp.MustCompile(validUsernameAllowed)

// validateModelName validates
func validateModelName(name string) error {
	return validateSyntax("Name", name, validModelNamePattern)
}

func validateUsername(username string) error {
	return validateSyntax("Username", username, validUsernamePattern)
}

func validateUserRole(role int) error {
	if role != datastore.Standard && role != datastore.Admin && role != datastore.Superuser {
		return errors.New("Role is not amongst valid ones")
	}
	return nil
}

func validateUserFullName(name string) error {
	return validateNotEmpty("Name", name)
}

func validateUserEmail(email string) error {
	return validateNotEmpty("Email", email)
}

func validateNotEmpty(fieldName, fieldValue string) error {
	if len(fieldValue) == 0 {
		return fmt.Errorf("%v must not be empty", fieldName)
	}
}

func validateSyntax(fieldName, fieldValue string, pattern *regexp.Regexp) error {
	if err := validateNotEmpty(fieldName, fieldValue); err != nil {
		return err
	}

	if strings.ToLower(fieldValue) != fieldValue {
		return fmt.Errorf("%v must not contain uppercase characters", fieldName)
	}

	if !pattern.MatchString(fieldValue) {
		return fmt.Errorf("%v contains invalid characters, allowed %q", fieldName, validModelAllowed)
	}

	return nil
}

// checkUserPermissions retrieves the user from the JWT.
// The user will be restricted by the accounts the username can access and their role i.e. only Admin and Superuser
// These are the rules:
//
// 	- If user authentication is turned off, the JWT will irrelevant. In this case the username is returned as "" if Admin
// 		is allowed, or error if only Superuser is allowed.
//	- If database user role is less than allowed role, an error is returned
//	- If there is no database user, role is considered Admin
//
func checkUserPermissions(w http.ResponseWriter, r *http.Request, minimumAuthorizedRole int) (string, error) {
	// User authentication is turned off
	if !datastore.Environ.Config.EnableUserAuth {
		// Superuser permissions don't allow turned off authentication
		if minimumAuthorizedRole == datastore.Superuser {
			return "", errors.New("A The user is not authorized")
		}
		return "", nil
	}

	// Check the authentication token
	token, err := JWTCheck(w, r)
	if err != nil {
		return "", err
	}

	// Get the user from the token
	claims := token.Claims.(jwt.MapClaims)
	username := claims[usso.ClaimsUsername].(string)

	// Check that the role is at least the authorized one.
	// NOTE: Take into account that RoleForUser() returns Admin in case username is empty
	role := datastore.Environ.DB.RoleForUser(username)
	if role < minimumAuthorizedRole {
		return username, errors.New("The user is not authorized")
	}

	return username, nil
}
