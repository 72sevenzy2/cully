package cully

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// Replay extraction. This turns one tool call into project-relative file
// operations and a coarse command label so `cully replay` can show what the
// agent did. Privacy design (see the journal.go header):
//
//   - Path/To are project-relative, at most journalPathMax characters, and are
//     dropped (never truncated, never stored absolute) when they fall outside
//     the project root or contain control characters.
//   - Op is one of read, write, edit, create, delete, move.
//   - Cmd is only the program and, for known multi-command tools, the first
//     subcommand, validated against journalCmdPattern.
//   - File contents, patch bodies, diffs, command arguments, output, prompts,
//     environment values and URLs are never read into the journal.
//   - When a command is not trivially parseable nothing is recorded.
//
// CULLY_JOURNAL_PATHS=0 disables all of it.

const (
	opRead   = "read"
	opWrite  = "write"
	opEdit   = "edit"
	opCreate = "create"
	opDelete = "delete"
	opMove   = "move"

	journalPathMax   = 200
	journalOpsPerHit = 50 // bound the files taken from one tool call
)

var journalCmdPattern = regexp.MustCompile(`^[A-Za-z0-9._/-]+( [A-Za-z0-9._-]+)?$`)

// journalPathsEnabled reports whether Path/Op/Cmd may be recorded.
func journalPathsEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("CULLY_JOURNAL_PATHS"))) {
	case "0", "false", "off", "no":
		return false
	}
	return true
}

// fileOp is one recorded file operation.
type fileOp struct {
	Path, Op, To string
}

// relProjectPath converts a path to a clean project-relative path, or "" when
// it must not be recorded (empty, control characters, too long, outside the
// project, or the project root itself).
func relProjectPath(cwd, p string) string {
	if p == "" || cwd == "" || len(p) > 4*journalPathMax {
		return ""
	}
	for _, r := range p {
		if unicode.IsControl(r) {
			return ""
		}
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(cwd, p)
	}
	p = filepath.Clean(p)
	roots := []string{filepath.Clean(cwd)}
	if real, err := filepath.EvalSymlinks(roots[0]); err == nil && real != roots[0] {
		roots = append(roots, real)
	}
	for _, root := range roots {
		rel, err := filepath.Rel(root, p)
		if err != nil {
			continue
		}
		rel = filepath.ToSlash(filepath.Clean(rel))
		if rel == "." || rel == ".." || strings.HasPrefix(rel, "../") || filepath.IsAbs(rel) {
			continue
		}
		if len(rel) > journalPathMax {
			return ""
		}
		return rel
	}
	return ""
}

func toolBaseName(name string) string {
	name = strings.ToLower(name)
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[i+1:]
	}
	if i := strings.LastIndex(name, "__"); i >= 0 {
		name = name[i+2:]
	}
	return name
}

// extractToolActivity returns the file operations and command label of a tool
// call. The input is only ever parsed for path-like fields and patch headers.
func extractToolActivity(event codexToolEvent) (ops []fileOp, cmd string) {
	name := toolBaseName(event.ToolName)
	add := func(op, path, to string) {
		if len(ops) >= journalOpsPerHit {
			return
		}
		rel := relProjectPath(event.Cwd, path)
		if rel == "" {
			return
		}
		o := fileOp{Path: rel, Op: op}
		if to != "" {
			if o.To = relProjectPath(event.Cwd, to); o.To == "" {
				return
			}
		}
		ops = append(ops, o)
	}
	switch name {
	case "read", "read_file", "readfile", "view":
		add(opRead, toolInputPath(event.ToolInput, "file_path", "path", "target_file"), "")
	case "write", "create_file":
		add(opWrite, toolInputPath(event.ToolInput, "file_path", "path", "target_file"), "")
	case "edit", "multiedit", "str_replace", "strreplace":
		add(opEdit, toolInputPath(event.ToolInput, "file_path", "path", "target_file"), "")
	case "notebookedit":
		add(opEdit, toolInputPath(event.ToolInput, "notebook_path", "file_path", "path"), "")
	case "delete":
		add(opDelete, toolInputPath(event.ToolInput, "file_path", "path", "target_file"), "")
	case "apply_patch":
		for _, o := range patchFileOps(toolInputText(event.ToolInput)) {
			add(o.Op, o.Path, o.To)
		}
		cmd = "apply_patch"
	case "bash", "exec_command", "shell":
		command := shellCommandOf(event.ToolInput)
		if strings.Contains(command, "*** Begin Patch") {
			for _, o := range patchFileOps(command) {
				add(o.Op, o.Path, o.To)
			}
			return ops, "apply_patch"
		}
		parsed := parseShell(command)
		cmd = parsed.label
		for _, o := range parsed.ops {
			if o.Op == opRead && event.Cwd != "" {
				// A search over a directory is not a file read.
				if info, err := os.Stat(filepath.Join(event.Cwd, o.Path)); err == nil && info.IsDir() {
					continue
				}
			}
			add(o.Op, o.Path, o.To)
		}
	}
	return ops, cmd
}

