package ui_client

import (
	"os"

	"github.com/rivo/tview"
	//"github.com/gdamore/tcell/v2"
)



type Filepanel struct {
	View     *tview.Table
	cwd      string
	app      *App
	entries  []File
	// Row index -> Selected
	selected map[int]bool
}

func NewFilepanel(title string, app *App) *Filepanel {
	t := tview.NewTable().
		SetSelectable(true, false).  // Set rows as selectable.
		SetBorders(false).
		SetFixed(1, 0)  // Freeze head row.

	t.SetBorder(true).SetTitle(" " + title + " ")

	p := &Filepanel{
		View:     t,
		app:      app,
		entries:  make([]File),
		selected: make(map[int]bool)
	}
	p.bindKeys()

}


// bindKeys	sets the keybinds for the filepanel.
func (p *Filepanel) bindKeys() {
	p.View.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.key() {
		case tcell.KeyEnter:
			//p.openSelected()
			// Override default enter behaviour.
			return nil
		case tcell.KeyDelete:
			//p.deleteSelected()
			return nil
		}

		switch event.Rune() {
		case ' ':
			// Could be done with "SetSelectedFunc", done like this for consistency.
			p.toggleSelection()
		case 'a', 'A':
			// TODO
			return nil
		}

		return event
	})
}


// SetActive changes the style of the FilePanel to denote the fact it's active.
func (p *FilePanel) SetActive(active bool) {
    if active {
        p.View.SetBorderColor(tcell.ColorGreen)
        p.View.SetBorderAttributes(tcell.AttrBold)
    } else {
        p.View.SetBorderColor(tcell.ColorDefault)
        p.View.SetBorderAttributes(tcell.AttrNone)
    }
}


func (p *Filepanel) renderRow(idx int) {
	e := p.entries[idx]
	row := idx + 1

	icon := "📄"
    if e.IsDir {
        icon = "📁"
    }
}


// toggleSelection toggles the selection status of the current row element.
func toggleSelection() {
	row, _ := p.View.GetSelection()
	// Offset for header row.
    entryIndex := row - 1 

    // Toggle selection.
    if p.selected[entryIndex] {
    	delete(p.selected, entryIndex)
    } else {
    	p.selected[entryIndex] = true
    }

    p.renderRow(entryIndex)
}