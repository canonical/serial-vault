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
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/snapcore/snapd/asserts"

	"gopkg.in/yaml.v2"
)

// Accepted service modes
var (
	ModeSigning = "signing"
	ModeAdmin   = "admin"
)

// Set the application version from a constant
const version = "0.7.0"

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
	if assertion == nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		response := SignResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, Signature: ""}

		// Encode the response as JSON
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Println("Error forming the signing response.")
			return err
		}
	} else {
		w.Header().Set("Content-Type", asserts.MediaType)
		w.WriteHeader(http.StatusOK)
		encoder := asserts.NewEncoder(w)
		err := encoder.Encode(assertion)
		if err != nil {
			// Not much we can do if we're here - apart from panic!
			log.Println("Error encoding the assertion.")
			return err
		}
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

func formatNonceResponse(success bool, errorCode, errorSubcode, message string, nonce DeviceNonce, w http.ResponseWriter) error {
	response := NonceResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, Nonce: nonce.Nonce}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the nonce response.")
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
