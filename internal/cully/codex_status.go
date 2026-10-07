//go:build !windows

package cully

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"
)

// Request public native instruments for this child only. The user's config file
// is untouched. Codex omits instruments it cannot supply (including rate limits).
const codexStatusConfig = `tui.status_line=["model-with-reasoning","context-remaining","total-input-tokens","total-output-tokens","five-hour-limit","weekly-limit","fast-mode"]`

type codexStatusView struct {
	Project, Branch, Model, Input, Output, FiveHour, Weekly, Fast string
	ContextLeft                                                   int
	ContextKnown                                                  bool
	Started                                                       time.Time
	Daemon                                                        bool
	LinesAdded, LinesRemoved                                      int64
	ChangesKnown                                                  bool
	ChangedFiles                                                  int
	AdvisorScroll                                                 int
	AdvisorFocused                                                bool
	Terminal                                                      terminalProfile
	Loops                                                         loopReport
	Agent                                                         string
}

var codexContextLeft = regexp.MustCompile(`^Context ([0-9]{1,3})% left$`)

// Read only the native footer in the bottom three terminal rows. Never open a
// transcript or persist screen text. Require a model segment and the exact
// context instrument; token and quota instruments can be clipped on narrow screens.
func readCodexFooter(emulator *vt.Emulator, view *codexStatusView) int {
	for y := max(0, emulator.Height()-3); y < emulator.Height(); y++ {
		var text strings.Builder
		for x := 0; x < emulator.Width(); x++ {
			if cell := emulator.CellAt(x, y); cell != nil {
				text.WriteString(cell.Content)
			}
		}
		parts := strings.Split(strings.TrimSpace(text.String()), " · ")
		if len(parts) < 2 || parts[0] == "" || codexContextLeft.FindStringSubmatch(parts[1]) == nil {
			continue
		}
		left, _ := strconv.Atoi(codexContextLeft.FindStringSubmatch(parts[1])[1])
		if left > 100 {
			continue
		}
		view.Model, view.ContextLeft, view.ContextKnown = parts[0], left, true
		// Narrow native footers omit instruments. Keep their last observed
		// values until Codex supplies an update instead of erasing them.
		for _, part := range parts[2:] {
			switch {
			case strings.HasSuffix(part, " in"):
				view.Input = strings.TrimSuffix(part, " in")
			case strings.HasSuffix(part, " out"):
				view.Output = strings.TrimSuffix(part, " out")
			case strings.HasPrefix(part, "5h"):
				view.FiveHour = part
			case strings.HasPrefix(part, "Weekly"), strings.HasPrefix(part, "Week"), strings.HasPrefix(part, "7d"):
				view.Weekly = part
			case strings.HasPrefix(part, "Fast "):
				view.Fast = part
			}
		}
		return y
	}
	return -1
}

type codexPanelContent struct {
	status, advice, controls []string
	heading                  string
}

func codexStatusRows(cols int, advice []string, stats codexToolStats, view codexStatusView) []string {
	content := codexStatusContent(cols, advice, stats, view)
	rows := append(content.status, content.heading, "")
	rows = append(rows, content.advice...)
	rows = append(rows, "")
	return append(rows, content.controls...)
}

