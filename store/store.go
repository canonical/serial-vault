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

package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"gopkg.in/macaroon.v1"

	"github.com/CanonicalLtd/serial-vault/datastore"

	"github.com/snapcore/snapd/asserts"
	"github.com/snapcore/snapd/overlord/auth"
)

const (
	ssoBaseURL   = "https://login.ubuntu.com/api/v2/"
	storeBaseURL = "https://dashboard.snapcraft.io/dev/api/"
)

// Permissions is the SSO authorization for the store
type Permissions struct {
	Permissions []string `json:"permissions"`
}

// ACL is the SSO authorization for the store
type ACL struct {
	Macaroon string `json:"macaroon"`
}

// Discharge is the SSO authorization for the store
type Discharge struct {
	Macaroon string `json:"discharge_macaroon"`
}

// Auth is the SSO authorization for the store
type Auth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	OTP      string `json:"otp"`
}

// KeyRegister is the request to submit a signing-key to the store
type KeyRegister struct {
	Auth
	AuthorityID string `json:"authority-id"`
	KeyName     string `json:"key-name"`
}

// RegisterKey registers an account-key assertion with the store
var RegisterKey = func(keyAuth KeyRegister, keypair datastore.Keypair) error {

	// Login to the store as the user
	// Permissions are documented at https://dashboard.snapcraft.io/docs/api/macaroon.html#request-a-macaroon
	permissions := []string{"edit_account", "modify_account_key"}
	m, discharge, err := LoginUser(keyAuth.Email, keyAuth.Password, keyAuth.OTP, permissions)
	if err != nil {
		log.Println("Error logging in to store", err)
		return err
	}

	// Generate the authorization header from the macaroons
	authHeader, err := AuthorizationHeader(m, discharge)
	if err != nil {
		return err
	}

	// Generate the account-key request assertion
	assert, err := generateAccountKeyRequest(keyAuth, keypair)
	if err != nil {
		return err
	}

	// The assertion needs to be sent as JSON (undocumented)
	data := map[string]string{
		"account_key_request": string(assert),
	}
	d, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshalling account-key assertion: %v", err)
		return err
	}

	// Submit the account-key assertion
	headers := map[string]string{
		"Authorization": authHeader,
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}
	_, err = submitPOSTRequest(storeBaseURL+"account/account-key", headers, d)
	if err != nil {
		log.Printf("Error submitting the account-key assertion: %v", err)
		return err
	}

	return nil
}

// LoginUser logs user in the store and returns the authentication macaroons.
func LoginUser(username, password, otp string, permissions []string) (string, string, error) {
	macaroon, err := requestStoreMacaroon(permissions)
	if err != nil {
		return "", "", err
	}
	deserializedMacaroon, err := auth.MacaroonDeserialize(macaroon)
	if err != nil {
		log.Println("Error deserializing macaroon", err)
		return "", "", err
	}

	// get SSO 3rd party caveat, and request discharge
	loginCaveat, err := loginCaveatID(deserializedMacaroon)
	if err != nil {
		log.Println("Error with login caveat", err)
		return "", "", err
	}

	discharge, err := dischargeAuthCaveat(loginCaveat, username, password, otp)
	if err != nil {
		log.Println("Error with discharge", err)
		return "", "", err
	}

	return macaroon, discharge, nil
}

func submitPOSTRequest(url string, headers map[string]string, data []byte) (*http.Request, error) {
	r, _ := http.NewRequest("POST", url, bytes.NewReader(data))

	for k, v := range headers {
		r.Header.Set(k, v)
	}
	client := http.Client{}
	_, err := client.Do(r)
	if err != nil {
		return r, err
	}

	return r, nil
}

func generateAccountKeyRequest(keyAuth KeyRegister, keypair datastore.Keypair) (string, error) {
	// Create a self-signed account-key assertion
	since := time.Now().AddDate(-1, 0, 0)
	headers := map[string]interface{}{
		"account-id":          keyAuth.AuthorityID,
		"name":                keyAuth.KeyName,
		"public-key-sha3-384": keypair.KeyID,
		"since":               since.Format(time.RFC3339),
	}

	// Loads the keypair into the memory keystore
	err := datastore.Environ.KeypairDB.LoadKeypair(keypair.AuthorityID, keypair.KeyID, keypair.SealedKey)
	if err != nil {
		log.Println("Error loading the keypair", err)
		return "", err
	}

	// Get the public key as it is the body of the assertion
	publicKey, err := datastore.Environ.KeypairDB.PublicKey(keypair.KeyID)
	if err != nil {
		log.Println("Error fetching the public key", err)
		return "", err
	}
	pubKeyEncoded, err := asserts.EncodePublicKey(publicKey)
	if err != nil {
		log.Println("Error encoding the public key", err)
		return "", err
	}

	accountKey, err := datastore.Environ.KeypairDB.SignAssertion(asserts.AccountKeyRequestType, headers, pubKeyEncoded, keypair.AuthorityID, keypair.KeyID, keypair.SealedKey)
	if err != nil {
		log.Printf("Error creating account-key assertion: %v", err)
		return "", err
	}

	assert := asserts.Encode(accountKey)
	return string(assert), nil
}

