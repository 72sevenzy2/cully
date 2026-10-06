package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Settings struct {
	Address      string
	DatabaseURL  string
	DataAPIURL   string
	DataAPIToken string
	Issuer       string
	Resource     string
	JWKSURL      string
	MCPPath      string
	Mem0URL      string
	Mem0APIKey   string
	MaxConns     int32
}

func value(key, defaultValue string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return defaultValue
}
func Load(data bool) (Settings, error) {
	port := "8080"
	if data {
		port = "8083"
	}
	port = value("CULLY_PORT", port)
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return Settings{}, fmt.Errorf("invalid CULLY_PORT")
	}
	s := Settings{Address: value("CULLY_HOST", "0.0.0.0") + ":" + port, DatabaseURL: os.Getenv("CULLY_DATABASE_URL"), DataAPIURL: os.Getenv("CULLY_DATA_API_URL"), DataAPIToken: os.Getenv("CULLY_DATA_API_TOKEN"), Issuer: strings.TrimRight(value("MCP_AUTH_ISSUER", value("CULLY_AUTH_ISSUER", "https://auth.mcpruntime.org/mcp-auth")), "/"), Resource: strings.TrimRight(value("MCP_AUTH_RESOURCE", value("CULLY_AUTH_RESOURCE", "https://mcp.mcpruntime.org/cully/mcp")), "/"), MCPPath: value("MCP_PATH", "/mcp"), Mem0URL: os.Getenv("CULLY_MEM0_URL"), Mem0APIKey: os.Getenv("CULLY_MEM0_API_KEY"), MaxConns: 8}
	s.JWKSURL = value("CULLY_JWKS_URL", s.Issuer+"/.well-known/jwks.json")
	for _, v := range []string{s.Issuer, s.Resource, s.JWKSURL} {
		u, e := url.Parse(v)
		if e != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			return s, fmt.Errorf("invalid OAuth URL")
		}
	}
	if !strings.HasPrefix(s.MCPPath, "/") || strings.ContainsAny(s.MCPPath, " {}?#") {
		return s, fmt.Errorf("invalid MCP_PATH")
	}
	if v := os.Getenv("CULLY_DB_MAX_CONNS"); v != "" {
		n, e := strconv.ParseInt(v, 10, 32)
		if e != nil || n < 2 || n > 100 {
			return s, fmt.Errorf("CULLY_DB_MAX_CONNS must be 2..100")
		}
		s.MaxConns = int32(n)
	}
	if data && s.DatabaseURL == "" {
		return s, fmt.Errorf("CULLY_DATABASE_URL is required")
	}
	return s, nil
}
