package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/user"

	"github.com/gorilla/mux"
	lxd "github.com/lxc/lxd/client"
	//"github.com/lxc/lxd/shared/api"
)

var lxcCurrentVersion = "19009"

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/listContainers", CheckUserIsAdminUserAuthMiddleware(listContainers))
	myRouter.HandleFunc("/listImages", AuthMiddleware(listImages))
	//log.Fatal(http.ListenAndServeTLS(":8081", "/etc/ssl/compsoc/star_compsoc_ie_crt_and_chain.crt", "/etc/ssl/compsoc/star_compsoc_ie.key", myRouter))
	log.Fatal(http.ListenAndServe(":8081", myRouter))
}

func main() {
	fmt.Println("I AM ALIVE! X)")
	handleRequests()
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage Endpoint Hit")
}

func lxcBind() (lxd.InstanceServer, error) {
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
	}

	TLSServerCert, err := ioutil.ReadFile(usr.HomeDir + "/snap/lxd/" + lxcCurrentVersion + "/.config/lxc/servercerts/hal.crt")
	TLSClientCert, err := ioutil.ReadFile(usr.HomeDir + "/snap/lxd/" + lxcCurrentVersion + "/.config/lxc/client.crt")
	TLSClientKey, err := ioutil.ReadFile(usr.HomeDir + "/snap/lxd/" + lxcCurrentVersion + "/.config/lxc/client.key")

	connectionArgs := &lxd.ConnectionArgs{
		TLSServerCert: string(TLSServerCert),
		TLSClientCert: string(TLSClientCert),
		TLSClientKey:  string(TLSClientKey),
	}

	return lxd.ConnectLXD("https://hal.compsoc.ie:8000", connectionArgs)
}

func listContainers(w http.ResponseWriter, r *http.Request) {
	c, err := lxcBind()
	if err != nil {
		fmt.Println(err)
	}

	instanceArray, err := c.GetInstances("container")
	if err != nil {
		fmt.Println(err)
	}

	jsonData, err := json.Marshal(instanceArray)
	fmt.Fprintf(w, string(jsonData))
}

func listImages(w http.ResponseWriter, r *http.Request) { // ListImages for /listimages

	c, err := lxcBind()
	if err != nil {
		fmt.Println(err)
	}

	allImages, err := c.GetImages()
	if err != nil {
		fmt.Println(err)
	}

	jsonData, err := json.Marshal(allImages)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Fprintf(w, string(jsonData))
}
