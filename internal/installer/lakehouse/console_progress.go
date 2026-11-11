package lakehouse

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/term"
)

// consoleProgress renders only controlled installer progress messages. It does
// not run checks or infer readiness from elapsed time. Unrecognised messages are
// kept above the live display, so warnings and failure diagnostics survive it.
type consoleProgress struct {
	mu          sync.Mutex
	out         io.Writer
	interactive bool
	finished    bool
	started     time.Time
	now         func() time.Time
	size        func() (int, int)
	stop        chan struct{}
	stopped     chan struct{}
	finishOnce  sync.Once
	rendered    int
	parents     []*consoleStage
	children    map[int][]*consoleStage
	parent      *consoleStage
	active      *consoleStage
	lastCheck   time.Time
	podChecked  time.Time
	detail      string
	warning     string
	limit       string
	logPath     string
	title       string
	writers     []*consoleLineWriter
}

type consoleStage struct {
	label   string
	status  string
	step    int
	started time.Time
	ended   time.Time
}

type consoleLineWriter struct {
	console *consoleProgress
	dst     io.Writer
	partial string // Protected by console.mu, including concurrent stdout/stderr.
}

var (
	consoleNumberedStep = regexp.MustCompile(`^\[([1-9][0-9]*)/([1-9][0-9]*)\] (.+)$`)
	consoleLimit        = regexp.MustCompile(`\(limit ([0-9]+s)\)`)
	consoleElapsed      = regexp.MustCompile(` \((?:[0-9]+s elapsed|phase elapsed=[0-9]+s; total elapsed=[0-9]+s)\)$`)
	consoleWaitTiming   = regexp.MustCompile(`(?:; |: )?(?:wait elapsed=[0-9]+s|[0-9]+s elapsed) \(limit [0-9]+s\)`)
	consoleWarnings     = regexp.MustCompile(`(?:^|[; ])recentWarnings=([^; ]+)`)
	consolePodAge       = regexp.MustCompile(`; Pod snapshot age=([0-9]+)s`)
)

func newConsoleProgress(out io.Writer, interactive bool) *consoleProgress {
	now := time.Now()
	c := &consoleProgress{
		out: out, interactive: interactive, started: now, now: time.Now,
		stop: make(chan struct{}), stopped: make(chan struct{}),
		children: map[int][]*consoleStage{}, title: "OceanBase Lakehouse installation",
	}
	c.size = func() (int, int) {
		if f, ok := out.(interface{ Fd() uintptr }); ok {
			if width, height, err := term.GetSize(int(f.Fd())); err == nil && width > 0 && height > 0 {
				return width, height
			}
		}
		return 80, 24
	}
	width, height := c.size()
	if width < 40 || height < 12 {
		c.interactive = false
	}
	if c.interactive {
		go c.refresh()
	} else {
		close(c.stopped)
	}
	return c
}

func (c *consoleProgress) refresh() {
	defer close(c.stopped)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			c.mu.Lock()
			if !c.finished && c.interactive && len(c.parents) > 0 {
				c.renderLocked(c.now())
			}
			c.mu.Unlock()
		}
	}
}

func (c *consoleProgress) Println(message string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.interactive {
		fmt.Fprintln(c.out, message)
		return
	}
	for _, line := range strings.Split(message, "\n") {
		c.lineLocked(line)
	}
}

func (c *consoleProgress) Live() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.interactive && !c.finished
}

// Interactive command streams share the display. Redirected output preserves
// the caller's original stream byte-for-byte and introduces no ANSI sequences.
func (c *consoleProgress) Writer(dst io.Writer) io.Writer {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.interactive {
		return dst
	}
	w := &consoleLineWriter{console: c, dst: dst}
	c.writers = append(c.writers, w)
	return w
}

