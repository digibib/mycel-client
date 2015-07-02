package window

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"unsafe"

	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/gdkpixbuf"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
)

// response struct to match JSON response from api/users/authentication
type response struct {
	Age           int
	Authenticated bool
	Message       string
	Minutes       int
	Type          string
}

// authenticate returns a user struct response from the mycel API
// given a username and password
func authenticate(API_HOST, API_PORT, username, password string) (r *response, err error) {
	u := "http://" + API_HOST + ":" + API_PORT + "/api/users/authenticate"
	resp, err := http.PostForm(u, url.Values{"username": {username}, "password": {password}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	r = new(response)
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		return nil, err
	}
	return
}

// Login creates a GTK fullscreen window where users can log inn.
// It returns when a user successfully authenticates.
func Login(API_HOST, API_PORT, client string, extraMinutes, agel, ageh int) (user string, minutes int, userType string) {
	// Inital window configuration
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	defer window.Destroy()
	window.Fullscreen()
	window.SetKeepAbove(true)
	window.SetTitle("Mycel Login")

	// Build GUI
	frame := gtk.NewFrame("Logg deg på " + client)
	frame.SetLabelAlign(0.5, 0.5)
	var imageLoader *gdkpixbuf.Loader
	imageLoader, _ = gdkpixbuf.NewLoaderWithMimeType("image/png")
	imageLoader.Write(logo_png())
	imageLoader.Close()
	logo := gtk.NewImageFromPixbuf(imageLoader.GetPixbuf())
	button := gtk.NewButtonWithLabel("Logg inn")
	userlabel := gtk.NewLabel("Lånenummer/brukernavn")
	pinlabel := gtk.NewLabel("PIN-kode/passord")
	table := gtk.NewTable(3, 2, false)
	userentry := gtk.NewEntry()
	userentry.SetMaxLength(10)
	userentry.SetSizeRequest(150, 23)
	pinentry := gtk.NewEntry()
	pinentry.SetVisibility(false)
	pinentry.SetMaxLength(10)

	table.Attach(userlabel, 0, 1, 0, 1, gtk.FILL, gtk.FILL, 7, 5)
	table.Attach(userentry, 1, 2, 0, 1, gtk.FILL, gtk.FILL, 7, 5)
	table.Attach(pinlabel, 0, 1, 1, 2, gtk.FILL, gtk.FILL, 7, 5)
	table.Attach(pinentry, 1, 2, 1, 2, gtk.FILL, gtk.FILL, 7, 5)
	table.Attach(button, 1, 2, 2, 3, gtk.FILL, gtk.FILL, 7, 5)

	error := gtk.NewLabel("")

	vbox := gtk.NewVBox(false, 20)
	vbox.SetBorderWidth(20)
	vbox.Add(logo)
	vbox.Add(table)
	vbox.Add(error)

	frame.Add(vbox)

	center := gtk.NewAlignment(0.5, 0.5, 0, 0)
	center.Add(frame)
	window.Add(center)

	// Functions to validate and check responses
	checkResponse := func(username, password string) {
		user, err := authenticate(API_HOST, API_PORT, username, password)
		if err != nil {
			println("DEBUG: call to api/users/authenticate failed")
			//error.SetMarkup("<span foreground='red'>Fikk ikke kontakt med server, vennligst prøv igjen!</span>")
			error.SetMarkup("<span foreground='red'>Feil lånenummer/brukernavn eller PIN/passord</span>")
			return
		}
		if !user.Authenticated {
			error.SetMarkup("<span foreground='red'>" + user.Message + "</span>")
			return
		}
		if user.Minutes+extraMinutes <= 0 && user.Type != "G" {
			error.SetMarkup("<span foreground='red'>Beklager, du har brukt opp kvoten din for i dag!</span>")
			return
		}
		if user.Type == "G" && user.Minutes <= 0 {
			error.SetMarkup("<span foreground='red'>Beklager, du har brukt opp kvoten din for i dag!</span>")
			return
		}
		if user.Age < agel || user.Age > ageh {
			error.SetMarkup("<span foreground='red'>Denne maskinen er kun for de mellom " +
				strconv.Itoa(agel) + " og " + strconv.Itoa(ageh) + "</span>")
			return
		}
		userType = user.Type
		minutes = user.Minutes
		gtk.MainQuit()
		return
	}
	validate := func(ctx *glib.CallbackContext) {
		arg := ctx.Args(0)
		kev := *(**gdk.EventKey)(unsafe.Pointer(&arg))
		username := userentry.GetText()
		password := pinentry.GetText()
		if kev.Keyval == gdk.KEY_Return {
			if username == "" && password == "" {
				return
			}
			if username != "" && password == "" {
				pinentry.GrabFocus()
				return
			}
			if password != "" && username == "" {
				userentry.GrabFocus()
				return
			}
			checkResponse(username, password)
			return
		}
	}

	// Connect GUI event signals to function callbacks
	pinentry.Connect("key-press-event", validate)
	userentry.Connect("key-press-event", validate)
	button.Connect("clicked", func() {
		username := userentry.GetText()
		password := pinentry.GetText()
		if (username == "") || (password == "") {
			error.SetMarkup("<span foreground='red'>Skriv inn ditt lånenummer og PIN-kode</span>")
			userentry.GrabFocus()
			return
		}
		checkResponse(username, password)
	})
	window.Connect("delete-event", func() bool {
		return true
	})

	window.ShowAll()
	gtk.Main()
	user = userentry.GetText()
	return
}
