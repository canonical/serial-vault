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
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"

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
				<label for="logfile">Filename:</label><br />
				<input type="file" name="logfile" accept="text/xml">
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
	file, handle, err := r.FormFile("logfile")
	if err != nil {
		formatUploadResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	defer file.Close()

	saveFile(w, file, handle)
}

func formatUploadResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", response.JSONHeader)
	w.WriteHeader(code)
	fmt.Fprint(w, message)
}

func saveFile(w http.ResponseWriter, file multipart.File, handle *multipart.FileHeader) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		formatUploadResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Set up the path for the file
	path := os.Getenv(paramsEnvVar)
	if len(path) == 0 {
		path = "."
	}
	path = filepath.Join(path, "/files")

	// Attempt to create the path, if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = os.Mkdir(path, 0755); err != nil {
			formatUploadResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// Generate a filename
	filename := fmt.Sprintf("%s/%d_%s", path, time.Now().UTC().Unix(), handle.Filename)

	// Save the file
	err = ioutil.WriteFile(filename, data, 0666)
	if err != nil {
		formatUploadResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	formatUploadResponse(w, http.StatusCreated, "File uploaded successfully")
}
