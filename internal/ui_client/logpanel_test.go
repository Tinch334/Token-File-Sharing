package ui_client

import (
	"strings"
	"sync"
	"testing"
)

func TestLogPanelCreation(t *testing.T) {
	lp := NewLogPanel()
	if lp == nil || lp.View == nil {
		t.Fatalf("expected non-nil LogPanel and View")
	}

	qp := NewQueuePanel()
	if qp == nil || qp.View == nil {
		t.Fatalf("expected non-nil QueuePanel and View")
	}
}

func TestLogPanelWriteAndLevels(t *testing.T) {
	lp := NewLogPanel("Test Log")

	lp.Info("information message")
	lp.Infof("formatted info %d", 42)
	lp.Success("success message")
	lp.Successf("formatted success %s", "ok")
	lp.Warn("warn message")
	lp.Warnf("formatted warn %s", "caution")
	lp.Error("error message")
	lp.Errorf("formatted error %d", 500)
	lp.Log("plain log")
	lp.Logf("plain formatted %s", "text")
	lp.LogRaw("raw text\n")

	// Test io.Writer
	n, err := lp.Write([]byte("io.Writer message\n"))
	if err != nil || n == 0 {
		t.Fatalf("unexpected write error: %v, written: %d", err, n)
	}

	text := lp.GetText(true)
	if !strings.Contains(text, "information message") {
		t.Errorf("expected 'information message' in log text, got: %s", text)
	}
	if !strings.Contains(text, "formatted info 42") {
		t.Errorf("expected 'formatted info 42' in log text, got: %s", text)
	}
	if !strings.Contains(text, "success message") {
		t.Errorf("expected 'success message' in log text, got: %s", text)
	}
	if !strings.Contains(text, "io.Writer message") {
		t.Errorf("expected 'io.Writer message' in log text, got: %s", text)
	}

	// Test Clear
	lp.Clear()
	clearedText := lp.GetText(true)
	if clearedText != "" {
		t.Errorf("expected empty text after clear, got: %q", clearedText)
	}
}

func TestQueuePanelItem(t *testing.T) {
	qp := NewQueuePanel()
	qp.AddQueueItem("UPLOAD", "myfile.txt", "1.5 MB")

	text := qp.GetText(true)
	if !strings.Contains(text, "UPLOAD") || !strings.Contains(text, "myfile.txt") {
		t.Errorf("expected queue entry in text, got: %s", text)
	}
}

func TestLogPanelConcurrency(t *testing.T) {
	lp := NewLogPanel("Concurrent Test")
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			lp.Logf("routine message %d", idx)
			lp.Infof("routine info %d", idx)
		}(i)
	}

	wg.Wait()
}

func TestLogPanelStateAndConfig(t *testing.T) {
	lp := NewLogPanel("Config Test")
	lp.SetActive(true)
	lp.SetActive(false)
	lp.SetTitle("Updated Title")
	lp.SetMaxLines(500)
	lp.ScrollToTop()
	lp.ScrollToEnd()
}
