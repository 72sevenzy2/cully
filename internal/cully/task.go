package cully

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// taskSuggestionPrefix starts the suggestion Cully writes when a session has
// work but no task. RunApply recognizes it and sets the task directly.
const taskSuggestionPrefix = "Looks like you're working on "

const taskMinEdits = 1

func sessionTaskFile(session string) string {
	return sessionReportFile(session) + ".task"
}

func readTask(session string) string {
	if session == "" {
		return ""
	}
	b, err := os.ReadFile(sessionTaskFile(session))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func writeTask(session, name string) error {
	name = strings.Join(strings.Fields(name), " ")
	if len(name) > 60 {
		name = strings.TrimSpace(name[:60])
	}
	if name == "" {
		_ = os.Remove(sessionTaskFile(session))
		return nil
	}
	return os.WriteFile(sessionTaskFile(session), []byte(name+"\n"), 0o644)
}

var branchPrefixRe = regexp.MustCompile(`^(feat|feature|fix|bugfix|hotfix|chore|docs|refactor|test|product)/`)

// suggestTaskName names the task from the Git branch, which carries no prompt
// or file content. It returns "" on a default branch or a detached head.
func suggestTaskName(branch string) string {
	branch = strings.TrimSpace(branch)
	switch branch {
	case "", "main", "master", "develop", "HEAD":
		return ""
	}
	branch = branchPrefixRe.ReplaceAllString(branch, "")
	name := strings.Join(strings.Fields(strings.NewReplacer("-", " ", "_", " ", "/", " ").Replace(branch)), " ")
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

// ensureTaskSuggestion adds one task suggestion to the session report when the
// session has edits, no task and no earlier task suggestion.
func ensureTaskSuggestion(session, cwd string, events []journalEvent) {
	if session == "" || readTask(session) != "" {
		return
	}
	edits := 0
	for _, e := range events {
		if e.Class == journalEdit {
			edits++
		}
	}
	if edits < taskMinEdits {
		return
	}
	name := suggestTaskName(gitBranch(cwd))
	if name == "" {
		return
	}
	lines := readSuggestionsLimit(session, maxReportLines)
	for _, ln := range lines {
		if strings.Contains(ln, taskSuggestionPrefix) {
			return
		}
	}
	text := fmt.Sprintf("%s%q. Add it as the task to Cully", taskSuggestionPrefix, name)
	_ = writeReportLines(session, cwd, append(lines, text))
}

// taskNameFromSuggestion extracts the quoted name from a task suggestion.
func taskNameFromSuggestion(s string) (string, bool) {
	i := strings.Index(s, taskSuggestionPrefix)
	if i < 0 {
		return "", false
	}
	rest := s[i+len(taskSuggestionPrefix):]
	start := strings.Index(rest, `"`)
	end := strings.LastIndex(rest, `"`)
	if start < 0 || end <= start {
		return "", false
	}
	return rest[start+1 : end], true
}

// RunTask shows, sets or clears the task for the current session.
func RunTask(w io.Writer, args []string) error {
	cwd, _ := os.Getwd()
	session := resolveSession(cwd)
	if session == "" {
		return fmt.Errorf("no session found for this directory")
	}
	if len(args) == 0 {
		if t := readTask(session); t != "" {
			fmt.Fprintln(w, t)
		} else {
			fmt.Fprintln(w, "No task set. Run: cully task NAME")
		}
		return nil
	}
	if len(args) == 1 && args[0] == "--clear" {
		return writeTask(session, "")
	}
	if err := writeTask(session, strings.Join(args, " ")); err != nil {
		return err
	}
	fmt.Fprintln(w, "Task set:", readTask(session))
	fmt.Fprintln(w, "Agent: record this task with cully_session (task name, current session_ref).")
	return nil
}
