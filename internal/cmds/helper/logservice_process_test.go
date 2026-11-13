package helper

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// An image without procps must still recognize an already running daemon.
func TestLogServiceDetectionWithoutPgrep(t *testing.T) {
	if _, err := os.Stat("/proc/self/stat"); err != nil {
		t.Skip("requires Linux procfs")
	}
	binary, err := os.ReadFile("/bin/sleep")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "oblogservice")
	if err := os.WriteFile(path, binary, 0755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(path, "30")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	t.Setenv("PATH", dir)
	pid, err := waitForDaemon(time.Now().Add(150*time.Millisecond), 150*time.Millisecond)
	if err != nil || pid != cmd.Process.Pid {
		t.Fatalf("running daemon must be detected without pgrep: pid=%d want=%d err=%v", pid, cmd.Process.Pid, err)
	}
}

func TestLogServiceProcSelection(t *testing.T) {
	root := t.TempDir()
	for pid, stat := range map[string]string{"15": "15 (oblogservice) S 1", "3": "3 (oblogservice) Z 1", "8": "8 (oblogservice) S 1", "7": "7 (oblogservice-x) S 1", "2": "2 (name with ) bracket) S 1"} {
		if err := os.Mkdir(filepath.Join(root, pid), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, pid, "stat"), []byte(stat), 0644); err != nil {
			t.Fatal(err)
		}
	}
	pid, err := findLogServiceDaemon(root)
	if err != nil || pid != 8 {
		t.Fatalf("pid=%d err=%v", pid, err)
	}
	running, err := logServiceProcessRunning(root, 3)
	if err != nil || running {
		t.Fatal("zombie must not count as running", err)
	}
	running, err = logServiceProcessRunning(root, 42)
	if err != nil || running {
		t.Fatal("exited PID must not count as running", err)
	}
}

func TestLogServiceProcInspectionErrors(t *testing.T) {
	if _, err := findLogServiceDaemon(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("procfs enumeration failure must be reported")
	}
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "123"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "123", "stat"), []byte("invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := findLogServiceDaemon(root); err == nil {
		t.Fatal("invalid process information must not become no-process")
	}
}
