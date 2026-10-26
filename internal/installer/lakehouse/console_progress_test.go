package lakehouse

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newTestConsole(t *testing.T, interactive bool) (*consoleProgress, *bytes.Buffer, *atomic.Int64) {
	t.Helper()
	var out bytes.Buffer
	c := newConsoleProgress(&out, interactive)
	base := time.Unix(1_000, 0)
	seconds := &atomic.Int64{}
	c.mu.Lock()
	c.started = base
	c.now = func() time.Time { return base.Add(time.Duration(seconds.Load()) * time.Second) }
	c.mu.Unlock()
	t.Cleanup(func() { c.Finish(nil) })
	return c, &out, seconds
}

func finalConsoleFrame(t *testing.T, c *consoleProgress, out *bytes.Buffer) string {
	t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()
	text := out.String()
	index := strings.LastIndex(text, "OceanBase Lakehouse")
	if index < 0 {
		t.Fatal("no progress frame was rendered", text)
	}
	return text[index:]
}

func TestConsolePlainPreservesStreamsAndHasNoANSI(t *testing.T) {
	c, out, _ := newTestConsole(t, false)
	var stderr bytes.Buffer
	c.Println("[1/5] Checking cluster and installation settings")
	w := c.Writer(&stderr)
	for _, text := range []string{"Warning: ", "keep this\n", "last partial line"} {
		if _, err := io.WriteString(w, text); err != nil {
			t.Fatal(err)
		}
	}
	c.Finish(errors.New("reported by caller"))
	c.Finish(nil)
	if got := out.String(); got != "[1/5] Checking cluster and installation settings\n" {
		t.Fatal("plain progress changed", got)
	}
	if got := stderr.String(); got != "Warning: keep this\nlast partial line" {
		t.Fatal("plain subprocess output changed", got)
	}
	if strings.Contains(out.String()+stderr.String(), "\x1b") {
		t.Fatal("plain output emitted terminal controls")
	}
}

func TestConsoleStagesUseObservedTransitionsAndKeepElapsedTimes(t *testing.T) {
	c, out, seconds := newTestConsole(t, true)
	c.Println("[1/5] Checking cluster and installation settings")
	seconds.Store(3)
	c.Println("[2/5] Checking images")
	seconds.Store(5)
	c.Println("[3/5] Checking Operator prerequisites")
	c.Println("[skip] Compatible Operator and admission webhooks are already ready.")
	c.Println("[4/5] Installing lakehouse resources and Dashboard")
	c.Println("[stage] Storage: apply resources and wait for initialization (0s elapsed)")
	seconds.Store(9)
	c.Println("[done] Storage: initialization Job and Helm resources ready (4s elapsed)")
	c.Println("[stage] LogService: apply configuration and wait for running status (4s elapsed)")
	seconds.Store(17)
	c.Println("[wait] oblogservicecluster/demo-ls: status=new; wait elapsed=8s (limit 1800s); resource status checked")
	seconds.Store(20)
	c.mu.Lock()
	c.renderLocked(c.now())
	c.mu.Unlock()
	frame := finalConsoleFrame(t, c, out)
	for _, want := range []string{
		"total 20s", "[done   ] 1/5 Cluster checks (3s)", "[done   ] 2/5 Images (2s)",
		"[skipped] 3/5 Operator (0s)", "[running] 4/5 Lakehouse (15s)",
		"[done   ] Storage (4s)", "[running] LogService (11s)",
		"[pending] 5/5 Access checks", "Current: LogService | elapsed 11s | timeout 1800s", "Latest check: 3s ago",
	} {
		if !strings.Contains(frame, want) {
			t.Fatalf("missing %q in final frame:\n%s", want, frame)
		}
	}
	if strings.Contains(frame, "%") || strings.Contains(frame, "ETA") {
		t.Fatal("progress must not invent a percentage or ETA", frame)
	}
}

func TestConsoleHeartbeatDoesNotFabricateFreshChecks(t *testing.T) {
	c, out, seconds := newTestConsole(t, true)
	c.Println("[4/5] Installing lakehouse")
	c.Println("[stage] OceanBase: verify database status (0s elapsed)")
	seconds.Store(5)
	c.Println("[wait] obcluster/demo-ob: status=new; resource status checked (limit 1800s)")
	seconds.Store(35)
	c.Println("[wait] Helm cluster stage: wait elapsed=30s (limit 1800s)")
	frame := finalConsoleFrame(t, c, out)
	if !strings.Contains(frame, "Latest check: 30s ago") {
		t.Fatal("a quiet-command heartbeat must not reset the actual check time", frame)
	}
	seconds.Store(60)
	c.mu.Lock()
	c.renderLocked(c.now())
	c.mu.Unlock()
	if frame = finalConsoleFrame(t, c, out); !strings.Contains(frame, "Latest check: 55s ago") {
		t.Fatal("a repaint must not claim a new resource check", frame)
	}
}

