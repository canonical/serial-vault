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

type SignSuite struct {
	keypairMgr asserts.KeypairManager
}

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

	assertionsWithSerial := append(assertSPlusM, []byte("\n"+serial)...)

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
		{false, "POST", "/v1/serial", assertionsWithSerial, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", []byte(wrongSignatureSerialReq), 400, response.JSONHeader, "ValidAPIKey"},
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
const newModelAssertion = `type: model
authority-id: mybrand
series: 16
brand-id: mybrand
model: alder-mybrand
display-name: Mybrand
architecture: amd64
gadget: mybrand-gadget
kernel: mybrand-linux
store: mybrand-store
body-length: 0
sign-key-sha3-384: Jv8_JiHiIzJVcO9M55pPdqSDWUvuhfDIBJUS-3VW7F_idjix7Ffn5qMxB21ZQuij
timestamp: 2018-05-07T14:38:51Z

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

const serial = `type: serial
authority-id: system
revision: 4
brand-id: system
model: alder
serial: Aduplicate
device-key:
    AcbATQRWhcGAAQgAzuDA7nxtfh77/XeX0UoIa8x0ILAtd4vEKRHXUmuf5LEUv0s7yQqtXjPQl5Rj
    Red4ssWFPFmanvgXjZMVmRfiTBW7VK86eG8E35TyIySWeT4dYqPzcEHLA4vPvEp0vS7IV0rr+tS5
    6NttX/oQmh/BSvBQruwIxpIJK0JhjSIzl6fO9RLFJe0eJCvpWPSSBiFeJAVeCfAIyrr4acANtyf5
    jkmwru1M+EpPj1VFN9yzPdspinnJW07w3RX5uqew6t325cTozKsrpV3OsK0QKQ7rttQpcKarY8rT
    QpfkVoSBFbgV6SlbragpaWWd0YTE4+YtXZ2OD0l8ipTyRDeE9lNotwARAQAB
device-key-sha3-384: majNNh3Nbgg0CozIjfLrvvEWG830dZmm76gJHnyyrW4g7udYbfVgt0WO15ayuN5l
timestamp: 2019-05-10T12:07:36+02:00
sign-key-sha3-384: UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO

AcLBUgQAAQoABgUCXNVNaAAAT0gQADiWK/pBSjDonvjiHNDiUYEMDfiu+WAuLo/k2SaWkoY0EgDy
Jk48BiIqsVcPQGOUIlFste+q7iqXhMO9PnzbdRUeEx5CtNioIzfoDmtuKUN4fweJ/J8v947oKmHg
DCdgf9dBDWU9mZGFa/FZqB5H+NefSV0dZo70hjsFbF60+zJN0RoQ9jk8zOA3LjF43u2qRak/yeie
4RxHDSGiKgvVouMoO8yla5897r9938Rv/zdN45LaXxhuEZ0HbFhnRYCmUw3ctgq0zkjkWCICFsLZ
bYT8tc+ZbcvyQoIND8r9QnV3DDw15fl037QBEhWH428bll3Pj2L0oP6ujGRVjKGcQND5KU3pga5l
96XFUPtWES8CJo7ekGHrmYihy6ThBmIrvaCCj5Pui2abUrBXb1aIhXHJzRRH+xN//68uaoYAjCTD
oa9BVIGHOGpIPO9ZpFE35SXiV8ktYfe3hbjPxDKYKqwMmwGwdQngcQ9NnOLvybv+uEfsF2eXFi1E
TlfGVlrz1fvafv98+mm5+GASPV3XOsL/HAhw0DxadCFOQOk+ALBypFDJLlocyb3PSyyr3RdvbnBH
NJJaotr3y52Cniv1XfM8sX2QWS/xfBNomEuySmjriKKapoFksigxPJm5TNjGQBNsfMDej+b5qdf6
SUH5uDA3J2ll8Oau8J2GSmB4Xt3C
`

const wrongSerial = `type: serial
authority-id: system
revision: 1
brand-id: system
model: ash
serial: A123456L
device-key:
    AcbATQRWhcGAAQgAzuDA7nxtfh77/XeX0UoIa8x0ILAtd4vEKRHXUmuf5LEUv0s7yQqtXjPQl5Rj
    Red4ssWFPFmanvgXjZMVmRfiTBW7VK86eG8E35TyIySWeT4dYqPzcEHLA4vPvEp0vS7IV0rr+tS5
    6NttX/oQmh/BSvBQruwIxpIJK0JhjSIzl6fO9RLFJe0eJCvpWPSSBiFeJAVeCfAIyrr4acANtyf5
    jkmwru1M+EpPj1VFN9yzPdspinnJW07w3RX5uqew6t325cTozKsrpV3OsK0QKQ7rttQpcKarY8rT
    QpfkVoSBFbgV6SlbragpaWWd0YTE4+YtXZ2OD0l8ipTyRDeE9lNotwARAQAB
device-key-sha3-384: majNNh3Nbgg0CozIjfLrvvEWG830dZmm76gJHnyyrW4g7udYbfVgt0WO15ayuN5l
timestamp: 2019-05-09T09:18:56+02:00
sign-key-sha3-384: UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO

AcLBUgQAAQoABgUCXNPUYAAAijMQABuugnuwvYzwtTXwKx6cyIpfcyn8SeQwmPuLkod4YzjZFopf
kk74iGshOKeMO1yMw2vHAK9t4bPLxuZDhG23GRehbO/8kddDbFmfuyj9OpWhP7SzJ/rhjgQMundN
aTh97pYoU5SGaovc3skCEvzLh8eUZB02NJELTuAaWWioTaI01kp78RXZv6QcGuT8cS8eD9Psdiga
hZfqClX6lDB2M7OrPTk3LgC/YIG4yRv8mz/B4fRaExxF+eg7m8KFIWCHRtL0poCWXqdUxkq5mt6B
fZR+YHKV9BEBD2Zi0xjy+UlZz2epavNRx8ksFvqnjo07t+TKp8J14DTiOlpGrJeU6XUdGscQDjMO
kExXzxAd12R1kKbtaAySFSu4mvFC6TnjJFU2lKtSO2yHKcptuSPhkQlQQH2gVHUnlgDx5hMDXV1w
ppwx/YRo+Oqc+5lwiSYafze2Q+X76K7UEPkqDP0Ck99sGk6XEhk9UIwH/aTQKvq94HFsDmplMhSY
wvEmhq8AwfwJ7Yrfl56q4QoQqRKLwTTU7dhG8FWwzeFsjQiQ/Mw63cBX39kxEiTD4LXBT40brGl+
B75/84prNnyI32T8ye/8IByM47KPiOBTwTk+iKYhJwR0sEmrEdB74p0jJiOAJGbvJzvuKlI+ntUt
v3IuKTcYKOSrpMBsArOm/D7mYxA8
`

const wrongDeviceKeySerial = `type: serial
authority-id: system
revision: 1
brand-id: system
model: alder
serial: A123456L
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
device-key-sha3-384: UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO
timestamp: 2019-05-13T05:57:41+02:00
sign-key-sha3-384: UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO

AcLBUgQAAQoABgUCXNjrNQAAYhoQAKFS/5a3B6wItn8IM7cSvu9Q++Bd6Tbc/64zghstJZj9QNrX
F370dRaP/UAbbGftjbZk3vm/K5UYZv65vODd9TDe/dz8RUODmnX7isT9VWCe4y5Vxr/asNF0u8k1
GrYzWs3PereAzmIw2Icm/5vcdBXIEYKaSsYOkPmWJzCTXi6ZEg0CQg6lf9wPuTHcU8XaEStZdAbm
7dQNHwlb+Y/jv/+9P5sPhDkxxcVKnp65ONoCisTiEhEOh55RjnUivaNj6rehChv+Pu8sng5VJksw
rS3nOxyxxrpgL9LFZhKcul+UKC4PFlsWGD/mOb5LsbRX9p0UGtXa/l98YQaGsSI91fOw2OA16vYQ
JXsID7Zxp/anTbwucnNwbWJA354ilP7zk2uTlTeKbG4dTkL+go4VNnar1/Ahgb3dEQ3CH+maLXnG
EIZO/VOpBoz78WZ0z/vL8fytdyz1Pq0kBVpeUnFixZCxxbvnlxb/vA+oUuRx6+ruEs61L3rTW9jJ
fCCQXNYUhKzUrSkNwMnfW7uhSilD6GGx8p8yWWmVjKm6PEaSXblT4dnVIJPNtC1fVASXWgtI0xnh
x3ltKuC0NdZQFRNUYNsuE3z3P99AZwMyPQJoFU6Le8Of8TF7AuvvKHzigKhFL4Obvp54JhJU9nlN
S26V/IDEdgkhpeF82yHnyst++Yre
`

const wrongBrandSerial = `type: serial
authority-id: systemX
revision: 4
brand-id: systemX
model: alder
serial: A123456L
device-key:
    AcbATQRWhcGAAQgAzuDA7nxtfh77/XeX0UoIa8x0ILAtd4vEKRHXUmuf5LEUv0s7yQqtXjPQl5Rj
    Red4ssWFPFmanvgXjZMVmRfiTBW7VK86eG8E35TyIySWeT4dYqPzcEHLA4vPvEp0vS7IV0rr+tS5
    6NttX/oQmh/BSvBQruwIxpIJK0JhjSIzl6fO9RLFJe0eJCvpWPSSBiFeJAVeCfAIyrr4acANtyf5
    jkmwru1M+EpPj1VFN9yzPdspinnJW07w3RX5uqew6t325cTozKsrpV3OsK0QKQ7rttQpcKarY8rT
    QpfkVoSBFbgV6SlbragpaWWd0YTE4+YtXZ2OD0l8ipTyRDeE9lNotwARAQAB
device-key-sha3-384: majNNh3Nbgg0CozIjfLrvvEWG830dZmm76gJHnyyrW4g7udYbfVgt0WO15ayuN5l
timestamp: 2019-05-10T12:07:36+02:00
sign-key-sha3-384: UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO

AcLBUgQAAQoABgUCXNVNaAAAT0gQADiWK/pBSjDonvjiHNDiUYEMDfiu+WAuLo/k2SaWkoY0EgDy
Jk48BiIqsVcPQGOUIlFste+q7iqXhMO9PnzbdRUeEx5CtNioIzfoDmtuKUN4fweJ/J8v947oKmHg
DCdgf9dBDWU9mZGFa/FZqB5H+NefSV0dZo70hjsFbF60+zJN0RoQ9jk8zOA3LjF43u2qRak/yeie
4RxHDSGiKgvVouMoO8yla5897r9938Rv/zdN45LaXxhuEZ0HbFhnRYCmUw3ctgq0zkjkWCICFsLZ
bYT8tc+ZbcvyQoIND8r9QnV3DDw15fl037QBEhWH428bll3Pj2L0oP6ujGRVjKGcQND5KU3pga5l
96XFUPtWES8CJo7ekGHrmYihy6ThBmIrvaCCj5Pui2abUrBXb1aIhXHJzRRH+xN//68uaoYAjCTD
oa9BVIGHOGpIPO9ZpFE35SXiV8ktYfe3hbjPxDKYKqwMmwGwdQngcQ9NnOLvybv+uEfsF2eXFi1E
TlfGVlrz1fvafv98+mm5+GASPV3XOsL/HAhw0DxadCFOQOk+ALBypFDJLlocyb3PSyyr3RdvbnBH
NJJaotr3y52Cniv1XfM8sX2QWS/xfBNomEuySmjriKKapoFksigxPJm5TNjGQBNsfMDej+b5qdf6
SUH5uDA3J2ll8Oau8J2GSmB4Xt3C
`

const wrongSignatureSerialReq = `type: serial-request
brand-id: system
device-key:
    AcbATQRWhcGAAQgAzuDA7nxtfh77/XeX0UoIa8x0ILAtd4vEKRHXUmuf5LEUv0s7yQqtXjPQl5Rj
    Red4ssWFPFmanvgXjZMVmRfiTBW7VK86eG8E35TyIySWeT4dYqPzcEHLA4vPvEp0vS7IV0rr+tS5
    6NttX/oQmh/BSvBQruwIxpIJK0JhjSIzl6fO9RLFJe0eJCvpWPSSBiFeJAVeCfAIyrr4acANtyf5
    jkmwru1M+EpPj1VFN9yzPdspinnJW07w3RX5uqew6t325cTozKsrpV3OsK0QKQ7rttQpcKarY8rT
    QpfkVoSBFbgV6SlbragpaWWd0YTE4+YtXZ2OD0l8ipTyRDeE9lNotwARAQAB
model: alder
request-id: REQID
serial: A123456L
sign-key-sha3-384: majNNh3Nbgg0CozIjfLrvvEWG830dZmm76gJHnyyrW4g7udYbfVgt0WO15ayuN5l

ACLAUgQAAQoABgUCXOPNuwAA9GUIAE9pYXStlAMhsFu21eq0jhpTzoXLBQ99o9r2XL06IcYc5z6+
yKIO8fynEVjRpfoa4LXPX/XVqADNFaMp+AYuwhQLyUzSg7eHZ/GOwI5oCVyHpMVNVesepMVnVoFi
esewraFvituhKNCGMey+XB6K4Nf9eHTy7SKOvYHwdHSaAmwTu45TUe8ziVa6odkuNqVB41/MWJ4m
gAM0VjpgwupwUK6oPpzxUtapJoo4aKiWtfsVcHjc1c4RNyE+GAgNjDfUf0EKGa1IGl02k2ViR26l
jyc5jpGvwkqzKmk2V1kLr5f+zsDlZyWXdjJHTHZL62qfTCGzMwNJRvQ4trTsZJyxE3k=
`

const wrongSignSerial = `type: serial
authority-id: system
revision: 1
brand-id: system
model: alder
serial: A123456L
device-key:
    AcbATQRWhcGAAQgAzuDA7nxtfh77/XeX0UoIa8x0ILAtd4vEKRHXUmuf5LEUv0s7yQqtXjPQl5Rj
    Red4ssWFPFmanvgXjZMVmRfiTBW7VK86eG8E35TyIySWeT4dYqPzcEHLA4vPvEp0vS7IV0rr+tS5
    6NttX/oQmh/BSvBQruwIxpIJK0JhjSIzl6fO9RLFJe0eJCvpWPSSBiFeJAVeCfAIyrr4acANtyf5
    jkmwru1M+EpPj1VFN9yzPdspinnJW07w3RX5uqew6t325cTozKsrpV3OsK0QKQ7rttQpcKarY8rT
    QpfkVoSBFbgV6SlbragpaWWd0YTE4+YtXZ2OD0l8ipTyRDeE9lNotwARAQAB
device-key-sha3-384: majNNh3Nbgg0CozIjfLrvvEWG830dZmm76gJHnyyrW4g7udYbfVgt0WO15ayuN5l
timestamp: 2019-05-21T14:25:47+02:00
sign-key-sha3-384: UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMO

acLBUgQAAQoABgUCXOPuSwAAh1MQAE3y2iWCVuj+Eg9t3VFiX5TDwmirjbzukGhm0T10b6mRIiN1
i+jClRwPo7+j5M1595dUen3nfqt/iz4Uf2n4b0OcGPAQxu64vrayVz59Zja3nO1FUdWXVlN276U9
6gArefbz7QtkqlICI4RQglqE13MCB9gNZWs1T0/dKyhdzEwtic9VVVL0ol1b5Sc+H2zOzv9cqRBv
TtR5VorbhGeMU2kKJjHK2hynVrOVMfkQ9PPuT1l1l8J9bCUFEDenhB389VVHyFFmh8Rkr0AVYNbf
LuH+XwOhv6bQ2DX6+vqBVj9OMZVd6iz9OeJ0pHN6aI1i5LfTSDuFKtTjDfUPOIC9hiS3SgxIoBvG
du6UaI4WIzIbaqMMQNjrPGR9pPJeCGXJBVvpBdKTRIaosFcKyMLrfi4iWbiIF9C1nitAC74wyQu5
QT/LjYDslGH7k9t0inW2hhBp6Y2tMD2DwC7IWidumZt2/YIo6RS+jZaeAhTRFbiLSp0JScIXv2cz
ElqoeUAz268SxYvsmtDcuyGoW8gfmG2hWlq2FitDyRq66O+aOZPSg7eEu7ZUy6BMv4Wyi897ddLX
HYr/WkB1fEJDBZ9VfBfG0efZBP05X95G1nVAQ0WYwiusp/tMHSQkUBX3bGxPGIXUhs5+OdwtsbEd
FyfuoDecWoGB8Fj4S/iVupegchyg
`

const wrongSignKey = `type: serial
authority-id: system
revision: 1
brand-id: system
model: alder
serial: A123456L
device-key:
    AcbATQRWhcGAAQgAzuDA7nxtfh77/XeX0UoIa8x0ILAtd4vEKRHXUmuf5LEUv0s7yQqtXjPQl5Rj
    Red4ssWFPFmanvgXjZMVmRfiTBW7VK86eG8E35TyIySWeT4dYqPzcEHLA4vPvEp0vS7IV0rr+tS5
    6NttX/oQmh/BSvBQruwIxpIJK0JhjSIzl6fO9RLFJe0eJCvpWPSSBiFeJAVeCfAIyrr4acANtyf5
    jkmwru1M+EpPj1VFN9yzPdspinnJW07w3RX5uqew6t325cTozKsrpV3OsK0QKQ7rttQpcKarY8rT
    QpfkVoSBFbgV6SlbragpaWWd0YTE4+YtXZ2OD0l8ipTyRDeE9lNotwARAQAB
device-key-sha3-384: majNNh3Nbgg0CozIjfLrvvEWG830dZmm76gJHnyyrW4g7udYbfVgt0WO15ayuN5l
timestamp: 2019-05-23T17:40:35+02:00
sign-key-sha3-384: UytTqTvREVhx0tSfYC6KkFHmLWllIIZbQ3NsEG7OARrWuaXSRJyey0vjIQkTEvMX

AcLBUgQAAQoABgUCXOa+8wAA2jsQAILmpL+ssEIuOusv4aD8FdQffTcUPhEXF+msR+b0flIn6L73
KqAiKYbtJVF0GI5qKB7oxJjJh62jmU3Vh9A570OZn3BLytogQk/EVoS5KnVVtdU1bzuMSdPcdfgj
qODPU4rchLSLe7VviRGTwfjNXkHn5HjEltCW2TQgNQm5KendddCb5oh6TPTDHRLLynx1he9uqMri
B02BMJSeqg9o07W6ctKZanS/iT7/apQ9uFakELnSFfFX1GLiMiwAYyrBa/crMe2LHTseJtZrQs+l
Ne5yo31W/5toEln+XZSukGxVO5MrSwKMgF9Y3BkWv6AQ6nqu0CxUBOZAlVuH+PKXtytos7n67ojF
MxufevNPq5Rpku4rrT+Etxh+b21hCBkqGeap8w7/OAE8+jPAtsaMzWFa3r5dgwLJ01CmRS1asUeE
N74lGpSGgoZ+F5pQ8d1zBSbHsApCUCifMgk1PFqCNiFmBBwjw/GsjkNCDs3rVx2NjFSV9tktWREI
FiVth2u6BqsWp+HNti/PdZprEU4+TbTxOSNARhJcn5DFZ5S4+6QRdMDKFVn22JmsZNPauoFCQylp
SvDgHhOkD52ud7larb8PsXNNY9BL+1apfARJkVEQM52zxzkIDFWfg0kSQluU+0p0qRDZzMtYzJqj
cvxVzSrgJm1PB96l9R3WCYmc5lAn`

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

func generateSerialRequestAssertionRemodeling(model, originalModel, serial, body string) ([]byte, error) {
	privateKey, _ := generatePrivateKey()
	encodedPubKey, _ := asserts.EncodePublicKey(privateKey.PublicKey())

	headers := map[string]interface{}{
		"brand-id":   "mybrand",
		"device-key": string(encodedPubKey),
		"request-id": "REQID",
		"model":      model,

		"original-brand-id": "system",
		"original-model":    originalModel,
		"original-serial":   serial,
	}

	if serial != "" {
		headers["serial"] = serial
	}

	sreq, err := asserts.SignWithoutAuthority(asserts.SerialRequestType, headers, []byte(body), privateKey)
	if err != nil {
		return nil, err
	}

	assertions := asserts.Encode(sreq)
	return assertions, nil
}

func (s *SignSuite) TestRemodeling(c *check.C) {
	// For the remodelling we need to send serial-request assertion + new model assertion + old serail assertion
	// First, get the serial assertion for the original model
	serialReq, err := generateSerialRequestAssertion("alder", "A123456L", "")
	c.Assert(err, check.IsNil)
	w := sendRequest("POST", "/v1/serial", bytes.NewReader(serialReq), "ValidAPIKey", c)
	c.Assert(w.Code, check.Equals, 200)
	c.Assert(w.Body, check.NotNil)
	serialAssertions := w.Body.String()
	serialReq, err = generateSerialRequestAssertionRemodeling("alder-mybrand", "alder", "A123456L", "")
	c.Assert(err, check.IsNil)

	assertionsOK := append(serialReq, []byte("\n"+newModelAssertion)...)
	assertionsOK = append(assertionsOK, []byte("\n"+serialAssertions)...)

	assertionsOnlyModel := append(serialReq, []byte("\n"+newModelAssertion)...)
	assertionsOnlySerial := append(serialReq, []byte("\n"+serialAssertions)...)

	assertionsWrong, err := generateSerialRequestAssertionRemodeling("alder-mybrand-x", "alder", "A123456L", "")
	c.Assert(err, check.IsNil)
	assertionsWrong = append(assertionsWrong, []byte("\n"+newModelAssertion)...)
	assertionsWrong = append(assertionsWrong, []byte("\n"+serialAssertions)...)

	wrongModel := append(serialReq, []byte("\n"+modelAssertion)...)
	wrongModel = append(wrongModel, []byte("\n"+serialAssertions)...)

	assertionsBad := append(assertionsOK, []byte("\nXXX")...)

	assertionsWrongSerial := append(serialReq, []byte("\n"+newModelAssertion)...)
	assertionsWrongSerial = append(assertionsWrongSerial, []byte("\n"+wrongSerial)...)

	assertionsBedSerialNumber, err := generateSerialRequestAssertionRemodeling("alder-mybrand", "alder", "XXX", "")
	c.Assert(err, check.IsNil)
	assertionsBedSerialNumber = append(assertionsBedSerialNumber, []byte("\n"+newModelAssertion)...)
	assertionsBedSerialNumber = append(assertionsBedSerialNumber, []byte("\n"+serialAssertions)...)

	assertionsWrongModel, err := generateSerialRequestAssertionRemodeling("alder-mybrand", "alder", "abc1234X", "")
	c.Assert(err, check.IsNil)
	assertionsWrongModel = append(assertionsWrongModel, []byte("\n"+newModelAssertion)...)
	assertionsWrongModel = append(assertionsWrongModel, []byte("\n"+serialAssertions)...)

	assertionsWrongDeviceKey := append(serialReq, []byte("\n"+newModelAssertion)...)
	assertionsWrongDeviceKey = append(assertionsWrongDeviceKey, []byte("\n"+wrongDeviceKeySerial)...)

	assertionWrongSerialNumber, err := generateSerialRequestAssertionRemodeling("alder-mybrand", "alder", "abc1234", "")
	c.Assert(err, check.IsNil)

	assertionWrongSerialNumber = append(assertionWrongSerialNumber, []byte("\n"+newModelAssertion)...)
	assertionWrongSerialNumber = append(assertionWrongSerialNumber, []byte("\n"+serialAssertions)...)

	assertionsWrongBrand := append(serialReq, []byte("\n"+newModelAssertion)...)
	assertionsWrongBrand = append(assertionsWrongBrand, []byte("\n"+wrongBrandSerial)...)

	assertionsWrongSignature := append([]byte(serialReq), []byte("\n"+newModelAssertion)...)
	assertionsWrongSignature = append(assertionsWrongSignature, []byte("\n"+wrongSignSerial)...)

	assertionsWrongSignKey := append([]byte(serialReq), []byte("\n"+newModelAssertion)...)
	assertionsWrongSignKey = append(assertionsWrongSignKey, []byte("\n"+wrongSignKey)...)

	tests := []SuiteTest{
		{false, "POST", "/v1/serial", assertionsOK, 200, asserts.MediaType, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertionsOK, 400, response.JSONHeader, "NoModelForApiKey"},
		{false, "POST", "/v1/serial", serialReq, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertionsOnlyModel, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertionsOnlySerial, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertionsWrong, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", wrongModel, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertionsBad, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertionsWrongSerial, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertionsBedSerialNumber, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertionsWrongModel, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertionsWrongDeviceKey, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertionWrongSerialNumber, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertionsWrongBrand, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertionsWrongSignature, 400, response.JSONHeader, "ValidAPIKey"},
		{false, "POST", "/v1/serial", assertionsWrongSignKey, 400, response.JSONHeader, "ValidAPIKey"},
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
