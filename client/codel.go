package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"

	"github.com/lxc/lxd/shared/api"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh/terminal"
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

// user information taken from KeyCloak
type userInfo struct {
	ID            string `json:"sub"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Username      string `json:"preferred_username"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
	Email         string `json:"email"`
	UIDNumber     string `json:"uidNumber"`
}

// define data structure
type tokenJSON struct {
	AccessToken  string `json:"accessToken"`
	TokenType    string `json:"tokenType"`
	RefreshToken string `json:"refreshToken"`
	Expiry       string `json:"expiry"`
}

func main() {
	if len(os.Args) == 1 {
		defaultPrintout()
	} else {
		switch os.Args[1] {
		case "account":
			switch os.Args[2] {
			case "info":
				accountInfo()
			case "login":
				accountLogin()
			}
		case "images":
			switch os.Args[2] {
			case "info":
				imagesInfo()
			}
		}
	}
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

func check(e error, str string) bool {
	if e != nil {
		fmt.Println(str)
		panic(e)
	} else {
		return false
	}
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

func accountLogin() {

	usr, error := user.Current()
	check(error, "Could not get active user")

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	check(err, "\nCould not parse password")
	password := string(bytePassword)

	config := ssoConfiguration{}
	getJSON("https://sso.compsoc.ie/auth/realms/base", &config)

	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID: "codel",
		Scopes:   []string{"openid"},
		Endpoint: oauth2.Endpoint{
			TokenURL: config.TokenService + "/token",
		},
	}

	// Resource Owner Password Credentials
	token, error := conf.PasswordCredentialsToken(ctx, os.Args[2], password)

	if error != nil {
		fmt.Println(error)
		fmt.Println("\nUnfortunetly we could not log you in. Please try again.")
	} else {
		tokenJSON, error := json.Marshal(token)
		check(error, "\nUnfortunetly we could not parse the response from the server. Please try again.")

		// If the file doesn't exist, create it, or write to the file
		os.Mkdir(usr.HomeDir+"/.codel/", 0700)
		f, error := os.OpenFile(usr.HomeDir+"/.codel/token.json", os.O_CREATE|os.O_WRONLY, 0600)
		check(error, "\nUnfortunetly we could not open "+usr.HomeDir+"/.codel/token.json. Please try again.")

		_, error = f.Write(tokenJSON)
		if !check(error, "\nUnfortunetly we could write token to "+usr.HomeDir+"/.codel/token.json. Please try again.") {
			fmt.Print("\nSuccessfully logged in. Session saved to " + usr.HomeDir + "/.codel/token.json\n")
		}
	}
}

func accountInfo() {
	data := &tokenJSON{}
	target := &userInfo{}

	usr, error := user.Current()
	check(error, "Could not get active user")
	file, _ := ioutil.ReadFile(usr.HomeDir + "/.codel/token.json")
	_ = json.Unmarshal([]byte(file), &data)

	getJSONAuth("https://sso.compsoc.ie/auth/realms/base/protocol/openid-connect/userinfo", data, target)

	fmt.Println(target)
}

func imagesInfo() {
	data := &tokenJSON{}
	images := &([]api.Image{})

	usr, error := user.Current()
	check(error, "Could not get active user")
	file, _ := ioutil.ReadFile(usr.HomeDir + "/.codel/token.json")
	_ = json.Unmarshal([]byte(file), &data)

	getJSONAuth("https://hal.compsoc.ie:8081/listImages", data, images)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ALIAS", "DESCRIPTION", "SIZE"})
	table.SetAutoWrapText(false)

	for _, i := range *images {
		table.Append([]string{i.ImagePut.Properties["os"], i.ImagePut.Properties["description"], strconv.FormatInt(i.Size/1000000, 10) + " MiB"})
	}
	table.Render() // Send output
}

func defaultPrintout() {
	fmt.Println("Manage CompSoc account and services from the command line.")
	fmt.Println("\nUsage:\n  codel [command]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  account\tManage CompSoc Account")
	fmt.Println("  container\tManage container(s)")
	fmt.Println("\nFlags:")
	fmt.Println("  -d, --debug\tPrint debug messages")
	fmt.Println("  -h, --help\tHelp for [command]")
}