func toolInputPath(raw json.RawMessage, keys ...string) string {
	var in map[string]json.RawMessage
	if json.Unmarshal(raw, &in) != nil {
		return ""
	}
	for _, k := range keys {
		var s string
		if json.Unmarshal(in[k], &s) == nil && s != "" {
			return s
		}
	}
	return ""
}

func shellCommandOf(raw json.RawMessage) string {
	var in struct {
		Command json.RawMessage `json:"command"`
		Cmd     string          `json:"cmd"`
	}
	if json.Unmarshal(raw, &in) != nil {
		return ""
	}
	var s string
	if json.Unmarshal(in.Command, &s) == nil && s != "" {
		return s
	}
	var list []string // local_shell style argv
	if json.Unmarshal(in.Command, &list) == nil {
		return strings.Join(list, " ")
	}
	return in.Cmd
}

// toolInputText gathers string values of a tool input so patch headers can be
// scanned. Only the headers are kept by patchFileOps.
func toolInputText(raw json.RawMessage) string {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return ""
	}
	var b strings.Builder
	var walk func(any, int)
	walk = func(x any, depth int) {
		if depth > 4 {
			return
		}
		switch t := x.(type) {
		case string:
			b.WriteString(t)
			b.WriteByte('\n')
		case []any:
			for _, e := range t {
				walk(e, depth+1)
			}
		case map[string]any:
			for _, e := range t {
				walk(e, depth+1)
			}
		}
	}
	walk(v, 0)
	return b.String()
}

// patchFileOps reads only the `*** Add/Update/Delete File:` and `*** Move to:`
// header lines of an apply_patch body. Patch bodies are never retained.
func patchFileOps(text string) []fileOp {
	var ops []fileOp
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "*** Add File: "):
			ops = append(ops, fileOp{Path: strings.TrimSpace(line[len("*** Add File: "):]), Op: opCreate})
		case strings.HasPrefix(line, "*** Update File: "):
			ops = append(ops, fileOp{Path: strings.TrimSpace(line[len("*** Update File: "):]), Op: opEdit})
		case strings.HasPrefix(line, "*** Delete File: "):
			ops = append(ops, fileOp{Path: strings.TrimSpace(line[len("*** Delete File: "):]), Op: opDelete})
		case strings.HasPrefix(line, "*** Move to: "):
			if n := len(ops); n > 0 && ops[n-1].Op == opEdit && ops[n-1].To == "" {
				ops[n-1].Op, ops[n-1].To = opMove, strings.TrimSpace(line[len("*** Move to: "):])
			}
		}
		if len(ops) >= journalOpsPerHit {
			break
		}
	}
	return ops
}

// shellSegment is one simple command of a command line.
type shellSegment struct {
	words    []string
	skip     []bool // word contains a glob, variable or other unparseable part
	redirect []string
	unsafe   bool // contains substitution, grouping, stdin redirect or heredoc
}