func codexStatusContent(cols int, advice []string, stats codexToolStats, view codexStatusView) codexPanelContent {
	width := max(1, cols-5)
	var rows []string
	add := func(line string) {
		for _, wrapped := range strings.Split(ansi.Wrap(line, width, ""), "\n") {
			rows = append(rows, "  "+ansi.Truncate(wrapped, width, "…")+rst)
		}
	}
	grid := func(left, right []string) {
		if cols < 110 {
			for _, row := range left {
				add(row)
			}
			rows = append(rows, "")
			for _, row := range right {
				add(row)
			}
			return
		}
		const gap = 8
		cellWidth := (width - gap) / 2
		for i := 0; i < max(len(left), len(right)); i++ {
			l, r := "", ""
			if i < len(left) {
				l = left[i]
			}
			if i < len(right) {
				r = right[i]
			}
			if r == "" {
				add(l)
			} else {
				add(l + strings.Repeat(" ", max(0, cellWidth-ansi.StringWidth(l))+gap) + r)
			}
		}
	}
	cellWidth := width
	if cols >= 110 {
		cellWidth = (width - 8) / 2
	}
	phase := codexSessionPhase(stats, view)
	elapsed := "unavailable"
	if !view.Started.IsZero() {
		elapsed = time.Since(view.Started).Truncate(time.Second).String()
	}
	grid([]string{cyan + bold + "✦ Cully" + rst + "    " + formatPhaseBadge(phase)}, []string{dim + "⏱️ elapsed " + rst + elapsed})
	rows = append(rows, "")
	context := dim + "waiting for Codex footer" + rst
	if view.ContextKnown {
		used := 100 - view.ContextLeft
		context = pctColor(used) + gauge(used) + bold + fmt.Sprintf("  %d%% used  ·  %d%% left", used, view.ContextLeft) + rst
		if view.ContextLeft <= 10 {
			context += red + "  /compact" + rst
		}
	}
	location := dim + "project unavailable" + rst
	if view.Project != "" {
		location = cyan + filepath.Base(view.Project) + rst
	}
	if view.Branch != "" {
		location += dim + "  /  " + rst + magenta + view.Branch + rst
	}
	model := blue + bold + view.Model + rst
	if view.Model == "" {
		model = dim + "waiting for Codex footer" + rst
	}
	grid(codexMetricRows("📁 Project", location, cellWidth), codexMetricRows("🤖 Model", model, cellWidth))
	if view.Terminal.Program != "" || view.Terminal.Type != "" {
		add(dim + "Terminal     " + rst + view.Terminal.label())
	}
	add("🧠 Context    " + context)
	rows = append(rows, "")
	left := []string{bold + "Session instruments" + rst, ""}
	right := []string{bold + "🔧 Activity" + rst + dim + "  ·  recent tool window" + rst, ""}
	metric := func(target *[]string, label, value string) {
		*target = append(*target, codexMetricRows(label, value, cellWidth)...)
	}
	metric(&left, "Tokens", "Input "+codexInstrument(view.Input)+"  ·  Output "+codexInstrument(view.Output))
	metric(&left, "5h limit", codexInstrument(strings.TrimPrefix(view.FiveHour, "5h ")))
	weekly := view.Weekly
	for _, prefix := range []string{"Weekly ", "Week ", "7d "} {
		weekly = strings.TrimPrefix(weekly, prefix)
	}
	metric(&left, "Weekly", codexInstrument(weekly))
	metric(&left, "Fast", codexInstrument(strings.TrimPrefix(view.Fast, "Fast ")))
	changes := codexInstrument("")
	if view.ChangesKnown {
		changes = green + fmt.Sprintf("+%d", view.LinesAdded) + rst + " / " + red + fmt.Sprintf("-%d", view.LinesRemoved) + rst + fmt.Sprintf("  ·  %d tracked files", view.ChangedFiles)
	}
	metric(&left, "Working tree", changes)
	for _, counter := range []struct {
		label string
		value int
	}{{"Tools", stats.Tools}, {"Searches", stats.Searches}, {"Edits", stats.Edits}, {"Checks", stats.Checks}, {"Errors", stats.Errors}} {
		value := fmt.Sprint(counter.value)
		if counter.label == "Errors" && counter.value > 0 {
			value = yellow + bold + value + rst
		}
		metric(&right, counter.label, value)
	}
	verification := green + "no pending edits" + rst
	if stats.EditsSinceCheck > 0 {
		verification = yellow + fmt.Sprintf("%d edits awaiting a successful check", stats.EditsSinceCheck) + rst
	}
	metric(&right, "Verification", verification)
	if view.Loops.Active() {
		metric(&right, "Loops", yellow+view.Loops.Summary()+rst)
	}
	grid(left, right)
	rows = append(rows, "")
	for _, line := range codexCullyInstrumentRows(stats, max(1, cols-5)) {
		add(line)
	}
	rows = append(rows, "")
	daemon := yellow + "● daemon offline" + rst
	if view.Daemon {
		daemon = green + "● daemon online" + rst
	}
	heading := magenta + bold + "▸ Advisor" + rst + "    " + daemon
	// Keep the section heading on one row so a constrained panel can always
	// reserve the following row for its first recommendation.
	headingRow := "  " + ansi.Truncate(heading, width, "…") + rst
	statusRows := rows
	var adviceRows []string
	for _, suggestion := range advice {
		if len(adviceRows) > 0 {
			adviceRows = append(adviceRows, "")
		}
		_, text, ok := strings.Cut(suggestion, "|")
		if !ok {
			text = suggestion
		}
		// Keep the hook message without decorative prefixes whose widths vary.
		text = strings.TrimSpace(strings.TrimLeft(text, "⚠️🔎✓ "))
		color, badge := codexAdviceBadge(suggestion)
		prefix := color + badge + rst + strings.Repeat(" ", 8-len(badge))
		for i, wrapped := range strings.Split(ansi.Wrap(text, max(1, width-10), ""), "\n") {
			indent := "        "
			if i == 0 {
				indent = prefix
			}
			adviceRows = append(adviceRows, "    "+indent+wrapped+rst)
		}
	}
	if len(adviceRows) == 0 {
		adviceRows = []string{"    " + dim + "note    " + rst + "Watching Codex tool activity."}
	}
	rows = nil
	add(dim + "Controls    " + rst + cyan + "/prompts:cully" + rst + dim + "    ·    " + rst + cyan + "cully suggestions" + rst)
	return codexPanelContent{status: statusRows, heading: headingRow, advice: adviceRows, controls: rows}
}

