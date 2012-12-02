package main

import (
	"github.com/mattn/go-gtk/gtk"
	// "os"
	// "path"
)

func main() {
	gtk.Init(nil)
	window := gtk.Window(gtk.GTK_WINDOW_TOPLEVEL)
	window.Fullscreen()
	window.SetTitle("Mycel")
	window.Connect("destroy", func() {
		println("quitting..")
		gtk.MainQuit()
	})

	// dir, _ := path.Split(os.Args[0])
	// imagefile = path.Join(dir, "logo.png")
	logo := gtk.ImageFromFile("logo.png")
	button := gtk.ButtonWithLabel("Log inn")
	frame := gtk.Frame("Logg deg på petter-samsung")
	//frame.label_xalign = 0.5
	userlabel := gtk.Label("Lånenummer/brukernavn")
	//userlabel.set_alignment 1, 0.5
	pinlabel := gtk.Label("PIN-kode/passord")
	// pinlabel.set_alignment 1, 0.5
	table := gtk.Table(3, 2, false)

	vbox := gtk.VBox(false, 20)
	userentry := gtk.Entry()
	userentry.SetMaxLength(10)
	userentry.SetSizeRequest(150, 23)

	pinentry := gtk.Entry()
	pinentry.SetVisibility(false)
	pinentry.SetMaxLength(10)

	// type GtkAttachOptions int

	// const (
	// 	GTK_EXPAND GtkAttachOptions = 1 << 0
	// 	GTK_SHRINK GtkAttachOptions = 1 << 1
	// 	GTK_FILL   GtkAttachOptions = 1 << 2
	// )
	table.Attach(userlabel, 0, 1, 0, 1, gtk.GTK_FILL, gtk.GTK_FILL, 7, 5)
	table.Attach(userentry, 1, 2, 0, 1, gtk.GTK_FILL, gtk.GTK_FILL, 7, 5)
	table.Attach(pinlabel, 0, 1, 1, 2, gtk.GTK_FILL, gtk.GTK_FILL, 7, 5)
	table.Attach(pinentry, 1, 2, 1, 2, gtk.GTK_FILL, gtk.GTK_FILL, 7, 5)
	table.Attach(button, 1, 2, 2, 3, gtk.GTK_FILL, gtk.GTK_FILL, 7, 5)
	// table.attach userlabel, 0, 1, 0, 1
	// table.attach @userentry, 1, 2, 0, 1
	// table.attach pinlabel, 0, 1, 1, 2
	// table.attach @pinentry, 1, 2, 1, 2
	// table.attach button, 1, 2, 2, 3

	// error = Gtk::Label.new
	vbox.Add(logo)
	vbox.Add(table)
	// vbox.pack_start_defaults logo
	// vbox.pack_start_defaults table
	// vbox.pack_end_defaults error
	// vbox.set_border_width 20
	vbox.SetBorderWidth(20)
	frame.Add(vbox)

	center := gtk.Alignment(0.5, 0.5, 0, 0)
	center.Add(frame)

	window.Add(center)

	window.ShowAll()
	gtk.Main()
}
