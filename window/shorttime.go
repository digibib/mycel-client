package window

import (
	"github.com/mattn/go-gtk/gdkpixbuf"
	"github.com/mattn/go-gtk/gtk"

	"strconv"
)

// ShortTime creates a GTK fullscreen window for the shorttime clients.
// No username/password required, only click 'start' button to log in
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
	var imageLoader *gdkpixbuf.Loader
	imageLoader, _ = gdkpixbuf.NewLoaderWithMimeType("image/png")
	imageLoader.Write(logo_png())
	imageLoader.Close()
	logo := gtk.NewImageFromPixbuf(imageLoader.GetPixbuf())
	info := gtk.NewLabel("")
	info.SetMarkup("<span foreground='red'>Dette er en korttidsmaskin\nMaks " +
		strconv.Itoa(minutes) + " minutter!</span>")
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
