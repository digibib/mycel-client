package window

import (
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/gtk"
	"strconv"
)

// Status struct represents the status window shown when users are logged in.
type Status struct {
	window    *gtk.Window
	client    string
	user      string
	minutes   int
	warned    bool
	timeLabel *gtk.Label
}

// Init acts as a constructor for the Status window struct
func (v *Status) Init(client, user string, minutes int) {

	// Initialize variables
	v.client = client
	v.user = user
	v.minutes = minutes
	v.warned = false
	v.window = gtk.NewWindow(gtk.WINDOW_TOPLEVEL)

	// Inital Window configuration
	v.window.SetKeepAbove(true)
	v.window.SetTitle(client)
	v.window.SetTypeHint(gdk.WINDOW_TYPE_HINT_MENU)
	v.window.SetSizeRequest(200, 180)
	v.window.SetResizable(false)

	// Build GUI
	userLabel := gtk.NewLabel(user)
	v.timeLabel = gtk.NewLabel("")
	v.timeLabel.SetMarkup("<span size='xx-large'>" + strconv.Itoa(v.minutes) + " min igjen</span>")
	button := gtk.NewButtonWithLabel("Logg ut")

	vbox := gtk.NewVBox(false, 20)
	vbox.SetBorderWidth(5)
	vbox.Add(userLabel)
	vbox.Add(v.timeLabel)
	vbox.Add(button)
	v.window.Add(vbox)

	// Connect GUI event signals to function callbacks
	v.window.Connect("delete-event", func() bool {
		// Don't allow user to quit by closing the window
		return true
	})
	button.Connect("clicked", func() {
		gtk.MainQuit()
	})

	// Position the window in lower right corner
	// TODO implement gtk.SetGravity()
	//window.SetGravity Gdk::Window::GRAVITY_SOUTH_WEST
	//window.Move (Gdk.screen_width - size[0] - 50), (Gdk.screen_height - size[1] - 50)
	return
}

func (v *Status) Show() {
	v.window.ShowAll()
	return
}

func (v *Status) SetMinutes(minutes int) {
	var bg string
	if minutes <= 5 {
		bg = "yellow"
	} else {
		bg = "#e0e0e0"
	}
	v.timeLabel.SetMarkup("<span background='" + bg + "' size='xx-large'>" + strconv.Itoa(minutes) + " min igjen</span>")

	if minutes <= 5 && v.warned == false {
		msg := "Du blir logget av om " + strconv.Itoa(minutes) + " minutter. Husk å lagre det du jobber med!"
		md := gtk.NewMessageDialog(v.window.GetTopLevelAsWindow(), gtk.DIALOG_MODAL,
			gtk.MESSAGE_WARNING, gtk.BUTTONS_OK, msg)
		md.SetTypeHint(gdk.WINDOW_TYPE_HINT_MENU)
		md.Connect("response", func() {
			md.Destroy()
		})
		md.ShowAll()
		v.warned = true
	}

	if minutes > 5 {
		v.warned = false
	}
}