// splitShell tokenizes a command line into simple commands. It understands
// quotes and backslashes and splits on ; & | and newlines. Anything it cannot
// reason about marks the segment unsafe, so nothing is recorded from it.
func splitShell(command string) []shellSegment {
	var segs []shellSegment
	var cur shellSegment
	var word strings.Builder
	inWord, wordSkip, redirNext := false, false, false
	rs := []rune(command)
	flushWord := func() {
		if !inWord {
			return
		}
		w := word.String()
		switch {
		case redirNext:
			if !wordSkip {
				cur.redirect = append(cur.redirect, w)
			}
			redirNext = false
		default:
			cur.words = append(cur.words, w)
			cur.skip = append(cur.skip, wordSkip)
		}
		word.Reset()
		inWord, wordSkip = false, false
	}
	endSeg := func() {
		flushWord()
		if len(cur.words) > 0 || len(cur.redirect) > 0 || cur.unsafe {
			segs = append(segs, cur)
		}
		cur = shellSegment{}
		redirNext = false
	}
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		switch {
		case r == '\'':
			inWord = true
			for i++; i < len(rs) && rs[i] != '\''; i++ {
				word.WriteRune(rs[i])
				if rs[i] == '$' || rs[i] == '`' {
					wordSkip = true
				}
			}
		case r == '"':
			inWord = true
			for i++; i < len(rs) && rs[i] != '"'; i++ {
				if rs[i] == '\\' && i+1 < len(rs) && strings.ContainsRune(`"\$`+"`", rs[i+1]) {
					i++
				} else if rs[i] == '$' || rs[i] == '`' {
					wordSkip = true
				}
				word.WriteRune(rs[i])
			}
		case r == '\\' && i+1 < len(rs):
			inWord = true
			i++
			word.WriteRune(rs[i])
		case r == ' ' || r == '\t':
			flushWord()
		case r == '\n' || r == ';' || r == '|' || r == '&':
			if r == '&' && i+1 < len(rs) && rs[i+1] == '>' { // &> file
				i++
				flushWord()
				redirNext = true
				if i+1 < len(rs) && rs[i+1] == '>' {
					i++
				}
				continue
			}
			endSeg()
			for i+1 < len(rs) && (rs[i+1] == '&' || rs[i+1] == '|') && r != '\n' {
				i++
			}
		case r == '>':
			if inWord && isAllDigits(word.String()) { // 2>file: drop the fd
				word.Reset()
				inWord = false
			} else {
				flushWord()
			}
			if i+1 < len(rs) && rs[i+1] == '>' {
				i++
			}
			if i+1 < len(rs) && rs[i+1] == '&' { // 2>&1: descriptor duplication
				i++
				for i+1 < len(rs) && rs[i+1] != ' ' && rs[i+1] != '\t' && rs[i+1] != ';' && rs[i+1] != '|' && rs[i+1] != '&' {
					i++
				}
				continue
			}
			redirNext = true
		case r == '<' || r == '(' || r == ')' || r == '`' || r == '{' || r == '}':
			cur.unsafe = true
			inWord = true
			wordSkip = true
			word.WriteRune(r)
		default:
			inWord = true
			if r == '$' || r == '*' || r == '?' || r == '[' || r == '~' && word.Len() == 0 {
				wordSkip = true
			}
			word.WriteRune(r)
		}
	}
	endSeg()
	return segs
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// subcommandPrograms are tools whose first non-flag argument is a subcommand
// rather than data, so it is safe and useful to record ("go test").
var subcommandPrograms = map[string]bool{
	"git": true, "go": true, "npm": true, "pnpm": true, "yarn": true, "cargo": true,
	"docker": true, "kubectl": true, "make": true, "pip": true, "pip3": true, "uv": true,
	"bun": true, "deno": true, "helm": true, "gh": true, "terraform": true, "dotnet": true,
	"mvn": true, "gradle": true, "poetry": true, "brew": true, "npx": false,
}

var (
	envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	subcmdPattern  = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
)

var skippedPrefixes = map[string]bool{"sudo": true, "time": true, "nohup": true, "command": true, "env": true, "exec": true}

// stripPrefix drops env assignments (their values are never kept) and wrapper
// programs from the front of a segment.
func stripPrefix(words []string, skip []bool) ([]string, []bool) {
	for len(words) > 0 {
		w := words[0]
		if i := strings.IndexByte(w, '='); i > 0 && envNamePattern.MatchString(w[:i]) || skippedPrefixes[w] {
			words, skip = words[1:], skip[1:]
			continue
		}
		break
	}
	return words, skip
}

type parsedShell struct {
	label string
	ops   []fileOp
}

// parseShell extracts a command label and best-effort file operations from a
// shell command line. When unsure it records nothing.
func parseShell(command string) parsedShell {
	var out parsedShell
	if len(command) > 8192 {
		return out
	}
	segs := splitShell(command)
	hasHeredoc := strings.Contains(command, "<<")
	changesDir := false
	for _, seg := range segs {
		words, _ := stripPrefix(seg.words, seg.skip)
		if len(words) > 0 && (words[0] == "cd" || words[0] == "pushd" || words[0] == "popd") {
			changesDir = true
		}
	}
	for _, seg := range segs {
		words, skip := stripPrefix(seg.words, seg.skip)
		if len(words) == 0 {
			continue
		}
		prog := filepath.Base(words[0])
		if prog == "cd" || prog == "pushd" || prog == "popd" || prog == "export" || prog == "set" || prog == "source" || prog == "." || prog == "true" {
			continue
		}
		if out.label == "" && !skip[0] {
			out.label = commandLabel(prog, words[1:])
		}
		if seg.unsafe || hasHeredoc || changesDir {
			continue
		}
		out.ops = append(out.ops, segmentOps(prog, words[1:], skip[1:])...)
		for _, target := range seg.redirect {
			out.ops = append(out.ops, fileOp{Path: target, Op: opWrite})
		}
	}
	return out
}

