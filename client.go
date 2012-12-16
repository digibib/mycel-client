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
	"strconv"
	"strings"
	"time"

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
	Hours          *oh     `json:"opening_hours"`
	AgeL           *int    `json:"age_limit_lower"`
	AgeH           *int    `json:"age_limit_higher"`
	Minutes        *int    `json:"time_limit"`
	ShortTimeLimit *int    `json:"shorttime_limit"`
	Printer        *string `json:"printeraddr"`
	Homepage       *string
}

// Client opening hours
type oh struct {
	MonOp *string `json:"monday_opens"`
	MonCl *string `json:"monday_closes"`
	MonX  *bool   `json:"monday_closed"`
	TueOp *string `json:"tuesday_opens"`
	TueCl *string `json:"tuesday_closes"`
	TueX  *bool   `json:"tuesday_closed"`
	SedOp *string `json:"wednsday_opens"`
	WedCl *string `json:"wednsday_closes"`
	WedX  *bool   `json:"wednsday_closed"`
	ThuOp *string `json:"thursday_opens"`
	ThuCl *string `json:"thursday_closes"`
	ThuX  *bool   `json:"thursday_closed"`
	FriOp *string `json:"friday_opens"`
	FriCl *string `json:"friday_closes"`
	FriX  *bool   `json:"friday_closed"`
	SatOp *string `json:"saturday_opens"`
	SatCl *string `json:"saturday_closes"`
	SatX  *bool   `json:"saturday_closed"`
	SunOp *string `json:"sunday_opens"`
	SunCl *string `json:"sunday_closes"`
	SunX  *bool   `json:"sunday_closed"`
	Min   *int    `json:"minutes_before_closing"`
}

// logOnOffMessage represent JSON message to request user to log on/off client
type logOnOffMessage struct {
	Action string `json:"action"`
	Client int    `json:"client"`
	User   string `json:"user"`
}

// message struct represents all websocket JSON messages other than log-on message
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
	url := fmt.Sprintf("http://%s:%s/api/clients/?mac=%s", API_HOST, API_PORT, MAC)
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

