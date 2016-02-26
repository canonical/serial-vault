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
	"time"

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

func formatAssertion(assertions *Assertions) (string, error) {
	timestamp := time.Now().UTC().String()
	assertion := DeviceAssertion{
		Type: "device", Brand: assertions.Brand, Model: assertions.Model,
		SerialNumber: assertions.SerialNumber, Timestamp: timestamp, Revision: assertions.Revision,
		PublicKey: assertions.PublicKey}

	dataToSign, err := yaml.Marshal(assertion)
	if err != nil {
		log.Println("Error formatting the assertions.")
		return "", err
	}
	return string(dataToSign), nil
}

// Return the armored private key as a string
func getPrivateKey(privateKeyFilePath string) ([]byte, error) {
	privateKey, err := ioutil.ReadFile(privateKeyFilePath)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func formatSignResponse(success bool, message, signature string, w http.ResponseWriter) error {
	response := SignResponse{Success: success, ErrorMessage: message, Signature: signature}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the signing response.")
		return err
	}
	return nil
}

func formatModelsResponse(success bool, message string, models []ModelDisplay, w http.ResponseWriter) error {
	response := ModelsResponse{Success: success, ErrorMessage: message, Models: models}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the models response.")
		return err
	}
	return nil
}

func formatBooleanResponse(success bool, message string, w http.ResponseWriter) error {
	response := BooleanResponse{Success: success, ErrorMessage: message}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error forming the boolean response.")
		return err
	}
	return nil
}
