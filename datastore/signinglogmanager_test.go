// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2018 Canonical Ltd
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

package datastore

import (
	"testing"

	check "gopkg.in/check.v1"
)

func TestSQL(t *testing.T) { check.TestingT(t) }

type sqlSuite struct{}

var _ = check.Suite(&sqlSuite{})

func (vs *sqlSuite) TestSigningLogSQLBuilder(c *check.C) {

	tests := []struct {
		username    string
		authorityID string
		params      *SigningLogParams
		wantSQL     string
		wantParams  []interface{}
	}{
		{
			authorityID: "admin",
			params:      &SigningLogParams{},
			wantSQL:     "SELECT *, count(*) OVER() AS total_count FROM signinglog s WHERE id < $1 AND make=$2 ORDER BY id DESC OFFSET 0",
			wantParams:  []interface{}{2147483647, "admin"},
		},
		{
			authorityID: "admin",
			params: &SigningLogParams{
				Offset: 150,
			},
			wantSQL:    "SELECT *, count(*) OVER() AS total_count FROM signinglog s WHERE id < $1 AND make=$2 ORDER BY id DESC OFFSET 150",
			wantParams: []interface{}{2147483647, "admin"},
		},
		{
			authorityID: "admin",
			params: &SigningLogParams{
				Offset: 250,
				Filter: []string{"foo", "bar"},
			},
			wantSQL:    "SELECT *, count(*) OVER() AS total_count FROM signinglog s WHERE id < $1 AND make=$2 AND model IN ($3,$4) ORDER BY id DESC OFFSET 250",
			wantParams: []interface{}{2147483647, "admin", "foo", "bar"},
		},
		{
			authorityID: "admin",
			params: &SigningLogParams{
				Limit:        123,
				Offset:       350,
				Serialnumber: "R1234567",
			},
			wantSQL:    "SELECT *, count(*) OVER() AS total_count FROM signinglog s WHERE id < $1 AND make=$2 AND serial_number LIKE $3 ORDER BY id DESC LIMIT 123 OFFSET 350",
			wantParams: []interface{}{2147483647, "admin", "R1234567%"},
		},
		{
			authorityID: "admin",
			params: &SigningLogParams{
				Offset:       350,
				Filter:       []string{"aaa"},
				Serialnumber: "000XXX12354",
			},
			wantSQL:    "SELECT *, count(*) OVER() AS total_count FROM signinglog s WHERE id < $1 AND make=$2 AND model IN ($3) AND serial_number LIKE $4 ORDER BY id DESC OFFSET 350",
			wantParams: []interface{}{2147483647, "admin", "aaa", "000XXX12354%"},
		},
		{
			authorityID: "admin",
			params: &SigningLogParams{
				Offset:       350,
				Filter:       []string{"aaa"},
				Serialnumber: "000XXX12354",
			},
			wantSQL:    "SELECT *, count(*) OVER() AS total_count FROM signinglog s WHERE id < $1 AND make=$2 AND model IN ($3) AND serial_number LIKE $4 ORDER BY id DESC OFFSET 350",
			wantParams: []interface{}{2147483647, "admin", "aaa", "000XXX12354%"},
		},

		{
			authorityID: "admin",
			username:    "bob",
			params:      &SigningLogParams{},
			wantSQL:     `SELECT *, count(*) OVER() AS total_count FROM signinglog s WHERE id < $1 AND make=$2 AND EXISTS ( SELECT * FROM account acc INNER JOIN useraccountlink ua on ua.account_id=acc.id INNER JOIN userinfo u on ua.user_id=u.id WHERE acc.authority_id=s.make AND u.username=$3 ) ORDER BY id DESC OFFSET 0`,
			wantParams:  []interface{}{2147483647, "admin", "bob"},
		},
		{
			authorityID: "admin",
			params: &SigningLogParams{
				Serialnumber: "Robert'); DROP TABLE signinglog;--",
			},
			wantSQL:    `SELECT *, count(*) OVER() AS total_count FROM signinglog s WHERE id < $1 AND make=$2 AND serial_number LIKE $3 ORDER BY id DESC OFFSET 0`,
			wantParams: []interface{}{2147483647, "admin", "Robert'); DROP TABLE signinglog;--%"},
		},
	}

	for _, tt := range tests {
		got := signingLogSQLBuilder(tt.username, tt.authorityID, tt.params)
		sql, args, err := got.ToSql()

		c.Assert(err, check.IsNil)
		c.Assert(sql, check.Equals, tt.wantSQL)
		c.Assert(args, check.DeepEquals, tt.wantParams)
	}
}
