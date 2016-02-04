package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"gopkg.in/yaml.v2"
)

// ConfigSettings defines the parsed config file settings.
type ConfigSettings struct {
	PrivateKeyPath string `yaml:"privateKeyPath"`
	Version        string `yaml:"version"`
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

func formatAssertion(assertions *Assertions) string {
	timestamp := time.Now().UTC().String()
	assertion := DeviceAssertion{
		Type: "device", Brand: assertions.Brand, Model: assertions.Model,
		SerialNumber: assertions.SerialNumber, Timestamp: timestamp, Revision: assertions.Revision,
		PublicKey: assertions.PublicKey}

	dataToSign, err := yaml.Marshal(assertion)
	if err != nil {
		panic(err)
	}
	return string(dataToSign)
}

// Return the armored private key as a string
func getPrivateKey(privateKeyFilePath string) ([]byte, error) {
	privateKey, err := ioutil.ReadFile(privateKeyFilePath)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func formatSignResponse(success bool, message, signature string, w http.ResponseWriter) {
	response := SignResponse{Success: success, ErrorMessage: message, Signature: signature}

	// Encode the response as JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}
