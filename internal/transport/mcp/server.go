// Package mcp exposes shared memory through the official Go MCP SDK.
package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"

	"github.com/mcp-runtime/cully/internal/memory"
	"github.com/modelcontextprotocol/go-sdk/auth"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

func add[I any](server *sdk.Server, service memory.Service, name, description string, write bool, request func(I) memory.Request) {
	sdk.AddTool(server, &sdk.Tool{Name: name, Description: description, Annotations: &sdk.ToolAnnotations{ReadOnlyHint: !write}}, func(ctx context.Context, _ *sdk.CallToolRequest, input I) (*sdk.CallToolResult, memory.Result, error) {
		ti := auth.TokenInfoFromContext(ctx)
		if ti == nil || ti.UserID == "" {
			return nil, memory.Result{}, memory.ErrForbidden
		}
		scope := "tools:read"
		if write {
			scope = "tools:write"
		}
		if !slices.Contains(ti.Scopes, scope) {
			return nil, memory.Result{}, memory.ErrForbidden
		}
		out, err := service.Execute(ctx, ti.UserID, request(input))
		return nil, out, err
	})
}
func Handler(service memory.Service, verifier auth.TokenVerifier, issuer, resource, path, version string) http.Handler {
	server := sdk.NewServer(&sdk.Implementation{Name: "Cully", Version: version}, &sdk.ServerOptions{Instructions: "Your companion for better work and everyday life. Recall what the user does and how they work to help with coding-agent orchestration, project management and workflow improvements. Search before answering history questions. Record substantive work, decisions, issues and lessons; remember personal context when the user asks. Local session controls use the cully CLI."})
	add(server, service, "cully_log", "Save a personal or company memory record.", true, func(v memory.LogInput) memory.Request { return memory.Request{Operation: "log", Log: &v} })
	add(server, service, "cully_search", "Search source records using full-text search and optional vectors.", false, func(v memory.SearchInput) memory.Request { return memory.Request{Operation: "search", Search: &v} })
	add(server, service, "cully_recall", "Recall live source records through self-hosted Mem0 semantic memory. Requires configured Mem0 indexing.", false, func(v memory.SearchInput) memory.Request { return memory.Request{Operation: "recall", Search: &v} })
	add(server, service, "cully_recent", "List recent memory records by section or project.", false, func(v memory.RecentInput) memory.Request { return memory.Request{Operation: "recent", Recent: &v} })
	add(server, service, "cully_get", "Get one owned memory record.", false, func(v memory.IDInput) memory.Request { return memory.Request{Operation: "get", ID: &v} })
	add(server, service, "cully_update", "Update selected fields. Empty optional text clears a field.", true, func(v memory.UpdateInput) memory.Request { return memory.Request{Operation: "update", Update: &v} })
	add(server, service, "cully_delete", "Delete one owned record and queue deletion of its Mem0 projection.", true, func(v memory.IDInput) memory.Request { return memory.Request{Operation: "delete", ID: &v} })
	add(server, service, "cully_projects", "List projects with recent memory activity.", false, func(v memory.ProjectsInput) memory.Request {
		return memory.Request{Operation: "projects", Projects: &v}
	})
	mux := http.NewServeMux()
	u, _ := url.Parse(resource)
	metadataPath := "/.well-known/oauth-protected-resource" + u.Path
	metadata := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(oauthex.ProtectedResourceMetadata{Resource: resource, AuthorizationServers: []string{issuer}, ScopesSupported: []string{"tools:read", "tools:write"}, BearerMethodsSupported: []string{"header"}, ResourceName: "Cully"})
	}
	mux.HandleFunc("GET "+metadataPath, metadata)
	if metadataPath != "/.well-known/oauth-protected-resource" {
		mux.HandleFunc("GET /.well-known/oauth-protected-resource", metadata)
	}
	endpoint := sdk.NewStreamableHTTPHandler(func(*http.Request) *sdk.Server { return server }, &sdk.StreamableHTTPOptions{Stateless: true, JSONResponse: true, MaxRequestBodyBytes: 256 << 10})
	mux.Handle(path, auth.RequireBearerToken(verifier, &auth.RequireBearerTokenOptions{ResourceMetadataURL: u.Scheme + "://" + u.Host + metadataPath})(endpoint))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	return mux
}
