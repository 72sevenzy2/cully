package e2e

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/mcp-runtime/cully/internal/memory"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestContainerStack crosses the live MCP, data API and PostgreSQL containers.
// The workflow supplies a disposable stack; local unit runs skip it.
func TestContainerStack(t *testing.T) {
	endpoint := os.Getenv("CULLY_E2E_MCP_URL")
	if endpoint == "" {
		t.Skip("CULLY_E2E_MCP_URL not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client := sdk.NewClient(&sdk.Implementation{Name: "cully-e2e", Version: "1"}, nil)
	session, err := client.Connect(ctx, &sdk.StreamableClientTransport{Endpoint: endpoint, DisableStandaloneSSE: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	tools, err := session.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) == 0 {
		t.Fatalf("list tools: %v, %+v", err, tools)
	}
	call := func(name string, input any) memory.Result {
		t.Helper()
		response, err := session.CallTool(ctx, &sdk.CallToolParams{Name: name, Arguments: input})
		if err != nil || response.IsError {
			t.Fatalf("%s: %v, %+v", name, err, response)
		}
		data, err := json.Marshal(response.StructuredContent)
		if err != nil {
			t.Fatal(err)
		}
		var result memory.Result
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	logged := call("cully_log", memory.LogInput{Summary: "Disposable container end-to-end check", Assistant: "ci", Section: "company"}).Entry
	if logged == nil || logged.ID == "" {
		t.Fatal("MCP write did not reach PostgreSQL")
	}
	id := memory.IDInput{EntryID: logged.ID}
	defer call("cully_delete", id)
	got := call("cully_get", id).Entry
	if got == nil || got.Summary != logged.Summary {
		t.Fatalf("MCP read did not return the written record: %+v", got)
	}
}
