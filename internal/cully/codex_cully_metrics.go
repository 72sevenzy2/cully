//go:build !windows

package cully

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// These are reduced counters for directly observed Cully MCP calls. They
// never contain tool arguments, responses, credentials or retrieved memory.
type codexCullyStats struct {
	Calls, Log, Context, Recall, Search, Get, Other int
	Health, Auth                                    string
	CheckedAt                                       string
}

func codexCullyStatsFile(session string) string {
	return filepath.Join(cullyDir(), safeSession(session)+".codex-mcp")
}

func codexCullyTool(name string) string {
	name = strings.ToLower(name)
	for _, prefix := range []string{"mcp__cully__", "mcp__cully.", "mcp.cully."} {
		if index := strings.LastIndex(name, prefix); index >= 0 {
			tool := name[index+len(prefix):]
			if strings.HasPrefix(tool, "cully_") && !strings.ContainsAny(tool, " .:/") {
				return tool
			}
		}
	}
	return ""
}

// Health is the last observed response, not a background network probe. A
// successful owner-scoped Cully tool establishes authentication at that time.
func codexCullyResponseState(raw json.RawMessage) (health, auth string) {
	var response struct {
		IsError      bool            `json:"isError"`
		IsErrorSnake bool            `json:"is_error"`
		Status       int             `json:"status_code"`
		HTTPStatus   int             `json:"status"`
		Error        json.RawMessage `json:"error"`
		Content      []struct {
			Text string `json:"text"`
		} `json:"content"`
		Structured json.RawMessage `json:"structuredContent"`
	}
	if json.Unmarshal(raw, &response) != nil {
		return "unknown", "unknown"
	}
	failed := response.IsError || response.IsErrorSnake || response.Status >= 400 || response.HTTPStatus >= 400 || len(response.Error) > 0 && string(response.Error) != "null"
	if failed {
		message := strings.ToLower(string(response.Error))
		for _, part := range response.Content {
			if len(message) < 2048 {
				message += " " + strings.ToLower(part.Text)
			}
		}
		unauth := response.Status == 401 || response.HTTPStatus == 401
		for _, phrase := range []string{"unauthenticated", "authentication required", "login required", "not authenticated", "not logged in", "unauthorized", "invalid token", "expired token", "invalid_token", "expired_token", "token_expired", "token is expired", "token has expired"} {
			unauth = unauth || strings.Contains(message, phrase)
		}
		if unauth {
			return "unhealthy", "unauthenticated"
		}
		return "unhealthy", "unknown"
	}
	if len(response.Content) > 0 || len(response.Structured) > 0 && string(response.Structured) != "null" {
		return "healthy", "authenticated"
	}
	return "unknown", "unknown"
}

func readCodexCullyStats(session string) codexCullyStats {
	if path := codexSessionStateFile(session); path != "" {
		if _, err := os.Stat(path); err == nil {
			return readCodexSessionState(path).Stats.Cully
		}
	}
	var stats codexCullyStats
	data, err := os.ReadFile(codexCullyStatsFile(session))
	if err == nil && len(data) <= 8192 {
		_ = json.Unmarshal(data, &stats)
	}
	return stats
}

func recordCodexCullyCall(session, tool, health, auth string) {
	if codexSessionStateFile(session) != "" {
		updateCodexSessionState(session, func(state *codexSessionState) {
			incrementCodexCullyStats(&state.Stats.Cully, tool, health, auth)
		})
		return
	}
	// Late asynchronous hooks must not recreate a closed pane's telemetry.
	if _, err := os.Stat(codexPaneRegistrationFile(session)); err != nil {
		return
	}
	path := codexCullyStatsFile(session)
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return
	}
	defer lock.Close()
	if syscall.Flock(int(lock.Fd()), syscall.LOCK_EX) != nil {
		return
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN) //nolint:errcheck
	if _, err := os.Stat(codexPaneRegistrationFile(session)); err != nil {
		return
	}
	stats := readCodexCullyStats(session)
	incrementCodexCullyStats(&stats, tool, health, auth)
	data, err := json.Marshal(stats)
	if err != nil {
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".cully-mcp-*")
	if err != nil {
		return
	}
	defer os.Remove(tmp.Name()) //nolint:errcheck
	_, writeErr := tmp.Write(data)
	closeErr := tmp.Close()
	if writeErr == nil && closeErr == nil {
		_ = os.Rename(tmp.Name(), path)
	}
}

func incrementCodexCullyStats(stats *codexCullyStats, tool, health, auth string) {
	stats.Calls++
	switch tool {
	case "cully_log":
		stats.Log++
	case "cully_context":
		stats.Context++
	case "cully_recall":
		stats.Recall++
	case "cully_search":
		stats.Search++
	case "cully_get":
		stats.Get++
	default:
		stats.Other++
	}
	if health != "unknown" {
		stats.Health = health
		stats.Auth = auth
		stats.CheckedAt = time.Now().UTC().Format(time.RFC3339)
	}
}

func clearCodexCullyStats(session string) {
	// Stop accepting new events before waiting for any in-flight writer. This
	// prevents a late rename from recreating telemetry after pane shutdown.
	_ = os.Remove(codexPaneRegistrationFile(session))
	path := codexCullyStatsFile(session)
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		_ = os.Remove(path)
		return
	}
	if syscall.Flock(int(lock.Fd()), syscall.LOCK_EX) == nil {
		_ = os.Remove(path)
		_ = syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	}
	_ = lock.Close()
	_ = os.Remove(path + ".lock")
}
