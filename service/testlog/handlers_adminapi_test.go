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

package testlog_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/CanonicalLtd/serial-vault/service/testlog"
	check "gopkg.in/check.v1"
)

type SyncTest struct {
	Method      string
	URL         string
	Data        []byte
	Code        int
	Type        string
	Permissions int
	EnableAuth  bool
	Success     bool
	MockError   bool
	Count       int
}

const exampleFile = "PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4NCjx0ZXN0X3JlcG9ydD4NCiAgICA8dXV0cz4NCiAgICAgICAgPHV1dD4NCiAgICAgICAgICAgIDxzdW1tYXJ5Pg0KICAgICAgICAgICAgICAgIDxwYXJ0X251bWJlcj44NjAtMDAwMTQ8L3BhcnRfbnVtYmVyPg0KICAgICAgICAgICAgICAgIDxzZXJpYWxfbnVtYmVyPmU0YmY3Y2ViLWY3YWYtNDQyZS1iNzM0LTU0MzJlZjMzMDZiNTwvc2VyaWFsX251bWJlcj4NCiAgICAgICAgICAgICAgICA8b3BlcmF0aW9uPlZhbGlkYXRpb24gVGVzdDwvb3BlcmF0aW9uPg0KICAgICAgICAgICAgICAgIDxzdGFydGVkX2F0PjIwMTgtMDQtMDlUMTc6NDQ6MTYrMDI6MDA8L3N0YXJ0ZWRfYXQ+DQogICAgICAgICAgICAgICAgPGVuZGVkX2F0PjIwMTgtMDQtMDlUMTc6NDQ6MTYrMDI6MDA8L2VuZGVkX2F0Pg0KICAgICAgICAgICAgICAgIDxzdGF0dXM+RmFpbGVkPC9zdGF0dXM+DQogICAgICAgICAgICA8L3N1bW1hcnk+DQogICAgICAgICAgICA8dGVzdHM+DQogICAgICAgICAgICAgICAgPHRlc3Q+DQogICAgICAgICAgICAgICAgICAgIDxuYW1lPmZhY3RvcnlfY3B1L2lNWDZVTEw8L25hbWU+DQogICAgICAgICAgICAgICAgICAgIDxzdGF0dXM+ZmFpbGVkPC9zdGF0dXM+DQogICAgICAgICAgICAgICAgPC90ZXN0Pg0KICAgICAgICAgICAgICAgIDx0ZXN0Pg0KICAgICAgICAgICAgICAgICAgICA8bmFtZT5mYWN0b3J5X2V0aGVybmV0L2NhcmQtZGV0ZWN0PC9uYW1lPg0KICAgICAgICAgICAgICAgICAgICA8c3RhdHVzPmZhaWxlZDwvc3RhdHVzPg0KICAgICAgICAgICAgICAgIDwvdGVzdD4NCiAgICAgICAgICAgICAgICA8dGVzdD4NCiAgICAgICAgICAgICAgICAgICAgPG5hbWU+ZmFjdG9yeV9oZGQvZHJpdmUtY291bnQ8L25hbWU+DQogICAgICAgICAgICAgICAgICAgIDxzdGF0dXM+ZmFpbGVkPC9zdGF0dXM+DQogICAgICAgICAgICAgICAgPC90ZXN0Pg0KICAgICAgICAgICAgICAgIDx0ZXN0Pg0KICAgICAgICAgICAgICAgICAgICA8bmFtZT5mYWN0b3J5X1JBTS9zaXplPC9uYW1lPg0KICAgICAgICAgICAgICAgICAgICA8c3RhdHVzPnBhc3NlZDwvc3RhdHVzPg0KICAgICAgICAgICAgICAgIDwvdGVzdD4NCiAgICAgICAgICAgICAgICA8dGVzdD4NCiAgICAgICAgICAgICAgICAgICAgPG5hbWU+ZmFjdG9yeV9SVEM8L25hbWU+DQogICAgICAgICAgICAgICAgICAgIDxzdGF0dXM+ZmFpbGVkPC9zdGF0dXM+DQogICAgICAgICAgICAgICAgPC90ZXN0Pg0KICAgICAgICAgICAgICAgIDx0ZXN0Pg0KICAgICAgICAgICAgICAgICAgICA8bmFtZT5mYWN0b3J5X3VzYi91c2IyLXJvb3QtaHViLXByZXNlbnQ8L25hbWU+DQogICAgICAgICAgICAgICAgICAgIDxzdGF0dXM+cGFzc2VkPC9zdGF0dXM+DQogICAgICAgICAgICAgICAgPC90ZXN0Pg0KICAgICAgICAgICAgPC90ZXN0cz4NCjwvdXV0Pg0KPC91dXRzPg0KPC90ZXN0X3JlcG9ydD4="

