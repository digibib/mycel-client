package window

import (
	"github.com/mattn/go-gtk/gtk"
	"strconv"
)

// Login creates a GTK fullscreen window where users can log inn.
// It returns when a user successfully authenticates.
func ShortTime(client string, minutes int) (user string) {
	// Inital window configuration
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	defer window.Destroy()
	window.Fullscreen()
	window.SetKeepAbove(true)
	window.SetTitle("Mycel Login")

	// Build GUI
	frame := gtk.NewFrame("Logg deg på " + client)
	frame.SetLabelAlign(0.5, 0.5)
	// infolabel = Gtk::Label.new
	// s = "Dette er en korttidsmaskin\nMaks #{@time_limit} minutter!"
	// infolabel.set_markup "<span foreground='red'>#{s}</span>"
	// infolabel.set_alignment 0.5, 0.5
	logo := gtk.NewImageFromFile("logo.png")
	info := gtk.NewLabel("")
	info.SetMarkup("<span foreground='red'>Dette er en korttidsmaskin\nMaks " +
		strconv.Itoa(minutes) + " minutter!</span>")
	//info.SetAlignment(0.5, 0.5)
	button := gtk.NewButtonWithLabel("\nStart\n")

	vbox := gtk.NewVBox(false, 20)
	vbox.SetBorderWidth(20)
	vbox.Add(logo)
	vbox.Add(info)
	vbox.Add(button)

	frame.Add(vbox)

	center := gtk.NewAlignment(0.5, 0.5, 0, 0)
	center.Add(frame)
	window.Add(center)

	// Connect GUI event signals to function callbacks
	button.Connect("clicked", func() {
		gtk.MainQuit()
	})
	window.Connect("delete-event", func() bool {
		return true
	})

	window.ShowAll()
	gtk.Main()
	return "Anonym"
}