type codexAdvisorViewport struct {
	status           []string
	pageSize         int
	spaced, controls bool
}

func codexAdvisorViewportForHeight(availableRows int, content codexPanelContent) codexAdvisorViewport {
	capacity := max(0, availableRows-1) // panel divider belongs to codexPanelLines
	viewport := codexAdvisorViewport{controls: capacity >= 4, spaced: capacity >= 6}
	reserved := 2 // heading plus at least one advice row
	if viewport.controls {
		reserved++
	}
	if viewport.spaced {
		reserved += 2
	}
	statusBudget := max(0, capacity-reserved)
	viewport.status = append([]string(nil), content.status[:min(len(content.status), statusBudget)]...)
	if statusBudget < len(content.status) && len(viewport.status) > 0 {
		contextVisible := false
		for _, row := range viewport.status {
			if strings.Contains(ansi.Strip(row), "Context ") {
				contextVisible = true
				break
			}
		}
		if !contextVisible {
			for _, row := range content.status {
				if strings.Contains(ansi.Strip(row), "Context ") {
					viewport.status[len(viewport.status)-1] = row
					break
				}
			}
		}
	}
	viewport.pageSize = max(0, capacity-len(viewport.status)-reserved+1)
	return viewport
}

// availableRows includes the divider row. Only wrapped advisor messages scroll;
// instruments retain their position while content expands to the panel budget.
func codexStatusRowsForHeight(cols, availableRows int, advice []string, stats codexToolStats, view codexStatusView) []string {
	capacity := max(0, availableRows-1)
	if capacity == 0 {
		return nil
	}
	content := codexStatusContent(cols, advice, stats, view)
	viewport := codexAdvisorViewportForHeight(availableRows, content)
	if viewport.pageSize == 0 {
		return []string{content.heading}
	}
	start := min(max(0, view.AdvisorScroll), max(0, len(content.advice)-viewport.pageSize))
	end := min(len(content.advice), start+viewport.pageSize)
	width := max(1, cols-5)
	position := fmt.Sprintf("%d–%d/%d", start+1, end, len(content.advice))
	if start > 0 {
		position = "↑ " + position
	}
	if end < len(content.advice) {
		position += " ↓"
	}
	focus := ""
	if view.AdvisorFocused {
		focus = " active"
	}
	heading := magenta + bold + "▸ Advisor" + focus + rst + "  " + dim + position + rst
	if cols < 55 {
		heading = magenta + bold + "Advisor" + rst + "  " + dim + position + rst
		if view.AdvisorFocused {
			heading += " *"
		}
	}
	if cols < 30 {
		heading = magenta + bold + "Advisor" + rst + fmt.Sprintf(" %d/%d", start+1, len(content.advice))
		if view.AdvisorFocused {
			heading += " *"
		}
	}
	if view.Daemon {
		heading += "    " + green + "● daemon online" + rst
	} else {
		heading += "    " + yellow + "● daemon offline" + rst
	}
	rows := append(viewport.status, "  "+ansi.Truncate(heading, width, "…")+rst)
	if viewport.spaced {
		rows = append(rows, "")
	}
	rows = append(rows, content.advice[start:end]...)
	if viewport.spaced {
		rows = append(rows, "")
	}
	if viewport.controls {
		control := "Ctrl+] / F6 focus advisor  ·  /prompts:cully  ·  cully suggestions"
		if view.AdvisorFocused {
			control = "↑↓ / PgUp/PgDn scroll  ·  Esc return to Codex"
		}
		if cols < 55 {
			control = "Ctrl+] / F6: advisor"
			if view.AdvisorFocused {
				control = "↑↓ PgUp/PgDn · Esc return"
			}
		}
		if len(viewport.status) < len(content.status) && !view.AdvisorFocused && cols >= 30 {
			control = "Ctrl+] advisor · enlarge for fields"
		}
		if cols < 30 && view.AdvisorFocused {
			control = "↑↓ · Esc return"
		}
		rows = append(rows, "  "+dim+ansi.Truncate(control, width, "…")+rst)
	}
	return rows[:min(capacity, len(rows))]
}

