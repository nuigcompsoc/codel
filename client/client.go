package main

import (
	"context"
	"encoding/json"
	"fmt"
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

// user information taken from KeyCloak
type userInfo struct {
	ID            string `json:"sub"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Username      string `json:"preferred_username"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
	Email         string `json:"email"`
	UIDNumber     int    `json:"uidNumber"`
}

var hostURL = "http://localhost:8081"

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
		case "containerd":
			if len(os.Args) >= 2 {
				switch os.Args[2] {
				case "list":
					switch os.Args[3] {
					case "all":
						containerdListAll()
					}
				case "images":
					switch os.Args[3] {
					case "list":
						containerdImagesList()
					}
				}
			} else {
				containerdPrintout()
			}
		}
	}
}

/*
 *
 * CONTAINERD
 *
 */

func containerdPrintout() {
	fmt.Println("Usage:\n  codel containerd [command]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  images\tList available container images")
	fmt.Println("  list\t\tList containerd")
}

func containerdListAll() {
	containers := &([]api.Container{})

	c, err := getHTTPClient()
	if err != nil {
		panic(err)
	}

	resp, err := c.Get(hostURL + "/listContainers")
	if err != nil {
		panic(err)
	}

	json.NewDecoder(resp.Body).Decode(containers)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "OS", "STATUS", "LAST SEEN"})
	table.SetAutoWrapText(false)

	for _, i := range *containers {
		table.Append([]string{i.Name, i.ContainerPut.Config["image.os"], i.Status, i.LastUsedAt.UTC().Format(time.UnixDate)})
	}

	table.Render() // Send output
}

func containerdImagesList() {
	images := &([]api.Image{})

	c, err := getHTTPClient()
	if err != nil {
		panic(err)
	}

	resp, err := c.Get(hostURL + "/listImages")
	if err != nil {
		panic(err)
	}

	json.NewDecoder(resp.Body).Decode(images)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ALIAS", "DESCRIPTION", "SIZE"})
	table.SetAutoWrapText(false)

	for _, i := range *images {
		table.Append([]string{i.ImagePut.Properties["os"], i.ImagePut.Properties["description"], strconv.FormatInt(i.Size/1000000, 10) + " MiB"})
	}
	table.Render() // Send output
}

/*
 *
 * ACCOUNT
 *
 */

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
	token, error := conf.PasswordCredentialsToken(ctx, os.Args[3], password)

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
	userInfo := &userInfo{}

	c, err := getHTTPClient()
	if err != nil {
		panic(err)
	}

	resp, err := c.Get("https://sso.compsoc.ie/auth/realms/base/protocol/openid-connect/userinfo")
	if err != nil {
		panic(err)
	}

	json.NewDecoder(resp.Body).Decode(userInfo)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"UIDNUMBER", "UID", "FIRST NAME", "LAST NAME", "EMAIL"})
	table.SetAutoWrapText(false)
	table.Append([]string{strconv.Itoa(userInfo.UIDNumber), userInfo.Username, userInfo.FirstName, userInfo.LastName, userInfo.Email})
	table.Render() // Send output
}

/*
 *
 * MISC
 *
 */

func defaultPrintout() {
	fmt.Println("Manage CompSoc account and services from the command line.")
	fmt.Println("\nUsage:\n  codel [command]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  account\tManage CompSoc Account")
	fmt.Println("  containerd\tManage container(s)")
	fmt.Println("\nFlags:")
	fmt.Println("  -d, --debug\tPrint debug messages")
	fmt.Println("  -h, --help\tHelp for [command]")
}

func check(e error, str string) bool {
	if e != nil {
		fmt.Println(str)
		panic(e)
	} else {
		return false
	}
}
