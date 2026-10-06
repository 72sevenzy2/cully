package main

import "testing"

func TestApplyFlagsAfterNumber(t *testing.T) {
	n, yes, dry, cwd, err := parseApply([]string{"2", "--dry-run", "--yes", "--cwd", "/tmp/my project"})
	if err != nil || n != 2 || !yes || !dry || cwd != "/tmp/my project" {
		t.Fatalf("parsed: %d %v %v %q %v", n, yes, dry, cwd, err)
	}
	for _, args := range [][]string{nil, {"0"}, {"x"}, {"1", "extra"}, {"1", "--unknown"}} {
		if _, _, _, _, err := parseApply(args); err == nil {
			t.Fatalf("accepted invalid arguments: %v", args)
		}
	}
}

func TestRemovedCommands(t *testing.T) {
	for _, name := range []string{"memory", "list", "systems", "plan", "checklist", "debrief", "daemon", "worker", "statusline", "analyze", "cleanup"} {
		if err := run([]string{name}); err == nil {
			t.Fatalf("old public command %s still exists", name)
		}
	}
}
