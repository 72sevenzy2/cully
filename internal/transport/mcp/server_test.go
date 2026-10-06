package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcp-runtime/cully/internal/memory"
	"github.com/mcp-runtime/cully/internal/store/remote"
	"github.com/mcp-runtime/cully/internal/transport/datahttp"
	"github.com/modelcontextprotocol/go-sdk/auth"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type fakeStore struct {
	owner string
	calls int
}

func (s *fakeStore) Execute(_ context.Context, owner string, r memory.Request) (memory.Result, error) {
	s.owner = owner
	s.calls++
	return memory.Result{Entry: &memory.Entry{ID: "00000000-0000-0000-0000-000000000001", Summary: r.Log.Summary, Section: r.Log.Section}}, nil
}

type bearerTransport struct{ token string }

func (t bearerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+t.token)
	return http.DefaultTransport.RoundTrip(r)
}
func TestMCPToPrivateAPI(t *testing.T) {
	backend := &fakeStore{}
	data := httptest.NewServer(datahttp.Handler(memory.Service{Store: backend}, "private-service-secret", nil))
	defer data.Close()
	store, err := remote.New(data.URL, "private-service-secret")
	if err != nil {
		t.Fatal(err)
	}
	verifier := func(_ context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
		scopes := []string{"tools:read"}
		if token == "writer" {
			scopes = append(scopes, "tools:write")
		}
		if token != "reader" && token != "writer" {
			return nil, auth.ErrInvalidToken
		}
		return &auth.TokenInfo{UserID: "owner-a", Scopes: scopes, Expiration: time.Now().Add(time.Hour)}, nil
	}
	server := httptest.NewServer(Handler(memory.Service{Store: store}, verifier, "https://issuer.test", "https://public.test/cully/mcp", "/mcp", "test"))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, token := range []string{"writer", "reader"} {
		client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "1"}, nil)
		session, err := client.Connect(ctx, &sdk.StreamableClientTransport{Endpoint: server.URL + "/mcp", HTTPClient: &http.Client{Transport: bearerTransport{token}}, DisableStandaloneSSE: true}, nil)
		if err != nil {
			t.Fatal(err)
		}
		tools, err := session.ListTools(ctx, nil)
		if err != nil || len(tools.Tools) != 8 {
			t.Fatalf("tools=%v err=%v", tools, err)
		}
		result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: "cully_log", Arguments: memory.LogInput{Summary: "Fixed OAuth", Assistant: "codex", Section: "company"}})
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError != (token == "reader") {
			t.Fatalf("scope not enforced: %s %+v", token, result)
		}
		_ = session.Close()
	}
	if backend.calls != 1 || backend.owner != "owner-a" {
		t.Fatalf("owner or permissions lost: %+v", backend)
	}
	for _, path := range []string{"/mcp", "/.well-known/oauth-protected-resource/cully/mcp"} {
		resp, err := http.Get(server.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		expected := 200
		if path == "/mcp" {
			expected = 401
		}
		if resp.StatusCode != expected {
			t.Fatalf("%s status %d", path, resp.StatusCode)
		}
		resp.Body.Close()
	}
}
