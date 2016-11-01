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

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"

	"log"

	"fmt"

	"errors"

	"github.com/snapcore/snapd/asserts"
	"github.com/ubuntu-core/identity-vault/service"
)

type requestParams struct {
	Brand        string
	Model        string
	SerialNumber string
	URL          string
	APIKey       string
}

var request requestParams

func main() {
	// Get the parameters from the command line
	parseCommandLine()

	// Create a serial-request assertion
	serialRequest, err := generateSerialRequestAssertion()
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	// Send it to the serial vault via HTTPS
	serialAssertion, err := getSerial(serialRequest)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(serialAssertion)
	os.Exit(0)
}

func parseCommandLine() {
	flag.StringVar(&request.Brand, "brand", "", "The brand-id of the device")
	flag.StringVar(&request.Model, "model", "", "The model name of the device")
	flag.StringVar(&request.SerialNumber, "serial", "", "The serial number of the device")
	flag.StringVar(&request.URL, "url", "", "The base URL of the serial vault API")
	flag.StringVar(&request.APIKey, "api", "", "The API Key for the serial vault")
	flag.Parse()
}

func generatePrivateKey() (asserts.PrivateKey, error) {
	signingKey, err := ioutil.ReadFile("./keystore/TestDeviceKey.asc")
	if err != nil {
		return nil, err
	}
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	privateKey, _, err := service.DeserializePrivateKey(encodedSigningKey)
	return privateKey, nil
}

func generateSerialRequestAssertion() (string, error) {
	privateKey, _ := generatePrivateKey()
	encodedPubKey, _ := asserts.EncodePublicKey(privateKey.PublicKey())

	// Generate a request-id
	r, _ := getRequestID()

	headers := map[string]interface{}{
		"brand-id":   request.Brand,
		"device-key": string(encodedPubKey),
		"request-id": r,
		"model":      request.Model,
		"serial":     request.SerialNumber,
	}

	sreq, err := asserts.SignWithoutAuthority(asserts.SerialRequestType, headers, []byte(""), privateKey)
	if err != nil {
		return "", err
	}

	assertSR := asserts.Encode(sreq)
	return string(assertSR), nil
}

func getRequestID() (string, error) {
	// Format the URL and headers for the HTTP call
	req := getHTTPRequest("request-id", "")

	// Call the /request-id API
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error fetching the request-id")
		return "", err
	}
	defer resp.Body.Close()

	// Parse the API response
	result := service.RequestIDResponse{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Println("Error parsing the request-id")
		return "", err
	}

	return result.RequestID, nil
}

func getSerial(serialRequest string) (string, error) {
	// Format the URL and headers for the HTTP call
	req := getHTTPRequest("serial", serialRequest)

	// Call the /request-id API
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error fetching the serial assertion")
		return "", err
	}
	defer resp.Body.Close()

	// Check the content-type to see if we have a JSON error response
	if resp.Header.Get("Content-Type") == "application/json; charset=UTF-8" {
		// Parse the API response
		result := service.SignResponse{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			log.Println("Error parsing the serial assertion error")
			return "", err
		}
		message := fmt.Sprintf("%s: %s", result.ErrorCode, result.ErrorMessage)
		return "", errors.New(message)
	}

	// Must have a valid assertion
	body, err := ioutil.ReadAll(resp.Body)

	return string(body), err
}

func getHTTPRequest(method, body string) *http.Request {
	// Format the URL and headers for the HTTP call
	url := fmt.Sprintf("%s%s", request.URL, method)
	req, _ := http.NewRequest("POST", url, bytes.NewBufferString(body))
	req.Header.Set("api-key", request.APIKey)
	req.Header.Set("Content-Type", "application/json")

	return req
}
