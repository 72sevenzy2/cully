package memory

import (
	"errors"
	"testing"
)

func TestValidation(t *testing.T) {
	good := func() Request {
		return Request{Operation: "log", Log: &LogInput{Summary: "Fixed authentication", Assistant: "Codex", Section: "company"}}
	}
	cases := []struct {
		name   string
		change func(*Request)
	}{
		{"credential", func(r *Request) { r.Log.Summary = "api_key=supersecretvalue" }},
		{"section", func(r *Request) { r.Log.Section = "shared" }},
		{"assistant", func(r *Request) { r.Log.Assistant = "anonymous-bot" }},
		{"timestamp", func(r *Request) { r.Log.OccurredAt = "2026-10-06T10:00:00" }},
		{"mismatched operation", func(r *Request) { r.Operation = "get" }},
		{"ambiguous input", func(r *Request) { r.ID = &IDInput{EntryID: "wrong"} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := good()
			tc.change(&r)
			if !errors.Is(r.Validate(), ErrInvalid) {
				t.Fatal("invalid input accepted")
			}
		})
	}
	r := good()
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
	if r.Log.Assistant != "codex" || r.Log.EntryType != "work" {
		t.Fatal("defaults missing")
	}
}
func TestProjectNormalization(t *testing.T) {
	for _, s := range []string{"git@github.com:MCP-Runtime/Cully.git", "https://github.com/MCP-Runtime/Cully.git"} {
		v, e := NormalizeProject(s)
		if e != nil || v != "https://github.com/mcp-runtime/cully" {
			t.Fatalf("%s: %s %v", s, v, e)
		}
	}
	for _, s := range []string{"https://github.com.evil.test/org/repo", "https://github.com/org/repo/issues/1", "https://user:secret@github.com/org/repo", "git@github.com:/repo"} {
		if _, e := NormalizeProject(s); e == nil {
			t.Fatalf("accepted %s", s)
		}
	}
}
func TestUpdateClearingAndLimits(t *testing.T) {
	empty := ""
	r := Request{Operation: "update", Update: &UpdateInput{EntryID: "00000000-0000-0000-0000-000000000001", Learning: &empty}}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
	r.Update.Summary = &empty
	if !errors.Is(r.Validate(), ErrInvalid) {
		t.Fatal("empty summary accepted")
	}
	s := Request{Operation: "search", Search: &SearchInput{Query: "oauth", Limit: 1000}}
	if err := s.Validate(); err != nil || s.Search.Limit != 50 {
		t.Fatal("limit not bounded")
	}
	missing := Request{Operation: "search", Search: &SearchInput{}}
	if !errors.Is(missing.Validate(), ErrInvalid) {
		t.Fatal("search without text accepted")
	}
}
