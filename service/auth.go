package service

import (
	"errors"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/usso"
	jwt "github.com/dgrijalva/jwt-go"
)

func checkIsStandardAndGetUserFromJWT(w http.ResponseWriter, r *http.Request) (datastore.User, error) {
	return checkPermissionsAndGetUserFromJWT(w, r, datastore.Standard)
}

func checkIsAdminAndGetUserFromJWT(w http.ResponseWriter, r *http.Request) (datastore.User, error) {
	return checkPermissionsAndGetUserFromJWT(w, r, datastore.Admin)
}

func checkIsSuperuserAndGetUserFromJWT(w http.ResponseWriter, r *http.Request) (datastore.User, error) {
	return checkPermissionsAndGetUserFromJWT(w, r, datastore.Superuser)
}

func checkPermissionsAndGetUserFromJWT(w http.ResponseWriter, r *http.Request, minimumAuthorizedRole int) (datastore.User, error) {
	user, err := getUserFromJWT(w, r)
	if err != nil {
		return user, err
	}
	err = checkUserPermissions(user, minimumAuthorizedRole)
	if err != nil {
		return user, err
	}
	return user, nil
}

func getUserFromJWT(w http.ResponseWriter, r *http.Request) (datastore.User, error) {
	token, err := JWTCheck(w, r)
	if err != nil {
		return datastore.User{}, err
	}

	// Null token means that auth is not enabled.
	if token == nil {
		return datastore.User{}, nil
	}

	claims := token.Claims.(jwt.MapClaims)
	username := claims[usso.ClaimsUsername].(string)
	role := int(claims[usso.ClaimsRole].(float64))

	return datastore.User{
		Username: username,
		Role:     role,
	}, nil
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
