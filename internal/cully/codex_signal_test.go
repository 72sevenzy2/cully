package cully

import (
	"os"
	"strings"
	"testing"
)

func TestCodexSignalHookStoresCountersWithoutPayloads(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	t.Setenv("CULLY_PANE_SESSION", "pane-a")
	RunCodexSignalHook(strings.NewReader(`{"tool_name":"Bash","tool_input":{"command":"go test ./... secret-value"},"tool_response":{"exit_code":1,"output":"private-output"}}`))
	data, err := os.ReadFile(codexSignalFile("pane-a"))
	if err != nil || string(data) != "T!\n" {
		t.Fatalf("unexpected signal data: %q, %v", data, err)
	}
	stats := readCodexToolStats("pane-a")
	if stats.Tools != 1 || stats.Errors != 1 || stats.Checks != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	for i := 0; i < 3; i++ {
		RunCodexSignalHook(strings.NewReader(`{"tool_name":"apply_patch","tool_response":{}}`))
	}
	stats = readCodexToolStats("pane-a")
	if stats.Edits != 3 || stats.Checks != 1 || stats.EditsSinceCheck != 3 {
		t.Fatalf("unexpected stats after edits: %+v", stats)
	}
	if got := codexAdvice(stats); len(got) != 1 || !strings.Contains(got[0], "focused check") {
		t.Fatalf("unexpected advice: %v", got)
	}
	RunCodexSignalHook(strings.NewReader(`{"tool_name":"Bash","tool_input":{"command":"go test ./..."},"tool_response":{"exit_code":0}}`))
	if got := codexAdvice(readCodexToolStats("pane-a")); len(got) != 1 || !strings.HasPrefix(got[0], "MEMO|") {
		t.Fatalf("successful check did not clear edit advice: %v", got)
	}
}

func TestCodexSignalHookOnlyRunsInsidePane(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	t.Setenv("CULLY_PANE_SESSION", "")
	RunCodexSignalHook(strings.NewReader(`{"tool_name":"Bash","tool_input":{"command":"rg secret"}}`))
	if _, err := os.Stat(codexSignalFile("pane-a")); !os.IsNotExist(err) {
		t.Fatalf("signal file created outside pane: %v", err)
	}
}

func TestCodexAdviceUsesBoundedCounters(t *testing.T) {
	got := codexAdvice(codexToolStats{Tools: 20, Errors: 3, Searches: 11, Edits: 2, EditsSinceCheck: 2})
	if len(got) != 3 || !strings.Contains(got[0], "failed") || !strings.Contains(got[1], "searches") || !strings.Contains(got[2], "check") {
		t.Fatalf("unexpected advice: %v", got)
	}
}
