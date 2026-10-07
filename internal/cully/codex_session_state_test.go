//go:build !windows

package cully

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCodexResumeRestoresThreadMetrics(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	t.Setenv("CULLY_PANE_SESSION", "")
	t.Setenv("MODEL_HINT_GUARD", "")
	cwd := t.TempDir()
	open := func(pane, thread string) {
		t.Helper()
		if err := registerCodexPane(pane, cwd); err != nil {
			t.Fatal(err)
		}
		bindCodexPane(cwd, thread)
	}
	event := func(thread, tool, response string) {
		t.Helper()
		RunCodexSignalHook(strings.NewReader(fmt.Sprintf(`{"cwd":%q,"session_id":%q,"tool_name":%q,"tool_response":%s}`, cwd, thread, tool, response)))
	}
	open("first-pane", "private-native-thread")
	event("private-native-thread", "apply_patch", `{}`)
	event("private-native-thread", "mcp__cully__cully_context", `{"content":[{"text":"private-memory-content"}]}`)
	view := codexStatusView{Model: "GPT-6.1-Sol high", Input: "123K", Output: "8K", ContextKnown: true, ContextLeft: 72, FiveHour: "5h 82% left", Weekly: "Weekly 91% left", Fast: "Fast off", Started: time.Now()}
	saveCodexSessionView("first-pane", view)
	path := codexSessionStateFile("first-pane")
	clearCodexCullyStats("first-pane")
	os.Remove(codexPaneBindingFile("first-pane"))
	os.Remove(codexSignalFile("first-pane"))
	open("resumed-pane", "private-native-thread")
	got := readCodexToolStats("resumed-pane")
	if got.Tools != 2 || got.Edits != 1 || got.EditsSinceCheck != 1 || got.Cully.Context != 1 || got.Cully.Health != "healthy" || got.Cully.Auth != "authenticated" {
		t.Fatal("resume lost metrics", got)
	}
	var restored codexStatusView
	if !restoreCodexSessionView("resumed-pane", &restored) || restored.Input != "123K" || restored.Output != "8K" || restored.ContextLeft != 72 || restored.Model != view.Model || restored.Fast != view.Fast || !restored.Started.Equal(view.Started) {
		t.Fatal("resume lost instruments", restored)
	}
	// New cumulative native totals replace the saved totals without double counting.
	restored.Input = "130K"
	saveCodexSessionView("resumed-pane", restored)
	event("private-native-thread", "mcp__cully__cully_log", `{"isError":true,"content":[{"text":"Unauthorized"}]}`)
	got = readCodexToolStats("resumed-pane")
	if got.Tools != 3 || got.Errors != 1 || got.Cully.Calls != 2 || got.Cully.Log != 1 || got.Cully.Auth != "unauthenticated" {
		t.Fatal("resume did not continue counts", got)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, private := range []string{"private-native-thread", "private-memory-content", "Unauthorized", cwd} {
		if strings.Contains(string(data), private) {
			t.Fatal("private payload saved", private)
		}
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0o600 {
		t.Fatal("state is not private", info.Mode())
	}
	clearCodexCullyStats("resumed-pane")
	os.Remove(codexPaneBindingFile("resumed-pane"))
	open("new-pane", "different-thread")
	if got := readCodexToolStats("new-pane"); got.Tools != 0 || got.Cully.Calls != 0 {
		t.Fatal("different thread inherited metrics", got)
	}
}

func TestCodexDurableTotalsSurviveConcurrentHooksAndWindow(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	t.Setenv("CULLY_PANE_SESSION", "durable")
	t.Setenv("MODEL_HINT_GUARD", "")
	cwd := t.TempDir()
	if err := registerCodexPane("durable", cwd); err != nil {
		t.Fatal(err)
	}
	bindCodexPane(cwd, "thread")
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			RunCodexSignalHook(strings.NewReader(`{"tool_name":"mcp__cully__cully_context","tool_response":{"structuredContent":{"notes":[]}}}`))
			saveCodexSessionView("durable", codexStatusView{Input: "2K"})
		}()
	}
	wg.Wait()
	// The bounded events file no longer determines a bound thread's totals.
	os.WriteFile(codexSignalFile("durable"), []byte(strings.Repeat("O.\n", 5000)), 0o600)
	got := readCodexToolStats("durable")
	if got.Tools != 40 || got.Cully.Calls != 40 || got.Cully.Context != 40 {
		t.Fatal("lost aggregate updates", got)
	}
	saveCodexSessionView("durable", codexStatusView{})
	var restored codexStatusView
	restoreCodexSessionView("durable", &restored)
	if restored.Input != "2K" {
		t.Fatal("missing footer erased saved tokens", restored)
	}
	clearCodexCullyStats("durable")
	RunCodexSignalHook(strings.NewReader(`{"tool_name":"mcp__cully__cully_context","tool_response":{"structuredContent":{"notes":[]}}}`))
	if after := readCodexToolStats("durable"); after.Tools != 40 || after.Cully.Calls != 40 {
		t.Fatal("late hook changed closed thread", after)
	}
}

func TestCodexSessionStateRejectsMalformedBinding(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	registerCodexPane("bad", t.TempDir())
	for _, value := range []string{"../../other", strings.Repeat("z", 32), ""} {
		os.WriteFile(codexPaneBindingFile("bad"), []byte(value), 0o600)
		if codexSessionStateFile("bad") != "" {
			t.Fatal("unsafe binding accepted", value)
		}
	}
}

func TestCodexSessionStateMigratesActiveLegacyPane(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	t.Setenv("CULLY_PANE_SESSION", "legacy")
	t.Setenv("MODEL_HINT_GUARD", "")
	registerCodexPane("legacy", t.TempDir())
	os.WriteFile(codexPaneBindingFile("legacy"), []byte(codexSessionKey("thread")), 0o600)
	os.WriteFile(codexSignalFile("legacy"), []byte("E.\nO.\n"), 0o600)
	os.WriteFile(codexCullyStatsFile("legacy"), []byte(`{"Calls":1,"Context":1,"Health":"healthy","Auth":"authenticated"}`), 0o600)
	RunCodexSignalHook(strings.NewReader(`{"tool_name":"mcp__cully__cully_log","tool_response":{"structuredContent":{"entry":{}}}}`))
	got := readCodexToolStats("legacy")
	if got.Tools != 3 || got.Edits != 1 || got.Cully.Calls != 2 || got.Cully.Context != 1 || got.Cully.Log != 1 {
		t.Fatal("upgrade lost or duplicated legacy metrics", got)
	}
}
