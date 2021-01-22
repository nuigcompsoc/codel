package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"time"

	"golang.org/x/oauth2"
)

// realm information taken from KeyCloak
type ssoConfiguration struct {
	Realm           string `json:"realm"`
	PublicKey       string `json:"public_key"`
	TokenService    string `json:"token-service"`
	AccountService  string `json:"account-service"`
	TokensNotBefore string `json:"tokens-not-before"`
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

func getJSON(url string, target interface{}) error {
	myClient := &http.Client{Timeout: 10 * time.Second}
	response, err := myClient.Get(url)
	if err != nil {
		return err
	}

	json.NewDecoder(response.Body).Decode(target)

	return nil
}

func getHTTPClient() (*http.Client, error) {
	usr, err := user.Current()
	file, err := ioutil.ReadFile(usr.HomeDir + "/.codel/token.json")
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{}
	err = json.Unmarshal([]byte(file), token)

	if err != nil {
		return nil, err
	}

	config := ssoConfiguration{}
	getJSON("https://sso.compsoc.ie/auth/realms/base", &config)

	conf := &oauth2.Config{
		ClientID: "codel",
		Scopes:   []string{"openid"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.TokenService + "/auth",
			TokenURL: config.TokenService + "/token",
		},
	}

	tokenSource := conf.TokenSource(context.Background(), token)
	token, err = tokenSource.Token()

	f, err := os.OpenFile(usr.HomeDir+"/.codel/token.json", os.O_WRONLY, 0600)
	tokenJSON, err := json.Marshal(token)
	_, err = f.Write(tokenJSON)

	if err != nil {
		return nil, err
	}

	return conf.Client(context.Background(), token), nil
}
