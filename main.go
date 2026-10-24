package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.NewWithID("io.github.gavin.relix")
	a.Settings().SetTheme(NewLabTheme())
	w := a.NewWindow("RELIX — Cell Simulator")
	w.Resize(fyne.NewSize(1500, 900))
	w.CenterOnScreen()
	NewUI(w)
	w.ShowAndRun()
}