func codexAdvisorScrollMax(cols, availableRows int, advice []string, stats codexToolStats, view codexStatusView) int {
	if availableRows <= 2 {
		return 0
	}
	content := codexStatusContent(cols, advice, stats, view)
	viewport := codexAdvisorViewportForHeight(availableRows, content)
	return max(0, len(content.advice)-max(1, viewport.pageSize))
}

func codexAdvisorPageSize(cols, availableRows int, advice []string, stats codexToolStats, view codexStatusView) int {
	content := codexStatusContent(cols, advice, stats, view)
	return max(1, codexAdvisorViewportForHeight(availableRows, content).pageSize)
}

// Keep values aligned while wrapping each column independently. Continuations
// stay under the value, never under its label or an adjacent instrument.
func codexMetricRows(label, value string, width int) []string {
	labelWidth := min(15, max(6, width/3))
	prefix := dim + label + rst + strings.Repeat(" ", max(2, labelWidth-ansi.StringWidth(label)))
	prefixWidth := ansi.StringWidth(prefix)
	if prefixWidth >= width-4 {
		rows := []string{dim + label + rst}
		for _, line := range strings.Split(ansi.Wrap(value, max(1, width-2), ""), "\n") {
			rows = append(rows, "  "+line)
		}
		return rows
	}
	var rows []string
	for i, line := range strings.Split(ansi.Wrap(value, max(1, width-prefixWidth), ""), "\n") {
		indent := strings.Repeat(" ", prefixWidth)
		if i == 0 {
			indent = prefix
		}
		rows = append(rows, indent+line)
	}
	return rows
}

func codexInstrument(value string) string {
	if value == "" {
		return dim + "unavailable" + rst
	}
	return value
}

// Busy work and context pressure alone do not establish a messy workflow.
// Use explicit failures and unresolved verification, both supplied by hooks.
func codexSessionPhase(stats codexToolStats, view codexStatusView) string {
	if workflowMessy(stats.Tools, stats.Errors, stats.EditsSinceCheck) {
		return "messy"
	}
	if view.ContextKnown && view.ContextLeft <= 10 {
		return "emergency"
	}
	if stats.Tools == 0 {
		return "preflight"
	}
	return "cruise"
}

// Share these visible categories with the expanded drawer. Configuration and
// tool changes expose Apply; ordinary workflow guidance remains Next.
func codexAdviceBadge(suggestion string) (color, badge string) {
	classified := classifySuggestion(suggestion, cullySnapshot{}, cullyState{})
	switch classified.Level {
	case AlertWarn:
		return red, "Warn"
	case AlertCaut:
		return yellow, "Watch"
	case AlertMemo:
		return dim, "Tip"
	}
	if isApplyable(classified) {
		lower := strings.ToLower(classified.Text)
		for _, action := range []string{"mcp", "skill", "graphify", "install", "configure", "enable", "register", "integration", "agents.md", "claude.md", "create ", "add ", "update ", "write ", "rule"} {
			if strings.Contains(lower, action) {
				return magenta, "Apply"
			}
		}
	}
	return cyan, "Next"
}

