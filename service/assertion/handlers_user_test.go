// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
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

package assertion_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/assertion"
	check "gopkg.in/check.v1"
)

type TestType struct {
	data       string
	statusCode int
	expected   bool
}

func (s *AssertionSuite) TestSystemUserAssertionHandler(c *check.C) {
	tests := []TestType{
		{generateSystemUserRequest(), 200, true},
		{"", 400, false},
		{"<invalid\\", 400, false},
		{generateSystemUserRequestInvalidModel(), 400, false},
		{generateSystemUserRequestInactiveModel(), 400, false},
		{generateSystemUserRequestInvalidAssertion(), 400, false},
		{generateSystemUserRequestInvalidSince(), 200, true},
	}

	for _, test := range tests {
		statusCode, result, _ := sendSystemUserAssertion(test.data, c)
		c.Assert(statusCode, check.Equals, test.statusCode)
		c.Assert(result, check.Equals, test.expected)
	}
}

func sendSystemUserAssertion(request string, c *check.C) (int, bool, string) {
	// Submit the serial-request assertion for signing
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/v1/assertions", bytes.NewBufferString(request))
	service.AdminRouter().ServeHTTP(w, r)

	// Check the JSON response
	result := assertion.SystemUserResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	c.Assert(err, check.IsNil)

	return w.Code, result.Success, result.ErrorMessage
}

func generateSystemUserRequest() string {
	request := assertion.SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 1, Since: "2017-03-24T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInvalidModel() string {
	request := assertion.SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 99, Since: "2017-03-24T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInactiveModel() string {
	request := assertion.SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 2, Since: "2017-03-24T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInvalidSince() string {
	request := assertion.SystemUserRequest{Email: "test@example.com", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 1, Since: "2024T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}

func generateSystemUserRequestInvalidAssertion() string {
	request := assertion.SystemUserRequest{Email: "test", Name: "John Doe", Username: "jdoe", Password: "super", ModelID: 1, Since: "2017-03-24T12:34:00Z"}
	req, _ := json.Marshal(request)

	return string(req)
}
