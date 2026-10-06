package main

import (
	"os"
	"path/filepath"
	"testing"
)

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

func TestSetupMCPOptionsFailBeforeChangingAgentSetup(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CODEX_HOME", dir)
	for _, args := range [][]string{
		{"codex", "--mcp-url", "not-a-url"},
		{"--mcp-url", "not-a-url", "codex"},
	} {
		if err := runSetup(args); err == nil {
			t.Fatalf("accepted invalid setup options: %v", args)
		}
		if _, err := os.Stat(filepath.Join(dir, "config.toml")); !os.IsNotExist(err) {
			t.Fatalf("modified agent config for invalid options: %v", err)
		}
	}
}

func TestSetupAgentFlagAndLegacyName(t *testing.T) {
	for _, test := range []struct {
		args    []string
		target  string
		oauth   bool
		prepare bool
	}{
		{args: []string{"--agent", "codex"}, target: "codex"},
		{args: []string{"--agent=claude", "--oauth"}, target: "claude", oauth: true},
		{args: []string{"--agent", "cursor", "--prepare"}, target: "cursor", prepare: true},
		{args: []string{"codex", "--oauth"}, target: "codex", oauth: true},
		{args: []string{"--prepare"}, prepare: true},
	} {
		target, oauth, prepare, err := parseSetup(test.args)
		if err != nil || target != test.target || oauth != test.oauth || prepare != test.prepare {
			t.Fatalf("parseSetup(%v) = %q, %v, %v, %v", test.args, target, oauth, prepare, err)
		}
	}
	for _, args := range [][]string{
		{"--agent"}, {"--agent", ""}, {"--agent", "all"}, {"--agent", "other"},
		{"codex", "--agent", "claude"}, {"--agent", "codex", "cursor"},
	} {
		if _, _, _, err := parseSetup(args); err == nil {
			t.Fatalf("accepted invalid setup args: %v", args)
		}
	}
}

func TestMCPAddRequiresDeploymentURL(t *testing.T) {
	if err := run([]string{"mcp", "add", "--agent", "codex"}); err == nil {
		t.Fatal("mcp add accepted a missing deployment URL")
	}
}
