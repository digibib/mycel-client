package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"log/syslog"
	"math"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/gtk"
	"golang.org/x/net/websocket"

	"github.com/digibib/mycel-client/window"
)

const DefaultMinutes = 60

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
	Printers  []printer `json:"printers"`
}

// Printer struct to match JSON response from Mycel api/clients.
type printer struct {
	Id			  int			`json:"id"`
	Name			*string `json:"name"`
	PPD				*string `json:"ppd_client"`
	URI       *string `json:"uri_client"`
	Location	*string `json:"location"`
	Info			*string `json:"info"`
	Options		*string `json:"poptions"`
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
	DefaultPrinterId	*int `json:"default_printer_id"`
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
func identify(hostAPI, MAC string) (client *Client, err error) {
	url := fmt.Sprintf("%s/api/clients/?mac=%s", hostAPI, MAC)
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
func connect(hostWS, username string, client int) (conn *websocket.Conn) {
	// Request Mycel server to log in
	for {
		var err error
		conn, err = websocket.Dial(fmt.Sprintf("%s/subscribe/clients/%d", hostWS, client), "", "http://localhost")
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

func init() {
	log.SetFlags(0)
	syslogW, err := syslog.New(syslog.LOG_ERR, "mycel-client")
	if err != nil {
		log.Println("failed to initialize syslog writer: ", err)
	} else {
		log.SetOutput(io.MultiWriter(syslogW, os.Stderr))
	}
}

func setPrinters(hostAPI, MAC string) {
	// Reloads client info to catch any printer setting updates
	url := fmt.Sprintf("%s/api/clients/?mac=%s", hostAPI, MAC)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("failed to reload client info: ", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("failed to reload client info")
		return
	}
	r := new(response)
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		log.Println("failed to parse client info")
		return
	}

	var client *Client = &r.Client

	if client.Printers != nil {
		for _, printer := range client.Printers {
			pms := ""

			if (printer.Name != nil) {
				pms += " -p '" + *printer.Name + "' "
			}

			if (printer.Options != nil) {
				pms += *printer.Options
			}

			if (printer.PPD != nil) {
				pms += " -m '" + *printer.PPD + "'"
			}

			if (printer.URI != nil) {
				pms += " -v '" + *printer.URI + "'"
			}

			if (printer.Location != nil) {
				pms += " -L '" + *printer.Location + "'"
			}

			if (printer.Info != nil) {
				pms += " -D '" + *printer.Info + "'"
			}

			fmt.Println(pms)
			cmd := exec.Command("/bin/sh", "-c", "/usr/bin/sudo -n /usr/sbin/lpadmin" + pms)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Println("failed to set network printer address:", string(output))
			}

			if (client.Options.DefaultPrinterId != nil && printer.Id == *client.Options.DefaultPrinterId) {
				// set default
				cmd := exec.Command("/bin/sh", "-c", "/usr/bin/sudo -n /usr/bin/lpoptions -d " + *printer.Name)
				output, err := cmd.CombinedOutput()
				if err != nil {
					log.Println("failed to set default printer: ", string(output))
				}
			}
		}
		} else if client.Options.Printer != nil { // this can be removed once the new scheme is fully established
			cmd := exec.Command("/bin/sh", "-c", "/usr/bin/sudo -n /usr/sbin/lpadmin -p publikumsskriver -v "+*client.Options.Printer)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Println("failed to set network printer address:", string(output))
			}
		}
}



func main() {
	hostAPI := flag.String("api", "http://mycel:9000", "mycel host (api)")
	hostWS := flag.String("ws", "ws://mycel:9001", "mycel host (ws)")
	flag.Parse()

	// Get the Mac-address of client
	//eth0, err := ioutil.ReadFile("/sys/class/net/enp0s3/address")
	eth0, err := ioutil.ReadFile("/sys/class/net/eth0/address")
	if err != nil {
		log.Fatal(err)
	}
	MAC := strings.TrimSpace(string(eth0))

	// Identify the client
	var client *Client
	for {
		client, err = identify(*hostAPI, MAC)
		if err != nil {
			if err.Error() == "404 Not Found" {
				log.Fatal("client MAC address not found in mycel DB: ", MAC)
			}
			log.Println("Couldn't reach Mycel server. Trying again in 1 seconds...")
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	// Send hardware specs to server
	commands := map[string]string {
		"ram": "-t 19 | grep 'Range Size:' | awk {'print $3'}",
		"manufacturer": "-s system-manufacturer",
		"product_name": "-s system-product-name",
		"product_version": "-s system-version",
		"serial_number": "-s system-serial-number",
		"uuid": "-s system-uuid",
		"cpu_family": "-s processor-family",
	}

	sysinfo := map[string]string{}
	sysinfo["mac"] = MAC

	for key, params := range commands {
		cmd := exec.Command("/bin/sh", "-c", "/usr/bin/sudo -n /usr/sbin/dmidecode " + params)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Println("failed to gather hw specs: ", string(output))
		}

		sysinfo[key] = string(output)
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(sysinfo)

	url := fmt.Sprintf("%s/api/client_specs", *hostAPI)
	resp, err := http.Post(url, "application/json; charset=utf-8", b)
	if err != nil {
		log.Println("Failed to post hw specs")
	}

	resp.Body.Close()



	// Create thread to send live signals to server
	ticker := time.NewTicker(5 * time.Minute)
	quit := make(chan struct{})
	keep_alive :=  fmt.Sprintf("%s/api/keep_alive/?mac=%s", *hostAPI, MAC)

	go func() {
		for {
			select {
			case <- ticker.C:
				resp, _ := http.Get(keep_alive)
				resp.Body.Close()
			case <- quit:
				ticker.Stop()
				return
			}
		}
		}()


	// Do local modifications to the client's environment

	// 1. Screen Resolution
	if client.ScreenRes != "auto" {
		xrandr, err := exec.Command("/usr/bin/xrandr").Output()
		if err != nil {
			log.Println("failed to run xrandr", err)
		}
		rgx := regexp.MustCompile(`([\w]+)\sconnected`)
		display := rgx.FindSubmatch(xrandr)[1]
		cmd := exec.Command("/bin/sh", "-c", "/usr/bin/xrandr --output "+string(display)+" --mode "+client.ScreenRes)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Println("failed to set screen resolution: ", string(output))
		}
	}

	// 2. Firefox homepage
	if client.Options.Homepage != nil {
		escHomepage := strings.Replace(*client.Options.Homepage, `/`, `\/`, -1)

		sed := `/bin/sed -i 's/user_pref("browser.startup.homepage",.*/user_pref("browser.startup.homepage","` +
			escHomepage + `");/' $HOME/.mozilla/firefox/*.default/prefs.js`
		cmd := exec.Command("/bin/sh", "-c", sed)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Println("failed to set Firefox startpage: ", string(output))
		}
	}

	// Get today's closing time from client API response
	var hm string
	now := time.Now()
	switch now.Weekday() {
	case time.Monday:
		hm = *client.Options.Hours.MonCl
	case time.Tuesday:
		hm = *client.Options.Hours.TueCl
	case time.Wednesday:
		hm = *client.Options.Hours.WedCl
	case time.Thursday:
		hm = *client.Options.Hours.ThuCl
	case time.Friday:
		hm = *client.Options.Hours.FriCl
	case time.Saturday:
		hm = *client.Options.Hours.SatCl
	case time.Sunday:
		hm = *client.Options.Hours.SunCl
	}

	// Convert closing time to datetime
	hour, _ := strconv.Atoi(hm[0:2])
	min, _ := strconv.Atoi(hm[3:])
	closingTime := time.Date(now.Year(), now.Month(), now.Day(), hour, min-*client.Options.Hours.Min, 0, 0, time.Local)

	// Show login screen
	gtk.Init(nil)
	var user, userType string
	var userMinutes, extraMinutes int
	if client.ShortTime {
		userMinutes = *client.Options.ShortTimeLimit
		extraMinutes = 0
		user = window.ShortTime(client.Name, userMinutes)
	} else {
		extraMinutes = *client.Options.Minutes - DefaultMinutes
		user, userMinutes, userType = window.Login(*hostAPI, client.Name, extraMinutes, *client.Options.AgeL, *client.Options.AgeH)
		if userType == "G" {
			// If guest user, minutes is user.minutes left or the minutes limit on the client
			tempMinutes := int(math.Min(float64(userMinutes), float64(*client.Options.Minutes)))
			extraMinutes = tempMinutes - userMinutes
		}
	}

	// Calculate how long until closing time.
	// Adjust minutes acording to closing hours, so that maximum minutes does
	// not exceed available minutes until closing
	untilClose := int(closingTime.Sub(now).Minutes())
	if userMinutes+extraMinutes > untilClose {
		extraMinutes = untilClose - userMinutes
	}

	// Show status window
	conn := connect(*hostWS, user, client.Id)
	// User has logged - set printers
	setPrinters(*hostAPI, MAC)
	gdk.ThreadsInit()
	status := new(window.Status)

	status.Init(client.Name, user, userMinutes+extraMinutes)
	status.Show()
	status.Move()

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
					conn = connect(*hostWS, user, client.Id)
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
