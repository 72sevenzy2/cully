package cully

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const HostedMCPURL = "https://mcp.mcpruntime.org/cully/mcp"

// AddMCP registers a remote service. It neither deploys servers nor stores OAuth tokens.
func AddMCP(w io.Writer, agent, endpoint string, oauth bool) error {
	if err := validateMCPURL(endpoint); err != nil {
		return err
	}
	if agent == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		var detected []string
		for _, candidate := range []string{"claude", "codex", "cursor"} {
			if codingAgentPresent(candidate, cwd) {
				detected = append(detected, candidate)
			}
		}
		if len(detected) != 1 {
			return fmt.Errorf("select one agent with --agent claude, --agent codex, or --agent cursor (detected %d)", len(detected))
		}
		agent = detected[0]
	}
	if agent != "claude" && agent != "codex" && agent != "cursor" {
		return fmt.Errorf("unknown MCP agent %q; choose one of claude, codex or cursor", agent)
	}
	for _, target := range []string{agent} {
		var path string
		var err error
		switch target {
		case "claude":
			path, err = claudeMCPConfigPath()
			if err == nil {
				err = addJSONMCP(path, endpoint, true)
			}
		case "cursor":
			path = filepath.Join(CursorConfigDir(), "mcp.json")
			err = addJSONMCP(path, endpoint, false)
		case "codex":
			path = codexConfigPath()
			err = addTOMLMCP(path, endpoint)
		}
		if err != nil {
			return fmt.Errorf("configure %s MCP: %w", target, err)
		}
		fmt.Fprintf(w, "Cully MCP configured for %s in %s\n", target, path)
		if oauth {
			switch target {
			case "codex":
				fmt.Fprintln(w, "Sign in: codex mcp login cully")
			case "claude":
				fmt.Fprintln(w, "Restart Claude Code, then use /mcp to sign in to cully.")
			case "cursor":
				fmt.Fprintln(w, "Restart Cursor, then sign in to cully in MCP settings.")
			}
		} else {
			fmt.Fprintln(w, "Connect to Cully directly; this server does not require sign-in.")
		}
	}
	fmt.Fprintf(w, "Remote MCP: %s\n", endpoint)
	fmt.Fprintln(w, "The server must be deployed before the connection can succeed. Data, PostgreSQL and Mem0 are configured by its operator.")
	return nil
}

func validateMCPURL(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil || u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.Fragment != "" || u.RawQuery != "" {
		return fmt.Errorf("MCP URL must be an absolute HTTP(S) URL without credentials, query or fragment")
	}
	return nil
}

func claudeMCPConfigPath() (string, error) {
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, ".claude.json"), nil
	}
	home, err := os.UserHomeDir()
	return filepath.Join(home, ".claude.json"), err
}

func addJSONMCP(path, endpoint string, httpType bool) error {
	config, err := loadSettings(path)
	if err != nil {
		return err
	}
	if config == nil {
		return fmt.Errorf("%s: configuration must be an object", path)
	}
	servers, ok := config["mcpServers"].(map[string]any)
	if _, exists := config["mcpServers"]; exists && !ok {
		return fmt.Errorf("%s: mcpServers must be an object", path)
	}
	if servers == nil {
		servers = map[string]any{}
	}
	if entry, exists := servers["cully"]; exists {
		old, ok := entry.(map[string]any)
		if ok && old["url"] == endpoint && old["command"] == nil && (!httpType || old["type"] == "http" || old["type"] == "streamable-http") {
			return nil
		}
		return fmt.Errorf("%s already has a different cully server; edit that entry explicitly before adding another endpoint", path)
	}
	entry := map[string]any{"url": endpoint}
	if httpType {
		entry["type"] = "http"
	}
	servers["cully"] = entry
	config["mcpServers"] = servers
	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return writeMCPConfig(path, append(b, '\n'))
}

func addTOMLMCP(path, endpoint string) error {
	b, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	var config map[string]any
	if err := toml.Unmarshal(b, &config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	servers, ok := config["mcp_servers"].(map[string]any)
	if _, exists := config["mcp_servers"]; exists && !ok {
		return fmt.Errorf("mcp_servers must be a table")
	}
	if entry, exists := servers["cully"]; exists {
		old, ok := entry.(map[string]any)
		if ok && old["url"] == endpoint && old["command"] == nil {
			return nil
		}
		return fmt.Errorf("%s already has a different cully server; edit that entry explicitly before adding another endpoint", path)
	}
	// Append only our table, preserving all existing formatting and comments.
	text := strings.TrimRight(string(b), "\n") + "\n\n[mcp_servers.cully]\nurl = " + strconv.Quote(endpoint) + "\n"
	if err := toml.Unmarshal([]byte(text), &config); err != nil {
		return fmt.Errorf("cannot safely append MCP table; configure cully manually: %w", err)
	}
	return writeMCPConfig(path, []byte(text))
}

func writeMCPConfig(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".cully-mcp-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
