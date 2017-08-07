package service

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/usso"
	"github.com/juju/usso/openid"

	check "gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type ServiceSuite struct{}

var _ = check.Suite(&ServiceSuite{})

func (s *ServiceSuite) SetUpSuite(c *check.C) {
	datastore.Environ.Config.EnableUserAuth = true
}

func (s *ServiceSuite) TestUsersHandler(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	result := s.sendRequestRepliesUsersList("GET", "/v1/users", nil, c)
	c.Assert(len(result.Users), check.Equals, 5)
	c.Assert(result.Users[0].Name, check.Equals, "Rigoberto Picaporte")
	c.Assert(result.Users[1].Name, check.Equals, "Nancy Reagan")
	c.Assert(result.Users[2].Name, check.Equals, "Steven Vault")
	c.Assert(result.Users[3].Name, check.Equals, "A")
	c.Assert(result.Users[4].Name, check.Equals, "Root User")
}

func (s *ServiceSuite) TestUsersHandlerWithError(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}

	s.sendRequestRepliesUsersListError("GET", "/v1/users", nil, c)
}

func (s *ServiceSuite) TestUsersHandlerWithoutPermissions(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	s.sendRequestWithoutPermissions("GET", "/v1/users", nil, c)
}

func (s *ServiceSuite) TestGetUserHandler(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	result := s.sendRequestRepliesUser("GET", "/v1/users/4", nil, c)
	c.Assert(result.User.ID, check.Equals, 4)
	c.Assert(result.User.Username, check.Equals, "a")
	c.Assert(result.User.Name, check.Equals, "A")
	c.Assert(result.User.OpenIDIdentity, check.Equals, "https://login.ubuntu.com/+id/AAAAAA")
	c.Assert(result.User.Email, check.Equals, "a@example.com")
	c.Assert(result.User.Role, check.Equals, datastore.Standard)
	c.Assert(len(result.User.Accounts), check.Equals, 0)
}

func (s *ServiceSuite) TestGetUserHandlerWithAccount(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	result := s.sendRequestRepliesUser("GET", "/v1/users/2", nil, c)
	c.Assert(result.User.ID, check.Equals, 2)
	c.Assert(result.User.Username, check.Equals, "user2")
	c.Assert(result.User.Name, check.Equals, "Nancy Reagan")
	c.Assert(result.User.Email, check.Equals, "nancy.reagan@usa.gov")
	c.Assert(result.User.Role, check.Equals, datastore.Standard)
	c.Assert(len(result.User.Accounts), check.Equals, 1)
	c.Assert(result.User.Accounts[0].ID, check.Equals, 2)
	c.Assert(result.User.Accounts[0].AuthorityID, check.Equals, "authority2")
	c.Assert(result.User.Accounts[0].Assertion, check.Equals, "assertioncontent2")
}

func (s *ServiceSuite) TestGetUserHandlerWithError(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}

	s.sendRequestRepliesUserError("GET", "/v1/users/2", nil, c)
}

func (s *ServiceSuite) TestGetUserHandlerWithoutPermissions(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	s.sendRequestWithoutPermissions("GET", "/v1/users/2", nil, c)
}

func (s *ServiceSuite) TestCreateUserHandler(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := UserRequest{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)

	result := s.sendRequestRepliesUser("POST", "/v1/users", bytes.NewReader(data), c)
	c.Assert(result.User.ID, check.Equals, 740)
	c.Assert(result.User.Username, check.Equals, "theusername")
}

func (s *ServiceSuite) TestCreateUserHandlerWithOneAccount(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := UserRequest{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
		Accounts: []string{"theauthorityid1"},
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)

	result := s.sendRequestRepliesUser("POST", "/v1/users", bytes.NewReader(data), c)
	c.Assert(result.User.ID, check.Equals, 740)
	c.Assert(result.User.Username, check.Equals, "theusername")
}

func (s *ServiceSuite) TestCreateUserHandlerWithAccounts(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := UserRequest{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
		Accounts: []string{"theauthorityid1", "theauthorityid2"},
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)

	result := s.sendRequestRepliesUser("POST", "/v1/users", bytes.NewReader(data), c)
	c.Assert(result.User.ID, check.Equals, 740)
	c.Assert(result.User.Username, check.Equals, "theusername")
}

func (s *ServiceSuite) TestCreateUserHandlerWithError(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}

	user := UserRequest{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
		Accounts: []string{"theauthorityid1"},
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)
	s.sendRequestRepliesUserError("POST", "/v1/users", bytes.NewReader(data), c)
}

func (s *ServiceSuite) TestCreateUserHandlerWithoutPermissions(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := UserRequest{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
		Accounts: []string{"theauthorityid1"},
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)
	s.sendRequestWithoutPermissions("POST", "/v1/users", bytes.NewReader(data), c)
}

func (s *ServiceSuite) TestUpdateUserHandler(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := UserRequest{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
		Accounts: []string{"theauthorityid1"},
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)

	result := s.sendRequestRepliesUser("PUT", "/v1/users/2", bytes.NewReader(data), c)
	c.Assert(result.User.ID, check.Equals, 2)
	c.Assert(result.User.Username, check.Equals, "theusername")
	c.Assert(result.User.Name, check.Equals, "The Name")
	c.Assert(result.User.Email, check.Equals, "theemail@mydb.com")
	c.Assert(result.User.Role, check.Equals, datastore.Standard)
}

