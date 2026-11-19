package oceanbase

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestOBClusterLogArguments(t *testing.T) {
	start, end := "2026-10-21 15:00:00", "2026-10-21 15:03:00"
	directory := "/shared/a ' quoted; $path"
	command := obClusterLogCommand(start, end, directory, directory+".tar.gz")
	if !reflect.DeepEqual(command[:2], []string{"/bin/sh", "-ec"}) || command[4] != "-c" || command[5] != collectOBClusterLogScript ||
		!reflect.DeepEqual(command[6:], []string{start, end, directory, directory + ".tar.gz"}) || strings.Contains(command[2], directory) {
		t.Fatal("collector arguments must be passed without shell interpolation")
	}
	init := obClusterLogInitCommand(directory, "test", "test-ob")
	if !reflect.DeepEqual(init[3:], []string{"generate-config", directory, "test", "test-ob"}) ||
		strings.Contains(init[2], directory) || !strings.Contains(init[2], "trap") {
		t.Fatal("initializer must quote positional parameters and clean credentials on failure")
	}
}

func TestOBClusterLogCollector(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Fatal("python3 is required to validate the embedded collector")
	}
	command := exec.Command(python, "-B", "obcluster_log_collect_test.py")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("collector regression tests: %v\n%s", err, output)
	}
}

func TestOBClusterLogPythonArguments(t *testing.T) {
	directory := t.TempDir()
	// A Python entry point records argv without executing the collector. Shell
	// metacharacters in all values must reach Python literally.
	python := filepath.Join(directory, "python3")
	if err := os.WriteFile(python, []byte("#!/bin/sh\nprintf '%s\\n' \"$@\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	values := []string{"start ' ; $literal", "end $(literal)", "/shared/a ' quoted; $path", "/shared/a.tar.gz"}
	args := obClusterLogCommand(values[0], values[1], values[2], values[3])
	command := exec.Command(args[0], args[1:]...)
	command.Env = []string{"PATH=" + directory}
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Python entry point: %v: %s", err, output)
	}
	expected := "-c\n" + collectOBClusterLogScript + "\n" + strings.Join(values, "\n") + "\n"
	if string(output) != expected {
		t.Fatal("Python script or argument boundaries changed")
	}
}
