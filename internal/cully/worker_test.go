package cully

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunClaudeFindsNativeInstallWithoutShellPath(t *testing.T) {
	for _, location := range []string{"config", "home"} {
		t.Run(location, func(t *testing.T) {
			root := t.TempDir()
			config := filepath.Join(root, "custom-claude")
			t.Setenv("HOME", root)
			t.Setenv("CLAUDE_CONFIG_DIR", config)
			t.Setenv("PATH", filepath.Join(root, "empty-path"))
			binDir := filepath.Join(config, "bin")
			if location == "home" {
				binDir = filepath.Join(root, ".claude", "bin")
			}
			if err := os.MkdirAll(binDir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(binDir, "claude"), []byte("#!/bin/sh\nprintf 'advisor ready\\n'\n"), 0o755); err != nil {
				t.Fatal(err)
			}
			out, err := runClaude("", "session signals")
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(out) != "advisor ready" {
				t.Fatalf("unexpected advisor output: %q", out)
			}
		})
	}
}