func (s *LogSuite) TestAPISyncHandler(c *check.C) {
	t1 := datastore.TestLog{
		Brand: "system", Model: "alder",
		Filename: "example.xml", Data: exampleFile,
	}
	tLog1, err := json.Marshal(t1)
	c.Assert(err, check.IsNil)

	t2 := t1
	t2.Data = ""
	tLog2, err := json.Marshal(t2)
	c.Assert(err, check.IsNil)

	t3 := t1
	t3.Data = "bad"
	tLog3, err := json.Marshal(t3)
	c.Assert(err, check.IsNil)

	t4 := t1
	t4.Filename = ""
	t4.Data = ""
	tLog4, err := json.Marshal(t4)
	c.Assert(err, check.IsNil)

	tests := []SyncTest{
		{"POST", "/api/testlog", []byte("bad"), 400, response.JSONHeader, datastore.SyncUser, false, false, false, 0},
		{"POST", "/api/testlog", tLog2, 400, response.JSONHeader, datastore.SyncUser, false, false, false, 0},
		{"POST", "/api/testlog", tLog3, 400, response.JSONHeader, datastore.SyncUser, false, false, false, 0},
		{"POST", "/api/testlog", tLog4, 400, response.JSONHeader, datastore.SyncUser, false, false, false, 0},
		{"POST", "/api/testlog", tLog1, 400, response.JSONHeader, 0, false, false, false, 0},
		{"POST", "/api/testlog", tLog1, 400, response.JSONHeader, datastore.SyncUser, true, false, true, 0},
		{"POST", "/api/testlog", tLog1, 400, response.JSONHeader, datastore.SyncUser, true, false, true, 0},
		{"POST", "/api/testlog", tLog1, 200, response.JSONHeader, datastore.SyncUser, true, true, false, 0},
		{"POST", "/api/testlog", tLog1, 200, response.JSONHeader, datastore.SyncUser, true, true, false, 0},
		{"POST", "/api/testlog", tLog1, 400, response.JSONHeader, datastore.Standard, true, false, false, 0},
		{"POST", "/api/testlog", tLog1, 400, response.JSONHeader, 0, true, false, false, 0},
	}

	for _, t := range tests {
		if t.EnableAuth {
			datastore.Environ.Config.EnableUserAuth = true
		}
		if t.MockError {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}

		w := sendAdminAPIRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := response.ParseStandardResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)

		datastore.Environ.Config.EnableUserAuth = false
		datastore.Environ.DB = &datastore.MockDB{}
	}
}

func (s *LogSuite) TestAPIListHandler(c *check.C) {
	tests := []SyncTest{
		{"GET", "/api/testlog", nil, 400, response.JSONHeader, datastore.Standard, false, false, false, 0},
		{"GET", "/api/testlog", nil, 400, response.JSONHeader, datastore.SyncUser, false, false, true, 0},
		{"GET", "/api/testlog", nil, 200, response.JSONHeader, datastore.SyncUser, false, true, false, 2},
	}

	for _, t := range tests {
		if t.EnableAuth {
			datastore.Environ.Config.EnableUserAuth = true
		}
		if t.MockError {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}

		w := sendAdminAPIRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := parseListResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)
		c.Assert(len(result.TestLog), check.Equals, t.Count)

		datastore.Environ.Config.EnableUserAuth = false
		datastore.Environ.DB = &datastore.MockDB{}
	}
}

func (s *LogSuite) TestAPIUpdateLogHandler(c *check.C) {
	tests := []SyncTest{
		{"PUT", "/api/testlog/1", nil, 200, response.JSONHeader, datastore.SyncUser, false, true, false, 0},
		{"PUT", "/api/testlog/1", nil, 200, response.JSONHeader, datastore.SyncUser, true, true, false, 0},
		{"PUT", "/api/testlog/1", nil, 400, response.JSONHeader, datastore.SyncUser, true, false, true, 0},
		{"PUT", "/api/testlog/1", nil, 400, response.JSONHeader, datastore.Standard, true, false, false, 0},
		{"PUT", "/api/testlog/1", nil, 400, response.JSONHeader, 0, false, false, false, 0},
	}

	for _, t := range tests {
		if t.EnableAuth {
			datastore.Environ.Config.EnableUserAuth = true
		}
		if t.MockError {
			datastore.Environ.DB = &datastore.ErrorMockDB{}
		}

		w := sendAdminAPIRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		datastore.Environ.Config.EnableUserAuth = false
		datastore.Environ.DB = &datastore.MockDB{}
	}
}

func sendAdminAPIRequest(method, url string, data io.Reader, permissions int, c *check.C) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	switch permissions {
	case datastore.SyncUser:
		r.Header.Set("user", "sync")
		r.Header.Set("api-key", "ValidAPIKey")
	case datastore.Standard:
		r.Header.Set("user", "user1")
		r.Header.Set("api-key", "ValidAPIKey")
	default:
		break
	}

	service.AdminRouter().ServeHTTP(w, r)

	return w
}

func parseListResponse(w *httptest.ResponseRecorder) (testlog.ListResponse, error) {
	// Check the JSON response
	result := testlog.ListResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	return result, err
}
