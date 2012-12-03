package windows

import "github.com/mattn/go-gtk/gtk"

func Status(client, user string, minutes int) (err error) {
	// Inital Window configuration
	window := gtk.Window(gtk.GTK_WINDOW_TOPLEVEL)
	window.SetKeepAbove(true)
	window.SetTitle(client)

	// Buid GUI
	return
}