func (w *consoleLineWriter) Write(data []byte) (int, error) {
	c := w.console
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.finished || !c.interactive {
		// Finish flushes the previous partial line. Late diagnostics must not
		// redraw or erase the final access summary.
		_, err := io.WriteString(c.out, consoleSafeText(string(data)))
		return len(data), err
	}
	w.partial += string(data)
	for {
		end := strings.IndexByte(w.partial, '\n')
		if end < 0 {
			// A broken subprocess should not grow an unlimited line buffer.
			if len(w.partial) > 64<<10 {
				c.lineLocked(w.partial)
				w.partial = ""
			}
			break
		}
		c.lineLocked(w.partial[:end])
		w.partial = w.partial[end+1:]
	}
	return len(data), nil
}

// Finish is synchronous and idempotent: after it returns no ticker or buffered
// command output can overwrite the final access instructions.
func (c *consoleProgress) Finish(err error) {
	c.finishOnce.Do(func() {
		close(c.stop)
		<-c.stopped
		c.mu.Lock()
		defer c.mu.Unlock()
		for _, w := range c.writers {
			if w.partial != "" {
				c.lineLocked(w.partial)
				w.partial = ""
			}
		}
		now := c.now()
		if err != nil {
			c.finishStageLocked(c.active, "failed", now)
			if c.parent != c.active {
				c.finishStageLocked(c.parent, "failed", now)
			}
		} else {
			c.finishStageLocked(c.active, "done", now)
			c.finishStageLocked(c.parent, "done", now)
		}
		c.finished = true
		if c.interactive {
			c.renderLocked(now)
			if err != nil {
				// The full error is printed outside the bounded frame. It is
				// never truncated, even on a narrow terminal.
				fmt.Fprintln(c.out, "Error: "+consoleSafeText(err.Error()))
			}
		}
	})
}

func (c *consoleProgress) finishStageLocked(stage *consoleStage, status string, now time.Time) {
	if stage != nil && stage.status == "running" {
		stage.status = status
		stage.ended = now
	}
}

