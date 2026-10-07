package cully

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// A stage is a run of related activity: first the agent explored, then it
// implemented, then it verified. Stages are built by fixed rules from the
// journal only; nothing is inferred beyond what was recorded.

const (
	stageExplore   = "Explore"
	stageImplement = "Implement"
	stageVerify    = "Verify"
	stageFix       = "Fix"
	stageRemember  = "Remember"
	stageRun       = "Run"
)

type stageFile struct {
	Path   string `json:"path"`
	Op     string `json:"op"`
	To     string `json:"to,omitempty"`
	Count  int    `json:"count"`
	Failed int    `json:"failed,omitempty"`
}

type stageCmd struct {
	Cmd    string `json:"cmd"`
	Count  int    `json:"count"`
	Failed int    `json:"failed"`
}

type stage struct {
	N        int         `json:"n"`
	Kind     string      `json:"kind"`
	Start    time.Time   `json:"start"`
	End      time.Time   `json:"end"`
	Seconds  int         `json:"seconds"`
	Reads    int         `json:"reads,omitempty"`
	Searches int         `json:"searches,omitempty"`
	Edits    int         `json:"edits,omitempty"`
	Checks   int         `json:"checks,omitempty"`
	Runs     int         `json:"runs,omitempty"`
	Memory   int         `json:"memory,omitempty"`
	Passed   int         `json:"passed,omitempty"`
	Failed   int         `json:"failed,omitempty"`
	Status   string      `json:"status"` // passed, failed, mixed or none
	Files    []stageFile `json:"files"`
	Cmds     []stageCmd  `json:"cmds"`
	Loop     bool        `json:"loop,omitempty"`
	Retries  int         `json:"retries,omitempty"` // on the first failed Verify of a retry cycle
	RetryOf  int         `json:"retry_of,omitempty"`
	RetryNo  int         `json:"retry_no,omitempty"`
}

func (s stage) Duration() time.Duration { return s.End.Sub(s.Start) }

// stageKindOf classifies one journal event; ok is false for neutral events
// (session markers, notes, untyped tool calls) which never split a stage.
func stageKindOf(e journalEvent, afterFailedCheck bool) (string, bool) {
	change := func() (string, bool) {
		if afterFailedCheck {
			return stageFix, true
		}
		return stageImplement, true
	}
	switch e.Class {
	case journalCheck:
		return stageVerify, true
	case journalEdit:
		return change()
	case journalMemory:
		return stageRemember, true
	case journalSearch:
		return stageExplore, true
	case journalFileOp:
		if e.Op == opRead {
			return stageExplore, true
		}
		if isChangeOp(e.Op) {
			return change()
		}
	case journalOther:
		switch {
		case e.Op == opRead:
			return stageExplore, true
		case isChangeOp(e.Op):
			return change()
		case e.Cmd != "":
			return stageRun, true
		}
	}
	return "", false
}

// buildStages segments a journal into stages. A kind change or a gap over five
// minutes starts a new stage; edits after a failed check are Fix stages; a
// failed Verify followed by Fix and another Verify is a retry cycle.
func buildStages(events []journalEvent) []stage {
	var stages []stage
	var lastT time.Time
	afterFailed := false
	for _, e := range events {
		kind, ok := stageKindOf(e, afterFailed)
		if !ok {
			continue
		}
		if n := len(stages); n == 0 || stages[n-1].Kind != kind || e.Time.Sub(lastT) > stageGap {
			stages = append(stages, stage{N: len(stages) + 1, Kind: kind, Start: e.Time, Files: []stageFile{}, Cmds: []stageCmd{}})
		}
		s := &stages[len(stages)-1]
		s.End, lastT = e.Time, e.Time
		switch {
		case e.Class == journalCheck:
			s.Checks++
			if e.Failed {
				s.Failed++
			} else {
				s.Passed++
			}
			addStageCmd(s, e.Cmd, e.Failed)
		case e.Class == journalMemory:
			s.Memory++
		case kind == stageRun:
			s.Runs++
			addStageCmd(s, e.Cmd, e.Failed)
		case kind == stageExplore:
			if e.Op == opRead {
				s.Reads++
			} else {
				s.Searches++
			}
		default:
			s.Edits++
		}
		if e.Op != "" && e.Path != "" {
			addStageFile(s, e)
		}
		if e.Class == journalCheck {
			afterFailed = e.Failed
		}
	}
	for i := range stages {
		s := &stages[i]
		s.Seconds = int(s.Duration().Seconds())
		switch {
		case s.Passed > 0 && s.Failed > 0:
			s.Status = "mixed"
		case s.Failed > 0:
			s.Status = "failed"
		case s.Passed > 0:
			s.Status = "passed"
		default:
			s.Status = "none"
		}
		sort.SliceStable(s.Files, func(a, b int) bool { return s.Files[a].Count > s.Files[b].Count })
	}
	markRetries(stages)
	markStageLoops(stages, events)
	return stages
}

