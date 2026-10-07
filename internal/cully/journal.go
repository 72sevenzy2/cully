package cully

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// The session journal is the semantic record of one coding session: when the
// agent edited, searched, ran checks, saved memory, or failed. Timeline,
// handoff, rescue, loop detection and status are all derived from it.
//
// It is deliberately coarse. It never stores prompts, commands, tool output,
// file names or file contents. A repeated failure is recognised by a one-way
// hash of the normalized command, never by the command itself.

// Journal classes. The first four match the single-letter hook counters.
const (
	journalEdit   = "E" // file edit or patch
	journalCheck  = "T" // test, vet or build check
	journalSearch = "S" // search or read
	journalOther  = "O" // any other tool
	journalMemory = "M" // a Cully memory tool call
	journalNote   = "N" // a Cully observation; Note holds a fixed keyword
	journalStart  = "B" // terminal session started
	journalEnd    = "X" // terminal session ended
)

// Fixed vocabulary for journalNote events. Free text is never recorded.
const (
	noteLoop      = "loop"
	noteVerifyGap = "verify-gap"
)

const (
	journalMaxBytes = 256 * 1024
	journalKeep     = 40
	journalMaxAge   = 30 * 24 * time.Hour
)

type journalEvent struct {
	Time   time.Time `json:"t"`
	Agent  string    `json:"a,omitempty"`
	Class  string    `json:"c"`
	Failed bool      `json:"f,omitempty"`
	Tool   string    `json:"tool,omitempty"`
	Sig    string    `json:"sig,omitempty"`
	Note   string    `json:"n,omitempty"`
}

func journalDir() string { return filepath.Join(cullyDir(), "journals") }

// projectKey identifies a project directory without storing its path.
func projectKey(cwd string) string {
	sum := sha256.Sum256([]byte(filepath.Clean(cwd)))
	return hex.EncodeToString(sum[:6])
}

func journalPath(cwd, session string) string {
	return filepath.Join(journalDir(), projectKey(cwd)+"-"+safeSession(session)+".jsonl")
}

// appendJournal records one event. It never fails the caller: hooks must stay
// silent and fast.
func appendJournal(cwd, session string, event journalEvent) {
	if session == "" || cwd == "" {
		return
	}
	if event.Time.IsZero() {
		event.Time = time.Now().UTC()
	}
	if err := os.MkdirAll(journalDir(), 0o700); err != nil {
		return
	}
	path := journalPath(cwd, session)
	line, err := json.Marshal(event)
	if err != nil {
		return
	}
	trimJournal(path)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	_, _ = f.Write(append(line, '\n'))
	_ = f.Close()
}

// trimJournal keeps a very long session bounded by discarding its oldest half.
func trimJournal(path string) {
	info, err := os.Stat(path)
	if err != nil || info.Size() < journalMaxBytes {
		return
	}
	events := readJournal(path)
	events = events[len(events)/2:]
	var out strings.Builder
	for _, event := range events {
		if line, err := json.Marshal(event); err == nil {
			out.Write(line)
			out.WriteByte('\n')
		}
	}
	_ = os.WriteFile(path, []byte(out.String()), 0o600)
}

func readJournal(path string) []journalEvent {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var events []journalEvent
	scan := bufio.NewScanner(f)
	scan.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for scan.Scan() {
		var event journalEvent
		if json.Unmarshal(scan.Bytes(), &event) == nil && event.Class != "" {
			events = append(events, event)
		}
	}
	return events
}

type journalFile struct {
	Path     string
	Session  string
	Modified time.Time
}

// projectJournals lists a project's journals, newest first.
func projectJournals(cwd string) []journalFile {
	prefix := projectKey(cwd) + "-"
	entries, _ := os.ReadDir(journalDir())
	var files []journalFile
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, journalFile{
			Path:     filepath.Join(journalDir(), name),
			Session:  strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".jsonl"),
			Modified: info.ModTime(),
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Modified.After(files[j].Modified) })
	return files
}

// latestJournal returns the newest journal for a project.
func latestJournal(cwd string) (journalFile, []journalEvent, bool) {
	files := projectJournals(cwd)
	if len(files) == 0 {
		return journalFile{}, nil, false
	}
	return files[0], readJournal(files[0].Path), true
}

// pruneJournals removes old journals, keeping the newest few for every project.
func pruneJournals() {
	entries, _ := os.ReadDir(journalDir())
	type item struct {
		path     string
		modified time.Time
	}
	var all []item
	for _, entry := range entries {
		info, err := entry.Info()
		if entry.IsDir() || err != nil || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		all = append(all, item{filepath.Join(journalDir(), entry.Name()), info.ModTime()})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].modified.After(all[j].modified) })
	for i, it := range all {
		if i >= journalKeep || time.Since(it.modified) > journalMaxAge {
			_ = os.Remove(it.path)
		}
	}
}

var commandWhitespace = regexp.MustCompile(`\s+`)

// commandSig returns a short one-way hash of a shell command after collapsing
// whitespace and case. Two runs of the same command share a signature; the
// command itself is never stored.
func commandSig(command string) string {
	command = strings.ToLower(strings.TrimSpace(commandWhitespace.ReplaceAllString(command, " ")))
	if command == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(command))
	return hex.EncodeToString(sum[:4])
}
