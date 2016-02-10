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
	"log"
	"net/http"
	"time"
)

// Logger Handle logging for the web service
func Logger(start time.Time, r *http.Request) {
	log.Printf(
		"%s\t%s\t%s",
		r.Method,
		r.RequestURI,
		time.Since(start),
	)
}

// Environ contains the parsed config file settings.
var Environ *Env

// Middleware to pre-process web service requests
func Middleware(inner http.Handler, env *Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if Environ == nil {
			Environ = env
		}

		// Log the request
		Logger(start, r)

		inner.ServeHTTP(w, r)
	})
}
