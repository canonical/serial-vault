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

package app

import (
	"net/http"
	"strings"
	"text/template"

	"github.com/CanonicalLtd/serial-vault/service/log"

	"github.com/CanonicalLtd/serial-vault/datastore"
)

// IndexTemplate is the path to the HTML template
var IndexTemplate = "/static/app.html"

// Page is the page details for the web application
type Page struct {
	Title string
	Logo  string
}

// Index is the front page of the web application
func Index(w http.ResponseWriter, r *http.Request) {
	page := Page{Title: datastore.Environ.Config.Title, Logo: datastore.Environ.Config.Logo}

	path := []string{datastore.Environ.Config.DocRoot, IndexTemplate}
	t, err := template.ParseFiles(strings.Join(path, ""))
	if err != nil {
		log.Printf("Error loading the application template: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
