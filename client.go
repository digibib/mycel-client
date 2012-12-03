package main

import (
	"encoding/json"
	"errors"
	"github.com/mattn/go-gtk/gtk"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/digibib/mycel-client/windows"
)

type response struct {
	Client Client
}

type Client struct {
	Id        int
	Name      string
	ScreenRes string `json:"screen_resolution"`
	ShortTime bool
	Options   options `json:"options_inherited"`
}

// These fields must be pointers, in case of null value from JSON
// When dereferencing check for nil pointers
type options struct {
	AgeL     *int    `json:"age_limit_lower"`
	AgeH     *int    `json:"age_limit_higher"`
	Printer  *string `json:"printeraddr"`
	Homepage *string
}

// Identifies the client from the mac-address and return Client struct
func Identify(MAC string) (client *Client, err error) {
	url := "http://localhost:9000/api/clients/?mac=" + MAC
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	r := new(response)
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		return nil, err
	}
	return &r.Client, nil
}

func main() {
	// Get the Mac-address of client
	eth0, err := ioutil.ReadFile("/sys/class/net/eth0/address")
	if err != nil {
		log.Fatal(err)
	}
	MAC := strings.TrimSpace(string(eth0))

	// Identify the client
	client, err := Identify(MAC)
	if err != nil {
		log.Fatal(err)
	}

	// Do local modifications to environment
	// 1. Screen Resolution
	// 2. Firefox homepage
	// 3. Printer address

	// Show login screen
	gtk.Init(nil)
	user, password := windows.Login(client.Name)
	println(user, password)
}
