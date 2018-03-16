// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018-2019 Canonical Ltd
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

package auth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/auth"
	"github.com/CanonicalLtd/serial-vault/usso"
	"github.com/juju/usso/openid"
	check "gopkg.in/check.v1"
)

func TestAuth(t *testing.T) { check.TestingT(t) }

type SuiteTest struct {
	User        datastore.User
	Permissions int
	Check       check.Checker
}

type authSuite struct{}

var _ = check.Suite(&authSuite{})

func (s *authSuite) TestGetUserAuthWhenAuthEnabled(c *check.C) {
	config := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)

	_, err := auth.GetUserFromJWT(w, r)
	c.Assert(err, check.NotNil)

	theRoles := []int{datastore.Standard, datastore.Admin, datastore.Superuser}
	for _, role := range theRoles {
		err := createJWTWithRole(r, role)
		c.Assert(err, check.IsNil)
		user, err := auth.GetUserFromJWT(w, r)
		c.Assert(err, check.IsNil)
		c.Assert(user.Username, check.Equals, "sv")
		c.Assert(user.Role, check.Equals, role)
	}
}

func (s *authSuite) TestGetUserAuthWhenAuthDisabled(c *check.C) {
	config := config.Settings{EnableUserAuth: false, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)

	user, err := auth.GetUserFromJWT(w, r)
	c.Assert(err, check.IsNil)
	c.Assert(user.Username, check.Equals, "")
	c.Assert(user.Role, check.Equals, 0)

	roles := []int{datastore.Standard, datastore.Admin, datastore.Superuser}
	for _, role := range roles {
		err := createJWTWithRole(r, role)
		c.Assert(err, check.IsNil)
		user, err := auth.GetUserFromJWT(w, r)
		c.Assert(err, check.IsNil)
		c.Assert(user.Username, check.Equals, "")
		c.Assert(user.Role, check.Equals, 0)
	}
}

func (s *authSuite) TestCheckStandardPermissionsWhenAuthEnabled(c *check.C) {
	config := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{Config: config}

	noRoleUser := datastore.User{Username: "auser", Role: 0}
	standardUser := datastore.User{Username: "auser", Role: datastore.Standard}
	adminUser := datastore.User{Username: "auser", Role: datastore.Admin}
	superUser := datastore.User{Username: "auser", Role: datastore.Superuser}

	tests := []SuiteTest{
		{noRoleUser, datastore.Standard, check.NotNil},
		{standardUser, datastore.Standard, check.IsNil},
		{adminUser, datastore.Standard, check.IsNil},
		{superUser, datastore.Standard, check.IsNil},
	}

	for _, t := range tests {
		err := auth.CheckUserPermissions(t.User, t.Permissions, false)
		c.Assert(err, t.Check)
	}
}