func (s *ServiceSuite) TestUpdateUserHandlerWithAccount(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := UserRequest{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
		Accounts: []string{"theauthorityid1"},
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)

	result := s.sendRequestRepliesUser("PUT", "/v1/users/2", bytes.NewReader(data), c)
	c.Assert(result.User.ID, check.Equals, 2)
	c.Assert(result.User.Username, check.Equals, "theusername")
	c.Assert(result.User.Name, check.Equals, "The Name")
	c.Assert(result.User.Email, check.Equals, "theemail@mydb.com")
	c.Assert(result.User.Role, check.Equals, datastore.Standard)
}

func (s *ServiceSuite) TestUpdateUserHandlerWithAccounts(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := UserRequest{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
		Accounts: []string{"theauthorityid1", "theauthorityid2"},
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)

	result := s.sendRequestRepliesUser("PUT", "/v1/users/2", bytes.NewReader(data), c)
	c.Assert(result.User.ID, check.Equals, 2)
	c.Assert(result.User.Username, check.Equals, "theusername")
	c.Assert(result.User.Name, check.Equals, "The Name")
	c.Assert(result.User.Email, check.Equals, "theemail@mydb.com")
	c.Assert(result.User.Role, check.Equals, datastore.Standard)
}

func (s *ServiceSuite) TestUpdateUserHandlerWithError(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}

	user := UserRequest{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)
	s.sendRequestRepliesUserError("PUT", "/v1/users/2", bytes.NewReader(data), c)
}

func (s *ServiceSuite) TestUpdateUserHandlerWithoutPermissions(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	user := UserRequest{
		Username: "theusername",
		Name:     "The Name",
		Email:    "theemail@mydb.com",
		Role:     datastore.Standard,
	}
	data, err := json.Marshal(user)
	c.Assert(err, check.IsNil)
	s.sendRequestWithoutPermissions("PUT", "/v1/users/2", bytes.NewReader(data), c)
}

func (s *ServiceSuite) TestDeleteUserHandler(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	s.sendRequestRepliesUser("DELETE", "/v1/users/2", nil, c)
}

func (s *ServiceSuite) TestDeleteUserHandlerWithError(c *check.C) {
	datastore.Environ.DB = &datastore.ErrorMockDB{}

	s.sendRequestRepliesUserError("DELETE", "/v1/users/2", nil, c)
}

func (s *ServiceSuite) TestDeleteUserHandlerWithoutPermissions(c *check.C) {
	datastore.Environ.DB = &datastore.MockDB{}

	s.sendRequestWithoutPermissions("DELETE", "/v1/users/2", nil, c)
}

func (s *ServiceSuite) createSuperuserJWT(r *http.Request, c *check.C) {
	sreg := map[string]string{"nickname": "root", "fullname": "Root User", "email": "the_root_user@thisdb.com"}
	resp := openid.Response{ID: "identity", Teams: []string{}, SReg: sreg}
	jwtToken, err := usso.NewJWTToken(&resp, datastore.Superuser)
	c.Assert(err, check.IsNil)

	r.Header.Set("Authorization", "Bearer "+jwtToken)
}

func (s *ServiceSuite) sendRequestRepliesUser(method, url string, data io.Reader, c *check.C) UserResponse {
	body := s.sendRequest(method, url, data, c)

	result := UserResponse{}
	err := json.NewDecoder(body).Decode(&result)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, true)

	return result
}

func (s *ServiceSuite) sendRequestRepliesUsersList(method, url string, data io.Reader, c *check.C) UsersResponse {
	body := s.sendRequest(method, url, data, c)

	result := UsersResponse{}
	err := json.NewDecoder(body).Decode(&result)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, true)

	return result
}

func (s *ServiceSuite) sendRequest(method, url string, data io.Reader, c *check.C) *bytes.Buffer {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)

	s.createSuperuserJWT(r, c)

	AdminRouter().ServeHTTP(w, r)

	return w.Body
}

func (s *ServiceSuite) sendRequestRepliesUserError(method, url string, data io.Reader, c *check.C) {
	body := s.sendRequest(method, url, data, c)
	result := UserResponse{}
	err := json.NewDecoder(body).Decode(&result)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, false)
}

func (s *ServiceSuite) sendRequestRepliesUsersListError(method, url string, data io.Reader, c *check.C) {
	body := s.sendRequest(method, url, data, c)
	result := UsersResponse{}
	err := json.NewDecoder(body).Decode(&result)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, false)
}

func (s *ServiceSuite) sendRequestWithoutPermissions(method, url string, data io.Reader, c *check.C) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, url, data)
	AdminRouter().ServeHTTP(w, r)

	result := UserResponse{}
	err := json.NewDecoder(w.Body).Decode(&result)
	c.Assert(err, check.IsNil)
	c.Assert(result.Success, check.Equals, false)
}