// The default panel keeps the full terminal width, condenses related metrics
// and previews two actionable comments. The expanded renderer retains every
// instrument and the complete advice text.
func codexCompactStatusRows(cols, availableRows int, advice []string, stats codexToolStats, view codexStatusView) []string {
	availableRows = min(availableRows, 22)
	capacity := max(0, availableRows-1)
	if capacity == 0 {
		return nil
	}
	width := max(1, cols-5)
	status, context := codexCompactInstrumentRows(cols, stats, view)
	selected := codexCompactAdvice(advice)
	var preview []string
	for _, suggestion := range selected {
		_, text, ok := strings.Cut(suggestion, "|")
		if !ok {
			text = suggestion
		}
		text = strings.TrimSpace(strings.TrimLeft(text, "⚠️🔎✓ℹ️ "))
		color, badge := codexAdviceBadge(suggestion)
		textWidth := max(1, width-10)
		wrapped := strings.Split(ansi.Wrap(text, textWidth, ""), "\n")
		for i := 0; i < min(2, len(wrapped)); i++ {
			indent := "        "
			if i == 0 {
				indent = color + badge + rst + strings.Repeat(" ", 8-len(badge))
			}
			message := wrapped[i]
			if i == 1 && len(wrapped) > 2 {
				message = ansi.Truncate(message, max(0, textWidth-1), "") + "…"
			}
			row := "    " + indent + message + rst
			preview = append(preview, row)
		}
	}
	if len(preview) == 0 {
		preview = []string{"    " + dim + "note    Awaiting advisor comments." + rst}
	}
	more := max(0, len(advice)-len(selected))
	suggestions, tips := 0, 0
	for _, item := range advice {
		if strings.TrimSpace(item) == "" {
			continue
		}
		if strings.HasPrefix(item, "MEMO|") {
			tips++
		} else {
			suggestions++
		}
	}
	heading := fmt.Sprintf("💡 Advisor has %d suggestion%s", suggestions, pluralSuffix(suggestions))
	if tips > 0 {
		heading += fmt.Sprintf(" · %d tip%s", tips, pluralSuffix(tips))
	}
	control := cyan + bold + "Ctrl+] / F6 Open advisor" + rst
	control += dim + fmt.Sprintf(" · %d suggestion%s", suggestions, pluralSuffix(suggestions)) + rst
	if tips > 0 {
		control += dim + fmt.Sprintf(" · %d tip%s", tips, pluralSuffix(tips)) + rst
	}
	if more > 0 {
		control = dim + fmt.Sprintf("+%d more  ·  ", more) + rst + control
	}
	footer := "  " + ansi.Truncate(control, width, "…") + rst
	if capacity <= 1 {
		return []string{footer}
	}
	if capacity == 2 {
		return []string{"  " + ansi.Truncate("🧠 Context  "+context, width, "…") + rst, footer}
	}
	// Section titles get breathing room; related metric rows stay together.
	// Smaller layouts shed optional whitespace before shedding information.
	if len(status)+len(preview)+5 <= capacity {
		rows := append([]string(nil), status...)
		rows = append(rows, "", "  "+cyan+ansi.Truncate(heading, width, "…")+rst, "")
		rows = append(rows, preview...)
		rows = append(rows, "", footer)
		return rows
	}
	if len(status)+1+len(preview)+1 <= capacity {
		rows := append(status, "")
		rows = append(rows, preview...)
		return append(rows, footer)
	}
	// Retain both previews when possible, while preserving context even when
	// secondary instruments need the expanded view on a very short terminal.
	previewCount := min(len(preview), max(1, capacity-2))
	statusCount := max(1, capacity-previewCount-1)
	status = status[:min(len(status), statusCount)]
	contextVisible := false
	for _, row := range status {
		if strings.Contains(ansi.Strip(row), "Context ") {
			contextVisible = true
			break
		}
	}
	if !contextVisible {
		status[len(status)-1] = "  " + ansi.Truncate("🧠 Context  "+context, width, "…") + rst
	}
	rows := append(status, preview[:previewCount]...)
	rows = append(rows, footer)
	return rows[:min(capacity, len(rows))]
}

