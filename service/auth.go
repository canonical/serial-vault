package service

import (
	"errors"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/auth"
	"github.com/CanonicalLtd/serial-vault/usso"
	jwt "github.com/dgrijalva/jwt-go"
)

func getUserFromJWT(w http.ResponseWriter, r *http.Request) (datastore.User, error) {
	token, err := auth.JWTCheck(w, r)
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
