package cully

import (
	"fmt"
	"strings"
	"time"
)

// sessionHealth holds only values measured from the journal.
type sessionHealth struct {
	Agent        string
	Start        time.Time
	End          time.Time // zero while the session is running
	Tools        int
	Edits        int
	ChecksPassed int
	ChecksFailed int
	Loops        int
	MemorySaves  int
	Verification string
}

func measureHealth(events []journalEvent) sessionHealth {
	var h sessionHealth
	lastEdit, lastCheck := -1, -1
	lastCheckFailed := false
	for i, e := range events {
		if h.Start.IsZero() || (e.Class == journalStart && h.Start.After(e.Time)) {
			h.Start = e.Time
		}
		if h.Agent == "" {
			h.Agent = e.Agent
		}
		switch e.Class {
		case journalEdit:
			h.Edits++
			h.Tools++
			lastEdit = i
		case journalCheck:
			h.Tools++
			if e.Failed {
				h.ChecksFailed++
			} else {
				h.ChecksPassed++
			}
			lastCheck, lastCheckFailed = i, e.Failed
		case journalSearch, journalOther:
			h.Tools++
		case journalMemory:
			h.MemorySaves++
		case journalEnd:
			h.End = e.Time
		}
	}
	switch {
	case lastEdit < 0:
		h.Verification = "no edits"
	case lastCheck < lastEdit:
		h.Verification = "none since last edit"
	case lastCheckFailed:
		h.Verification = "failed"
	default:
		h.Verification = "passed since last edit"
	}
	h.Loops = loopCount(events)
	return h
}

func healthDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return "under 1m"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
}

// renderSessionHealth formats the newest session's measured health. gitFiles is
// the live uncommitted file count, or negative when Git could not be read.
func renderSessionHealth(events []journalEvent, gitFiles int, now time.Time) string {
	if len(events) == 0 {
		return "Session health: nothing recorded yet — start one with cully run AGENT\n"
	}
	h := measureHealth(events)
	var b strings.Builder
	agent := h.Agent
	if agent == "" {
		agent = "unknown"
	}
	b.WriteString("Session health\n")
	fmt.Fprintf(&b, "  Agent          %s\n", agent)
	if h.End.IsZero() {
		fmt.Fprintf(&b, "  Duration       %s (running)\n", healthDuration(now.Sub(h.Start)))
	} else {
		fmt.Fprintf(&b, "  Duration       %s (ended)\n", healthDuration(h.End.Sub(h.Start)))
	}
	fmt.Fprintf(&b, "  Tool calls     %d\n", h.Tools)
	fmt.Fprintf(&b, "  Edits          %d\n", h.Edits)
	fmt.Fprintf(&b, "  Checks         %d passed, %d failed\n", h.ChecksPassed, h.ChecksFailed)
	fmt.Fprintf(&b, "  Verification   %s\n", h.Verification)
	fmt.Fprintf(&b, "  Loops detected %d\n", h.Loops)
	if gitFiles >= 0 {
		fmt.Fprintf(&b, "  Uncommitted    %d files\n", gitFiles)
	}
	fmt.Fprintf(&b, "  Memory saves   %d\n", h.MemorySaves)
	return b.String()
}