func (c *consoleProgress) lineLocked(raw string) {
	line := consoleSafeText(strings.TrimSuffix(raw, "\r"))
	if c.finished || !c.interactive {
		fmt.Fprintln(c.out, line)
		return
	}
	now := c.now()
	if match := consoleNumberedStep.FindStringSubmatch(line); match != nil {
		step, _ := strconv.Atoi(match[1])
		total, _ := strconv.Atoi(match[2])
		// Bound the display to installer-supported stage counts. Never let a
		// child command create an arbitrarily large terminal frame.
		if (total == 2 || total == 5) && step <= total {
			if len(c.parents) == 0 {
				labels := []string{"Cluster checks", "Images", "Operator", "Lakehouse", "Access checks"}
				if total == 2 {
					c.title = "OceanBase Lakehouse preflight checks"
				}
				for i := 0; i < total; i++ {
					c.parents = append(c.parents, &consoleStage{step: i + 1, label: labels[i], status: "pending"})
				}
			}
			if step <= len(c.parents) {
				c.finishStageLocked(c.active, "done", now)
				c.finishStageLocked(c.parent, "done", now)
				c.parent = c.parents[step-1]
				c.parent.status, c.parent.started = "running", now
				c.parent.ended = time.Time{}
				c.active = c.parent
				c.detail = match[3]
				c.limit, c.warning, c.lastCheck, c.podChecked = "", "", time.Time{}, time.Time{}
				c.renderLocked(now)
				return
			}
		}
	}
	if strings.HasPrefix(line, "Detailed log: ") {
		c.logPath = strings.TrimSuffix(strings.TrimPrefix(line, "Detailed log: "), " (private, mode 0600)")
		c.persistentLocked(line)
		c.renderLocked(now)
		return
	}
	for _, prefix := range []string{"[stage] ", "[done] ", "[skip] "} {
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		message := consoleElapsed.ReplaceAllString(strings.TrimPrefix(line, prefix), "")
		if numbered := consoleNumberedStep.FindStringSubmatch(message); numbered != nil {
			message = numbered[3]
		}
		name, _, _ := strings.Cut(message, ":")
		switch name {
		case "Preflight", "Storage", "LogService", "OceanBase", "Observer Pods", "Dashboard":
			if c.parent != nil {
				var stage *consoleStage
				for _, candidate := range c.children[c.parent.step] {
					if candidate.label == name {
						stage = candidate
					}
				}
				if stage == nil {
					stage = &consoleStage{label: name, status: "pending"}
					c.children[c.parent.step] = append(c.children[c.parent.step], stage)
				}
				if prefix == "[stage] " {
					c.finishStageLocked(c.active, "done", now)
					// Entering a child does not complete its parent.
					c.parent.status, c.parent.ended = "running", time.Time{}
					stage.status, stage.started, stage.ended = "running", now, time.Time{}
					c.limit, c.warning, c.lastCheck, c.podChecked = "", "", time.Time{}, time.Time{}
				} else if prefix == "[skip] " {
					stage.status = "skipped"
					stage.ended = now
				} else {
					stage.status = "done"
					stage.ended = now
					c.lastCheck = now
				}
				c.active, c.detail = stage, message
				c.renderLocked(now)
				return
			}
		}
	}
	if strings.HasPrefix(line, "[wait] ") || strings.HasPrefix(line, "[ready] ") || strings.HasPrefix(line, "[done] ") || strings.HasPrefix(line, "[checked] ") || strings.HasPrefix(line, "[skip] ") {
		c.detail = strings.TrimSpace(line[strings.IndexByte(line, ']')+1:])
		if match := consoleLimit.FindStringSubmatch(line); match != nil {
			c.limit = match[1]
		}
		if strings.Contains(line, "resource status checked") || strings.HasPrefix(line, "[ready] ") || strings.HasPrefix(line, "[done] ") || strings.HasPrefix(line, "[checked] ") {
			c.lastCheck = now
		}
		if match := consolePodAge.FindStringSubmatch(line); match != nil {
			if age, err := strconv.ParseInt(match[1], 10, 32); err == nil {
				c.podChecked = now.Add(-time.Duration(age) * time.Second)
				// Every successful Pod snapshot reports its complete current
				// warning summary; absent recentWarnings must clear old data.
				if strings.Contains(line, "resource status checked") {
					c.warning = ""
				}
			}
		} else if strings.Contains(line, "Pod details unavailable") {
			c.podChecked = time.Time{}
			c.warning = "unknown (Pod details unavailable)"
		}
		if match := consoleWarnings.FindStringSubmatch(line); match != nil {
			c.warning = match[1]
		}
		c.detail = consoleWaitTiming.ReplaceAllString(c.detail, "")
		c.detail = strings.ReplaceAll(c.detail, "; resource status checked", "")
		c.detail = consolePodAge.ReplaceAllString(c.detail, "")
		c.detail = consoleElapsed.ReplaceAllString(c.detail, "")
		c.detail = strings.ReplaceAll(c.detail, "status=new;", "status=new (initializing);")
		if strings.HasPrefix(line, "[skip] ") {
			c.finishStageLocked(c.active, "skipped", now)
		}
		c.renderLocked(now)
		return
	}
	c.persistentLocked(line)
	c.renderLocked(now)
}

func (c *consoleProgress) persistentLocked(line string) {
	c.clearLocked()
	fmt.Fprintln(c.out, line)
}

func (c *consoleProgress) clearLocked() {
	for i := 0; i < c.rendered; i++ {
		io.WriteString(c.out, "\x1b[1A\r\x1b[2K")
	}
	c.rendered = 0
}

