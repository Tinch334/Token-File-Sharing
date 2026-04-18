package main

import (
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
)


func main() {
	app := tview.NewApplication()

	createClient(app)

	if err := app.Run(); err != nil {
		panic(err)
	}
}


func createClient(app *tview.Application) {
	frame := tview.NewFrame(tview.NewBox().SetBackgroundColor(tcell.ColorCadetBlue)).
		SetBorders(0, 0, 2, 2, 4, 4).
		AddText("Header left", true, tview.AlignLeft, tcell.ColorWhite).
		AddText("Header bottom 1", false, tview.AlignLeft, tcell.ColorWhite)

	app.SetRoot(frame, true)
}