func codexCompactInstrumentRows(cols int, stats codexToolStats, view codexStatusView) ([]string, string) {
	width := max(1, cols-5)
	var rows []string
	add := func(text string) { rows = append(rows, "  "+ansi.Truncate(text, width, "…")+rst) }
	elapsed := "unavailable"
	if !view.Started.IsZero() {
		elapsed = time.Since(view.Started).Truncate(time.Second).String()
	}
	daemon := yellow + "● daemon offline" + rst
	if view.Daemon {
		daemon = green + "● daemon online" + rst
	}
	title := cyan + bold + "✦ Cully" + rst + "  " + formatPhaseBadge(codexSessionPhase(stats, view))
	state := daemon + dim + " · elapsed " + elapsed + rst
	if cols >= 120 {
		add(title + strings.Repeat(" ", max(2, width-ansi.StringWidth(title)-ansi.StringWidth(state))) + state)
	} else {
		add(title + "  " + state)
	}
	rows = append(rows, "")
	model := blue + bold + view.Model + rst
	if view.Model == "" {
		model = dim + "waiting for Codex" + rst
	}
	errors := fmt.Sprint(stats.Errors)
	if stats.Errors > 0 {
		errors = yellow + bold + errors + rst
	}
	verification := green + "no pending edits" + rst
	if stats.EditsSinceCheck > 0 {
		verification = yellow + fmt.Sprintf("%d edits need check", stats.EditsSinceCheck) + rst
	}
	activity := []string{bold + "Model & activity" + rst, "", "Model " + model, fmt.Sprintf("Tools %d · Searches %d", stats.Tools, stats.Searches), fmt.Sprintf("Edits %d · Checks %d · Errors %s", stats.Edits, stats.Checks, errors), "Verification " + verification}
	loopRow := ""
	if view.Loops.Active() {
		loopRow = "Loops " + yellow + view.Loops.Summary() + rst
		activity = append(activity, loopRow)
	}
	location := dim + "Project unavailable" + rst
	if view.Project != "" {
		location = "Project " + cyan + filepath.Base(view.Project) + rst
	}
	if view.Branch != "" {
		location += " / " + magenta + view.Branch + rst
	}
	context := dim + "waiting for Codex footer" + rst
	if view.ContextKnown {
		used := 100 - view.ContextLeft
		context = pctColor(used) + gauge(used) + fmt.Sprintf(" %d%% used · %d%% left", used, view.ContextLeft) + rst
		if view.ContextLeft <= 10 {
			context += red + " /compact" + rst
		}
	}
	tokens := "Tokens " + dim + "unavailable" + rst
	if view.Input != "" || view.Output != "" {
		tokens = "Tokens in " + codexInstrument(view.Input) + " / out " + codexInstrument(view.Output)
	}
	quota := dim + "5h / Weekly unavailable" + rst
	if view.FiveHour != "" || view.Weekly != "" {
		weekly := view.Weekly
		for _, prefix := range []string{"Weekly ", "Week ", "7d "} {
			weekly = strings.TrimPrefix(weekly, prefix)
		}
		quota = "5h " + codexInstrument(strings.TrimPrefix(view.FiveHour, "5h ")) + " · Week " + codexInstrument(weekly)
	}
	changes := dim + "unavailable" + rst
	if view.ChangesKnown {
		changes = green + fmt.Sprintf("+%d", view.LinesAdded) + rst + "/" + red + fmt.Sprintf("-%d", view.LinesRemoved) + rst + fmt.Sprintf(" (%d tracked)", view.ChangedFiles)
	}
	work := "Fast " + codexInstrument(strings.TrimPrefix(view.Fast, "Fast ")) + " · Git " + changes
	if view.Fast == "" && !view.ChangesKnown {
		work = dim + "Fast / Git unavailable" + rst
	}
	usage := []string{bold + "Project & usage" + rst, "", location, "Context " + context, tokens, quota, work}
	cully := codexCullyInstrumentRows(stats, width)
	groups := [][]string{activity, usage, cully}
	if cols >= 120 {
		const gap = 6
		cellWidth := max(1, (width-gap*2)/3)
		var wrapped [3][]string
		for column, group := range groups {
			for _, value := range group {
				wrapped[column] = append(wrapped[column], strings.Split(ansi.Wrap(value, cellWidth, ""), "\n")...)
			}
		}
		for y := 0; y < max(len(wrapped[0]), len(wrapped[1]), len(wrapped[2])); y++ {
			var line strings.Builder
			for column := 0; column < 3; column++ {
				value := ""
				if y < len(wrapped[column]) {
					value = wrapped[column][y]
				}
				line.WriteString(value)
				if column < 2 {
					line.WriteString(strings.Repeat(" ", max(0, cellWidth-ansi.StringWidth(value))+gap))
				}
			}
			add(line.String())
		}
	} else {
		// Stack the same groups, combining adjacent counters and usages so a
		// normal narrow terminal still leaves most rows to the conversation.
		rows = rows[:len(rows)-1] // title and first group belong to one section
		groups = [][]string{
			{activity[0], "Model " + model, fmt.Sprintf("Tools %d · Searches %d · Edits %d · Checks %d · Errors %s", stats.Tools, stats.Searches, stats.Edits, stats.Checks, errors), "Verification " + verification, loopRow},
			{usage[0], location, "Context " + context, tokens + " · " + quota, work},
			{cully[0], fmt.Sprintf("Calls %d observed", stats.Cully.Calls) + " · Auth " + codexObservedState(stats.Cully.Auth, true), fmt.Sprintf("Log %d · Context %d · Recall %d · Search %d · Get %d · Other %d", stats.Cully.Log, stats.Cully.Context, stats.Cully.Recall, stats.Cully.Search, stats.Cully.Get, stats.Cully.Other)},
		}
		for i, group := range groups {
			if i > 0 {
				rows = append(rows, "")
			}
			for _, value := range group {
				if value == "" {
					continue
				}
				for _, line := range strings.Split(ansi.Wrap(value, width, ""), "\n") {
					add(line)
				}
			}
		}
	}
	return rows, context
}

