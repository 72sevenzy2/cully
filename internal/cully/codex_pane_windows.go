//go:build windows

package cully

import (
	"fmt"
	"os"
)

func RunCodexPane(_ []string, _, _ *os.File) error {
	return fmt.Errorf("cully codex terminal pane is not available on Windows")
}
