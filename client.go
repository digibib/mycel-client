package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/gtk"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"github.com/digibib/mycel-client/window"
)

type response struct {
	Client Client
}

// Client struct to match JSON response from Mycel api/clients.
type Client struct {
	Id        int
	Name      string
	ScreenRes string `json:"screen_resolution"`
	ShortTime bool
	Options   options `json:"options_inherited"`
}

// These fields must be pointers, in case of null value from JSON
// When dereferencing check for nil pointers.
type options struct {
	AgeL     *int    `json:"age_limit_lower"`
	AgeH     *int    `json:"age_limit_higher"`
	Printer  *string `json:"printeraddr"`
	Homepage *string
	Minutes  *int `json:"time_limit"`
}

// logOnMessage represent JSON message to request user to log on client
type logOnMessage struct {
	Action string `json:"action"`
	Client int    `json:"client"`
	User   string `json:"user"`
}

// message struct represents all websocket JSON messages other than log-on message
// TODO rethink this, do I need all fields?
type message struct {
	Status string  `json:"status"`
	User   msgUser `json:"user"`
}

type msgUser struct {
	Username string `json:"username"`
	Minutes  int    `json:"minutes"`
}

// identify sends the client's mac-address to the Mycel API and returns a Client struct.
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

// localMods makes modifications to the client's environment.
func localMods(screenRes, homepage, printer string) {
	// 1. Screen Resolution
	xrandr, err := exec.Command("/usr/bin/xrandr").Output()
	if err != nil {
		println("DEBUG: couldn't find or access xrandr")
	}
	r, _ := regexp.Compile(`([\w]+)\sconnected`)
	display := r.FindSubmatch(xrandr)[1]
	cmd := exec.Command("/bin/sh", "-c", "/usr/bin/xrandr --output "+string(display)+" --mode "+screenRes)
	err = cmd.Run()
	if err != nil {
		println("DEBUG: xrandr change mode failed")
	}

	// 2. Firefox homepage
	escHomepage := strings.Replace(homepage, `/`, `\/`, -1)

	sed := `/bin/sed -i 's/user_pref("browser.startup.homepage",.*/user_pref("browser.startup.homepage","` +
		escHomepage + `");/' $HOME/.mozilla/firefox/*.default/prefs.js`
	cmd = exec.Command("/bin/sh", "-c", sed)
	err = cmd.Run()
	if err != nil {
		println("DEBUG: failed to set Firefox startpage")
	}

	// 3. Printer address
	cmd = exec.Command("/bin/sh", "-c", "/usr/bin/sudo -n /usr/sbin/lpadmin -p skranken -v "+printer)
	err = cmd.Run()
	if err != nil {
		println("DEBUG: failed to set network printer address")
	}
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
	//localMods("1600x900", "http://morgenbladet.no", "socket://10.172.2.31:9000")

	// Show login screen
	gtk.Init(nil)
	user, minutes := window.Login(client.Name, *client.Options.Minutes-60, *client.Options.AgeL, *client.Options.AgeH)

	// Request Mycel server to log in
	conn, err := websocket.Dial(fmt.Sprintf("ws://%s:%d/subscribe/clients/%d", HOST, PORT, client.Id), "", "http://localhost")
	if err != nil {
		panic(err)
	}
	logonMsg := logOnMessage{Action: "log-on", Client: client.Id, User: user}
	err = websocket.JSON.Send(conn, logonMsg)
	if err != nil {
		fmt.Println("Couldn't send message " + err.Error())
	}
	// Wait for "logged-on" confirmation
	var msg message
	for {
		err := websocket.JSON.Receive(conn, &msg)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Couldn't receive msg " + err.Error())
		}

		if msg.Status == "logged-on" {
			break
		}
	}

	gdk.ThreadsInit()
	status := new(window.Status)
	extraMinutes := *client.Options.Minutes - 60
	status.Init(client.Name, user, minutes+extraMinutes)
	status.Show()
	// goroutine to check for websocket messages and update status window
	// with number of minutes leeft
	go func() {
		for {
			err := websocket.JSON.Receive(conn, &msg)
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("Couldn't receive msg " + err.Error())
			}

			if msg.Status == "ping" {
				gdk.ThreadsEnter()
				status.SetMinutes(msg.User.Minutes + extraMinutes)
				gdk.ThreadsLeave()
			}
		}
	}()
	gtk.Main()
}
