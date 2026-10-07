//go:build !windows

package cully

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
	"github.com/google/uuid"
	"golang.org/x/term"
)

// RunCodexPane owns the physical terminal and gives Codex a smaller virtual
// terminal. Codex never writes directly to the parent screen, so its clear and
// cursor sequences cannot erase the advisor area.
func RunCodexPane(args []string, input, output *os.File) error {
	if !term.IsTerminal(int(input.Fd())) || !term.IsTerminal(int(output.Fd())) {
		return fmt.Errorf("cully codex requires an interactive terminal")
	}
	cols, rows, err := term.GetSize(int(output.Fd()))
	if err != nil {
		return err
	}
	if cols < 20 || rows < 8 {
		return fmt.Errorf("terminal is too small for cully codex")
	}
	state, err := term.MakeRaw(int(input.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(input.Fd()), state) //nolint:errcheck
	if _, err := io.WriteString(output, "\x1b[?1049h\x1b[?25l"); err != nil {
		return err
	}
	defer io.WriteString(output, "\x1b[0m\x1b[?25h\x1b[?1049l") //nolint:errcheck

	session := uuid.NewString()
	defer os.Remove(codexSignalFile(session))   //nolint:errcheck
	defer os.Remove(sessionReportFile(session)) //nolint:errcheck
	cmd := exec.Command("codex", args...)
	cmd.Env = append(os.Environ(), "CULLY_PANE_SESSION="+session, "CULLY_SESSION="+session, "CULLY_AGENT=codex")
	top := codexPaneTop(rows)
	child, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: uint16(top), Cols: uint16(cols)})
	if err != nil {
		return fmt.Errorf("start codex: %w", err)
	}
	emulator := vt.NewEmulator(cols, top)
	repliesDone := make(chan struct{})
	defer func() {
		_ = child.Close()
		if pipe, ok := emulator.InputPipe().(io.Closer); ok {
			_ = pipe.Close()
		}
		<-repliesDone
		_ = emulator.Close()
	}()
	go io.Copy(child, input) //nolint:errcheck
	go func() {
		_, _ = io.Copy(child, emulator) // terminal query replies
		close(repliesDone)
	}()

	chunks := make(chan []byte, 16)
	go func() {
		defer close(chunks)
		buf := make([]byte, 32*1024)
		for {
			n, err := child.Read(buf)
			if n > 0 {
				chunks <- bytes.Clone(buf[:n])
			}
			if err != nil {
				return
			}
		}
	}()
	resize := make(chan os.Signal, 1)
	stop := make(chan os.Signal, 1)
	signal.Notify(resize, syscall.SIGWINCH)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(resize)
	defer signal.Stop(stop)
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	dirty, lastAdvice := true, ""
	var advice []string
	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				if dirty {
					_, _ = io.WriteString(output, renderCodexPane(emulator, cols, rows, advice, readCodexToolStats(session)))
				}
				return cmd.Wait()
			}
			_, _ = emulator.Write(chunk)
			dirty = true
		case <-resize:
			if width, height, sizeErr := term.GetSize(int(output.Fd())); sizeErr == nil && width > 0 && height > 0 {
				cols, rows = width, height
				top = codexPaneTop(rows)
				emulator.Resize(cols, top)
				_ = pty.Setsize(child, &pty.Winsize{Rows: uint16(top), Cols: uint16(cols)})
				dirty = true
			}
		case sig := <-stop:
			_ = cmd.Process.Signal(sig)
		case <-ticker.C:
			stats := readCodexToolStats(session)
			advice = codexAdvice(stats)
			joined := strings.Join(advice, "\n")
			if joined != lastAdvice {
				lastAdvice = joined
				_ = writeReportLines(session, currentDir(), advice)
				dirty = true
			}
			if dirty {
				if _, err := io.WriteString(output, renderCodexPane(emulator, cols, rows, advice, stats)); err != nil {
					_ = cmd.Process.Kill()
					_ = cmd.Wait()
					return err
				}
				dirty = false
			}
		}
	}
}

func currentDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

func codexPaneTop(rows int) int {
	if rows < 16 {
		return rows // keep Codex usable in a small terminal
	}
	return rows / 2
}

func renderCodexPane(emulator *vt.Emulator, cols, rows int, advice []string, stats codexToolStats) string {
	var out strings.Builder
	limit := max(0, cols-1) // avoid triggering automatic terminal wrap
	top := codexPaneTop(rows)
	out.WriteString("\x1b[?25l")
	for y := 0; y < top; y++ {
		fmt.Fprintf(&out, "\x1b[%d;1H\x1b[0m\x1b[2K", y+1)
		line := uv.NewLine(limit)
		for x := 0; x < limit; x++ {
			if cell := emulator.CellAt(x, y); cell != nil {
				line.Set(x, cell)
			}
		}
		out.WriteString(line.Render())
	}
	if top < rows {
		panel := codexPanelLines(cols, rows-top, advice, stats)
		for i, line := range panel {
			fmt.Fprintf(&out, "\x1b[%d;1H\x1b[0m\x1b[2K\x1b[48;5;235m%s\x1b[0m", top+i+1, line)
		}
	}
	cursor := emulator.CursorPosition()
	fmt.Fprintf(&out, "\x1b[%d;%dH\x1b[?25h", min(max(cursor.Y+1, 1), top), min(max(cursor.X+1, 1), cols))
	return out.String()
}

func codexPanelLines(cols, height int, advice []string, stats codexToolStats) []string {
	lines := make([]string, height)
	if height == 0 {
		return lines
	}
	width := max(0, cols-1)
	lines[0] = "\x1b[38;5;81m" + strings.Repeat("─", width)
	if height > 1 {
		lines[1] = "\x1b[1;38;5;81m" + ansi.Truncate(" CULLY  /  Codex advisor", width, "")
	}
	if height > 2 {
		lines[2] = "\x1b[38;5;245m" + ansi.Truncate(fmt.Sprintf(" %d tools  ·  %d searches  ·  %d edits  ·  %d checks", stats.Tools, stats.Searches, stats.Edits, stats.Checks), width, "")
	}
	for i, suggestion := range advice {
		row := 4 + i*2
		if row >= height-1 {
			break
		}
		color := "\x1b[38;5;114m"
		if strings.HasPrefix(suggestion, "CAUT|") || strings.HasPrefix(suggestion, "WARN|") {
			color = "\x1b[38;5;220m"
		}
		if _, text, ok := strings.Cut(suggestion, "|"); ok {
			suggestion = text
		}
		lines[row] = color + ansi.Truncate("  "+suggestion, width, "")
	}
	if height > 3 {
		lines[height-1] = "\x1b[38;5;245m" + ansi.Truncate(" /prompts:cully for controls  ·  cully suggestions for details", width, "")
	}
	return lines
}