func TestConsoleStructuredWaitKeepsPodAgeAndMovesTimingToFooter(t *testing.T) {
	c, out, seconds := newTestConsole(t, true)
	c.Println("[4/5] Installing lakehouse")
	c.Println("[stage] OceanBase: verify database status (phase elapsed=0s; total elapsed=10s)")
	seconds.Store(35)
	c.Println("[wait] obcluster/demo: status=new; podsReady=0/1; scheduled=1/1; waiting=ContainerCreating; recentWarnings=FailedMount; wait elapsed=30s (limit 1800s); resource status checked; Pod snapshot age=7s")
	frame := finalConsoleFrame(t, c, out)
	for _, want := range []string{"timeout 1800s", "Latest check: 0s ago", "status=new (initializing)", "Pod snapshot: 7s old", "Warnings: FailedMount"} {
		if !strings.Contains(frame, want) {
			t.Fatalf("missing structured wait field %q:\n%s", want, frame)
		}
	}
	for _, redundant := range []string{"wait elapsed=", "phase elapsed=", "total elapsed=", "resource status checked"} {
		if strings.Contains(frame, redundant) {
			t.Fatal("detail repeats timing already shown by the live footer", frame)
		}
	}
	c.Println("[wait] obcluster/demo: status=pending; podsReady=1/1; scheduled=1/1; waiting=ClusterInitialization; wait elapsed=35s (limit 1800s); resource status checked; Pod snapshot age=0s")
	if strings.Contains(finalConsoleFrame(t, c, out), "Warnings: FailedMount") {
		t.Fatal("cleared structured warnings must not remain as current warnings")
	}
	seconds.Store(40)
	c.mu.Lock()
	c.renderLocked(c.now())
	c.mu.Unlock()
	if frame = finalConsoleFrame(t, c, out); !strings.Contains(frame, "Pod snapshot: 5s old") {
		t.Fatal("cached Pod snapshot age must advance as the terminal repaints", frame)
	}
	c.Println("[wait] obcluster/demo: status=pending; Pod details unavailable; resource status checked")
	if frame = finalConsoleFrame(t, c, out); strings.Contains(frame, "Pod snapshot:") || strings.Contains(frame, "podsReady=") || !strings.Contains(frame, "Warnings: unknown (Pod details unavailable)") {
		t.Fatal("failed Pod snapshot reads must not leave old readiness counts as current", frame)
	}
	c.Println("[wait] obcluster/demo: status=pending; podsReady=1/1; scheduled=1/1; waiting=ClusterInitialization; wait elapsed=40s (limit 1800s); resource status checked; Pod snapshot age=0s")
	if frame = finalConsoleFrame(t, c, out); strings.Contains(frame, "Warnings:") || !strings.Contains(frame, "Pod snapshot: 0s old") {
		t.Fatal("recovered snapshot must replace unknown warnings and snapshot age", frame)
	}
}

func TestConsoleFailureRetainsDiagnosticsAndFlushesPartialLines(t *testing.T) {
	c, out, seconds := newTestConsole(t, true)
	w := c.Writer(io.Discard)
	c.Println("[4/5] Installing lakehouse")
	c.Println("Detailed log: /tmp/private-install/install.log (private, mode 0600)")
	_, _ = io.WriteString(w, "[stage] OceanBase: verify database status (0s elapsed)\n")
	seconds.Store(4)
	_, _ = io.WriteString(w, "[wait] obcluster/demo: status=pending; resource status checked (limit 1800s)\n")
	_, _ = io.WriteString(w, "Warning: FailedMount\nlast diagnostic without newline")
	seconds.Store(12)
	c.Finish(errors.New("database timed out; inspect the PVC"))
	c.Finish(nil)
	frame := finalConsoleFrame(t, c, out)
	for _, want := range []string{"[failed ] 4/5 Lakehouse", "[failed ] OceanBase", "[pending] 5/5 Access checks", "timeout 1800s", "Latest check: 8s ago", "/tmp/private-install/install.log", "database timed out; inspect the PVC"} {
		if !strings.Contains(frame, want) {
			t.Fatalf("failure omitted %q:\n%s", want, frame)
		}
	}
	for _, want := range []string{"Warning: FailedMount", "last diagnostic without newline", "Detailed log: /tmp/private-install/install.log (private, mode 0600)"} {
		if strings.Count(out.String(), want) != 1 {
			t.Fatal("diagnostics must be kept once, outside the replaceable frame", want)
		}
	}
	if !strings.Contains(out.String(), "\x1b[1A\r\x1b[2K") {
		t.Fatal("interactive output never refreshed in place")
	}
	// Later diagnostics and the caller's access summary remain below the final
	// frame, without another cursor movement.
	before := out.Len()
	c.Println("Final caller summary")
	_, _ = io.WriteString(w, "late diagnostic\n")
	if tail := out.String()[before:]; strings.Contains(tail, "\x1b") || !strings.Contains(tail, "Final caller summary\nlate diagnostic\n") {
		t.Fatal("late messages must not redraw the finished frame", tail)
	}
}