func (s *authSuite) TestCheckStandardPermissionsWhenAuthDisabled(c *check.C) {
	config := config.Settings{EnableUserAuth: false, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{Config: config}

	noRoleUser := datastore.User{Username: "auser", Role: 0}
	standardUser := datastore.User{Username: "auser", Role: datastore.Standard}
	adminUser := datastore.User{Username: "auser", Role: datastore.Admin}
	superUser := datastore.User{Username: "auser", Role: datastore.Superuser}

	tests := []SuiteTest{
		{noRoleUser, datastore.Standard, check.IsNil},
		{standardUser, datastore.Standard, check.IsNil},
		{adminUser, datastore.Standard, check.IsNil},
		{superUser, datastore.Standard, check.IsNil},
	}

	for _, t := range tests {
		err := auth.CheckUserPermissions(t.User, t.Permissions, false)
		c.Assert(err, t.Check)
	}
}

func (s *authSuite) TestCheckAdminPermissionsWhenAuthEnabled(c *check.C) {
	config := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{Config: config}

	noRoleUser := datastore.User{Username: "auser", Role: 0}
	standardUser := datastore.User{Username: "auser", Role: datastore.Standard}
	adminUser := datastore.User{Username: "auser", Role: datastore.Admin}
	superUser := datastore.User{Username: "auser", Role: datastore.Superuser}

	tests := []SuiteTest{
		{noRoleUser, datastore.Admin, check.NotNil},
		{standardUser, datastore.Admin, check.NotNil},
		{adminUser, datastore.Admin, check.IsNil},
		{superUser, datastore.Admin, check.IsNil},
	}

	for _, t := range tests {
		err := auth.CheckUserPermissions(t.User, t.Permissions, false)
		c.Assert(err, t.Check)
	}
}

func (s *authSuite) TestCheckAdminPermissionsWhenAuthDisabled(c *check.C) {
	config := config.Settings{EnableUserAuth: false, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{Config: config}

	noRoleUser := datastore.User{Username: "auser", Role: 0}
	standardUser := datastore.User{Username: "auser", Role: datastore.Standard}
	adminUser := datastore.User{Username: "auser", Role: datastore.Admin}
	superUser := datastore.User{Username: "auser", Role: datastore.Superuser}

	tests := []SuiteTest{
		{noRoleUser, datastore.Admin, check.IsNil},
		{standardUser, datastore.Admin, check.IsNil},
		{adminUser, datastore.Admin, check.IsNil},
		{superUser, datastore.Admin, check.IsNil},
	}

	for _, t := range tests {
		err := auth.CheckUserPermissions(t.User, t.Permissions, false)
		c.Assert(err, t.Check)
	}
}

func (s *authSuite) TestCheckSuperuserPermissionsWhenAuthEnabled(c *check.C) {
	config := config.Settings{EnableUserAuth: true, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{Config: config}

	noRoleUser := datastore.User{Username: "auser", Role: 0}
	standardUser := datastore.User{Username: "auser", Role: datastore.Standard}
	adminUser := datastore.User{Username: "auser", Role: datastore.Admin}
	superUser := datastore.User{Username: "auser", Role: datastore.Superuser}

	tests := []SuiteTest{
		{noRoleUser, datastore.Superuser, check.NotNil},
		{standardUser, datastore.Superuser, check.NotNil},
		{adminUser, datastore.Superuser, check.NotNil},
		{superUser, datastore.Superuser, check.IsNil},
	}

	for _, t := range tests {
		err := auth.CheckUserPermissions(t.User, t.Permissions, false)
		c.Assert(err, t.Check)
	}
}

func (s *authSuite) TestCheckSuperuserPermissionsWhenAuthDisabled(c *check.C) {

	config := config.Settings{EnableUserAuth: false, JwtSecret: "SomeTestSecretValue"}
	datastore.Environ = &datastore.Env{Config: config}

	noRoleUser := datastore.User{Username: "auser", Role: 0}
	standardUser := datastore.User{Username: "auser", Role: datastore.Standard}
	adminUser := datastore.User{Username: "auser", Role: datastore.Admin}
	superUser := datastore.User{Username: "auser", Role: datastore.Superuser}

	tests := []SuiteTest{
		{noRoleUser, datastore.Superuser, check.NotNil},
		{standardUser, datastore.Superuser, check.NotNil},
		{adminUser, datastore.Superuser, check.NotNil},
		{superUser, datastore.Superuser, check.NotNil},
	}

	for _, t := range tests {
		err := auth.CheckUserPermissions(t.User, t.Permissions, false)
		c.Assert(err, t.Check)
	}

}

func createJWTWithRole(r *http.Request, role int) error {
	sreg := map[string]string{"nickname": "sv", "fullname": "Steven Vault", "email": "sv@example.com"}
	resp := openid.Response{ID: "identity", Teams: []string{}, SReg: sreg}
	jwtToken, err := usso.NewJWTToken(&resp, role)
	if err != nil {
		return fmt.Errorf("Error creating a JWT: %v", err)
	}
	r.Header.Set("Authorization", "Bearer "+jwtToken)
	return nil
}
