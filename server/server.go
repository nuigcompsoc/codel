package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	lxd "github.com/lxc/lxd/client"
	//"github.com/lxc/lxd/shared/api"
)

var lxcSocketLocation = "/var/snap/lxd/common/lxd/unix.socket"

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage Endpoint Hit")
}

func listContainers(w http.ResponseWriter, r *http.Request) {

	c, err := lxd.ConnectLXDUnix(lxcSocketLocation, nil)
	if err != nil {
		fmt.Println(err)
	}

	instanceFullArray, err := c.GetInstancesFull("container")
	if err != nil {
		fmt.Println(err)
	}

	jsonData, err := json.MarshalIndent(instanceFullArray, "", "    ")
	fmt.Fprintf(w, string(jsonData))
}

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

func listImages(w http.ResponseWriter, r *http.Request) { // ListImages for /listimages

	c, err := lxd.ConnectLXDUnix("/var/snap/lxd/common/lxd/unix.socket", nil)
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
