//go:build !windows

package cully

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestCullyScopePanelKeepsBothCommentsAndCompleteDetails(t *testing.T) {
	s := codexToolStats{Cully: codexCullyStats{Calls: 30, Foreground: 20, Advisor: 8, Startup: 2, SemanticRecall: 6, Health: "healthy", Auth: "authenticated", CheckedAt: "2026-10-07T12:34:56Z", ByTool: map[string]int{"cully_update": 4, "cully_recent": 3, "cully_projects": 2}}}
	advice := []string{"CAUT|" + strings.Repeat("Inspect failures before repeating work. ", 4), "ADV|" + strings.Repeat("Run verification before finishing. ", 4)}
	rows := codexCompactStatusRows(149, 43, advice, s, codexStatusView{})
	text := normalizedCodexPanel(strings.Join(rows, "\n"))
	for _, want := range []string{"Calls 30 observed", "Foreground 20", "Advisor 8", "Startup 2", "Semantic recall 6", "Inspect failures", "Run verification", "Open advisor"} {
		if !strings.Contains(text, want) {
			t.Fatal("compact panel lost", want, text)
		}
	}
	if len(rows)+1 > 22 {
		t.Fatal("panel grew beyond cap")
	}
	details := strings.Join(codexCullyToolRows(s.Cully, 144), "\n")
	for _, want := range []string{"cully_update  4", "cully_recent  3", "cully_projects  2", "cully_delete  0"} {
		if !strings.Contains(details, want) {
			t.Fatal("tool details lost", want)
		}
	}
}

func TestCullyEveryToolAndOriginPersist(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	t.Setenv("CULLY_PANE_SESSION", "every")
	t.Setenv("MODEL_HINT_GUARD", "")
	cwd := t.TempDir()
	registerCodexPane("every", cwd)
	bindCodexPane(cwd, "native-thread")
	for _, tool := range []string{"cully_log", "cully_context", "cully_search", "cully_recall", "cully_get", "cully_update", "cully_delete", "cully_recent", "cully_projects", "cully_future_tool"} {
		RunCodexSignalHook(strings.NewReader(`{"tool_name":"mcp__cully__` + tool + `","tool_response":{"content":[{"text":"private-result"}]}}`))
	}
	RunCodexSignalHook(strings.NewReader(`{"tool_name":"mcp__other__cully_log","tool_response":{}}`))
	t.Setenv("MODEL_HINT_GUARD", "1")
	t.Setenv("CULLY_MCP_METRICS_SESSION", "every")
	t.Setenv("CULLY_MCP_ORIGIN", "advisor")
	RunCodexSignalHook(strings.NewReader(`{"tool_name":"mcp__cully__cully_context","tool_input":{"mode":"semantic","query":"private-query"},"tool_response":{"structuredContent":{"notes":[]}}}`))
	t.Setenv("CULLY_MCP_ORIGIN", "startup")
	RunCodexSignalHook(strings.NewReader(`{"tool_name":"mcp__cully__cully_context","tool_response":{"structuredContent":{"notes":[]}}}`))
	got := readCodexToolStats("every")
	if got.Tools != 11 || got.Cully.Calls != 12 || got.Cully.Foreground != 10 || got.Cully.Advisor != 1 || got.Cully.Startup != 1 || got.Cully.SemanticRecall != 2 || got.Cully.ByTool["cully_future_tool"] != 1 || got.Cully.ByTool["cully_context"] != 3 {
		t.Fatal(got)
	}
	data, _ := os.ReadFile(codexSessionStateFile("every"))
	if strings.Contains(string(data), "private-") || strings.Contains(string(data), "mcp__other") {
		t.Fatal("payload/other server saved")
	}
	clearCodexCullyStats("every")
	os.Remove(codexPaneBindingFile("every"))
	registerCodexPane("resumed", cwd)
	bindCodexPane(cwd, "native-thread")
	if after := readCodexToolStats("resumed"); after.Cully.Calls != 12 || after.Cully.ByTool["cully_projects"] != 1 || after.Cully.Advisor != 1 {
		t.Fatal("resume lost counts", after)
	}
}

func TestCullyScopedConcurrentOriginsAndLegacyCounts(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	cwd := t.TempDir()
	registerCodexPane("origins", cwd)
	bindCodexPane(cwd, "thread")
	updateCodexSessionState("origins", func(s *codexSessionState) { s.Stats.Cully = codexCullyStats{Calls: 5, Log: 2, Context: 1, Other: 2} })
	var wg sync.WaitGroup
	for _, origin := range []string{"foreground", "advisor", "startup"} {
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(origin string) {
				defer wg.Done()
				recordCodexCullyOriginCall("origins", "cully_get", "healthy", "authenticated", origin, false)
			}(origin)
		}
	}
	wg.Wait()
	s := readCodexCullyStats("origins")
	if s.Calls != 65 || s.Foreground != 25 || s.Advisor != 20 || s.Startup != 20 || s.ByTool["cully_get"] != 60 || s.ByTool["earlier_other"] != 2 || s.ByTool["cully_log"] != 2 {
		t.Fatal(s)
	}
	for _, name := range []string{"mcp__cully__cully_log\nsecret", "mcp__other__cully_log", "mcp__cully__cully_log;private"} {
		if codexCullyTool(name) != "" {
			t.Fatal("invalid name accepted", name)
		}
	}
}

func TestAdvisorCullyScopeDoesNotLeakToOtherWorkers(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CODEX_HOME", root)
	t.Setenv("PATH", root)
	t.Setenv("CULLY_MCP_METRICS_SESSION", "inherited-wrong-session")
	script := `#!/bin/sh
[ -z "$CULLY_PANE_SESSION" ] || exit 10
[ "$MODEL_HINT_GUARD" = 1 ] || exit 11
printf '%s:%s' "$CULLY_MCP_METRICS_SESSION" "$CULLY_MCP_ORIGIN"
`
	os.WriteFile(filepath.Join(root, "codex"), []byte(script), 0o755)
	out, err := runAdvisorAgentContext(withCodexMCPScope(context.Background(), "owner-pane", "advisor"), "codex", root, false, "bounded")
	if err != nil || out != "owner-pane:advisor" {
		t.Fatal(out, err)
	}
	out, err = runAdvisorAgentContext(context.Background(), "codex", root, false, "bounded")
	if err != nil || out != ":" {
		t.Fatal("scope leaked", out, err)
	}
}
