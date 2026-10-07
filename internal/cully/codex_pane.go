//go:build !windows

package cully

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
	"github.com/google/uuid"
	"golang.org/x/term"
)

// RunCodexPane owns the physical terminal and gives Codex a smaller virtual
// terminal. Codex never writes directly to the parent screen, so its clear and
// cursor sequences cannot erase the status area.
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
	if _, err := io.WriteString(output, "\x1b[?1049h\x1b[?25l\x1b[?1000h\x1b[?1006h"); err != nil {
		return err
	}
	defer io.WriteString(output, "\x1b[?1000l\x1b[?1006l\x1b[0m\x1b[?25h\x1b[?1049l") //nolint:errcheck
	reviewContext, cancelReviews := context.WithCancel(context.Background())
	defer cancelReviews()

	session := uuid.NewString()
	if err := registerCodexPane(session, currentDir()); err != nil {
		return fmt.Errorf("register Codex advisor pane: %w", err)
	}
	defer os.Remove(codexPaneRegistrationFile(session)) //nolint:errcheck
	defer os.Remove(codexPaneBindingFile(session))      //nolint:errcheck
	defer os.Remove(codexSignalFile(session))           //nolint:errcheck
	defer os.Remove(sessionReportFile(session))         //nolint:errcheck
	defer os.Remove(sessionSnapshotFile(session))       //nolint:errcheck
	defer os.Remove(sessionSignalsFile(session))        //nolint:errcheck
	defer os.Remove(sessionSeenFile(session))           //nolint:errcheck
	cmd := exec.Command("codex", append([]string{"-c", codexStatusConfig}, args...)...)
	cmd.Env = append(os.Environ(), "CULLY_PANE_SESSION="+session, "CULLY_SESSION="+session, "CULLY_AGENT=codex")
	view := codexStatusView{Project: currentDir(), Branch: gitBranch(currentDir()), Started: time.Now(), Daemon: isDaemonRunning()}
	view.Terminal = detectTerminalProfile()
	readCodexGitChanges(&view)
	stats := readCodexToolStats(session)
	advice := codexCombinedAdvice(session, stats, view)
	panelBudget := func() int { return rows - codexPaneTop(rows, rows) }
	statusRows := func() []string { return codexCompactStatusRows(cols, panelBudget(), advice, stats, view) }
	content := statusRows()
	top := codexPaneTop(rows, len(content))
	child, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: uint16(top), Cols: uint16(cols)})
	if err != nil {
		return fmt.Errorf("start codex: %w", err)
	}
	// Wait for the process itself, not PTY EOF: a descendant can retain the
	// terminal after Codex exits. The wrapper must still release its session.
	waitDone := make(chan struct{})
	var waitErr error
	go func() {
		waitErr = cmd.Wait()
		close(waitDone)
	}()
	readerDone := make(chan struct{})
	emulator := vt.NewEmulator(cols, top)
	repliesDone := make(chan struct{})
	defer func() {
		close(readerDone)
		// StartWithSize creates a child process group. Never leave that group
		// running when the owning wrapper exits or loses its terminal.
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		_ = child.Close()
		select {
		case <-waitDone:
		case <-time.After(2 * time.Second):
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			<-waitDone
		}
		// Descendants may ignore TERM even after the main process has exited.
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if pipe, ok := emulator.InputPipe().(io.Closer); ok {
			_ = pipe.Close()
		}
		<-repliesDone
		_ = emulator.Close()
	}()
	inputChunks := make(chan []byte, 16)
	go func() {
		defer close(inputChunks)
		buf := make([]byte, 4096)
		for {
			n, err := input.Read(buf)
			if n > 0 {
				select {
				case inputChunks <- bytes.Clone(buf[:n]):
				case <-readerDone:
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()
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
				select {
				case chunks <- bytes.Clone(buf[:n]):
				case <-readerDone:
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()
	resize := make(chan os.Signal, 1)
	stop := make(chan os.Signal, 1)
	signal.Notify(resize, syscall.SIGWINCH)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(resize)
	defer signal.Stop(stop)
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	dirty, lastAdvice := true, ""
	lastContent := ""
	lastBeat := time.Now()
	lastAnalysis := time.Time{}
	lastAnalysisTools := 0
	var panelInput codexPanelInput
	lastInput := time.Now()
	var drawer codexAdvisorDrawer
	var cancelPreview context.CancelFunc
	defer func() {
		if cancelPreview != nil {
			cancelPreview()
		}
	}()
	type reviewResult struct {
		generation int
		review     advisorReview
		err        error
	}
	reviews := make(chan reviewResult, 1)
	paint := func() string {
		if drawer.Open {
			return drawer.render(cols, rows, stats, view)
		}
		return renderCodexPane(emulator, cols, rows, content)
	}
	preview := func() {
		if drawer.Busy || drawer.Details || len(drawer.Items) == 0 {
			return
		}
		if drawer.Review != nil {
			review := *drawer.Review
			if review.Handoff != "" {
				// Paste a request rather than submit an arbitrary action or discard the
				// user's current draft. The foreground agent retains its normal review.
				_, _ = child.Write([]byte("\x1b[200~\n" + terminalSafeText(review.Handoff) + "\x1b[201~"))
				drawer.close()
			} else if err := acceptAdvisorReview(view.Project, review); err != nil {
				drawer.Message = red + "Could not apply: " + err.Error() + rst
			} else {
				drawer.Message = green + "✓ Applied · Esc returns to suggestions" + rst
				drawer.Review = nil
			}
			dirty = true
			return
		}
		line := drawer.Items[drawer.Selected]
		if strings.HasPrefix(line, "MEMO|") {
			drawer.Message = dim + "Informational tip · no action to apply" + rst
			dirty = true
			return
		}
		// Routine workflow actions belong in the coding session. Configuration
		// advice can produce a small local-file preview using this agent's adapter.
		_, category := codexAdviceBadge(line)
		if category != "Apply" {
			review := advisorHandoff(line)
			drawer.Review = &review
			drawer.Scroll = 0
			dirty = true
			return
		}
		drawer.Busy = true
		dirty = true
		generation := drawer.Generation
		project := view.Project
		if cancelPreview != nil {
			cancelPreview()
		}
		previewCtx, cancel := context.WithCancel(reviewContext)
		cancelPreview = cancel
		go func() {
			review, err := prepareAdvisorReviewContext(previewCtx, "codex", project, line)
			select {
			case reviews <- reviewResult{generation, review, err}:
			case <-reviewContext.Done():
			}
		}()
	}
	handleKeys := func(keys []codexPanelKey) {
		for _, key := range keys {
			if key.Action == "focus" {
				if drawer.Open {
					if cancelPreview != nil {
						cancelPreview()
					}
					drawer.close()
				} else {
					drawer.open(advice)
				}
				dirty = true
				continue
			}
			if key.Mouse != nil {
				mouse := key.Mouse
				if !mouse.Release {
					if !drawer.Open && mouse.Button == 0 && mouse.Y > top {
						drawer.open(advice)
					} else if drawer.Open {
						switch mouse.Button {
						case 64:
							drawer.move(-1)
						case 65:
							drawer.move(1)
						case 0:
							if index, ok := drawer.HitRows[mouse.Y]; ok && !drawer.Busy {
								drawer.Selected = index
								drawer.Message = ""
							}
							if mouse.Y == drawer.AcceptRow {
								preview()
							}
						}
					}
				}
				dirty = true
				continue
			}
			if drawer.Open {
				switch key.Action {
				case "escape":
					if drawer.Busy && cancelPreview != nil {
						cancelPreview()
					}
					drawer.escape()
				case "up":
					drawer.move(-1)
				case "down":
					drawer.move(1)
				case "pageup":
					drawer.move(-max(1, rows/4))
				case "pagedown":
					drawer.move(max(1, rows/4))
				case "home":
					drawer.move(-10000)
				case "end":
					drawer.move(10000)
				case "enter":
					preview()
				case "details":
					if !drawer.Busy {
						drawer.Details = !drawer.Details
						drawer.Review = nil
						drawer.Scroll = 0
					}
				default:
					if bytes.Equal(key.Data, []byte{3}) {
						drawer.close()
					} // return control without killing Codex
				}
				dirty = true
				continue
			}
			_, _ = child.Write(key.Data)
		}
	}
	var shutdown <-chan time.Time
	paintFinal := func() {
		readCodexFooter(emulator, &view)
		stats = readCodexToolStats(session)
		advice = codexCombinedAdvice(session, stats, view)
		content = statusRows()
		_, _ = io.WriteString(output, paint())
	}
	for {
		select {
		case result := <-reviews:
			if drawer.Open && result.generation == drawer.Generation {
				drawer.Busy = false
				drawer.Review = &result.review
				drawer.Scroll = 0
				if result.err != nil {
					drawer.Message = yellow + result.err.Error() + " · session handoff is available" + rst
				}
				dirty = true
			}
		case chunk, ok := <-inputChunks:
			if !ok {
				inputChunks = nil
				handleKeys(panelInput.feed(nil, true))
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
				if shutdown == nil {
					shutdown = time.After(2 * time.Second)
				}
				continue
			}
			lastInput = time.Now()
			handleKeys(panelInput.feed(chunk, false))
		case <-waitDone:
			// Preserve any output already delivered before the child exited.
			for {
				select {
				case chunk, ok := <-chunks:
					if !ok {
						paintFinal()
						return waitErr
					}
					_, _ = emulator.Write(chunk)
				default:
					paintFinal()
					return waitErr
				}
			}
		case chunk, ok := <-chunks:
			if !ok {
				if dirty {
					readCodexFooter(emulator, &view)
					stats = readCodexToolStats(session)
					advice = codexCombinedAdvice(session, stats, view)
					content = statusRows()
					_, _ = io.WriteString(output, paint())
				}
				chunks = nil
				continue
			}
			_, _ = emulator.Write(chunk)
			dirty = true
		case <-resize:
			if width, height, sizeErr := term.GetSize(int(output.Fd())); sizeErr == nil && width > 0 && height > 0 {
				cols, rows = width, height
				content = statusRows()
				top = codexPaneTop(rows, len(content))
				emulator.Resize(cols, top)
				_ = pty.Setsize(child, &pty.Winsize{Rows: uint16(top), Cols: uint16(cols)})
				dirty = true
			}
		case sig := <-stop:
			_ = syscall.Kill(-cmd.Process.Pid, sig.(syscall.Signal))
			if sig != syscall.SIGINT && shutdown == nil {
				shutdown = time.After(2 * time.Second)
			}
		case <-shutdown:
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			shutdown = nil
		case <-ticker.C:
			if time.Since(lastInput) >= 80*time.Millisecond {
				handleKeys(panelInput.feed(nil, true))
			}
			if time.Since(lastBeat) >= 2*time.Second {
				_ = os.Chtimes(codexPaneRegistrationFile(session), time.Now(), time.Now())
				view.Daemon = isDaemonRunning()
				view.Branch = gitBranch(view.Project)
				readCodexGitChanges(&view)
				lastBeat = time.Now()
			}
			stats = readCodexToolStats(session)
			readCodexFooter(emulator, &view)
			if stats.Tools > lastAnalysisTools && time.Since(lastAnalysis) >= 2*time.Minute {
				snap := readSnapshot(session)
				snap.Cwd = view.Project
				snap.Searches, snap.ToolErrors = stats.Searches, stats.Errors
				if view.ContextKnown {
					snap.ContextUsedPct = 100 - view.ContextLeft
				}
				writeSnapshot(session, snap)
				dispatchAdvisor(codexAdvisorSignals(stats, view), session, view.Project)
				lastAnalysis, lastAnalysisTools = time.Now(), stats.Tools
			}
			advice = codexCombinedAdvice(session, stats, view)
			joined := strings.Join(advice, "\n")
			if joined != lastAdvice {
				lastAdvice = joined
				dirty = true
			}
			content = statusRows()
			if joinedContent := strings.Join(content, "\n"); joinedContent != lastContent {
				lastContent, dirty = joinedContent, true
			}
			// Expand for new content, but do not keep reflowing Codex when
			// temporary advice clears. A physical resize recalculates both areas.
			if nextTop := codexPaneTop(rows, len(content)); nextTop < top {
				top = nextTop
				emulator.Resize(cols, top)
				_ = pty.Setsize(child, &pty.Winsize{Rows: uint16(top), Cols: uint16(cols)})
				dirty = true
			}
			if dirty {
				if _, err := io.WriteString(output, paint()); err != nil {
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

func renderCodexPane(emulator *vt.Emulator, cols, rows int, content []string) string {
	var out strings.Builder
	limit := max(0, cols-1) // avoid triggering automatic terminal wrap
	top := min(emulator.Height(), rows)
	var native codexStatusView
	footerRow := readCodexFooter(emulator, &native)
	out.WriteString("\x1b[?25l")
	for y := 0; y < top; y++ {
		fmt.Fprintf(&out, "\x1b[%d;1H\x1b[0m\x1b[2K", y+1)
		if y == footerRow && top < rows {
			continue // the native instruments now live in the Cully status view
		}
		line := uv.NewLine(limit)
		for x := 0; x < limit; x++ {
			if cell := emulator.CellAt(x, y); cell != nil {
				line.Set(x, cell)
			}
		}
		out.WriteString(line.Render())
	}
	if top < rows {
		panel := codexPanelLines(cols, rows-top, content)
		for i, line := range panel {
			fmt.Fprintf(&out, "\x1b[%d;1H\x1b[0m\x1b[2K%s\x1b[0m", top+i+1, line)
		}
	}
	cursor := emulator.CursorPosition()
	fmt.Fprintf(&out, "\x1b[%d;%dH\x1b[?25h", min(max(cursor.Y+1, 1), top), min(max(cursor.X+1, 1), cols))
	return out.String()
}
