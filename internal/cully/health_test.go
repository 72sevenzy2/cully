package cully

import (
	"strings"
	"testing"
	"time"
)

func TestRenderSessionHealthNoJournal(t *testing.T) {
	got := renderSessionHealth(nil, 0, rescueNow)
	if strings.Count(strings.TrimSpace(got), "\n") != 0 || !strings.Contains(got, "cully run AGENT") {
		t.Fatalf("got %q", got)
	}
}

func TestRenderSessionHealthRunning(t *testing.T) {
	events := []journalEvent{
		{Time: rescueNow.Add(-90 * time.Minute), Agent: "codex", Class: journalStart},
		{Time: rescueNow.Add(-80 * time.Minute), Class: journalEdit},
		{Time: rescueNow.Add(-70 * time.Minute), Class: journalCheck, Failed: true, Sig: "a"},
		{Time: rescueNow.Add(-60 * time.Minute), Class: journalSearch},
		{Time: rescueNow.Add(-50 * time.Minute), Class: journalMemory},
		{Time: rescueNow.Add(-40 * time.Minute), Class: journalNote, Note: noteLoop},
	}
	got := renderSessionHealth(events, 4, rescueNow)
	for _, want := range []string{"codex", "1h30m (running)", "Tool calls     3", "Edits          1", "0 passed, 1 failed", "Verification   failed", "Loops detected 1", "4 files", "Memory saves   1"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in\n%s", want, got)
		}
	}
}

func TestSessionHealthVerification(t *testing.T) {
	edit := journalEvent{Class: journalEdit}
	pass := journalEvent{Class: journalCheck}
	fail := journalEvent{Class: journalCheck, Failed: true}
	cases := map[string][]journalEvent{
		"no edits":               {pass},
		"none since last edit":   {pass, edit},
		"passed since last edit": {edit, fail, pass},
		"failed":                 {edit, pass, edit, fail},
	}
	for want, events := range cases {
		if got := measureHealth(events).Verification; got != want {
			t.Fatalf("%v: got %q want %q", events, got, want)
		}
	}
}

func TestRenderSessionHealthEndedHidesUnknownGit(t *testing.T) {
	events := []journalEvent{
		{Time: rescueNow.Add(-time.Hour), Class: journalStart},
		{Time: rescueNow.Add(-30 * time.Minute), Class: journalEnd},
	}
	got := renderSessionHealth(events, -1, rescueNow)
	if !strings.Contains(got, "30m (ended)") || strings.Contains(got, "Uncommitted") {
		t.Fatalf("got\n%s", got)
	}
}
