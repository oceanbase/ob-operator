package oceanbase

import (
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestOBClusterLogArguments(t *testing.T) {
	start, end := "2026-10-21 15:00:00", "2026-10-21 15:03:00"
	directory := "/shared/a ' quoted; $path"
	command := obClusterLogCommand(start, end, directory, directory+".tar.gz")
	if !reflect.DeepEqual(command[:2], []string{"python3", "-c"}) || command[2] != collectOBClusterLogScript ||
		!reflect.DeepEqual(command[3:], []string{start, end, directory, directory + ".tar.gz"}) {
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