func addStageCmd(s *stage, cmd string, failed bool) {
	if cmd == "" {
		cmd = "check"
		if s.Kind == stageRun {
			cmd = "command"
		}
	}
	for i := range s.Cmds {
		if s.Cmds[i].Cmd == cmd {
			s.Cmds[i].Count++
			if failed {
				s.Cmds[i].Failed++
			}
			return
		}
	}
	c := stageCmd{Cmd: cmd, Count: 1}
	if failed {
		c.Failed = 1
	}
	s.Cmds = append(s.Cmds, c)
}

func addStageFile(s *stage, e journalEvent) {
	for i := range s.Files {
		f := &s.Files[i]
		if f.Path == e.Path && f.Op == e.Op && f.To == e.To {
			f.Count++
			if e.Failed {
				f.Failed++
			}
			return
		}
	}
	f := stageFile{Path: e.Path, Op: e.Op, To: e.To, Count: 1}
	if e.Failed {
		f.Failed = 1
	}
	s.Files = append(s.Files, f)
}

// markRetries links each re-run of a failing Verify, after a Fix, back to the
// first failure of that cycle.
func markRetries(stages []stage) {
	origin, fixed := -1, false
	for i := range stages {
		switch stages[i].Kind {
		case stageFix:
			fixed = true
		case stageVerify:
			failing := stages[i].Status == "failed" || stages[i].Status == "mixed"
			if origin >= 0 && fixed {
				stages[origin].Retries++
				stages[i].RetryOf, stages[i].RetryNo = stages[origin].N, stages[origin].Retries
			}
			switch {
			case !failing:
				origin = -1
			case origin < 0:
				origin = i
			}
			fixed = false
		}
	}
}