func TestConsoleCheckModeDoesNotClaimInstallationOrCompleteUnstartedStages(t *testing.T) {
	c, out, _ := newTestConsole(t, true)
	c.Println("[1/2] Checking cluster and installation settings")
	c.Finish(nil)
	frame := finalConsoleFrame(t, c, out)
	if !strings.Contains(frame, "OceanBase Lakehouse preflight checks") || !strings.Contains(frame, "[pending] 2/2 Images") || strings.Contains(frame, "Installation complete") || strings.Contains(frame, "5/5") {
		t.Fatal("preflight display overstates completed work", frame)
	}
}

func TestConsoleSanitizesControlsAndBoundsFrameWidth(t *testing.T) {
	c, out, _ := newTestConsole(t, true)
	c.Println("[1/5] Checking cluster")
	c.Println("[wait] " + strings.Repeat("long-field", 60) + "\x1b[2J\x07\r\u202eevil")
	c.Println("Warning: malicious\x1b[2J\x07\r\u202efield")
	c.Finish(errors.New("error\x1b]0;title\x07\r\u202efield"))
	frame := finalConsoleFrame(t, c, out)
	for _, line := range strings.Split(frame, "\n") {
		if len(line) > 79 {
			t.Fatal("table line can wrap and corrupt cursor tracking", len(line), line)
		}
	}
	for _, unsafe := range []string{"\x1b[2J", "\x1b]0;", "\x07", "\u202e"} {
		if strings.Contains(out.String(), unsafe) {
			t.Fatal("untrusted terminal control was emitted", unsafe)
		}
	}
}

func TestConsoleSmallTerminalFallsBackWithoutLosingMessages(t *testing.T) {
	c, out, _ := newTestConsole(t, true)
	c.mu.Lock()
	c.size = func() (int, int) { return 32, 8 }
	c.mu.Unlock()
	c.Println("[1/5] Checking cluster")
	c.Println("Warning: preserve this full diagnostic on a small terminal")
	if c.Live() {
		t.Fatal("small terminal must report that live display is disabled")
	}
	c.Finish(nil)
	if strings.Contains(out.String(), "\x1b") || !strings.Contains(out.String(), "Warning: preserve this full diagnostic on a small terminal") {
		t.Fatal("small-terminal fallback lost diagnostics or emitted cursor controls", out.String())
	}
}

func TestConsoleConcurrentStreamsAndFinishAreRaceFree(t *testing.T) {
	c, out, _ := newTestConsole(t, true)
	c.Println("[4/5] Installing lakehouse")
	writers := []io.Writer{c.Writer(io.Discard), c.Writer(io.Discard)}
	var wg sync.WaitGroup
	for i, w := range writers {
		wg.Add(1)
		go func(i int, w io.Writer) {
			defer wg.Done()
			for n := 0; n < 15; n++ {
				fmt.Fprintf(w, "Warning: stream-%d-event-%d\n", i, n)
			}
		}(i, w)
	}
	wg.Wait()
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); c.Finish(nil) }()
	}
	wg.Wait()
	for i := range writers {
		for n := 0; n < 15; n++ {
			if count := strings.Count(out.String(), fmt.Sprintf("Warning: stream-%d-event-%d\n", i, n)); count != 1 {
				t.Fatalf("concurrent diagnostic lost or repeated: stream=%d event=%d count=%d", i, n, count)
			}
		}
	}
}

func TestConsoleTickerRefreshesThenStopsBeforeSummary(t *testing.T) {
	c, out, seconds := newTestConsole(t, true)
	c.Println("[1/5] Checking cluster")
	seconds.Store(3)
	deadline := time.Now().Add(2500 * time.Millisecond)
	refreshed := false
	for time.Now().Before(deadline) {
		c.mu.Lock()
		refreshed = strings.Contains(out.String(), "total 3s")
		c.mu.Unlock()
		if refreshed {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !refreshed {
		t.Fatal("quiet operation did not repaint elapsed time within one ticker interval")
	}
	c.Finish(nil)
	select {
	case <-c.stopped:
	default:
		t.Fatal("Finish returned before the refresh goroutine stopped")
	}
	before := out.Len()
	seconds.Store(30)
	time.Sleep(50 * time.Millisecond)
	if out.Len() != before {
		t.Fatal("ticker overwrote the completed display")
	}
}
