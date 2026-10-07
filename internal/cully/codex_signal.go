package cully

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Codex hook events are reduced to single-letter counters. Commands, prompts,
// tool output, and transcript paths are never written to disk.
type codexToolEvent struct {
	Cwd          string          `json:"cwd"`
	SessionID    string          `json:"session_id"`
	ToolName     string          `json:"tool_name"`
	ToolInput    json.RawMessage `json:"tool_input"`
	ToolResponse json.RawMessage `json:"tool_response"`
}

type codexToolStats struct {
	Tools, Errors, Searches, Edits, Checks, EditsSinceCheck int
	Cully                                                   codexCullyStats
}

var (
	codexSearchCommand = regexp.MustCompile(`(?i)(^|[;&|[:space:]])(rg|grep|find|ls)([[:space:]]|$)`)
	codexCheckCommand  = regexp.MustCompile(`(?i)(^|[;&|[:space:]])(go test|go vet|npm test|npm run test|cargo test|pytest)([[:space:]]|$)`)
)

func codexSignalFile(session string) string {
	return filepath.Join(cullyDir(), safeSession(session)+".codex-events")
}

type codexPaneRegistration struct {
	Session string `json:"session"`
	Cwd     string `json:"cwd"`
	PID     int    `json:"pid"`
}

func codexPaneRegistrationFile(session string) string {
	return filepath.Join(cullyDir(), safeSession(session)+".codex-pane")
}

func codexPaneBindingFile(session string) string {
	return filepath.Join(cullyDir(), safeSession(session)+".codex-pane-session")
}

func registerCodexPane(session, cwd string) error {
	b, err := json.Marshal(codexPaneRegistration{Session: session, Cwd: filepath.Clean(cwd), PID: os.Getpid()})
	if err != nil {
		return err
	}
	return os.WriteFile(codexPaneRegistrationFile(session), b, 0o600)
}

// Codex may execute hooks in a long-running app server that predates the pane.
// Its SessionStart hook binds the Codex session to the sole live pane for this
// working directory; later tool hooks use that binding, not inherited env.
func soleActiveCodexPane(cwd string) (codexPaneRegistration, bool) {
	if cwd == "" {
		return codexPaneRegistration{}, false
	}
	entries, err := os.ReadDir(cullyDir())
	if err != nil {
		return codexPaneRegistration{}, false
	}
	var match codexPaneRegistration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".codex-pane") {
			continue
		}
		path := filepath.Join(cullyDir(), entry.Name())
		info, err := entry.Info()
		if err != nil || time.Since(info.ModTime()) > 10*time.Second {
			continue
		}
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var pane codexPaneRegistration
		if json.Unmarshal(b, &pane) != nil || pane.Session == "" || pane.PID <= 0 ||
			filepath.Clean(pane.Cwd) != filepath.Clean(cwd) || !processAlive(pane.PID) {
			continue
		}
		if match.Session != "" {
			return codexPaneRegistration{}, false
		}
		match = pane
	}
	return match, match.Session != ""
}

func codexSessionKey(id string) string {
	sum := sha256.Sum256([]byte("codex\x00" + id))
	return hex.EncodeToString(sum[:16])
}

func bindCodexPane(cwd, codexSessionID string) {
	if codexSessionID == "" {
		return
	}
	pane, ok := soleActiveCodexPane(cwd)
	if !ok {
		return
	}
	previous := readCodexToolStats(pane.Session)
	f, err := os.OpenFile(codexPaneBindingFile(pane.Session), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return // first SessionStart for this pane keeps its binding
	}
	_, _ = f.WriteString(codexSessionKey(codexSessionID))
	_ = f.Close()
	updateCodexSessionState(pane.Session, func(state *codexSessionState) {
		// Seed early unbound observations only on a new thread; never replace
		// saved totals when opening a resumed session.
		if _, err := os.Stat(codexSessionStateFile(pane.Session)); os.IsNotExist(err) {
			state.Stats = previous
		}
	})
}

func activeCodexPaneSession(cwd, codexSessionID string) string {
	if codexSessionID == "" {
		return ""
	}
	pane, ok := soleActiveCodexPane(cwd)
	if !ok {
		return ""
	}
	b, err := os.ReadFile(codexPaneBindingFile(pane.Session))
	if err != nil || string(b) != codexSessionKey(codexSessionID) {
		return ""
	}
	return pane.Session
}

// RunCodexSignalHook is installed as an asynchronous PostToolUse hook. It is
// inactive outside the opt-in pane, so ordinary Codex sessions have no local
// advisor artifacts.
func RunCodexSignalHook(r io.Reader) {
	var event codexToolEvent
	if json.NewDecoder(io.LimitReader(r, 1<<20)).Decode(&event) != nil || event.ToolName == "" {
		return
	}
	if os.Getenv("MODEL_HINT_GUARD") != "" {
		session := os.Getenv("CULLY_MCP_METRICS_SESSION")
		origin := os.Getenv("CULLY_MCP_ORIGIN")
		if session == "" {
			session, origin = os.Getenv("CULLY_MCP_PROBE_SESSION"), "startup"
		}
		if tool := codexCullyTool(event.ToolName); session != "" && tool != "" {
			health, auth := codexCullyResponseState(event.ToolResponse)
			recordCodexCullyOriginCall(session, tool, health, auth, origin, codexCullySemantic(event.ToolInput))
		}
		return
	}
	session := os.Getenv("CULLY_PANE_SESSION")
	if session == "" {
		session = activeCodexPaneSession(event.Cwd, event.SessionID)
	}
	if session == "" {
		return
	}
	class := codexToolClass(event)
	failure := byte('.')
	if codexToolFailed(event.ToolResponse) {
		failure = '!'
	}
	if tool := codexCullyTool(event.ToolName); tool != "" {
		health, auth := codexCullyResponseState(event.ToolResponse)
		if health == "unhealthy" {
			failure = '!'
		}
		recordCodexCullyOriginCall(session, tool, health, auth, "foreground", codexCullySemantic(event.ToolInput))
	}
	recordCodexSessionTool(session, class, failure)
	f, err := os.OpenFile(codexSignalFile(session), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	_, _ = f.Write([]byte{class, failure, '\n'})
	_ = f.Close()
}

func codexToolClass(event codexToolEvent) byte {
	name := strings.ToLower(event.ToolName)
	// Tool names can include transport namespaces. Recognize the actual shell
	// tool rather than guessing commands embedded in orchestration source text.
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[i+1:]
	}
	if i := strings.LastIndex(name, "__"); i >= 0 {
		name = name[i+2:]
	}
	if name == "apply_patch" || name == "edit" || name == "write" {
		return 'E'
	}
	if name == "bash" || name == "exec_command" {
		var input struct {
			Command string `json:"command"`
			Cmd     string `json:"cmd"`
		}
		_ = json.Unmarshal(event.ToolInput, &input)
		command := input.Command
		if name == "exec_command" {
			command = input.Cmd
		}
		if codexCheckCommand.MatchString(command) {
			return 'T'
		}
		if codexSearchCommand.MatchString(command) {
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
	if path := codexSessionStateFile(session); path != "" {
		if _, err := os.Stat(path); err == nil {
			return readCodexSessionState(path).Stats
		}
	}
	stats := codexToolStats{Cully: readCodexCullyStats(session)}
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
			return []string{"MEMO|Awaiting Codex tool signals. Workflow checks are unavailable until a tool event arrives."}
		}
		return []string{"MEMO|No workflow warning in the observed tool signals."}
	}
	return lines
}
