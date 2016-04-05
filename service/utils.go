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
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ubuntu-core/snappy/asserts"

	"gopkg.in/yaml.v2"
)

// ConfigSettings defines the parsed config file settings.
type ConfigSettings struct {
	PrivateKeyPath string `yaml:"privateKeyPath"`
	Version        string `yaml:"version"`
	Title          string `yaml:"title"`
	Logo           string `yaml:"logo"`
	Driver         string `yaml:"driver"`
	DataSource     string `yaml:"datasource"`
	KeyStoreType   string `yaml:"keystore"`
	KeyStorePath   string `yaml:"keystorePath"`
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
	Config         ConfigSettings
	DB             Datastore
	AuthorizedKeys AuthorizedKeystore
	KeypairDB      *asserts.Database
}

var settingsFile string

// ParseArgs checks the command line arguments
func ParseArgs() {
	flag.StringVar(&settingsFile, "config", "./settings.yaml", "Path to the config file")
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
	return nil
}

func formatSignResponse(success bool, errorCode, errorSubcode, message string, assertion asserts.Assertion, w http.ResponseWriter) error {
	if assertion == nil {
		response := SignResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, Signature: ""}

		// Encode the response as JSON
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Println("Error forming the signing response.")
			return err
		}
	} else {
		log.Println("-------------------------")
		log.Printf("%v", assertion)
	}

	return nil
}

func formatModelsResponse(success bool, errorCode, errorSubcode, message string, models []ModelDisplay, w http.ResponseWriter) error {
	response := ModelsResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, Models: models}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the models response.")
		return err
	}
	return nil
}

func formatBooleanResponse(success bool, errorCode, errorSubcode, message string, w http.ResponseWriter) error {
	response := BooleanResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the boolean response.")
		return err
	}
	return nil
}

func formatModelResponse(success bool, errorCode, errorSubcode, message string, model ModelDisplay, w http.ResponseWriter) error {
	response := ModelResponse{Success: success, ErrorCode: errorCode, ErrorSubcode: errorSubcode, ErrorMessage: message, Model: model}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the model response.")
		return err
	}
	return nil
}
