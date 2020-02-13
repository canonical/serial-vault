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

package signinglog_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"time"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service"
	"github.com/CanonicalLtd/serial-vault/service/signinglog"
	check "gopkg.in/check.v1"
)

func (s *SigningLogSuite) TestAPISigningLogHandler(c *check.C) {
	log1 := datastore.SigningLog{ID: 1, Make: "system", Model: "alder", SerialNumber: "abcd1234", Fingerprint: "aaaabbbbccccdddd", Revision: 1, Created: time.Now()}
	l1, _ := json.Marshal(log1)

	tests := []SigningLogTest{
		{"GET", "/api/signinglog", nil, 400, "application/json; charset=UTF-8", 0, false, false, 0},
		{"GET", "/api/signinglog", nil, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 4},
		{"GET", "/api/signinglog", nil, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
		{"GET", "/api/signinglog", nil, 400, "application/json; charset=UTF-8", 0, true, false, 0},
		{"POST", "/api/signinglog", l1, 400, "application/json; charset=UTF-8", 0, false, false, 0},
		{"POST", "/api/signinglog", l1, 200, "application/json; charset=UTF-8", datastore.Admin, true, true, 0},
		{"POST", "/api/signinglog", l1, 400, "application/json; charset=UTF-8", datastore.Standard, true, false, 0},
		{"POST", "/api/signinglog", l1, 400, "application/json; charset=UTF-8", 0, true, false, 0},
	}

	for _, t := range tests {
		if t.EnableAuth {
			datastore.Environ.Config.EnableUserAuth = true
		}

		w := sendAdminAPIRequest(t.Method, t.URL, bytes.NewReader(t.Data), t.Permissions, c)
		c.Assert(w.Code, check.Equals, t.Code)
		c.Assert(w.Header().Get("Content-Type"), check.Equals, t.Type)

		result, err := parseListResponse(w)
		c.Assert(err, check.IsNil)
		c.Assert(result.Success, check.Equals, t.Success)
		c.Assert(len(result.SigningLog), check.Equals, t.List)

		datastore.Environ.Config.EnableUserAuth = false
	}
}

func sendAdminAPIRequest(method, url string, data io.Reader, permissions int, c *check.C) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	switch permissions {
	case datastore.Admin:
		r.Header.Set("user", "sv")
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

func (s *SigningLogSuite) TestGetSigningLogParams(c *check.C) {
	tests := []struct {
		name string
		url  string
		want *datastore.SigningLogParams
	}{
		{
			name: "case 1",
			url:  `/ping?offset=12&serialnumber=R12300&filter=foo,bar`,
			want: &datastore.SigningLogParams{
				Offset:       12,
				Serialnumber: "R12300",
				Filter:       []string{"foo", "bar"},
				Limit:        datastore.ListSigningLogDefaultLimit,
			},
		},
		{
			name: "case 2",
			url:  `/ping?offset=xxx`,
			want: &datastore.SigningLogParams{
				Offset: 0,
				Limit:  datastore.ListSigningLogDefaultLimit,
			},
		},
		{
			name: "case 3",
			url:  `/ping`,
			want: &datastore.SigningLogParams{
				Limit: datastore.ListSigningLogDefaultLimit,
			},
		},
		{
			name: "case 4",
			url:  `/ping?filter&serialnumber&filter&all=xxx`,
			want: &datastore.SigningLogParams{
				Limit: datastore.ListSigningLogDefaultLimit,
			},
		},
		{
			name: "case 5",
			url:  `/ping?all=true`,
			want: &datastore.SigningLogParams{
				Limit: 0,
			},
		},
	}
	for _, tt := range tests {
		r, _ := http.NewRequest("GET", tt.url, nil)

		if got := signinglog.GetSigningLogParams(r); !reflect.DeepEqual(got, tt.want) {
			c.Errorf("getSigningLogParams() = %#v, want %#v", got, tt.want)
		}
	}
}
