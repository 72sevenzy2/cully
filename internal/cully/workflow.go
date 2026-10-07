package cully

import "time"

// A workflow hint compares this session with the project's earlier journals.
// It only fires on a clear habit: at least workflowMinSessions earlier sessions
// with edits, a check after the last edit in at least workflowMinShare of them.
const (
	workflowMinSessions = 3
	workflowMinSharePct = 80
	workflowMaxFiles    = 20 // newest journals read; keeps the panel cheap
)

// editsCheckedAfter reports whether a journal has edits and whether a check ran
// after the last one.
func editsCheckedAfter(events []journalEvent) (hasEdits, checked bool) {
	for _, ev := range events {
		switch ev.Class {
		case journalEdit:
			hasEdits, checked = true, false
		case journalCheck:
			if hasEdits {
				checked = true
			}
		}
	}
	return hasEdits, checked
}

func workflowHints(cwd string, current []journalEvent) []string {
	hasEdits, checked := editsCheckedAfter(current)
	if !hasEdits || checked {
		return nil
	}
	var currentStart time.Time
	if len(current) > 0 {
		currentStart = current[0].Time
	}
	sessions, withCheck := 0, 0
	for i, file := range projectJournals(cwd) {
		if i >= workflowMaxFiles {
			break
		}
		events := readJournal(file.Path)
		if len(events) > 0 && events[0].Time.Equal(currentStart) {
			continue // the current session's own journal
		}
		if edits, ok := editsCheckedAfter(events); edits {
			sessions++
			if ok {
				withCheck++
			}
		}
	}
	if sessions < workflowMinSessions || withCheck*100 < sessions*workflowMinSharePct {
		return nil
	}
	return []string{"ADV|You usually run checks after editing in this project. None have run since your last edit."}
}