// markStageLoops flags stages that overlap a detected loop or hold a loop note.
func markStageLoops(stages []stage, events []journalEvent) {
	for _, f := range detectLoops(events).Loops {
		for i := range stages {
			if !stages[i].End.Before(f.First) && !stages[i].Start.After(f.Last) {
				stages[i].Loop = true
			}
		}
	}
	for _, e := range events {
		if e.Class != journalNote || e.Note != noteLoop {
			continue
		}
		for i := len(stages) - 1; i >= 0; i-- {
			if !stages[i].Start.After(e.Time) {
				stages[i].Loop = true
				break
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Rendering

func opBadge(op string) string {
	switch op {
	case opCreate:
		return "+"
	case opEdit, opWrite:
		return "~"
	case opDelete:
		return "−"
	case opMove:
		return "→"
	}
	return "·"
}

func stageGlyph(s stage) string {
	switch s.Status {
	case "passed":
		return " ✓"
	case "failed":
		return " ✕"
	case "mixed":
		return " ±"
	}
	return ""
}

func stageColor(s stage) string {
	switch s.Kind {
	case stageExplore:
		return cyan
	case stageImplement:
		return blue
	case stageFix:
		return yellow
	case stageRemember:
		return magenta
	case stageVerify:
		switch s.Status {
		case "passed":
			return green
		case "failed":
			return red
		case "mixed":
			return yellow
		}
	}
	return dim
}

func stageTimes(s stage) string {
	return stampOf(s.Start) + " · " + formatDuration(s.Duration())
}

func cmdLabel(c stageCmd) string {
	out := safeText(c.Cmd)
	switch {
	case c.Failed == 0:
		out += " ✓"
	case c.Failed == c.Count:
		out += " ✕"
	default:
		out += fmt.Sprintf(" ✓%d ✕%d", c.Count-c.Failed, c.Failed)
	}
	if c.Count > 1 && (c.Failed == 0 || c.Failed == c.Count) {
		out += fmt.Sprintf(" ×%d", c.Count)
	}
	return out
}

func fileLabel(f stageFile) string {
	name := safeText(f.Path)
	if f.To != "" {
		name += " → " + safeText(f.To)
	}
	label := opBadge(f.Op) + " " + name
	if f.Count > 1 {
		label += fmt.Sprintf(" ×%d", f.Count)
	}
	if f.Failed > 0 {
		label += " ✕"
	}
	return label
}

func countsLine(s stage) string {
	var parts []string
	add := func(n int, word string) {
		if n > 0 {
			parts = append(parts, fmt.Sprintf("%s ×%d", word, n))
		}
	}
	add(s.Reads, "read")
	add(s.Searches, "searched")
	if s.Kind == stageImplement || s.Kind == stageFix {
		add(s.Edits, "tool calls")
	}
	add(s.Checks, "checks")
	add(s.Runs, "commands")
	add(s.Memory, "memory calls")
	return strings.Join(parts, " · ")
}

// stageSummaryLines is the collapsed body of a stage box.
func stageSummaryLines(s stage, color bool) []string {
	var lines []string
	switch s.Kind {
	case stageExplore, stageRemember:
		lines = append(lines, countsLine(s))
	case stageImplement, stageFix:
		for i, f := range s.Files {
			if i >= 3 {
				lines = append(lines, fmt.Sprintf("… and %d more", len(s.Files)-3))
				break
			}
			lines = append(lines, fileLabel(f))
		}
		if len(s.Files) == 0 {
			lines = append(lines, countsLine(s)+" (file names not recorded)")
		}
	default:
		var cmds []string
		for _, c := range s.Cmds {
			cmds = append(cmds, cmdLabel(c))
		}
		lines = append(lines, strings.Join(cmds, "  ·  "))
	}
	if s.Loop {
		lines[len(lines)-1] += "   " + paint(color, yellow, "⚠ loop")
	}
	return lines
}

// stageDetailLines is the expanded body: every file with its operation, every
// command with its result, and timestamps.
func stageDetailLines(s stage, color bool) []string {
	lines := []string{stampOf(s.Start) + " → " + stampOf(s.End) + " · " + formatDuration(s.Duration())}
	if c := countsLine(s); c != "" {
		lines = append(lines, c)
	}
	for _, f := range s.Files {
		lines = append(lines, fileLabel(f))
	}
	for _, c := range s.Cmds {
		lines = append(lines, "ran "+cmdLabel(c))
	}
	if s.Loop {
		lines = append(lines, paint(color, yellow, "⚠ loop span: the same thing failed repeatedly"))
	}
	if s.RetryOf > 0 {
		lines = append(lines, fmt.Sprintf("↺ retry %dx of stage %d", s.RetryNo, s.RetryOf))
	}
	return lines
}

type stageGraphOptions struct {
	Width    int
	Color    bool
	All      bool         // expand every stage and show no selection
	Selected int          // selected stage index, -1 for none
	Expanded map[int]bool // stage index -> expanded
	Filter   string       // show only stages touching a file containing this text
	NoDetail bool         // collapsed bodies only (detail lives in a side panel)
}

type stageGraph struct {
	Lines   []string
	StageAt []int    // stage index under each line, -1 for none
	Rows    [][2]int // first and last line of each stage box
}

func padTo(s string, w int) string {
	if d := w - ansi.StringWidth(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}

func fitTo(s string, w int) string {
	if w < 1 {
		return ""
	}
	return padTo(ansi.Truncate(s, w, "…"), w)
}

func stageMatches(s stage, filter string) bool {
	if filter == "" {
		return true
	}
	for _, f := range s.Files {
		if strings.Contains(f.Path, filter) || strings.Contains(f.To, filter) {
			return true
		}
	}
	return false
}

// boxLines draws one titled box of the given outer width.
func boxLines(title, right string, body []string, width int, selected, connect bool, colorCode string, color bool, rounded bool) []string {
	tl, tr, bl, br, h, v, down := "┌", "┐", "└", "┘", "─", "│", "┬"
	switch {
	case selected:
		tl, tr, bl, br, h, v, down = "╔", "╗", "╚", "╝", "═", "║", "╩"
	case rounded:
		tl, tr, bl, br = "╭", "╮", "╰", "╯"
	}
	style := colorCode
	if selected {
		style = bold + colorCode
	}
	line := func(s string) string { return paint(color, style, s) }
	left := tl + h + " " + title + " "
	rt := " " + right + " " + tr
	room := width - ansi.StringWidth(left) - ansi.StringWidth(rt)
	if room < 1 {
		left = ansi.Truncate(left, max(4, width-ansi.StringWidth(rt)-1), "… ")
		room = max(1, width-ansi.StringWidth(left)-ansi.StringWidth(rt))
	}
	var out []string
	out = append(out, line(left+strings.Repeat(h, room)+rt))
	for _, b := range body {
		out = append(out, line(v)+" "+fitTo(b, width-4)+" "+line(v))
	}
	bottom := []rune(bl + strings.Repeat(h, width-2) + br)
	if connect {
		bottom[width/2] = []rune(down)[0]
	}
	out = append(out, line(string(bottom)))
	return out
}

// renderStageGraph draws the stage flow chart. It is pure: the interactive
// player and the plain transcript both use it.
func renderStageGraph(doc replayDoc, o stageGraphOptions) stageGraph {
	const margin = 2
	g := stageGraph{}
	add := func(line string, stageIdx int) {
		g.Lines = append(g.Lines, line)
		g.StageAt = append(g.StageAt, stageIdx)
	}
	if len(doc.Stages) == 0 {
		add("No stages recorded: the journal has no tool calls yet.", -1)
		return g
	}
	hasRetry := false
	for _, s := range doc.Stages {
		if s.RetryOf > 0 {
			hasRetry = true
		}
	}
	gutter := 0
	if hasRetry {
		gutter = 20
	}
	boxW := min(o.Width-margin-gutter, 76)
	boxW = max(boxW, 30)
	rows := make([][2]int, len(doc.Stages))
	visible := 0
	var shown []int
	for i, s := range doc.Stages {
		if stageMatches(s, o.Filter) {
			shown = append(shown, i)
			visible++
		}
	}
	if o.Filter != "" {
		add(fmt.Sprintf("Filter %q: %d of %d stages", safeText(o.Filter), visible, len(doc.Stages)), -1)
		add("", -1)
	}
	for _, i := range shown {
		s := doc.Stages[i]
		selected := !o.All && i == o.Selected
		body := stageSummaryLines(s, o.Color)
		if (o.All || o.Expanded[i]) && !o.NoDetail {
			body = stageDetailLines(s, o.Color)
		}
		title := fmt.Sprintf("%d %s%s", s.N, s.Kind, stageGlyph(s))
		if s.Loop {
			title += " ⚠"
		}
		box := boxLines(title, stageTimes(s), body, boxW, selected, true, stageColor(s), o.Color, false)
		rows[i][0] = len(g.Lines)
		for _, l := range box {
			add(strings.Repeat(" ", margin)+l, i)
		}
		rows[i][1] = len(g.Lines) - 1
		if selected {
			g.Lines[rows[i][0]] = paint(o.Color, bold, "▶") + " " + g.Lines[rows[i][0]][margin:]
		}
		pad := strings.Repeat(" ", margin+boxW/2)
		add(pad+paint(o.Color, dim, "▼"), -1)
	}
	// Final summary node.
	sum := doc.Files
	var parts []string
	for _, p := range []struct {
		n    int
		word string
	}{{len(sum.Created), "created"}, {len(sum.Edited), "edited"}, {len(sum.Deleted), "deleted"}, {len(sum.Moved), "moved"}} {
		if p.n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", p.n, p.word))
		}
	}
	files := "Files: none recorded"
	if len(parts) > 0 {
		files = "Files: " + strings.Join(parts, " · ")
	}
	body := []string{files,
		fmt.Sprintf("Checks: %d passed / %d failed · Loops: %d", doc.Stats.ChecksPassed, doc.Stats.ChecksFailed, doc.Stats.Loops)}
	code := green
	if doc.Stats.ChecksFailed > 0 || doc.Stats.Loops > 0 {
		code = yellow
	}
	for _, l := range boxLines("Summary", formatDuration(doc.Stats.Duration), body, boxW, false, false, code, o.Color, true) {
		add(strings.Repeat(" ", margin)+l, -1)
	}
	// Retry back-edges share one lane in the right gutter: an edge leaves the
	// first failed Verify of a cycle and is labelled at each retry.
	if hasRetry {
		for i := range g.Lines {
			g.Lines[i] = padTo(g.Lines[i], margin+boxW)
		}
		retries := map[int][]stage{}
		for _, s := range doc.Stages {
			if s.RetryOf > 0 && stageMatches(s, o.Filter) && stageMatches(doc.Stages[s.RetryOf-1], o.Filter) {
				retries[s.RetryOf] = append(retries[s.RetryOf], s)
			}
		}
		edge := map[int]string{}
		for origin, list := range retries {
			from := rows[origin-1][0]
			last := rows[list[len(list)-1].N-1][0]
			edge[from] = "◄──┐"
			for r := from + 1; r < last; r++ {
				edge[r] = "   │"
			}
			for k, s := range list {
				corner := "┤"
				if k == len(list)-1 {
					corner = "┘"
				}
				edge[rows[s.N-1][0]] = fmt.Sprintf("───%s ↺ retry %dx", corner, s.RetryNo)
			}
		}
		for row, text := range edge {
			if row >= 0 && row < len(g.Lines) {
				g.Lines[row] += paint(o.Color, yellow, " "+text)
			}
		}
	}
	for i := range g.Lines {
		g.Lines[i] = strings.TrimRight(g.Lines[i], " ")
	}
	g.Rows = rows
	return g
}
