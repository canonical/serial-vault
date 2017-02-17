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
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/snapcore/snapd/asserts"

	"gopkg.in/yaml.v2"
)

// Accepted service modes
var (
	ModeSigning = "signing"
	ModeAdmin   = "admin"
)

// Set the application version from a constant
const version = "1.3.0"

// Set the nonce expiry time
const nonceMaximumAge = 600

// ConfigSettings defines the parsed config file settings.
type ConfigSettings struct {
	Version        string
	Title          string   `yaml:"title"`
	Logo           string   `yaml:"logo"`
	DocRoot        string   `yaml:"docRoot"`
	Driver         string   `yaml:"driver"`
	DataSource     string   `yaml:"datasource"`
	KeyStoreType   string   `yaml:"keystore"`
	KeyStorePath   string   `yaml:"keystorePath"`
	KeyStoreSecret string   `yaml:"keystoreSecret"`
	Mode           string   `yaml:"mode"`
	APIKeys        []string `yaml:"apiKeys"`
	APIKeysMap     map[string]struct{}
}

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

// Env Environment struct that holds the config and data store details.
type Env struct {
	Config    ConfigSettings
	DB        Datastore
	KeypairDB *KeypairDatabase
}

var settingsFile string

// ServiceMode is whether we are running the user or admin service
var ServiceMode string

// BooleanResponse is the JSON response from an API method, indicating success or failure.
type BooleanResponse struct {
	Success      bool   `json:"success"`
	ErrorCode    string `json:"error_code"`
	ErrorSubcode string `json:"error_subcode"`
	ErrorMessage string `json:"message"`
}

// ParseArgs checks the command line arguments
func ParseArgs() {
	flag.StringVar(&settingsFile, "config", "./settings.yaml", "Path to the config file")
	flag.StringVar(&ServiceMode, "mode", "", "Mode of operation: signing service or admin service")
	flag.Parse()
}

// ReadConfig parses the config file
func ReadConfig(config *ConfigSettings) error {
	source, err := ioutil.ReadFile(settingsFile)
	if err != nil {
		log.Println("Error opening the config file.")
		return err
	}

	err = yaml.Unmarshal(source, &config)
	if err != nil {
		log.Println("Error parsing the config file.")
		return err
	}

	// Set the application version from the constant
	config.Version = version

	// Set the service mode from the config file if it is not set
	if ServiceMode == "" {
		ServiceMode = config.Mode
	}

	// Migrate the API keys to a map for more efficient lookups
	config.APIKeysMap = make(map[string]struct{})
	for _, key := range config.APIKeys {
		config.APIKeysMap[key] = struct{}{}
	}

	return nil
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

func formatKeypairsResponse(success bool, errorCode, errorSubcode, message string, keypairs []Keypair, w http.ResponseWriter) error {
	response := KeypairsResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, Keypairs: keypairs}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the keypairs response.")
		return err
	}
	return nil
}

func formatSigningLogResponse(success bool, errorCode, errorSubcode, message string, logs []SigningLog, w http.ResponseWriter) error {
	response := SigningLogResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, SigningLog: logs}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the signing log response.")
		return err
	}
	return nil
}

func formatSigningLogFiltersResponse(success bool, errorCode, errorSubcode, message string, filters SigningLogFilters, w http.ResponseWriter) error {
	response := SigningLogFiltersResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, SigningLogFilters: filters}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the signing log response.")
		return err
	}
	return nil
}

func formatRequestIDResponse(success bool, message string, nonce DeviceNonce, w http.ResponseWriter) error {
	response := RequestIDResponse{Success: success, ErrorMessage: message, RequestID: nonce.Nonce}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error forming the request-id response: %v\n", err)
		return err
	}
	return nil
}

// padRight truncates a string to a specific length, padding with a named
// character for shorter strings.
func padRight(str, pad string, length int) string {
	for {
		str += pad
		if len(str) > length {
			return str[0:length]
		}
	}
}

// checkAPIKey the API key header to make sure it is an allowed header
func checkAPIKey(apiKey string) error {
	if Environ.Config.APIKeys == nil || len(Environ.Config.APIKeys) == 0 {
		log.Println("No API key authorisation defined - default policy is allow")
		return nil
	}

	if _, ok := Environ.Config.APIKeysMap[apiKey]; !ok {
		return errors.New("Unauthorized API key used")
	}

	return nil
}

// logMessage logs a message in a fixed format so it can be analyzed by log handlers
// e.g. "METHOD CODE descriptive reason"
func logMessage(method, code, reason string) {
	log.Printf("%s %s %s\n", method, code, reason)
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

// define pattern for model name validation
const validModelAllowed = "^[a-zA-Z0-9](?:-?[a-zA-Z0-9])*$"

var validModelNamePattern = regexp.MustCompile(validModelAllowed)

// validateModelName validates
func validateModelName(name string) error {
	if len(name) == 0 {
		return errors.New("Name must not be empty")
	}

	if !validModelNamePattern.MatchString(name) {
		return fmt.Errorf("Name contains invalid characters, allowed %q", validModelAllowed)
	}

	// TODO: support the concept of case insensitive/preserving string headers
	if strings.ToLower(name) != name {
		return errors.New("Name must not contain uppercase characters")
	}
	return nil
}
