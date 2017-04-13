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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/snapcore/snapd/asserts"
)

func TestUserIndexHandler(t *testing.T) {

	userIndexTemplate = "../static/app_user.html"

	config := ConfigSettings{Title: "Site Title", Logo: "/url"}
	Environ = &Env{Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	http.HandlerFunc(UserIndexHandler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got: %d", http.StatusOK, w.Code)
	}
}

func TestUserIndexHandlerInvalidTemplate(t *testing.T) {

	userIndexTemplate = "../static/does_not_exist.html"

	config := ConfigSettings{Title: "Site Title", Logo: "/url"}
	Environ = &Env{Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	http.HandlerFunc(UserIndexHandler).ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got: %d", http.StatusInternalServerError, w.Code)
	}
}

func generateSystemUserRequest() string {

	request := SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 1, Since: "20170324T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInvalidModel() string {

	request := SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 99, Since: "20170324T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInactiveModel() string {

	request := SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 2, Since: "20170324T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInvalidSince() string {

	request := SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 1, Since: "2024T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInvalidAssertion() string {

	request := SystemUserRequest{Email: "test", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 1, Since: "20170324T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func TestSystemUserAssertionHandler(t *testing.T) {
	type TestType struct {
		data       string
		statusCode int
		expected   bool
	}

	tests := []TestType{
		{generateSystemUserRequest(), 200, true},
		{"", 400, false},
		{"<invalid\\", 400, false},
		{generateSystemUserRequestInvalidModel(), 200, false},
		{generateSystemUserRequestInactiveModel(), 200, false},
		{generateSystemUserRequestInvalidAssertion(), 200, false},
		{generateSystemUserRequestInvalidSince(), 200, true},
	}

	for _, test := range tests {
		statusCode, result, message := sendSystemUserAssertion(test.data, t)
		if statusCode != test.statusCode {
			t.Errorf("Unexpected status code from request '%v': %d", test.data, statusCode)
		}
		if result != test.expected {
			t.Errorf("Unexpected result from request '%v': %s", test.data, message)
		}
	}
}

func sendSystemUserAssertion(request string, t *testing.T) (int, bool, string) {
	// Mock the database
	config := ConfigSettings{KeyStoreType: "filesystem", KeyStorePath: "../keystore", KeyStoreSecret: "secret code to encrypt the auth-key hash"}
	Environ = &Env{DB: &mockDB{}, Config: config}
	Environ.KeypairDB, _ = GetKeyStore(config)

	// Mock the retrieval of the assertion from the store (using a fixed assertion)
	FetchAssertionFromStore = func(modelType *asserts.AssertionType, headers []string) (asserts.Assertion, error) {
		headersMap := map[string]interface{}{
			"type":              "account",
			"authority-id":      "canonical",
			"account-id":        "canonical",
			"display-name":      "Canonical",
			"username":          "Canonical",
			"timestamp":         "2016-04-01T00:00:00.0Z",
			"validation":        "certified",
			"sign-key-sha3-384": "-CvQKAwRQ5h3Ffn10FILJoEZUXOv6km9FwA80-Rcj-f-6jadQ89VRswHNiEB9Lxk",
		}

		signature := []byte(`AcLDXAQAAQoABgUCV7UYzwAKCRDUpVvql9g3IK7uH/4udqNOurx5WYVknzXdwekp0ovHCQJ0iBPw
TSFxEVr9faZSzb7eqJ1WicHsShf97PYS3ClRYAiluFsjRA8Y03kkSVJHjC+sIwGFubsnkmgflt6D
WEmYIl0UBmeaEDS8uY4Xvp9NsLTzNEj2kvzy/52gKaTc1ZSl5RDL9ppMav+0V9iBYpiDPBWH2rJ+
aDSD8Rkyygm0UscfAKyDKH4lrvZ0WkYyi1YVNPrjQ/AtBySh6Q4iJ3LifzKa9woIyAuJET/4/FPY
oirqHAfuvNod36yNQIyNqEc20AvTvZNH0PSsg4rq3DLjIPzv5KbJO9lhsasNJK1OdL6x8Yqrdsbk
ldZp4qkzfjV7VOMQKaadfcZPRaVVeJWOBnBiaukzkhoNlQi1sdCdkBB/AJHZF8QXw6c7vPDcfnCV
1lW7ddQ2p8IsJbT6LzpJu3GW/P4xhNgCjtCJ1AJm9a9RqLwQYgdLZwwDa9iCRtqTbRXBlfy3apps
1VjbQ3h5iCd0hNfwDBnGVm1rhLKHCD1DUdNE43oN2ZlE7XGyh0HFV6vKlpqoW3eoXCIxWu+HBY96
+LSl/jQgCkb0nxYyzEYK4Reb31D0mYw1Nji5W+MIF5E09+DYZoOT0UvR05YMwMEOeSdI/hLWg/5P
k+GDK+/KopMmpd4D1+jjtF7ZvqDpmAV98jJGB2F88RyVb4gcjmFFyTi4Kv6vzz/oLpbm0qrizC0W
HLGDN/ymGA5sHzEgEx7U540vz/q9VX60FKqL2YZr/DcyY9GKX5kCG4sNqIIHbcJneZ4frM99oVDu
7Jv+DIx/Di6D1ULXol2XjxbbJLKHFtHksR97ceaFvcZwTogC61IYUBJCvvMoqdXAWMhEXCr0QfQ5
Xbi31XW2d4/lF/zWlAkRnGTzufIXFni7+nEuOK0SQEzO3/WaRedK1SGOOtTDjB8/3OJeW96AUYK5
oTIynkYkEyHWMNCXALg+WQW6L4/YO7aUjZ97zOWIugd7Xy63aT3r/EHafqaY2nacOhLfkeKZ830b
o/ezjoZQAxbh6ce7JnXRgE9ELxjdAhBTpGjmmmN2sYrJ7zP9bOgly0BnEPXGSQfFA+NNNw1FADx1
MUY8q9DBjmVtgqY+1KGTV5X8KvQCBMODZIf/XJPHdCRAHxMd8COypcwgL2vDIIXpOFbi1J/B0GF+
eklxk9wzBA8AecBMCwCzIRHDNpD1oa2we38bVFrOug6e/VId1k1jYFJjiLyLCDmV8IMYwEllHSXp
LQAdm3xZ7t4WnxYC8YSCk9mXf3CZg59SpmnV5Q5Z6A5Pl7Nc3sj7hcsMBZEsOMPzNC9dPsBnZvjs
WpPUffJzEdhHBFhvYMuD4Vqj6ejUv9l3oTrjQWVC`)

		return asserts.Assemble(headersMap, nil, nil, signature)
	}

	// Submit the serial-request assertion for signing
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/assertions", bytes.NewBufferString(request))
	SystemUserRouter(Environ).ServeHTTP(w, r)

	// Check the JSON response
	result := SystemUserResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding the version response: %v", err)
		t.Log(w.Body.String())
	}

	return w.Code, result.Success, result.ErrorMessage
}