func codexCullyInstrumentRows(stats codexToolStats, width int) []string {
	c := stats.Cully
	health := dim + "Awaiting first MCP response" + rst
	if c.Health == "healthy" {
		health = green + "● Healthy" + rst
	} else if c.Health == "unhealthy" {
		health = red + "● Unhealthy" + rst
	} else if c.Calls > 0 {
		health = dim + "Awaiting response evidence" + rst
	}
	auth := codexObservedState(c.Auth, true)
	rows := []string{bold + "Cully MCP" + rst + " · " + health, "", fmt.Sprintf("Calls %d observed", c.Calls), fmt.Sprintf("Log %d · Context %d · Recall %d", c.Log, c.Context, c.Recall), fmt.Sprintf("Search %d · Get %d · Other %d", c.Search, c.Get, c.Other), "Auth " + auth}
	if c.Foreground+c.Advisor+c.Startup > 0 {
		rows = append(rows, fmt.Sprintf("Foreground %d · Advisor %d · Startup %d", c.Foreground, c.Advisor, c.Startup))
	}
	if c.SemanticRecall > 0 {
		rows = append(rows, fmt.Sprintf("Semantic recall %d", c.SemanticRecall))
	}
	if c.CheckedAt != "" {
		rows = append(rows, "Last observed "+c.CheckedAt)
	}
	return rows
}

