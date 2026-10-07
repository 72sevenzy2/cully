//go:build !windows

package cully

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// Only reduced instruments live here: never prompts, screen text, tool payloads,
// credentials, or retrieved memory. The opaque key follows the native thread.
type codexSessionState struct {
	Version                                      int
	Stats                                        codexToolStats
	Model, Input, Output, FiveHour, Weekly, Fast string
	ContextLeft                                  int
	ContextKnown                                 bool
	Started                                      time.Time
}

// An explicit native ID hydrates the first frame before the child opens. Names,
// flags and the interactive resume picker bind through the actual SessionStart.
func codexExplicitResumeID(args []string) string {
	if len(args) < 2 || args[0] != "resume" {
		return ""
	}
	if id, err := uuid.Parse(args[1]); err == nil {
		return id.String()
	}
	return ""
}

func codexSessionStateFile(session string) string {
	b, err := os.ReadFile(codexPaneBindingFile(session))
	if err != nil || len(b) != 32 {
		return ""
	}
	if _, err := hex.DecodeString(string(b)); err != nil {
		return ""
	}
	return filepath.Join(cullyDir(), "codex-session-"+string(b)+".json")
}

func readCodexSessionState(path string) codexSessionState {
	var state codexSessionState
	b, err := os.ReadFile(path)
	if err != nil || len(b) > 16384 || json.Unmarshal(b, &state) != nil || state.Version != 1 {
		return codexSessionState{Version: 1}
	}
	return state
}

func updateCodexSessionState(session string, update func(*codexSessionState)) {
	path := codexSessionStateFile(session)
	if path == "" {
		return
	}
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return
	}
	defer lock.Close()
	if syscall.Flock(int(lock.Fd()), syscall.LOCK_EX) != nil {
		return
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN) //nolint:errcheck
	if _, err := os.Stat(codexPaneRegistrationFile(session)); err != nil {
		return
	}
	state := readCodexSessionState(path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Upgrade an already open pane without discarding its available counters.
		state.Stats = readCodexToolStats(session)
	}
	update(&state)
	b, err := json.Marshal(state)
	if err != nil {
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".codex-session-*")
	if err != nil {
		return
	}
	defer os.Remove(tmp.Name()) //nolint:errcheck
	_, writeErr := tmp.Write(b)
	closeErr := tmp.Close()
	if writeErr == nil && closeErr == nil {
		_ = os.Rename(tmp.Name(), path)
	}
}

func recordCodexSessionTool(session string, class, failure byte) {
	updateCodexSessionState(session, func(state *codexSessionState) {
		state.Stats.Tools++
		switch class {
		case 'S':
			state.Stats.Searches++
		case 'E':
			state.Stats.Edits++
			state.Stats.EditsSinceCheck++
		case 'T':
			state.Stats.Checks++
			if failure != '!' {
				state.Stats.EditsSinceCheck = 0
			}
		}
		if failure == '!' {
			state.Stats.Errors++
		}
	})
}

func restoreCodexSessionView(session string, view *codexStatusView) bool {
	path := codexSessionStateFile(session)
	if path == "" {
		return false
	}
	state := readCodexSessionState(path)
	if state.Model != "" {
		view.Model = state.Model
	}
	if state.Input != "" {
		view.Input = state.Input
	}
	if state.Output != "" {
		view.Output = state.Output
	}
	if state.FiveHour != "" {
		view.FiveHour = state.FiveHour
	}
	if state.Weekly != "" {
		view.Weekly = state.Weekly
	}
	if state.Fast != "" {
		view.Fast = state.Fast
	}
	if state.ContextKnown {
		view.ContextLeft, view.ContextKnown = state.ContextLeft, true
	}
	if !state.Started.IsZero() {
		view.Started = state.Started
	}
	return true
}

func saveCodexSessionView(session string, view codexStatusView) {
	updateCodexSessionState(session, func(state *codexSessionState) {
		if state.Started.IsZero() {
			state.Started = view.Started
		}
		if view.Model != "" {
			state.Model = strings.TrimSpace(view.Model)
		}
		if view.ContextKnown {
			state.ContextLeft, state.ContextKnown = view.ContextLeft, true
		}
		// A narrow footer may temporarily omit instruments. Keep the last native
		// total until a new observed value replaces it; never sum cumulative totals.
		if view.Input != "" {
			state.Input = view.Input
		}
		if view.Output != "" {
			state.Output = view.Output
		}
		if view.FiveHour != "" {
			state.FiveHour = view.FiveHour
		}
		if view.Weekly != "" {
			state.Weekly = view.Weekly
		}
		if view.Fast != "" {
			state.Fast = view.Fast
		}
	})
}
