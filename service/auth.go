package service

import (
	"errors"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/usso"
	jwt "github.com/dgrijalva/jwt-go"
)

func checkIsAdminAndGetAuthUser(w http.ResponseWriter, r *http.Request) (datastore.User, error) {
	return checkPermissionsAndGetAuthUser(w, r, datastore.Admin)
}

func checkIsSuperuserAndGetAuthUser(w http.ResponseWriter, r *http.Request) (datastore.User, error) {
	return checkPermissionsAndGetAuthUser(w, r, datastore.Superuser)
}

func checkPermissionsAndGetAuthUser(w http.ResponseWriter, r *http.Request, minimumAuthorizedRole int) (datastore.User, error) {
	user, err := getAuthUser(w, r)
	if err != nil {
		return user, err
	}
	err = checkUserPermissions(user, minimumAuthorizedRole)
	if err != nil {
		return user, err
	}
	return user, nil
}

func getAuthUser(w http.ResponseWriter, r *http.Request) (datastore.User, error) {
	token, err := JWTCheck(w, r)
	if err != nil {
		return datastore.User{}, err
	}

	// nil token means auth is not enabled.
	if token == nil {
		return datastore.User{}, nil
	}

	claims := token.Claims.(jwt.MapClaims)
	username := claims[usso.ClaimsUsername].(string)

	if len(username) == 0 {
		return datastore.User{}, nil
	}

	return datastore.User{
		Username: username,
		Role:     roleForUser(username),
	}, nil
}

func roleForUser(username string) int {
	if len(username) == 0 {
		return 0
	}

	user, err := datastore.Environ.DB.GetUserByUsername(username)
	if err != nil {
		return 0
	}
	return user.Role
}

func checkUserPermissions(user datastore.User, minimumAuthorizedRole int) error {
	// User authentication is turned off
	if !datastore.Environ.Config.EnableUserAuth {
		// Superuser permissions don't allow turned off authentication
		if minimumAuthorizedRole == datastore.Superuser {
			return errors.New("The user is not authorized")
		}
		return nil
	}

	if user.Role < minimumAuthorizedRole {
		return errors.New("The user is not authorized")
	}
	return nil
}
