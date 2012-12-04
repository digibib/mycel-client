package main

import (
	"encoding/json"
	"errors"
	"github.com/mattn/go-gtk/gtk"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
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
func identify(MAC string) (client *Client, err error) {
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

// Do local modifications to the client's environment
func localMods(screenRes string) {
	// 1. Screen Resolution
	xrandr, err := exec.Command("/usr/bin/xrandr").Output()
	if err != nil {
		//log.Fatal(err)
	}
	r, _ := regexp.Compile(`([\w]+)\sconnected`)
	display := r.FindSubmatch(xrandr)[1]
	cmd := exec.Command("/bin/sh", "-c", "/usr/bin/xrandr --output "+string(display)+" --mode "+screenRes)
	err = cmd.Run()
	if err != nil {
		print("DEBUG: xrandr change mode failed")
	}

	// 2. Firefox homepage
	// 3. Printer address

}

func main() {
	// Get the Mac-address of client
	eth0, err := ioutil.ReadFile("/sys/class/net/eth0/address")
	if err != nil {
		log.Fatal(err)
	}
	MAC := strings.TrimSpace(string(eth0))

	// Identify the client
	client, err := identify(MAC)
	if err != nil {
		log.Fatal(err)
	}

	// Do local mods (screenRes, Firefox Homepage, Printer adress)
	localMods("1600x900")

	// Show login screen
	gtk.Init(nil)
	user, password := windows.Login(client.Name)
	println(user, password)
}
