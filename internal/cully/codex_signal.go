package cully

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Codex hook events are reduced to single-letter counters. Commands, prompts,
// tool output, and transcript paths are never written to disk.
type codexToolEvent struct {
	ToolName     string          `json:"tool_name"`
	ToolInput    json.RawMessage `json:"tool_input"`
	ToolResponse json.RawMessage `json:"tool_response"`
}

type codexToolStats struct {
	Tools, Errors, Searches, Edits, Checks, EditsSinceCheck int
}

var (
	codexSearchCommand = regexp.MustCompile(`(?i)(^|[;&|[:space:]])(rg|grep|find|ls)([[:space:]]|$)`)
	codexCheckCommand  = regexp.MustCompile(`(?i)(^|[;&|[:space:]])(go test|go vet|go fmt|gofmt|npm test|npm run test|cargo test|pytest)([[:space:]]|$)`)
)

func codexSignalFile(session string) string {
	return filepath.Join(cullyDir(), safeSession(session)+".codex-events")
}

// RunCodexSignalHook is installed as an asynchronous PostToolUse hook. It is
// inactive outside the opt-in pane, so ordinary Codex sessions have no local
// advisor artifacts.
func RunCodexSignalHook(r io.Reader) {
	session := os.Getenv("CULLY_PANE_SESSION")
	if session == "" {
		return
	}
	var event codexToolEvent
	if json.NewDecoder(io.LimitReader(r, 1<<20)).Decode(&event) != nil || event.ToolName == "" {
		return
	}
	class := codexToolClass(event)
	failure := byte('.')
	if codexToolFailed(event.ToolResponse) {
		failure = '!'
	}
	f, err := os.OpenFile(codexSignalFile(session), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	_, _ = f.Write([]byte{class, failure, '\n'})
	_ = f.Close()
}

func codexToolClass(event codexToolEvent) byte {
	name := strings.ToLower(event.ToolName)
	if name == "apply_patch" || name == "edit" || name == "write" {
		return 'E'
	}
	if name == "bash" {
		var input struct {
			Command string `json:"command"`
		}
		_ = json.Unmarshal(event.ToolInput, &input)
		if codexCheckCommand.MatchString(input.Command) {
			return 'T'
		}
		if codexSearchCommand.MatchString(input.Command) {
			return 'S'
		}
	}
	if strings.Contains(name, "search") || strings.Contains(name, "read_file") {
		return 'S'
	}
	return 'O'
}

func codexToolFailed(raw json.RawMessage) bool {
	var response struct {
		ExitCode      *int `json:"exit_code"`
		ExitCodeCamel *int `json:"exitCode"`
		IsError       bool `json:"isError"`
		IsErrorSnake  bool `json:"is_error"`
	}
	if json.Unmarshal(raw, &response) != nil {
		return false
	}
	return response.IsError || response.IsErrorSnake ||
		(response.ExitCode != nil && *response.ExitCode != 0) ||
		(response.ExitCodeCamel != nil && *response.ExitCodeCamel != 0)
}

func readCodexToolStats(session string) codexToolStats {
	var stats codexToolStats
	f, err := os.Open(codexSignalFile(session))
	if err != nil {
		return stats
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return stats
	}
	// The pane uses a bounded recent window even during a long session.
	reader := bufio.NewReader(f)
	if info.Size() > 8192 {
		_, _ = f.Seek(info.Size()-8192, io.SeekStart)
		reader.Reset(f)
		_, _ = reader.ReadBytes('\n') // discard a potentially partial first record
	}
	scan := bufio.NewScanner(reader)
	for scan.Scan() {
		line := scan.Bytes()
		if len(line) != 2 {
			continue
		}
		stats.Tools++
		switch line[0] {
		case 'S':
			stats.Searches++
		case 'E':
			stats.Edits++
			stats.EditsSinceCheck++
		case 'T':
			stats.Checks++
			if line[1] != '!' {
				stats.EditsSinceCheck = 0
			}
		}
		if line[1] == '!' {
			stats.Errors++
		}
	}
	return stats
}

func codexAdvice(stats codexToolStats) []string {
	var lines []string
	if stats.Errors >= 3 {
		lines = append(lines, "CAUT|⚠️ Several tool calls failed. Inspect the first failure before retrying.")
	}
	if stats.Searches >= 10 {
		lines = append(lines, "ADV|🔎 Many searches this session. Narrow the path or query before continuing.")
	}
	if stats.EditsSinceCheck > 0 {
		lines = append(lines, "ADV|✓ Files changed. Run a focused check before finishing.")
	}
	if len(lines) == 0 {
		if stats.Tools == 0 {
			return []string{"MEMO|Watching Codex. Advice appears as the session progresses."}
		}
		return []string{"MEMO|Session looks steady. No workflow warning right now."}
	}
	return lines
}
