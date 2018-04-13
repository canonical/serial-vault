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

package testlog

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/response"
)

const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<title>Test Log Upload</title>
		<style>
		body {font-family: Ubuntu,Roboto,sans-serif}
		h1   {margin-top: 0; padding: 10px 10px; background-color: #E95420; color: #fff; font-weight: normal}
		div { margin: 40px}
		fieldset { padding: 20px}
		</style>
	</head>
	<body>
		<h1>Serial Vault</h1>
		<div>
			<h2>Test Log Upload</h2>
			<form action="/testlog" method="POST" enctype="multipart/form-data">
				<fieldset>
				<label for="brand">Brand:</label><br />
				<input type="text" name="brand" /><br />
				<label for="model">Model:</label><br />
				<input type="text" name="model" /><br />
				<label for="logfile">Filename:</label><br />
				<input type="file" name="logfile" accept="text/xml" />
				<input type="Submit">
				</fieldset>
			</form>
		</div>
	</body>
</html>
`
const paramsEnvVar = "SNAP_DATA"

// Index is the form of the test log upload web application
func Index(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("testlog").Parse(tpl)
	if err != nil {
		log.Printf("Error loading the application template: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Submit is the POST request for caching a test log
func Submit(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(r.FormValue("brand")) == "" || strings.TrimSpace(r.FormValue("model")) == "" {
		formatUploadResponse(w, http.StatusBadRequest, "The 'brand' and 'model' must be supplied")
		return
	}

	// Check that the model exists
	if found := datastore.Environ.DB.CheckModelExists(r.FormValue("brand"), r.FormValue("model")); !found {
		formatUploadResponse(w, http.StatusBadRequest, "The model does not exist")
		return
	}

	// Get the file from the request
	filename, base64File, err := readFile(r)
	if err != nil {
		formatUploadResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Store the testlog
	t := datastore.TestLog{
		Brand: r.FormValue("brand"), Model: r.FormValue("model"),
		Filename: filename, Data: base64File,
	}
	err = datastore.Environ.DB.CreateTestLog(t)
	if err != nil {
		formatUploadResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	formatUploadResponse(w, http.StatusCreated, "File uploaded successfully")
}

func formatUploadResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", response.JSONHeader)
	w.WriteHeader(code)
	fmt.Fprint(w, message)
}

func readFile(r *http.Request) (string, string, error) {
	file, handle, err := r.FormFile("logfile")
	if err != nil {
		return "", "", err
	}
	if handle.Size == 0 {
		return "", "", errors.New("The file cannot be empty")
	}

	defer file.Close()

	// Read the file
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", "", err
	}

	// Encode the file for storage
	return handle.Filename, base64.StdEncoding.EncodeToString(data), nil
}
