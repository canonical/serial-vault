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
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/response"
	check "gopkg.in/check.v1"
)

func TestLogSuite(t *testing.T) { check.TestingT(t) }

type LogSuite struct{}

type SuiteTest struct {
	Method   string
	URL      string
	Data     []byte
	Path     string
	Filename string
	Code     int
	Type     string
	Message  string
}

var _ = check.Suite(&LogSuite{})

func (s *LogSuite) SetUpTest(c *check.C) {
	// Mock the database
	config := config.Settings{Driver: "sqlite3", KeyStoreType: "filesystem", KeyStorePath: "../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)
}

func sendSigningRequest(method, url string, data io.Reader, path, filename string, c *check.C) *httptest.ResponseRecorder {
	var r *http.Request
	w := httptest.NewRecorder()

	if len(path) != 0 {
		ctype, body, err := createFile(path, filename, c)
		c.Assert(err, check.IsNil)
		r, _ = http.NewRequest(method, url, body)
		r.Header.Add("Content-Type", ctype)
	} else {
		r, _ = http.NewRequest(method, url, data)
	}

	service.SigningRouter().ServeHTTP(w, r)

	return w
}

func (s *LogSuite) TestLogHandler(c *check.C) {
	tests := []SuiteTest{
		{"GET", "/testlog", nil, "", "", 200, "text/html; charset=utf-8", ""},
		{"POST", "/testlog", nil, "", "", 400, response.JSONHeader, ""},
		{"POST", "/testlog", []byte(""), "", "", 400, response.JSONHeader, ""},
		{"POST", "/testlog", []byte(""), "../../keystore/example_report.xml", "logfile", 201, response.JSONHeader, ""},
	}

	for _, t := range tests {
		w := sendSigningRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Path, t.Filename, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)
	}
}

func createFile(path, filename string, c *check.C) (string, *bytes.Buffer, error) {
	file, err := os.Open(path)
	c.Assert(err, check.IsNil)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(filename, filepath.Base(path))
	if err != nil {
		return "", nil, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", nil, err
	}

	if err = writer.Close(); err != nil {
		return "", nil, err
	}
	return writer.FormDataContentType(), body, nil
}