func requestStoreMacaroon(permissions []string) (string, error) {

	perm := Permissions{Permissions: permissions}
	macaroonJSONData, err := json.Marshal(perm)
	if err != nil {
		log.Println("Error marshalling the macaroon", err)
		return "", err
	}

	// Submit the account-key assertion
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	r, err := submitPOSTRequest(storeBaseURL+"acl/", headers, macaroonJSONData)
	if err != nil {
		log.Printf("Error submitting the ACL request: %v", err)
		return "", err
	}

	defer r.Body.Close()

	acl := ACL{}
	err = json.NewDecoder(r.Body).Decode(&acl)
	if err != nil {
		log.Printf("Error decoding the ACL request: %v", err)
		return "", err
	}

	return acl.Macaroon, nil
}

// loginCaveatID returns the 3rd party caveat from the macaroon to be discharged by Ubuntuone
func loginCaveatID(m *macaroon.Macaroon) (string, error) {
	caveatID := ""

	for _, caveat := range m.Caveats() {
		if caveat.Location == "login.ubuntu.com" {
			caveatID = caveat.Id
			break
		}
	}
	if caveatID == "" {
		return "", fmt.Errorf("missing login caveat")
	}
	return caveatID, nil
}

// dischargeAuthCaveat returns a macaroon with the store auth caveat discharged.
func dischargeAuthCaveat(caveat, username, password, otp string) (string, error) {
	data := map[string]string{
		"email":     username,
		"password":  password,
		"caveat_id": caveat,
	}
	if otp != "" {
		data["otp"] = otp
	}

	return requestDischargeMacaroon(ssoBaseURL+"tokens/discharge", data)
}

// refreshDischargeMacaroon returns a soft-refreshed discharge macaroon.
func refreshDischargeMacaroon(discharge string) (string, error) {
	data := map[string]string{
		"discharge_macaroon": discharge,
	}

	return requestDischargeMacaroon(ssoBaseURL+"tokens/refresh", data)
}

func requestDischargeMacaroon(endpoint string, data map[string]string) (string, error) {
	const errorPrefix = "cannot authenticate to snap store: "

	var err error
	dischargeJSONData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf(errorPrefix+"%v", err)
	}

	resp, err := postRequestDecodeJSON(endpoint, dischargeJSONData)
	if err != nil {
		return "", fmt.Errorf(errorPrefix+"%v", err)
	}

	responseData := Discharge{}
	err = json.NewDecoder(resp.Body).Decode(&responseData)

	if responseData.Macaroon == "" {
		return "", fmt.Errorf(errorPrefix + "empty macaroon returned")
	}
	return responseData.Macaroon, nil
}

func postRequestDecodeJSON(url string, data []byte) (*http.Response, error) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error sending request: %v", err)
	}
	return resp, err
}

// AuthorizationHeader generates the authorization headers from the serialized macaroons
func AuthorizationHeader(macaroon, discharge string) (string, error) {
	var buf bytes.Buffer

	root, err := auth.MacaroonDeserialize(macaroon)
	if err != nil {
		log.Printf("Error deserializing macaroon: %v", err)
		return "", err
	}

	dischargeMacaroon, err := auth.MacaroonDeserialize(discharge)
	if err != nil {
		log.Printf("Error deserializing discharge: %v", err)
		return "", err
	}

	dischargeMacaroon.Bind(root.Signature())

	serializedMacaroon, err := auth.MacaroonSerialize(root)
	if err != nil {
		log.Printf("Error serializing root macaroon: %v", err)
		return "", err
	}
	serializedDischarge, err := auth.MacaroonSerialize(dischargeMacaroon)
	if err != nil {
		log.Printf("Error serializing discharge macaroon: %v", err)
		return "", err
	}

	fmt.Fprintf(&buf, `Macaroon root="%s", discharge="%s"`, serializedMacaroon, serializedDischarge)
	return buf.String(), nil
}
