package main

import (
	"encoding/json"
	"net/http"
	"os"
	"os/user"
	"time"
)

// realm information taken from KeyCloak
type ssoConfiguration struct {
	Realm           string `json:"realm"`
	PublicKey       string `json:"public_key"`
	TokenService    string `json:"token-service"`
	AccountService  string `json:"account-service"`
	TokensNotBefore string `json:"tokens-not-before"`
}

// define data structure
type tokenJSON struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Expiry       string `json:"expiry"`
}

var myClient = &http.Client{Timeout: 10 * time.Second}

func getJSON(url string, target interface{}) error {
	response, err := myClient.Get(url)
	if err != nil {
		return err
	}

	json.NewDecoder(response.Body).Decode(target)

	return nil
}

func getJSONAuth(url string, data *tokenJSON, target interface{}) error {

	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add("Authorization", "Bearer "+data.AccessToken)

	response, err := myClient.Do(request)

	if err != nil {
		return err
	}

	json.NewDecoder(response.Body).Decode(target)

	return nil
}

func verifyTokenExists() bool {
	usr, error := user.Current()
	check(error, "Could not get active user directory")

	_, error = os.Stat(usr.HomeDir + "/.codel/token.json")
	if os.IsNotExist(error) {
		return false
	}

	return true
}