func (c *consoleProgress) renderLocked(now time.Time) {
	if len(c.parents) == 0 {
		return
	}
	width, height := c.size()
	lines := []string{fmt.Sprintf("%s | total %s", c.title, consoleDuration(now.Sub(c.started)))}
	for _, parent := range c.parents {
		lines = append(lines, c.stageLine(parent, fmt.Sprintf("%d/%d ", parent.step, len(c.parents)), now))
		for _, child := range c.children[parent.step] {
			lines = append(lines, "  "+c.stageLine(child, "", now))
		}
	}
	if c.active != nil {
		elapsed := time.Duration(0)
		if !c.active.started.IsZero() {
			end := now
			if !c.active.ended.IsZero() {
				end = c.active.ended
			}
			elapsed = end.Sub(c.active.started)
		}
		limit := c.limit
		if limit == "" {
			limit = "not reported"
		}
		lines = append(lines, fmt.Sprintf("Current: %s | elapsed %s | timeout %s", c.active.label, consoleDuration(elapsed), limit))
	}
	last := "not reported yet"
	if !c.lastCheck.IsZero() {
		last = consoleDuration(now.Sub(c.lastCheck)) + " ago"
	}
	checkLine := "Latest check: " + last
	if !c.podChecked.IsZero() {
		checkLine += " | Pod snapshot: " + consoleDuration(now.Sub(c.podChecked)) + " old"
	}
	lines = append(lines, checkLine)
	if c.detail != "" {
		lines = append(lines, consoleDetailLines(c.detail, width-1)...)
	}
	if c.warning != "" {
		lines = append(lines, "Warnings: "+c.warning)
	}
	if c.logPath != "" {
		lines = append(lines, "Log: "+c.logPath)
	}
	if width < 40 || height < len(lines)+1 {
		// Do not move the cursor beyond a small screen's visible history.
		// The current snapshot is retained and subsequent events stay plain.
		c.clearLocked()
		for _, line := range lines {
			fmt.Fprintln(c.out, line)
		}
		c.interactive = false
		return
	}
	c.clearLocked()
	for _, line := range lines {
		fmt.Fprintln(c.out, consoleFrameLine(line, width-1))
	}
	c.rendered = len(lines)
}

func (c *consoleProgress) stageLine(stage *consoleStage, prefix string, now time.Time) string {
	duration := ""
	if !stage.started.IsZero() {
		end := now
		if !stage.ended.IsZero() {
			end = stage.ended
		}
		duration = " (" + consoleDuration(end.Sub(stage.started)) + ")"
	}
	return fmt.Sprintf("[%-7s] %s%s%s", stage.status, prefix, stage.label, duration)
}

func consoleDuration(d time.Duration) string {
	seconds := int64(d / time.Second)
	if seconds < 0 {
		seconds = 0
	}
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm %02ds", seconds/60, seconds%60)
	}
	return fmt.Sprintf("%dh %02dm %02ds", seconds/3600, seconds/60%60, seconds%60)
}

// Terminal controls, including escape and bidi controls, cannot enter the live
// display through resource names, paths, subprocess diagnostics or errors.
func consoleSafeText(text string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' {
			return r
		}
		if r == '\t' {
			return ' '
		}
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, text)
}

func consoleFrameLine(line string, width int) string {
	// The compact table is ASCII so its byte width is its terminal width. Full
	// diagnostics and paths are preserved separately as persistent text.
	line = strings.Map(func(r rune) rune {
		if r < 32 || r > 126 {
			return '?'
		}
		return r
	}, consoleSafeText(line))
	if len(line) > width && width >= 3 {
		return line[:width-3] + "..."
	}
	return line
}

func consoleDetailLines(detail string, width int) []string {
	// Pod and CR status summaries commonly need more than one line. Preserve
	// three compact lines before truncating, without letting long user fields
	// wrap the cursor-tracked table unpredictably.
	if width < 16 {
		return []string{"Detail: " + detail}
	}
	detail = consoleFrameLine(detail, len(detail)+1)
	var lines []string
	for i := 0; i < 3 && detail != ""; i++ {
		prefix := "        "
		if i == 0 {
			prefix = "Detail: "
		}
		available := width - len(prefix)
		if len(detail) <= available || i == 2 {
			lines = append(lines, consoleFrameLine(prefix+detail, width))
			break
		}
		cut := strings.LastIndexByte(detail[:available], ' ')
		if cut < available/2 {
			cut = available
		}
		lines = append(lines, prefix+detail[:cut])
		detail = strings.TrimSpace(detail[cut:])
	}
	return lines
}