func codexCullyToolRows(stats codexCullyStats, width int) []string {
	rows := []string{bold + "Cully MCP tools · entire session" + rst, ""}
	counts := stats.ByTool
	if counts == nil {
		counts = map[string]int{"cully_log": stats.Log, "cully_context": stats.Context, "cully_recall": stats.Recall, "cully_search": stats.Search, "cully_get": stats.Get, "earlier_other": stats.Other}
	}
	names := map[string]bool{}
	for _, name := range []string{"cully_log", "cully_context", "cully_search", "cully_recall", "cully_get", "cully_update", "cully_delete", "cully_recent", "cully_projects"} {
		names[name] = true
	}
	for name := range counts {
		names[name] = true
	}
	ordered := make([]string, 0, len(names))
	for name := range names {
		ordered = append(ordered, name)
	}
	sort.Strings(ordered)
	for _, name := range ordered {
		if name == "earlier_other" && counts[name] == 0 {
			continue
		}
		rows = append(rows, strings.Split(ansi.Wrap(fmt.Sprintf("%s  %d", name, counts[name]), max(1, width), ""), "\n")...)
	}
	return rows
}

func codexObservedState(state string, auth bool) string {
	if auth {
		switch state {
		case "authenticated":
			return green + state + rst
		case "unauthenticated":
			return red + state + rst
		}
	} else {
		switch state {
		case "healthy":
			return green + state + rst
		case "unhealthy":
			return red + state + rst
		}
	}
	return dim + "unknown" + rst
}

func codexCompactAdvice(advice []string) []string {
	var selected []string
	for priority := 0; priority < 4 && len(selected) < 2; priority++ {
		for _, suggestion := range advice {
			kind, _, _ := strings.Cut(suggestion, "|")
			rank := 3
			switch kind {
			case "WARN":
				rank = 0
			case "CAUT":
				rank = 1
			case "ADV":
				rank = 2
			}
			if rank == priority && strings.TrimSpace(suggestion) != "" {
				selected = append(selected, suggestion)
			}
			if len(selected) == 2 {
				break
			}
		}
	}
	return selected
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// Expand to fit wrapped content while leaving enough space for the
// conversation. RunCodexPane retains growth until a real terminal resize
// so advice clearing does not repeatedly expand and shrink the child viewport.
func codexPaneTop(rows, statusRows int) int {
	if rows < 12 {
		return rows
	}
	minimumCodex := 8
	if rows >= 24 {
		minimumCodex = 12
	}
	budget := rows - minimumCodex
	return rows - min(max(0, statusRows)+1, budget)
}

func codexPanelLines(cols, height int, content []string, hud string) []string {
	lines := make([]string, max(0, height))
	if height <= 0 {
		return lines
	}
	width := max(0, cols-1)
	lines[0] = hudRule(width, hud)
	capacity := height - 1
	if len(content) > capacity {
		// Keep advice visible on shorter terminals. Drop secondary instruments
		// before the advisor heading and first recommendation.
		advisor := -1
		for i, row := range content {
			if strings.HasPrefix(strings.TrimSpace(ansi.Strip(row)), "▸ Advisor") {
				advisor = i
				break
			}
		}
		if advisor >= 0 && capacity >= 2 {
			// Retain status instruments, reserving a heading, recommendation and
			// explicit overflow control instead of discarding half the status block.
			reserve := 2
			if capacity >= 4 {
				reserve = 3
			}
			statusCount := min(advisor, capacity-reserve)
			advisorRows := []string{content[advisor]}
			for _, row := range content[advisor+1:] {
				if strings.TrimSpace(ansi.Strip(row)) != "" {
					advisorRows = append(advisorRows, row)
				}
			}
			advisorCount := min(len(advisorRows), capacity-statusCount)
			selected := append([]string(nil), content[:min(advisor, statusCount)]...)
			if statusCount == 1 {
				// In a tiny panel, context is more useful than its title badge.
				for _, row := range content[:advisor] {
					if strings.Contains(ansi.Strip(row), "Context ") {
						selected[0] = row
						break
					}
				}
			}
			selected = append(selected, advisorRows[:advisorCount]...)
			if advisorCount < len(advisorRows) || statusCount < advisor {
				last := len(selected) - 1
				if advisorCount >= 3 {
					selected[last] = dim + ansi.Truncate(" … enlarge terminal for all fields", width, "…") + rst
				} else {
					selected[last] = ansi.Truncate(selected[last], max(0, width-1), "") + "…" + rst
				}
			}
			content = selected
		}
	}
	for i := 1; i < height && i <= len(content); i++ {
		lines[i] = ansi.Truncate(content[i-1], width, "…")
	}
	return lines
}
