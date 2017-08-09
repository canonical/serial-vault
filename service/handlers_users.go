package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/gorilla/mux"
)

// UserRequest is the JSON version of the request to create a user
type UserRequest struct {
	Username string   `json:"username"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Role     int      `json:"role"`
	Accounts []string `json:"accounts"`
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

	_, err := checkIsSuperuserAndGetUserFromJWT(w, r)
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

// UserCreateHandler is the API method to create a user
func UserCreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	_, err := checkIsSuperuserAndGetUserFromJWT(w, r)
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

	err = validateUser(userRequest)
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
		Accounts: datastore.BuildAccountsFromAuthorityIDs(userRequest.Accounts),
	}
	user.ID, err = datastore.Environ.DB.CreateUser(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatUserResponse(false, "error-creating-user", "", err.Error(), datastore.User{}, w)
		return
	}

	// Format the user for output and return JSON response
	w.WriteHeader(http.StatusOK)
	formatUserResponse(true, "", "", "", user, w)
}

// UserGetHandler is the API method to retrieve user info
func UserGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	_, err := checkIsSuperuserAndGetUserFromJWT(w, r)
	if err != nil {
		formatUserResponse(false, "error-auth", "", "", datastore.User{}, w)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("%v", vars)
		formatUserResponse(false, "error-invalid-user", "", errorMessage, datastore.User{}, w)
		return
	}

	user, err := datastore.Environ.DB.GetUser(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("User ID: %d.", id)
		formatUserResponse(false, "error-get-user", "", errorMessage, datastore.User{ID: id}, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	formatUserResponse(true, "", "", "", user, w)
}

// UserOtherAccountsGetHandler is the API method to retrieve accounts not belonging to the user
func UserOtherAccountsGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	_, err := checkIsSuperuserAndGetUserFromJWT(w, r)
	if err != nil {
		formatUserResponse(false, "error-auth", "", "", datastore.User{}, w)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("%v", vars)
		formatAccountsResponse(false, "error-invalid-user", "", errorMessage, nil, w)
		return
	}

	user, err := datastore.Environ.DB.GetUser(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("User ID: %d.", id)
		formatAccountsResponse(false, "error-get-user", "", errorMessage, nil, w)
		return
	}

	accounts, err := datastore.Environ.DB.ListNotUserAccounts(user.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		formatAccountsResponse(false, "error-get-non-user-accounts", "", err.Error(), nil, w)
		return
	}

	// Format the model for output and return JSON response
	w.WriteHeader(http.StatusOK)
	formatAccountsResponse(true, "", "", "", accounts, w)
}

// UserUpdateHandler is the API method to update user info
func UserUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	_, err := checkIsSuperuserAndGetUserFromJWT(w, r)
	if err != nil {
		formatUserResponse(false, "error-auth", "", "", datastore.User{}, w)
		return
	}

	// Get the user primary key
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("%v", vars["id"])
		formatUserResponse(false, "error-invalid-user", "", errorMessage, datastore.User{}, w)
		return
	}

	// Check that we have a message body
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		formatUserResponse(false, "error-nil-data", "", "Uninitialized POST data", datastore.User{}, w)
		return
	}
	defer r.Body.Close()

	userRequest := UserRequest{}
	err = json.NewDecoder(r.Body).Decode(&userRequest)
	switch {
	// Check we have some data
	case err == io.EOF:
		w.WriteHeader(http.StatusBadRequest)
		formatUserResponse(false, "error-user-data", "", "No user data supplied.", datastore.User{}, w)
		return
		// Check for parsing errors
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		formatUserResponse(false, "error-decode-json", "", err.Error(), datastore.User{}, w)
		return
	}

	err = validateUser(userRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatUserResponse(false, "error-updating-user", "", err.Error(), datastore.User{}, w)
		return
	}

	// Update the database
	user := datastore.User{
		ID:       id,
		Username: userRequest.Username,
		Name:     userRequest.Name,
		Email:    userRequest.Email,
		Role:     userRequest.Role,
		Accounts: datastore.BuildAccountsFromAuthorityIDs(userRequest.Accounts),
	}
	err = datastore.Environ.DB.UpdateUser(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatUserResponse(false, "error-updating-user", "", err.Error(), datastore.User{}, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	formatUserResponse(true, "", "", "", user, w)
}

// UserDeleteHandler is the API method to delete a user
func UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	_, err := checkIsSuperuserAndGetUserFromJWT(w, r)
	if err != nil {
		formatUserResponse(false, "error-auth", "", "", datastore.User{}, w)
		return
	}

	// Get the user primary key
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		errorMessage := fmt.Sprintf("%v", vars["id"])
		formatUserResponse(false, "error-invalid-user", "", errorMessage, datastore.User{}, w)
		return
	}

	err = datastore.Environ.DB.DeleteUser(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		formatUserResponse(false, "error-deleting-user", "", err.Error(), datastore.User{}, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	formatUserResponse(true, "", "", "", datastore.User{}, w)
}

func validateUser(userRequest UserRequest) error {
	// Validate username; the rule is: lowercase with no spaces
	err := validateUsername(userRequest.Username)
	if err != nil {
		return err
	}

	// Validate name; the rule is: not empty
	err = validateUserFullName(userRequest.Name)
	if err != nil {
		return err
	}

	// Validate email; the rule is: not empty
	err = validateUserEmail(userRequest.Email)
	if err != nil {
		return err
	}

	// Validate role; the rule is the role is 100, 200 or 300
	err = validateUserRole(userRequest.Role)
	if err != nil {
		return err
	}
	return nil
}
