//go:build !windows

package cully

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
)

func TestCodexPaneConfinesChildScreen(t *testing.T) {
	emulator := vt.NewEmulator(60, 12)
	defer emulator.Close()
	_, _ = emulator.WriteString("\x1b[2J\x1b[1;1HCODEX\x1b[12;1Hlower edge")
	frame := renderCodexPane(emulator, 60, 24, []string{"ADV|Run a focused check."}, codexToolStats{Tools: 3})
	for _, want := range []string{"CODEX", "lower edge", "CULLY", "Run a focused check.", "\x1b[13;1H"} {
		if !strings.Contains(frame, want) {
			t.Fatalf("pane frame missing %q", want)
		}
	}
	if strings.Contains(frame, "\x1b[2J") {
		t.Fatal("child's whole-screen clear escaped the virtual terminal")
	}
	if codexPaneTop(12) != 12 || codexPaneTop(24) != 12 {
		t.Fatal("unexpected pane sizing")
	}
}

func TestCodexPaneLaunchesChildInPTY(t *testing.T) {
	dir := t.TempDir()
	command := filepath.Join(dir, "codex")
	if err := os.WriteFile(command, []byte("#!/bin/sh\nprintf '\\033[2J\\033[1;1Hcodex child\\n'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	master, slave, err := pty.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer master.Close()
	defer slave.Close()
	if err := pty.Setsize(slave, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		t.Fatal(err)
	}
	output := make(chan string, 1)
	go func() {
		var captured strings.Builder
		buf := make([]byte, 4096)
		for {
			n, readErr := master.Read(buf)
			captured.Write(buf[:n])
			if strings.Contains(captured.String(), "\x1b[?1049l") || readErr != nil {
				output <- captured.String()
				return
			}
		}
	}()
	done := make(chan error, 1)
	go func() { done <- RunCodexPane(nil, slave, slave) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Codex pane did not exit with its child")
	}
	select {
	case got := <-output:
		if !strings.Contains(got, "codex child") || !strings.Contains(got, "CULLY") || !strings.Contains(got, "\x1b[?1049l") {
			t.Fatalf("pane did not render child and advisor or restore terminal: %q", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("pane output did not close")
	}
}