func commandLabel(prog string, args []string) string {
	if !journalCmdPattern.MatchString(prog) || strings.Contains(prog, "..") {
		return ""
	}
	label := prog
	if subcommandPrograms[prog] && len(args) > 0 {
		sub := args[0]
		if !strings.HasPrefix(sub, "-") && subcmdPattern.MatchString(sub) {
			label += " " + sub
		}
	}
	if !journalCmdPattern.MatchString(label) {
		return ""
	}
	return label
}

type argSplit struct {
	flags []string
	paths []string
	ok    bool
}

// plainArgs separates flags from path arguments. Globs and expansions are
// dropped. valueFlags lists short flags that consume the next word; seeing one
// makes the command ambiguous only when skipValue is false.
func plainArgs(args []string, skip []bool, valueFlags map[string]bool) argSplit {
	var s argSplit
	s.ok = true
	rest := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case rest:
			if !skip[i] {
				s.paths = append(s.paths, a)
			}
		case a == "--":
			rest = true
		case strings.HasPrefix(a, "-") && len(a) > 1:
			s.flags = append(s.flags, a)
			if valueFlags[a] {
				i++ // its value is not a path
			}
		case skip[i]:
		default:
			s.paths = append(s.paths, a)
		}
	}
	return s
}

func hasFlag(flags []string, want ...string) bool {
	for _, f := range flags {
		for _, w := range want {
			if f == w {
				return true
			}
		}
	}
	return false
}

func segmentOps(prog string, args []string, skip []bool) []fileOp {
	var ops []fileOp
	paths := func(op string, s argSplit) {
		for _, p := range s.paths {
			ops = append(ops, fileOp{Path: p, Op: op})
		}
	}
	switch prog {
	case "rm", "unlink":
		paths(opDelete, plainArgs(args, skip, nil))
	case "git":
		if len(args) == 0 {
			return nil
		}
		switch args[0] {
		case "rm":
			s := plainArgs(args[1:], skip[1:], nil)
			if !hasFlag(s.flags, "--cached") {
				paths(opDelete, s)
			}
		case "mv":
			s := plainArgs(args[1:], skip[1:], nil)
			if len(s.paths) == 2 {
				ops = append(ops, fileOp{Path: s.paths[0], Op: opMove, To: s.paths[1]})
			}
		}
	case "mv":
		s := plainArgs(args, skip, nil)
		if len(s.paths) == 2 {
			ops = append(ops, fileOp{Path: s.paths[0], Op: opMove, To: s.paths[1]})
		}
	case "cat", "less", "more", "bat", "nl":
		paths(opRead, plainArgs(args, skip, nil))
	case "head", "tail":
		paths(opRead, plainArgs(args, skip, map[string]bool{"-n": true, "-c": true}))
	case "sed":
		s := plainArgs(args, skip, nil)
		// Only the read-only `sed -n SCRIPT FILE...` form is understood.
		if hasFlag(s.flags, "-n") && len(s.flags) == 1 && len(s.paths) >= 2 {
			s.paths = s.paths[1:]
			paths(opRead, s)
		}
	case "rg":
		ops = searchOps(args, skip, "e", "f", "g", "t", "T", "A", "B", "C", "m", "j", "r", "E", "d", "M")
	case "grep":
		ops = searchOps(args, skip, "e", "f", "A", "B", "C", "m", "d", "D")
	case "touch":
		paths(opWrite, plainArgs(args, skip, nil))
	case "tee":
		paths(opWrite, plainArgs(args, skip, nil))
	}
	return ops
}

// searchOps handles rg/grep: the first plain argument is the pattern, the rest
// are paths. Flags that take a value make the command ambiguous, so nothing is
// recorded for them.
func searchOps(args []string, skip []bool, valueShort ...string) []fileOp {
	for _, a := range args {
		if len(a) == 2 && a[0] == '-' {
			for _, v := range valueShort {
				if a[1:] == v {
					return nil
				}
			}
		}
		if strings.HasPrefix(a, "--") && !strings.Contains(a, "=") && len(a) > 2 && valueLongFlags[a] {
			return nil
		}
	}
	s := plainArgs(args, skip, nil)
	if len(s.paths) < 2 {
		return nil
	}
	var ops []fileOp
	for _, p := range s.paths[1:] {
		ops = append(ops, fileOp{Path: p, Op: opRead})
	}
	return ops
}

var valueLongFlags = map[string]bool{"--regexp": true, "--file": true, "--glob": true, "--type": true, "--max-count": true, "--replace": true, "--after-context": true, "--before-context": true, "--context": true, "--include": true, "--exclude": true}
