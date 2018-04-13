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
	Method    string
	URL       string
	ModelName string
	Path      string
	WithFile  bool
	Code      int
	Type      string
	Message   string
}

var _ = check.Suite(&LogSuite{})

func (s *LogSuite) SetUpTest(c *check.C) {
	// Mock the database
	config := config.Settings{Driver: "sqlite3", KeyStoreType: "filesystem", KeyStorePath: "../keystore", JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.OpenKeyStore(config)
}

func sendSigningRequest(method, url string, modelName, path string, withFile bool, c *check.C) *httptest.ResponseRecorder {
	var r *http.Request
	w := httptest.NewRecorder()

	if withFile {
		ctype, body, err := createFile(modelName, path, c)
		c.Assert(err, check.IsNil)
		r, _ = http.NewRequest(method, url, body)
		r.Header.Add("Content-Type", ctype)
	} else {
		r, _ = http.NewRequest(method, url, nil)
	}

	service.SigningRouter().ServeHTTP(w, r)

	return w
}

func (s *LogSuite) TestLogHandler(c *check.C) {
	tests := []SuiteTest{
		{"GET", "/testlog", "alder", "", false, 200, "text/html; charset=utf-8", ""},
		{"POST", "/testlog", "", "", true, 400, response.JSONHeader, "The 'brand' and 'model' must be supplied"},
		{"POST", "/testlog", "invalid", "", true, 400, response.JSONHeader, "The model does not exist"},
		{"POST", "/testlog", "alder", "", true, 400, response.JSONHeader, "http: no such file"},
		{"POST", "/testlog", "alder", "../../keystore/empty_report.xml", true, 400, response.JSONHeader, "The file cannot be empty"},
		{"POST", "/testlog", "alder", "../../keystore/example_report.xml", true, 201, response.JSONHeader, "File uploaded successfully"},
	}

	for _, t := range tests {
		w := sendSigningRequest(t.Method, t.URL, t.ModelName, t.Path, t.WithFile, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)
		if t.Method != "GET" {
			c.Assert(w.Body.String(), check.Equals, t.Message)
		}
	}
}

func createFile(modelName, path string, c *check.C) (string, *bytes.Buffer, error) {
	var file *os.File
	var err error
	if len(path) > 0 {
		file, err = os.Open(path)
		c.Assert(err, check.IsNil)
		defer file.Close()
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if len(path) > 0 {
		part, err := writer.CreateFormFile("logfile", filepath.Base(path))
		if err != nil {
			return "", nil, err
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return "", nil, err
		}
	}

	if err := writer.WriteField("brand", "system"); err != nil {
		return "", nil, err
	}
	if err := writer.WriteField("model", modelName); err != nil {
		return "", nil, err
	}

	if err := writer.Close(); err != nil {
		return "", nil, err
	}
	return writer.FormDataContentType(), body, nil
}
