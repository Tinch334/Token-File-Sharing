package ui_client

import (
	"github.com/rivo/tview"
	//"github.com/gdamore/tcell/v2"
)


// Stores all UI elements in the main interface along with the connection to the TFS server.
type App struct {
	app        tview.*Application
	pages      tview.*Pages
	layout     tview.*Layout

	titleBar   *tview.TextView
	localPane  *Filepanel
	remotePane *Filepanel
	queue      *LogPanel
	log        *LogPanel
	statusBar  *tview.TextView
}

// NewApp creates a App structure along a new tview application.
func NewApp() *App {
	app := App {
		app:    tview.NewApplication()
		pages:  tview.NewPages()
	}

	app.buildLayout()
	app.bindKeys()
	return &a
}

// https://claude.ai/chat/6ca39e6e-3c05-4d81-b095-175f9bea7da3
func (a* App) buildLayout() {
	a.localPane = NewFilePanel("Local", a)
	a.remotePane = NewFilePanel("Remote", a)
	a.queue = NewQueuePanel()
	a.log = NewLogPanel()

	a.titleBar = tview.NewTextView().
		SetText("[red]TFS client[white] | Version: 0.6.1").
		SetDynamicColors()

	a.statusBar = tview.NewTextView().
		SetText("[Tab] Switch  [Enter] Open  [Space] Select  [U] Upload  [D] Download  [Del] Delete  [Q] Quit").
		SetDynamicColors()

	fileFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(localPane,  0, 1, true).
		AddItem(remotePane, 0, 1, false)

	a.layout = tview.NewFlex.SetDirection(tview.FlexRow).
		AddItem(a.titleBar,   1, 0, false).
		AddItem(fileFlex,     0, 3, true).
		AddItem(a.queue.View, 6, 0, false).
		AddItem(a.log.View,   8, 0, false).
		AddItem(a.statusBar,  1, 0, false)

	a.pages.AddPage("main",    a.layout,             true, true)
	a.pages.AddPage("connect", a.buildConnectForm(), true, false)

	a.app.SetRoot(a.pages, true)
}


// nextUIElem switches the application's element to the next one on the list and sets focus to it.
func (ui *uiElems) nextUIElem() {
	// Adjust element pointer and check overflow.
	ui.elemPointer++
	if (ui.elemPointer >= len(ui.elems)) {
		ui.elemPointer = 0
	}

	ui.app.SetFocus(ui.elems[ui.elemPointer])
}

// prevUIElem switches the application's element to the previous one on the list and sets focus to it.
func (ui *uiElems) prevUIElem() {
	// Adjust element pointer and check overflow.
	ui.elemPointer--
	if (ui.elemPointer < 0) {
		ui.elemPointer = len(ui.elems) - 1
	}

	ui.app.SetFocus(ui.elems[ui.elemPointer])
}


func CreateClient(app tview.*Application) {



	pages := tview.NewPages()

	app.SetRoot(pages, true)
}


func mainInterface() *tview.Flex {
	fileFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Local files"), 0, 1, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Server files"), 0, 1, false)

	topBanner := tview.NewTextView().SetText("[red]TFS client[white] | Version: 0.6.1")
	topBanner.SetScrollable(false).SetRegions(false).SetDynamicColors(true)//.SetBackgroundColor(tcell.NewRGBColor(176, 113, 86))

	console := tview.NewTextView().SetText("[red]TFS client[white] | Version: 0.6.1")
	console.SetScrollable(false).SetRegions(false).SetDynamicColors(true).SetTextAlign(tview.AlignLeft).SetWordWrap(true)
	console.SetBorder(true).SetTitle("TFS log")

	input := tview.NewInputField().
		SetLabel("> ")

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topBanner, 1, 1, false).
		AddItem(fileFlex, 0, 3, false).
		AddItem(console, 0, 1, false).
		AddItem(input, 1, 1, true)

	return flex	
}


func connectInterface() *tview.Form {
	form := tview.NewForm().
		AddInputField("Server address", "", 40, nil, nil).
		AddInputField("Username", "", 50, nil, nil).
		AddPasswordField("Password", "", 60, "*", nil)
		AddButton("Connect", nil).
		AddButton("Quit")
}