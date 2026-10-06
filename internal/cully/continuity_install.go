package cully

import (
	"fmt"
	"os"
	"path/filepath"
)

func continuityCommand(exe, agent, event string) string {
	return quote(exe) + " _internal continuity " + agent + " " + event
}

func executablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(exe)
}

func codexHooksPath() string  { return filepath.Join(CodexConfigDir(), "hooks.json") }
func cursorHooksPath() string { return filepath.Join(CursorConfigDir(), "hooks.json") }

func installCodexContinuityHooks() error {
	exe, err := executablePath()
	if err != nil {
		return err
	}
	path := codexHooksPath()
	m, err := loadHookConfig(path)
	if err != nil {
		return err
	}
	setEventHook(m, "SessionStart", continuityCommand(exe, "codex", "start"), "continuity codex start")
	setEventHook(m, "Stop", continuityCommand(exe, "codex", "stop"), "continuity codex stop")
	return writeSettings(path, m)
}

func uninstallCodexContinuityHooks() error {
	path := codexHooksPath()
	if !fileExists(path) {
		return nil
	}
	m, err := loadHookConfig(path)
	if err != nil {
		return err
	}
	removeEventHook(m, "SessionStart", "", "continuity codex start")
	removeEventHook(m, "Stop", "", "continuity codex stop")
	return writeSettings(path, m)
}

func installCursorContinuityHooks() error {
	exe, err := executablePath()
	if err != nil {
		return err
	}
	path := cursorHooksPath()
	m, err := loadHookConfig(path)
	if err != nil {
		return err
	}
	if version, exists := m["version"]; exists && version != float64(1) {
		return fmt.Errorf("%s: unsupported Cursor hooks version", path)
	}
	m["version"] = 1
	setCursorHook(m, "sessionStart", continuityCommand(exe, "cursor", "start"), "continuity cursor start")
	setCursorHook(m, "afterAgentResponse", continuityCommand(exe, "cursor", "response"), "continuity cursor response")
	setCursorHook(m, "stop", continuityCommand(exe, "cursor", "stop"), "continuity cursor stop")
	setCursorHook(m, "sessionEnd", continuityCommand(exe, "cursor", "end"), "continuity cursor end")
	return writeSettings(path, m)
}

func uninstallCursorContinuityHooks() error {
	path := cursorHooksPath()
	if !fileExists(path) {
		return nil
	}
	m, err := loadHookConfig(path)
	if err != nil {
		return err
	}
	removeCursorHook(m, "sessionStart", "continuity cursor start")
	removeCursorHook(m, "afterAgentResponse", "continuity cursor response")
	removeCursorHook(m, "stop", "continuity cursor stop")
	removeCursorHook(m, "sessionEnd", "continuity cursor end")
	return writeSettings(path, m)
}

func loadHookConfig(path string) (map[string]any, error) {
	m, err := loadSettings(path)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("%s: hooks config must be an object", path)
	}
	if v, ok := m["hooks"]; ok {
		hooks, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s: hooks must be an object", path)
		}
		for event, raw := range hooks {
			if _, ok := raw.([]any); !ok {
				return nil, fmt.Errorf("%s: hooks.%s must be a list", path, event)
			}
		}
	}
	return m, nil
}

func setCursorHook(m map[string]any, event, cmd, sub string) {
	hooks, _ := m["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}
	list := filterCursorHooks(toList(hooks[event]), sub)
	entry := map[string]any{"command": cmd}
	if event == "stop" {
		entry["loop_limit"] = nil // our per-conversation marker prevents the hook's own follow-up loop
	}
	hooks[event] = append(list, entry)
	m["hooks"] = hooks
}

func removeCursorHook(m map[string]any, event, sub string) {
	hooks, _ := m["hooks"].(map[string]any)
	if hooks == nil {
		return
	}
	list := filterCursorHooks(toList(hooks[event]), sub)
	if len(list) == 0 {
		delete(hooks, event)
	} else {
		hooks[event] = list
	}
	if len(hooks) == 0 {
		delete(m, "hooks")
	}
}

func filterCursorHooks(list []any, sub string) []any {
	out := make([]any, 0, len(list))
	for _, item := range list {
		entry, ok := item.(map[string]any)
		if !ok {
			out = append(out, item)
			continue
		}
		cmd, _ := entry["command"].(string)
		if !isCullySubcommand(cmd, sub) {
			out = append(out, item)
		}
	}
	return out
}

func cursorHookConfigured(m map[string]any, event, sub string) bool {
	hooks, _ := m["hooks"].(map[string]any)
	for _, item := range toList(hooks[event]) {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		cmd, _ := entry["command"].(string)
		if isCullySubcommand(cmd, sub) {
			return true
		}
	}
	return false
}
