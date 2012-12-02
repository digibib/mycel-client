package main

import (
	"github.com/mattn/go-gtk/gtk"
	// "log"

	"github.com/digibib/mycel-client/windows"
)

func main() {
	gtk.Init(nil)
	windows.Login("petter-samsung")
	gtk.Main()
}
