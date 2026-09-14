package ui_client

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// LogPanel wraps a tview.TextView to provide a thread-safe, colored log or queue display.
type LogPanel struct {
	View *tview.TextView
	app  *tview.Application
	mu   sync.Mutex
}

// NewLogPanel creates a new LogPanel with an optional custom title.
// If no title is specified, it defaults to "TFS Log".
func NewLogPanel(title ...string) *LogPanel {
	t := "TFS Log"
	if len(title) > 0 && strings.TrimSpace(title[0]) != "" {
		t = strings.TrimSpace(title[0])
	}

	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetMaxLines(1000)

	tv.SetBorder(true).SetTitle(" " + t + " ")

	return &LogPanel{
		View: tv,
	}
}

// NewQueuePanel creates a LogPanel specifically titled for the transfer queue.
func NewQueuePanel() *LogPanel {
	return NewLogPanel("Transfer Queue")
}

// SetApp associates a tview.Application with this panel to support thread-safe redraws.
func (p *LogPanel) SetApp(app *tview.Application) *LogPanel {
	p.app = app
	return p
}

// SetTitle updates the title displayed on the panel border.
func (p *LogPanel) SetTitle(title string) {
	p.View.SetTitle(" " + strings.TrimSpace(title) + " ")
}

// SetMaxLines sets the maximum number of lines retained in the buffer.
func (p *LogPanel) SetMaxLines(max int) *LogPanel {
	p.View.SetMaxLines(max)
	return p
}

// SetActive changes the border style to indicate whether the panel is active/focused.
func (p *LogPanel) SetActive(active bool) {
	if active {
		p.View.SetBorderColor(tcell.ColorGreen)
		p.View.SetBorderAttributes(tcell.AttrBold)
	} else {
		p.View.SetBorderColor(tcell.ColorDefault)
		p.View.SetBorderAttributes(tcell.AttrNone)
	}
}

func (p *LogPanel) formatTimestamp() string {
	return time.Now().Format("15:04:05")
}

func (p *LogPanel) appendText(text string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.app != nil {
		p.app.QueueUpdateDraw(func() {
			fmt.Fprint(p.View, text)
			p.View.ScrollToEnd()
		})
	} else {
		fmt.Fprint(p.View, text)
		p.View.ScrollToEnd()
	}
}

// Write implements io.Writer, allowing LogPanel to be used directly as a log destination.
func (p *LogPanel) Write(data []byte) (n int, err error) {
	p.appendText(string(data))
	return len(data), nil
}

// Log writes a plain message prefixed with the current timestamp.
func (p *LogPanel) Log(msg string) {
	p.appendText(fmt.Sprintf("[gray]%s[-] %s\n", p.formatTimestamp(), tview.Escape(msg)))
}

// Logf writes a formatted message prefixed with the current timestamp.
func (p *LogPanel) Logf(format string, args ...any) {
	p.Log(fmt.Sprintf(format, args...))
}

// LogRaw appends raw text directly to the view without escaping or timestamp prefixing.
func (p *LogPanel) LogRaw(text string) {
	p.appendText(text)
}

// Info writes an info-level log message with a blue badge.
func (p *LogPanel) Info(msg string) {
	p.appendText(fmt.Sprintf("[gray]%s[-] [dodgerblue][INFO][-] %s\n", p.formatTimestamp(), tview.Escape(msg)))
}

// Infof writes a formatted info-level log message with a blue badge.
func (p *LogPanel) Infof(format string, args ...any) {
	p.Info(fmt.Sprintf(format, args...))
}

// Success writes a success-level log message with a green badge.
func (p *LogPanel) Success(msg string) {
	p.appendText(fmt.Sprintf("[gray]%s[-] [green][OK][-] %s\n", p.formatTimestamp(), tview.Escape(msg)))
}

// Successf writes a formatted success-level log message with a green badge.
func (p *LogPanel) Successf(format string, args ...any) {
	p.Success(fmt.Sprintf(format, args...))
}

// Warn writes a warning-level log message with a yellow badge.
func (p *LogPanel) Warn(msg string) {
	p.appendText(fmt.Sprintf("[gray]%s[-] [yellow][WARN][-] %s\n", p.formatTimestamp(), tview.Escape(msg)))
}

// Warnf writes a formatted warning-level log message with a yellow badge.
func (p *LogPanel) Warnf(format string, args ...any) {
	p.Warn(fmt.Sprintf(format, args...))
}

// Error writes an error-level log message with a red badge.
func (p *LogPanel) Error(msg string) {
	p.appendText(fmt.Sprintf("[gray]%s[-] [red][ERROR][-] %s\n", p.formatTimestamp(), tview.Escape(msg)))
}

// Errorf writes a formatted error-level log message with a red badge.
func (p *LogPanel) Errorf(format string, args ...any) {
	p.Error(fmt.Sprintf(format, args...))
}

// AddQueueItem logs a structured transfer queue event.
func (p *LogPanel) AddQueueItem(action, filename, size string) {
	p.appendText(fmt.Sprintf("[gray]%s[-] [cyan]%-8s[-] [white]%-30s[-] [gray]%s[-]\n",
		p.formatTimestamp(), action, tview.Escape(filename), size))
}

// Clear clears all text from the panel.
func (p *LogPanel) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.app != nil {
		p.app.QueueUpdateDraw(func() {
			p.View.Clear()
		})
	} else {
		p.View.Clear()
	}
}

// ScrollToTop scrolls the view to the top of the log buffer.
func (p *LogPanel) ScrollToTop() {
	if p.app != nil {
		p.app.QueueUpdateDraw(func() {
			p.View.ScrollToBeginning()
		})
	} else {
		p.View.ScrollToBeginning()
	}
}

// ScrollToEnd scrolls the view to the latest log line.
func (p *LogPanel) ScrollToEnd() {
	if p.app != nil {
		p.app.QueueUpdateDraw(func() {
			p.View.ScrollToEnd()
		})
	} else {
		p.View.ScrollToEnd()
	}
}

// GetText returns the text content of the panel.
func (p *LogPanel) GetText(stripTags bool) string {
	return p.View.GetText(stripTags)
}