// connect logs on user. Blocks until successfull and
// returns the websocket connection
func connect(username string, client int) (conn *websocket.Conn) {
	// Request Mycel server to log in
	for {
		var err error
		conn, err = websocket.Dial(fmt.Sprintf("ws://%s:%s/subscribe/clients/%d", HOST, PORT, client), "", "http://localhost")
		if err != nil {
			fmt.Println("Can't connect to Mycel websocket server. Trying reconnect in 1 second...")
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	// Create and send log-on request
	logonMsg := logOnOffMessage{Action: "log-on", Client: client, User: username}
	err := websocket.JSON.Send(conn, logonMsg)
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
	return
}

func shutdown() {
	// Department is closed; force shutdown!
	cmd := exec.Command("/bin/sh", "-c", "sudo shutdown -P now")
	err := cmd.Run()
	if err != nil {
		// Do nothing
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
	var client *Client
	for {
		client, err = identify(MAC)
		if err != nil {
			fmt.Println("Couldn't reach Mycel server. Trying again in 1 second...")
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	// Do local modifications to the client's environment

	// 1. Screen Resolution
	if client.ScreenRes != "auto" {
		xrandr, err := exec.Command("/usr/bin/xrandr").Output()
		if err != nil {
			println("DEBUG: couldn't find or access xrandr")
		}
		r, _ := regexp.Compile(`([\w]+)\sconnected`)
		display := r.FindSubmatch(xrandr)[1]
		cmd := exec.Command("/bin/sh", "-c", "/usr/bin/xrandr --output "+string(display)+" --mode "+client.ScreenRes)
		err = cmd.Run()
		if err != nil {
			println("DEBUG: xrandr change mode failed")
		}
	}

	// 2. Firefox homepage
	if client.Options.Homepage != nil {
		escHomepage := strings.Replace(*client.Options.Homepage, `/`, `\/`, -1)

		sed := `/bin/sed -i 's/user_pref("browser.startup.homepage",.*/user_pref("browser.startup.homepage","` +
			escHomepage + `");/' $HOME/.mozilla/firefox/*.default/prefs.js`
		cmd := exec.Command("/bin/sh", "-c", sed)
		err = cmd.Run()
		if err != nil {
			println("DEBUG: failed to set Firefox startpage")
		}
	}

	// 3. Printer address
	if client.Options.Printer != nil {
		cmd := exec.Command("/bin/sh", "-c", "/usr/bin/sudo -n /usr/sbin/lpadmin -p skranken -v "+*client.Options.Printer)
		err = cmd.Run()
		if err != nil {
			println("DEBUG: failed to set network printer address")
		}
	}

	// Get today's closing time from client API response, and force a shutdown
	// if the department is closed that day
	var hm string
	now := time.Now()
	switch now.Weekday() {
	case time.Monday:
		if *client.Options.Hours.MonX {
			shutdown()
		}
		hm = *client.Options.Hours.MonCl
	case time.Tuesday:
		if *client.Options.Hours.TueX {
			shutdown()
		}
		hm = *client.Options.Hours.TueCl
	case time.Wednesday:
		if *client.Options.Hours.WedX {
			shutdown()
		}
		hm = *client.Options.Hours.WedCl
	case time.Thursday:
		if *client.Options.Hours.ThuX {
			shutdown()
		}
		hm = *client.Options.Hours.ThuCl
	case time.Friday:
		if *client.Options.Hours.FriX {
			shutdown()
		}
		hm = *client.Options.Hours.FriCl
	case time.Saturday:
		if *client.Options.Hours.SatX {
			shutdown()
		}
		hm = *client.Options.Hours.SatCl
	case time.Sunday:
		if *client.Options.Hours.SunX {
			shutdown()
		}
		hm = *client.Options.Hours.SunCl
	}

	// Convert closing time to datetime, and shut down if allready closed
	hour, _ := strconv.Atoi(hm[0:2])
	min, _ := strconv.Atoi(hm[3:])
	closingTime := time.Date(now.Year(), now.Month(), now.Day(), hour, min-*client.Options.Hours.Min, 0, 0, time.Local)
	if now.After(closingTime) {
		shutdown()
	}

	// Show login screen
	gtk.Init(nil)
	var user string
	var userMinutes, extraMinutes int
	if client.ShortTime {
		userMinutes = *client.Options.ShortTimeLimit
		extraMinutes = 0
		user = window.ShortTime(client.Name, userMinutes)
	} else {
		user, userMinutes = window.Login(API_HOST, API_PORT, client.Name, *client.Options.Minutes-60, *client.Options.AgeL, *client.Options.AgeH)
		extraMinutes = *client.Options.Minutes - 60
	}

	// Calculate how long until closing time.
	// Adjust minutes acording to closing hours, so that maximum minutes does
	// not exceed available minutes until closing
	untilClose := int(closingTime.Sub(now).Minutes())
	if userMinutes+extraMinutes > untilClose {
		extraMinutes = untilClose - userMinutes
	}
	// This function will trigger and shutdown at closingTime, if the client is
	// still running.
	time.AfterFunc(closingTime.Sub(now), func() {
		shutdown()
	})

	// Show status window
	conn := connect(user, client.Id)
	gdk.ThreadsInit()
	status := new(window.Status)

	status.Init(client.Name, user, userMinutes+extraMinutes)
	status.Show()

	// goroutine to check for websocket messages and update status window
	// with number of minutes left
	go func() {
		var msg message
		for {
			err := websocket.JSON.Receive(conn, &msg)
			if err != nil {
				if err == io.EOF {
					println("ws disconnected")
					// reconnect
					conn = connect(user, client.Id)
				}
				continue
			}

			if msg.Status == "ping" {
				if msg.User.Minutes+extraMinutes <= 0 {
					gtk.MainQuit()
				}
				gdk.ThreadsEnter()
				status.SetMinutes(msg.User.Minutes + extraMinutes)
				gdk.ThreadsLeave()
			}
		}
	}()

	// This blocks until the 'logg out' button is clicked, or until the user
	// has spent all minutes
	gtk.Main()

	// Send log-out message to server
	logOffMsg := logOnOffMessage{Action: "log-off", Client: client.Id, User: user}
	err = websocket.JSON.Send(conn, logOffMsg)
	if err != nil {
		// Don't bother to resend. Server will log off user anyway, when the
		// connection is closed
	}

	// Force session restart
	cmd := exec.Command("/bin/sh", "-c", "/usr/bin/killall /usr/bin/lxsession")
	err = cmd.Run()
}
