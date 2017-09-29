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

package config

import (
	"flag"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// Set the application version from a constant
const version = "2.0-1"

// Settings defines the parsed config file settings.
type Settings struct {
	Version        string
	Title          string `yaml:"title"`
	Logo           string `yaml:"logo"`
	DocRoot        string `yaml:"docRoot"`
	Driver         string `yaml:"driver"`
	DataSource     string `yaml:"datasource"`
	KeyStoreType   string `yaml:"keystore"`
	KeyStorePath   string `yaml:"keystorePath"`
	KeyStoreSecret string `yaml:"keystoreSecret"`
	Mode           string `yaml:"mode"`
	CSRFAuthKey    string `yaml:"csrfAuthKey"`
	URLHost        string `yaml:"urlHost"`
	URLScheme      string `yaml:"urlScheme"`
	EnableUserAuth bool   `yaml:"enableUserAuth"`
	JwtSecret      string `yaml:"jwtSecret"`
}

// SettingsFile is the path to the YAML configuration file
var SettingsFile string

// ServiceMode is whether we are running the user or admin service
var ServiceMode string

// ParseArgs checks the command line arguments
func ParseArgs() {
	flag.StringVar(&SettingsFile, "config", "./settings.yaml", "Path to the config file")
	flag.StringVar(&ServiceMode, "mode", "", "Mode of operation: signing, admin or system-user service ")
	flag.Parse()
}

// ReadConfig parses the config file
func ReadConfig(settings *Settings, filePath string) error {
	source, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("Error opening the config file.")
		return err
	}

	err = yaml.Unmarshal(source, &settings)
	if err != nil {
		log.Println("Error parsing the config file.")
		return err
	}

	// Set the application version from the constant
	settings.Version = version

	// Set the service mode from the config file if it is not set
	if ServiceMode == "" {
		ServiceMode = settings.Mode
	}

	return nil
}
