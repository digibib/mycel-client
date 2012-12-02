package windows

import "github.com/mattn/go-gtk/gtk"

func Login(client string) {
	// Inital Window configuration
	window := gtk.Window(gtk.GTK_WINDOW_TOPLEVEL)
	window.Fullscreen()
	window.SetKeepAbove(true)
	window.SetTitle("Mycel")

	// Buid GUI
	logo := gtk.ImageFromFile("logo.png")
	button := gtk.ButtonWithLabel("Log inn")
	frame := gtk.Frame("Logg deg på " + client)
	frame.SetLabelAlign(0.5, 0.5)
	userlabel := gtk.Label("Lånenummer/brukernavn")
	pinlabel := gtk.Label("PIN-kode/passord")
	table := gtk.Table(3, 2, false)

	userentry := gtk.Entry()
	userentry.SetMaxLength(10)
	userentry.SetSizeRequest(150, 23)

	pinentry := gtk.Entry()
	pinentry.SetVisibility(false)
	pinentry.SetMaxLength(10)

	// 	GTK_EXPAND / GTK_SHRINK / GTK_FILL
	table.Attach(userlabel, 0, 1, 0, 1, gtk.GTK_FILL, gtk.GTK_FILL, 7, 5)
	table.Attach(userentry, 1, 2, 0, 1, gtk.GTK_FILL, gtk.GTK_FILL, 7, 5)
	table.Attach(pinlabel, 0, 1, 1, 2, gtk.GTK_FILL, gtk.GTK_FILL, 7, 5)
	table.Attach(pinentry, 1, 2, 1, 2, gtk.GTK_FILL, gtk.GTK_FILL, 7, 5)
	table.Attach(button, 1, 2, 2, 3, gtk.GTK_FILL, gtk.GTK_FILL, 7, 5)

	error := gtk.Label("")

	vbox := gtk.VBox(false, 20)
	vbox.Add(logo)
	vbox.Add(table)
	vbox.Add(error)
	vbox.SetBorderWidth(20)
	frame.Add(vbox)

	center := gtk.Alignment(0.5, 0.5, 0, 0)
	center.Add(frame)

	window.Add(center)

	// Connect signals (GUI events)
	window.Connect("destroy", func() {
		println("quitting..")
		gtk.MainQuit()
	})
	button.Connect("clicked", func() {
		println("clik")
		gtk.MainQuit()
	})

	window.ShowAll()
}
