package service

import (
	"net/http"
	"net/http/httptest"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	check "gopkg.in/check.v1"
)

type authSuite struct{}

var _ = check.Suite(&authSuite{})

func (s *authSuite) TestGetUserAuthWhenAuthEnabled(c *check.C) {
	config := config.Settings{EnableUserAuth: true}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)

	_, err := getUserFromJWT(w, r)
	c.Assert(err, check.NotNil)

	theRoles := []int{datastore.Standard, datastore.Admin, datastore.Superuser}
	for _, role := range theRoles {
		err := createJWTWithRole(r, role)
		c.Assert(err, check.IsNil)
		user, err := getUserFromJWT(w, r)
		c.Assert(err, check.IsNil)
		c.Assert(user.Username, check.Equals, "sv")
		c.Assert(user.Role, check.Equals, role)
	}
}

func (s *authSuite) TestGetUserAuthWhenAuthDisabled(c *check.C) {
	config := config.Settings{EnableUserAuth: false}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)

	user, err := getUserFromJWT(w, r)
	c.Assert(err, check.IsNil)
	c.Assert(user.Username, check.Equals, "")
	c.Assert(user.Role, check.Equals, 0)

	roles := []int{datastore.Standard, datastore.Admin, datastore.Superuser}
	for _, role := range roles {
		err := createJWTWithRole(r, role)
		c.Assert(err, check.IsNil)
		user, err := getUserFromJWT(w, r)
		c.Assert(err, check.IsNil)
		c.Assert(user.Username, check.Equals, "")
		c.Assert(user.Role, check.Equals, 0)
	}
}

func (s *authSuite) TestCheckStandardPermissionsWhenAuthEnabled(c *check.C) {

	config := config.Settings{EnableUserAuth: true}
	datastore.Environ = &datastore.Env{Config: config}

	noRoleUser := datastore.User{Username: "auser", Role: 0}
	err := checkUserPermissions(noRoleUser, datastore.Standard)
	c.Assert(err, check.NotNil)

	standardUser := datastore.User{Username: "auser", Role: datastore.Standard}
	err = checkUserPermissions(standardUser, datastore.Standard)
	c.Assert(err, check.IsNil)

	adminUser := datastore.User{Username: "auser", Role: datastore.Admin}
	err = checkUserPermissions(adminUser, datastore.Standard)
	c.Assert(err, check.IsNil)

	superUser := datastore.User{Username: "auser", Role: datastore.Superuser}
	err = checkUserPermissions(superUser, datastore.Standard)
	c.Assert(err, check.IsNil)
}

func (s *authSuite) TestCheckStandardPermissionsWhenAuthDisabled(c *check.C) {

	config := config.Settings{EnableUserAuth: false}
	datastore.Environ = &datastore.Env{Config: config}

	noRoleUser := datastore.User{Username: "auser", Role: 0}
	err := checkUserPermissions(noRoleUser, datastore.Standard)
	c.Assert(err, check.IsNil)

	standardUser := datastore.User{Username: "auser", Role: datastore.Standard}
	err = checkUserPermissions(standardUser, datastore.Standard)
	c.Assert(err, check.IsNil)

	adminUser := datastore.User{Username: "auser", Role: datastore.Admin}
	err = checkUserPermissions(adminUser, datastore.Standard)
	c.Assert(err, check.IsNil)

	superUser := datastore.User{Username: "auser", Role: datastore.Superuser}
	err = checkUserPermissions(superUser, datastore.Standard)
	c.Assert(err, check.IsNil)
}

func (s *authSuite) TestCheckAdminPermissionsWhenAuthEnabled(c *check.C) {

	config := config.Settings{EnableUserAuth: true}
	datastore.Environ = &datastore.Env{Config: config}

	noRoleUser := datastore.User{Username: "auser", Role: 0}
	err := checkUserPermissions(noRoleUser, datastore.Admin)
	c.Assert(err, check.NotNil)

	standardUser := datastore.User{Username: "auser", Role: datastore.Standard}
	err = checkUserPermissions(standardUser, datastore.Admin)
	c.Assert(err, check.NotNil)

	adminUser := datastore.User{Username: "auser", Role: datastore.Admin}
	err = checkUserPermissions(adminUser, datastore.Admin)
	c.Assert(err, check.IsNil)

	superUser := datastore.User{Username: "auser", Role: datastore.Superuser}
	err = checkUserPermissions(superUser, datastore.Admin)
	c.Assert(err, check.IsNil)
}

func (s *authSuite) TestCheckAdminPermissionsWhenAuthDisabled(c *check.C) {

	config := config.Settings{EnableUserAuth: false}
	datastore.Environ = &datastore.Env{Config: config}

	noRoleUser := datastore.User{Username: "auser", Role: 0}
	err := checkUserPermissions(noRoleUser, datastore.Admin)
	c.Assert(err, check.IsNil)

	standardUser := datastore.User{Username: "auser", Role: datastore.Standard}
	err = checkUserPermissions(standardUser, datastore.Admin)
	c.Assert(err, check.IsNil)

	adminUser := datastore.User{Username: "auser", Role: datastore.Admin}
	err = checkUserPermissions(adminUser, datastore.Admin)
	c.Assert(err, check.IsNil)

	superUser := datastore.User{Username: "auser", Role: datastore.Superuser}
	err = checkUserPermissions(superUser, datastore.Admin)
	c.Assert(err, check.IsNil)
}

func (s *authSuite) TestCheckSuperuserPermissionsWhenAuthEnabled(c *check.C) {

	config := config.Settings{EnableUserAuth: true}
	datastore.Environ = &datastore.Env{Config: config}

	noRoleUser := datastore.User{Username: "auser", Role: 0}
	err := checkUserPermissions(noRoleUser, datastore.Superuser)
	c.Assert(err, check.NotNil)

	standardUser := datastore.User{Username: "auser", Role: datastore.Standard}
	err = checkUserPermissions(standardUser, datastore.Superuser)
	c.Assert(err, check.NotNil)

	adminUser := datastore.User{Username: "auser", Role: datastore.Admin}
	err = checkUserPermissions(adminUser, datastore.Superuser)
	c.Assert(err, check.NotNil)

	superUser := datastore.User{Username: "auser", Role: datastore.Superuser}
	err = checkUserPermissions(superUser, datastore.Superuser)
	c.Assert(err, check.IsNil)
}

func (s *authSuite) TestCheckSuperuserPermissionsWhenAuthDisabled(c *check.C) {

	config := config.Settings{EnableUserAuth: false}
	datastore.Environ = &datastore.Env{Config: config}

	noRoleUser := datastore.User{Username: "auser", Role: 0}
	err := checkUserPermissions(noRoleUser, datastore.Superuser)
	c.Assert(err, check.NotNil)

	standardUser := datastore.User{Username: "auser", Role: datastore.Standard}
	err = checkUserPermissions(standardUser, datastore.Superuser)
	c.Assert(err, check.NotNil)

	adminUser := datastore.User{Username: "auser", Role: datastore.Admin}
	err = checkUserPermissions(adminUser, datastore.Superuser)
	c.Assert(err, check.NotNil)

	superUser := datastore.User{Username: "auser", Role: datastore.Superuser}
	err = checkUserPermissions(superUser, datastore.Superuser)
	c.Assert(err, check.NotNil)

}
