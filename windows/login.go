package windows

import (
	"encoding/json"
	"errors"
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"net/http"
	"net/url"
	"unsafe"
)

// response struct to match JSON response from api/users/authentication
type response struct {
	Age           int
	Authenticated bool
	Message       string
	Minutes       int
}

// authenticate returns a user struct response from the mycel API
// given a username and password
func authenticate(username, password string) (r *response, err error) {
	u := "http://localhost:9000/api/users/authenticate"
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
func Login(client string) (user, password string) {
	// Inital window configuration
	window := gtk.Window(gtk.GTK_WINDOW_TOPLEVEL)
	window.Fullscreen()
	window.SetKeepAbove(true)
	window.SetTitle("Mycel Login")

	// Buid GUI
	frame := gtk.Frame("Logg deg på " + client)
	frame.SetLabelAlign(0.5, 0.5)
	logo := gtk.ImageFromFile("logo.png")
	button := gtk.ButtonWithLabel("Log inn")
	userlabel := gtk.Label("Lånenummer/brukernavn")
	pinlabel := gtk.Label("PIN-kode/passord")
	table := gtk.Table(3, 2, false)
	userentry := gtk.Entry()
	userentry.SetMaxLength(10)
	userentry.SetSizeRequest(150, 23)
	pinentry := gtk.Entry()
	pinentry.SetVisibility(false)
	pinentry.SetMaxLength(10)

	table.Attach(userlabel, 0, 1, 0, 1, gtk.GTK_FILL, gtk.GTK_FILL, 7, 5)
	table.Attach(userentry, 1, 2, 0, 1, gtk.GTK_FILL, gtk.GTK_FILL, 7, 5)
	table.Attach(pinlabel, 0, 1, 1, 2, gtk.GTK_FILL, gtk.GTK_FILL, 7, 5)
	table.Attach(pinentry, 1, 2, 1, 2, gtk.GTK_FILL, gtk.GTK_FILL, 7, 5)
	table.Attach(button, 1, 2, 2, 3, gtk.GTK_FILL, gtk.GTK_FILL, 7, 5)

	error := gtk.Label("")

	vbox := gtk.VBox(false, 20)
	vbox.SetBorderWidth(20)
	vbox.Add(logo)
	vbox.Add(table)
	vbox.Add(error)

	frame.Add(vbox)

	center := gtk.Alignment(0.5, 0.5, 0, 0)
	center.Add(frame)
	window.Add(center)

	// Connect signal callbacks (GUI events)
	window.Connect("destroy", func() {
		println("quitting..")
		gtk.MainQuit()
	})
	checkResponse := func(username, password string) {
		user, err := authenticate(username, password)
		if err != nil {
			println("DEBUG: authentication call failed")
			error.SetMarkup("<span foreground='red'>Fikk ikke kontakt med server, vennligst prøv igjen!</span>")
			return
		}
		if user.Authenticated {
			gtk.MainQuit()
		} else {
			error.SetMarkup("<span foreground='red'>" + user.Message + "</span>")
		}
	}
	validate := func(ctx *glib.CallbackContext) {
		arg := ctx.Args(0)
		kev := *(**gdk.EventKey)(unsafe.Pointer(&arg))
		username := userentry.GetText()
		password := pinentry.GetText()
		if kev.Keyval == gdk.GDK_KEY_Return {
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
		}

	}
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

	window.ShowAll()
	gtk.Main()
	return
}
