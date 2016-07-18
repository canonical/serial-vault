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
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/openpgp/packet"

	"github.com/snapcore/snapd/asserts"

	"gopkg.in/yaml.v2"
)

// Accepted service modes
var (
	ModeSigning = "signing"
	ModeAdmin   = "admin"
)

// ConfigSettings defines the parsed config file settings.
type ConfigSettings struct {
	Version        string   `yaml:"version"`
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

func generateAuthKey(authorityID, keyID string) string {
	return strings.Join([]string{authorityID, "/", keyID}, "")
}

func createSecret(length int) (string, error) {
	rb := make([]byte, length)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(rb), nil
}

// encryptKey uses symmetric encryption to encrypt the data for storage
func encryptKey(plainTextKey, keyText string) ([]byte, error) {
	// The AES key needs to be 16 or 32 bytes i.e. AES-128 or AES-256
	aesKey := padRight(keyText, "x", 32)

	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		log.Printf("Error creating the cipher block: %v", err)
		return nil, err
	}

	// The IV needs to be unique, but not secure. Including it at the start of the plaintext
	ciphertext := make([]byte, aes.BlockSize+len(plainTextKey))
	iv := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		log.Printf("Error creating the IV for the cipher: %v", err)
		return nil, err
	}

	// Use CFB mode for the encryption
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plainTextKey))

	return ciphertext, nil
}

func decryptKey(sealedKey []byte, keyText string) ([]byte, error) {
	aesKey := padRight(keyText, "x", 32)

	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		log.Printf("Error creating the cipher block: %v", err)
		return nil, err
	}

	if len(sealedKey) < aes.BlockSize {
		return nil, errors.New("Cipher text too short")
	}

	iv := sealedKey[:aes.BlockSize]
	sealedKey = sealedKey[aes.BlockSize:]

	// Use CFB mode for the decryption
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(sealedKey, sealedKey)

	return sealedKey, nil
}

// decodePublicKey replicates a private method in snapcore asserts to convert the device-key header
// into a usable PublicKey format.
func decodePublicKey(pubKey []byte) (asserts.PublicKey, error) {
	pkt, err := decodeOpenpgp(pubKey, "public key")
	if err != nil {
		return nil, err
	}
	pubk, ok := pkt.(*packet.PublicKey)
	if !ok {
		return nil, fmt.Errorf("expected public key, got instead: %T", pkt)
	}
	return asserts.OpenPGPPublicKey(pubk), nil
}

func decodeOpenpgp(formatAndBase64 []byte, kind string) (packet.Packet, error) {
	if len(formatAndBase64) == 0 {
		return nil, fmt.Errorf("empty %s", kind)
	}
	format, data, err := splitFormatAndBase64Decode(formatAndBase64)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", kind, err)
	}
	if format != "openpgp" {
		return nil, fmt.Errorf("unsupported %s format: %q", kind, format)
	}
	pkt, err := packet.Read(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("could not decode %s data: %v", kind, err)
	}
	return pkt, nil
}

func splitFormatAndBase64Decode(formatAndBase64 []byte) (string, []byte, error) {
	parts := bytes.SplitN(formatAndBase64, []byte(" "), 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("expected format and base64 data separated by space")
	}
	buf := make([]byte, base64.StdEncoding.DecodedLen(len(parts[1])))
	n, err := base64.StdEncoding.Decode(buf, parts[1])
	if err != nil {
		return "", nil, fmt.Errorf("could not decode base64 data: %v", err)
	}
	return string(parts[0]), buf[:n], nil
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
