package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
)

// UserRequest is the JSON version of the request to create a user
type UserRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     int    `json:"role"`
}

// UserResponse is the response from a user creation/update
type UserResponse struct {
	Success      bool           `json:"success"`
	ErrorCode    string         `json:"error_code"`
	ErrorSubcode string         `json:"error_subcode"`
	ErrorMessage string         `json:"message"`
	User         datastore.User `json:"user"`
}

// UsersResponse is the response from a user list request
type UsersResponse struct {
	Success      bool             `json:"success"`
	ErrorCode    string           `json:"error_code"`
	ErrorSubcode string           `json:"error_subcode"`
	ErrorMessage string           `json:"message"`
	Users        []datastore.User `json:"users"`
}

// UsersHandler is the API method to list the users
func UsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get the user from the JWT
	username, err := checkUserPermissions(w, r, datastore.Superuser)
	if err != nil {
		formatUsersResponse(false, "error-auth", "", "", nil, w)
		return
	}

	users, err := datastore.Environ.DB.ListUsers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		formatUsersResponse(false, "error-fetch-users", "", err.Error(), nil, w)
		return
	}

	// Return successful JSON response with the list of users
	w.WriteHeader(http.StatusOK)
	formatUsersResponse(true, "", "", "", users, w)
}

// UsersCreateHandler is the API method to create or update a user
func UsersCreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get the user from the JWT
	username, err := checkUserPermissions(w, r, datastore.Superuser)
	if err != nil {
		formatUserResponse(false, "error-auth", "", "", datastore.User{}, w)
		return
	}

	// Check that we have a message body
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		formatUserResponse(false, "error-nil-data", "", "Uninitialized POST data", datastore.User{}, w)
		return
	}
	defer r.Body.Close()

	// Decode the JSON body
	userRequest := UserRequest{}
	err = json.NewDecoder(r.Body).Decode(&userRequest)
	switch {

	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatUserResponse(false, "error-user-data", "", "No user data supplied", datastore.User{}, w)
		return
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatUserResponse(false, "error-decode-json", "", errorMessage, datastore.User{}, w)
		return
	}

	// Validate username; the rule is: lowercase with no spaces
	err = validateUsername(userRequest.Username)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatUserResponse(false, "error-creating-user", "", err.Error(), datastore.User{}, w)
		return
	}

	// Validate role; the rule is the role is 100, 200 or 300
	err = validateUserRole(userRequest.Role)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatUserResponse(false, "error-creating-user", "", err.Error(), datastore.User{}, w)
		return
	}

	// Create a new user
	user := datastore.User{
		Username: userRequest.Username,
		Name:     userRequest.Name,
		Email:    userRequest.Email,
		Role:     userRequest.Role,
	}
	user.ID, err = datastore.Environ.DB.CreateUser(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorMessage := fmt.Sprintf("%v", err)
		formatUserResponse(false, "error-creating-user", "", err.Error(), datastore.User{}, w)
		return
	}

	// Format the model for output and return JSON response
	w.WriteHeader(http.StatusOK)
	formatUserResponse(true, "", "", "", user, w)
}
