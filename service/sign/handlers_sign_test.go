// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2018 Canonical Ltd
 * License granted by Canonical Limited
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

package sign_test

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/crypt"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/snapcore/snapd/asserts"
	check "gopkg.in/check.v1"
)

func TestSignSuite(t *testing.T) { check.TestingT(t) }

type SignSuite struct{}

type SuiteTest struct {
	MockError bool
	Method    string
	URL       string
	Data      []byte
	Code      int
	Type      string
	APIKey    string
}

var _ = check.Suite(&SignSuite{})

func (s *SignSuite) SetUpTest(c *check.C) {
	// Mock the database
	config := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)
}

func sendRequest(method, url string, data io.Reader, apiKey string, c *check.C) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)
	r.Header.Set("api-key", apiKey)

	service.SigningRouter().ServeHTTP(w, r)

	return w
}

func (s *SignSuite) TestSerial(c *check.C) {
	// Generate a test serial-request assertion
	assert, err := generateSerialRequestAssertion("alder", "A123456L", "")
	c.Assert(err, check.IsNil)
	assertInactive, err := generateSerialRequestAssertion("inactive", "A123456L", "")
	c.Assert(err, check.IsNil)
	assertSerialInBody, err := generateSerialRequestAssertion("alder", "", "serial: A123456L")
	c.Assert(err, check.IsNil)
	assertSPlusM, err := serialRequestPlusModelAssertion(c)
	c.Assert(err, check.IsNil)
	assertSPlusBad, err := serialRequestPlusBadModelAssertion(c)
	c.Assert(err, check.IsNil)
	assertSPlusMPlusBad, err := serialRequestPlusModelPlusGarbage(c)
	c.Assert(err, check.IsNil)
	assertSPlusMPlusExtra, err := serialRequestPlusModelPlusExtra(c)
	c.Assert(err, check.IsNil)
	assertSPlusMPlusWrong, err := serialRequestPlusModelPlusWrongModel(c)
	c.Assert(err, check.IsNil)
	assertNoSerial, err := generateSerialRequestAssertion("alder", "", "")
	c.Assert(err, check.IsNil)
	assertFakeModel, err := generateSerialRequestAssertion("invalid", "A123456L", "")
	c.Assert(err, check.IsNil)
	assertDuplicate, err := generateSerialRequestAssertion("alder", "Aduplicate", "")
	c.Assert(err, check.IsNil)
	assertSigningLogError, err := generateSerialRequestAssertion("alder", "AsigninglogError", "")
	c.Assert(err, check.IsNil)

	tests := []SuiteTest{
		{false, "POST", "/v1/serial", assert, 200, asserts.MediaType, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertSerialInBody, 200, asserts.MediaType, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertSPlusM, 200, asserts.MediaType, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertSPlusBad, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertSPlusMPlusBad, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertSPlusMPlusExtra, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertSPlusMPlusWrong, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertNoSerial, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertFakeModel, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assert, 400, response.JSONHeader, "NoModelForApiKey"},
		{false, "POST", "/v1/serial", assertSigningLogError, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertDuplicate, 200, asserts.MediaType, "ValidAPIKey"},
		{false, "POST", "/v1/serial", nil, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", []byte(""), 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assert, 400, response.JSONHeader, "InvalidAPIKey"},
		{false, "POST", "/v1/serial", assertInactive, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", []byte(badSerialRequest), 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", []byte(assertionWrongType), 400, response.JSONHeader, "ValidAPIKey"},
		{true, "POST", "/v1/serial", assert, 400, response.JSONHeader, "ValidAPIKey"},
	}

	for _, t := range tests {
		if t.MockError {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}

		w := sendRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.APIKey, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		datastore.Environ.DB = &datastore.MockDB{}
	}
}

func (s *SignSuite) TestRequestIDHandler(c *check.C) {
	tests := []SuiteTest{
		{false, "POST", "/v1/request-id", nil, 200, response.JSONHeader, "InbuiltAPIKey"},
		{false, "POST", "/v1/request-id", nil, 400, response.JSONHeader, "InvalidAPIKey"},
		{true, "POST", "/v1/request-id", nil, 400, response.JSONHeader, "InbuiltAPIKey"},
	}

	for _, t := range tests {
		if t.MockError {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}

		w := sendRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.APIKey, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		datastore.Environ.DB = &datastore.MockDB{}
	}
}

func generatePrivateKey() (asserts.PrivateKey, error) {
	signingKey, err := ioutil.ReadFile("../../keystore/TestDeviceKey.asc")
	if err != nil {
		return nil, err
	}
	encodedSigningKey := base64.StdEncoding.EncodeToString(signingKey)

	privateKey, _, err := crypt.DeserializePrivateKey(encodedSigningKey)
	return privateKey, err
}

func generateSerialRequestAssertion(model, serial, body string) ([]byte, error) {
	privateKey, _ := generatePrivateKey()
	encodedPubKey, _ := asserts.EncodePublicKey(privateKey.PublicKey())

	headers := map[string]interface{}{
		"brand-id":   "system",
		"device-key": string(encodedPubKey),
		"request-id": "REQID",
		"model":      model,
	}

	if serial != "" {
		headers["serial"] = serial
	}

	sreq, err := asserts.SignWithoutAuthority(asserts.SerialRequestType, headers, []byte(body), privateKey)
	if err != nil {
		return nil, err
	}

	return asserts.Encode(sreq), nil
}

func serialRequestPlusModelAssertion(c *check.C) ([]byte, error) {
	// Generate a test serial-request assertion
	assertions, err := generateSerialRequestAssertion("alder", "A123456L", "")
	if err != nil {
		return nil, err
	}

	assertions = append(assertions, []byte("\n"+modelAssertion)...)
	return assertions, nil
}

func serialRequestPlusBadModelAssertion(c *check.C) ([]byte, error) {
	// Generate a test serial-request assertion
	assertions, err := generateSerialRequestAssertion("alder", "A123456L", "")
	if err != nil {
		return nil, err
	}

	assertions = append(assertions, []byte("\n\nxyz")...)
	return assertions, nil
}

func serialRequestPlusModelPlusExtra(c *check.C) ([]byte, error) {
	// Generate a test serial-request assertion
	assertions, err := generateSerialRequestAssertion("alder", "A123456L", "")
	if err != nil {
		return nil, err
	}

	assertions = append(assertions, []byte("\n"+modelAssertion+"\n"+modelAssertion)...)
	return assertions, nil
}

func serialRequestPlusModelPlusGarbage(c *check.C) ([]byte, error) {
	// Generate a test serial-request assertion
	assertions, err := generateSerialRequestAssertion("alder", "A123456L", "")
	if err != nil {
		return nil, err
	}

	assertions = append(assertions, []byte("\n"+modelAssertion+"\n\nxyz")...)
	return assertions, nil
}

func serialRequestPlusModelPlusWrongModel(c *check.C) ([]byte, error) {
	// Generate a test serial-request assertion
	assertions, err := generateSerialRequestAssertion("alder", "A123456L", "")
	if err != nil {
		return nil, err
	}

	assertions = append(assertions, []byte("\n"+string(assertions))...)
	return assertions, nil
}

const modelAssertion = `type: model
authority-id: system
series: 16
brand-id: system
model: alder
display-name: Alder
architecture: amd64
gadget: alder-gadget
kernel: alder-linux
store: brand-store
body-length: 0
sign-key-sha3-384: Jv8_JiHiIzJVcO9M55pPdqSDWUvuhfDIBJUS-3VW7F_idjix7Ffn5qMxB21ZQuij
timestamp: 2016-01-02T15:04:05Z

AXNpZw==
`

const badSerialRequest = `type: serial-request
brand-id: System
device-key:
    AcbBTQRWhcGAARAAx6VJoV9ZKASKa1pFA0G6hQimQT7ym8EZFN7+SzZhWSWLIwFd06oRQVKetQB6
    a+ab0zMN3yfI94aB9aH/q6vA7T7Yo1KaBFy4aaztUvDmMzEGaVwJvDSBUBFr4yUCJEtLXAw5fMkS
    DGvNUFRacLifAfGU5mLHJl7WXY2e7T+VjJPoSU3nAZjvGd2YQnQ1fNfQ0X+zuQVDGrtmJJF3x0CM
    8LL0XF4UCTBYyLZK2YvSKrrk2qmIUVr3PXoY+fH9Bs5AZAAZ91GIrt0qc0uradXxI6kq8zy8bVl8
    GTazEmkBE9Y7snAqWJWGXt9K4tO7h+4Xgprvf27dddp68XS2KHT3r86qC/1i9mTGMbHWJ5NKd/No
    Jnawjc1qo2tnVVyw+GKwMhukpvmtuejhtk395dNczGZ2sw2yPHORUHUyq/sPLoAWyWLQFHL3MxQq
    qyxgxWNnRYhcs6wmWEf2nNFlllld6YzS7It+cA+I04j5h85DGO6+knn1J7X4WuORDx3nn3bEQKik
    v4uu1xFJYk6N14B/ofMoUCzbPtgkNpmV0NmgFeogx+I5yRuF0EF5U+LfMuAE+ROoYHHwiBHeSttr
    YewdunntDyeRUc3CTwsvfq2zARObr5He5z4ldSASuzxbzEEXVd6UERPN+zeJGyctKIYEqvpSNNuu
    4Fs8Ctp6yar9KucAEQEAAQ==
request-id: REQID
body-length: 10
sign-key-sha3-384: UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO

HW-DETAILS

AcLBUgQAAQoABgUCV7R2CQAAYx4QAB2vxjMYFb5nmQdkeX4pjbD6sjheD5PZV6h3DDDznZccMP+v
y3x8PtTA7h1oN04nzMBqilPH01buSVSSzVAy789oecAwSMhpUi50lVIWdye2zeE+G3DEbZdHOBod
+rxK0LTuDxf8dCCt2zbGlm/4wSORGPsn4dR+G6Da+ZEEAORQuHCdVGNe9LgFi7ZIX5ZkvK5oNTyH
Ebgf4VLVpHpBZ2sl6sNPwLDpH1LOmMFgq3tEZXaKaa9QAn6g/S/hgTbv6eDfKHTX99ynpqgu6+am
+HZ28PG39kbJoKpexzIxhxhR42hKso3xUHJfwFSeTxLIlRK0KlDRsDOAe6MzjhTnA8b/xMjw8NaF
6q60hgS8Qytyvu1/7f75CTy4cTwenmUuw/v2mcO98FurVpDFzXSb5HK44Ej6gYXpTOtE4lSH0oP4
7VL/JAjhP3qncgDMVh0URIqh6FDCD7bb2USP4Fo2yvkVfLHCS80vZGury+rGxV2bPRcOTfbnoKZy
cwmwjJS6vKEYIIlMwVaHsPd9ZBvyYBwTzfGKtoazjm44mByBG0AEUZrZ7MWnf7lWwU+Ze3g3GNQF
9EEnrN8E9yYxFgCGaYA7kBFhkhJElafMQNr/EYU3bwLKHa++1iKmNKcGePRZyy9kyUpmgtRaQt/6
ic2Xx1ds+umMC5AHW9wZAWNPDI/T
`

const assertionWrongType = `type: model
authority-id: System
brand-id: System
model: alder
serial: A1234-L
series: Alder
revision: 1
os: 14.04
core: apple
architecture: i686
gadget: magic wand
kernel: 4.2.0-35-generic
store: Canonical
class: Class
allowed-modes: all
required-snaps:
  - gadget
timestamp: 2016-01-02T15:04:05Z
sign-key-sha3-384: UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO
device-key: openpgp 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

openpgp PvKbwRkU3v5LNmFZJYsjAV3TqhFBUp61AHpr5pvTMw3fJ8j3h
`

func (s *SignSuite) TestSignHandlerErrorKeyStore(c *check.C) {
	// Mock the database and the keystore
	settings := config.Settings{KeyStoreType: "filesystem", KeyStorePath: "../../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: settings}
	datastore.Environ.KeypairDB, _ = datastore.GetErrorMockKeyStore(settings)

	// Generate a test serial-request assertion
	assertions, err := generateSerialRequestAssertion("alder", "A1234L", "")
	c.Assert(err, check.IsNil)

	w := sendRequest("POST", "/v1/serial", bytes.NewReader(assertions), "ValidAPIKey", c)
	c.Assert(w.Code, check.Equals, 400)
	c.Assert(w.Header().Get("Content-Type"), check.Equals, response.JSONHeader)
}
