// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
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

package usso

import (
	"net/http"
	"time"

	"github.com/juju/httprequest"
	"github.com/juju/loggo"
	"golang.org/x/net/context"
)

var logger = loggo.GetLogger("serial-vault.usso.utils")

const (
	// identityMacaroonDuration is the length of time for which an
	// identity macaroon is valid.
	identityMacaroonDuration = 28 * 24 * time.Hour
)

// URL creates a URL addressed to the given path within the given path
// and adds the given waitid (when specified).
func URL(ctx Context, path, waitid string) string {
	callback := ctx.URL(path)
	if waitid != "" {
		callback += "?waitid=" + waitid
	}
	return callback
}

// WaitID gets the wait ID from the given request using the standard form value.
func WaitID(req *http.Request) string {
	return req.Form.Get("waitid")
}

// RequestParams creates an httprequest.Params object from the given fields.
func RequestParams(ctx context.Context, w http.ResponseWriter, req *http.Request) httprequest.Params {
	return httprequest.Params{
		Response: w,
		Request:  req,
		Context:  ctx,
	}
}

// LoginUser completes a successful login for the specified user. A new
// identity macaroon is generated for the user and an appropriate message
// will be returned for the login request.
func LoginUser(ctx RequestContext, waitid string, w http.ResponseWriter, u *User) {
	if !ctx.LoginSuccess(waitid, Username(u.Username), time.Now().Add(identityMacaroonDuration)) {
		return
	}
	t := ctx.Template("login")
	if t == nil {
		return
	}
	w.Header().Set("Content-Type", "text/html;charset=utf-8")
	if err := t.Execute(w, u); err != nil {
		logger.Errorf("error processing login template: %s", err)
	}
}
