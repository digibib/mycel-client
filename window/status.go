package window

import (
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/gtk"
)

type Status struct {
	window  *gtk.GtkWindow
	title   string
	user    string
	minutes int
}

func (v *Status) Init(client, user string, minutes int) {
	// Inital Window configuration
	v.window = gtk.Window(gtk.GTK_WINDOW_TOPLEVEL)
	v.window.SetKeepAbove(true)
	v.window.SetTitle(client)
	v.window.SetTypeHint(gdk.GDK_WINDOW_TYPE_HINT_MENU)
	v.window.SetSizeRequest(200, 180)

	// Buid GUI
	userLabel := gtk.Label(user)
	timeLabel := gtk.Label("")
	timeLabel.SetMarkup("<span size='xx-large'>" + string(minutes) + " min igjen</span>")
	button := gtk.ButtonWithLabel("Logg ut")

	vbox := gtk.VBox(false, 20)
	vbox.SetBorderWidth(5)
	vbox.Add(userLabel)
	vbox.Add(timeLabel)
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
	//window.SetGravity Gdk::Window::GRAVITY_SOUTH_WEST
	//window.Move (Gdk.screen_width - size[0] - 50), (Gdk.screen_height - size[1] - 50)
	return
}

func (v *Status) Show() {
	v.window.ShowAll()
	return
}
