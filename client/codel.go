package main

import (
	"context"
	"golang.org/x/oauth2"
	"encoding/json"
	"os/user"
	"os"
    "fmt"
	"syscall"
	
	"golang.org/x/crypto/ssh/terminal"
	"github.com/tkanos/gonfig"
)

type Configuration struct {
	TokenURL	string
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func main() {
	switch (os.Args[1]) {
	case "login":
		auth()
	}

}

func auth() {
	usr, error := user.Current()
	check(error)

	config := Configuration{}
	err := gonfig.GetConf(usr.HomeDir + "/.codel/config.json", &config)
	check(err)

	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     "codel",
		Scopes:       []string{"openid"},
		Endpoint: oauth2.Endpoint{
			TokenURL: config.TokenURL,
		},
	}

    fmt.Print("Enter Password: ")
    bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
    check(err)
    password := string(bytePassword)

	// Resource Owner Password
	token, error := conf.PasswordCredentialsToken(ctx, os.Args[2], password)
	check(error)

	tokenJSON, error := json.Marshal(token)
	check(error)

	// If the file doesn't exist, create it, or append to the file
	os.Mkdir(usr.HomeDir + "/.codel/", 0700)
	f, error := os.OpenFile(usr.HomeDir + "/.codel/token.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	check(error)
	
	_, error = f.Write(tokenJSON)
	check(error)
}
