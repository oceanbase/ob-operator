package lakehouse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Opt-in only: run against a disposable daemon with its own root/state/socket.
// The launcher must create the marker beside that socket, never beside the live
// Kubernetes runtime socket. This test imports images but creates no containers.
func TestImagesIsolatedContainerdColdImport(t *testing.T) {
	bundle, socket := os.Getenv("LAKEHOUSE_COLD_IMPORT_BUNDLE"), os.Getenv("LAKEHOUSE_COLD_IMPORT_SOCKET")
	if bundle == "" && socket == "" {
		t.Skip("requires an explicitly isolated containerd daemon and extracted bundle")
	}
	if !filepath.IsAbs(bundle) || !filepath.IsAbs(socket) {
		t.Fatal("bundle and isolated socket must both be explicit absolute paths")
	}
	marker := filepath.Join(filepath.Dir(socket), "ISOLATED-COLD-IMPORT")
	if data, err := os.ReadFile(marker); err != nil || string(data) != "disposable-containerd-no-workloads\n" {
		t.Fatal("isolated runtime marker is missing; refusing to access a potentially shared runtime")
	}
	lock, manifest, err := readImageDelivery(bundle)
	if err != nil {
		t.Fatal(err)
	}
	for _, archive := range manifest.Archives {
		if archive.Base != "bundle" {
			t.Fatalf("self-contained package still depends on delivery-side archive %s", archive.Name)
		}
	}
	if expected := os.Getenv("LAKEHOUSE_COLD_IMPORT_EXPECT_IMAGES"); expected != "" {
		count, err := strconv.Atoi(expected)
		if err != nil || count != len(lock.Images) {
			t.Fatalf("expected %s locked images, got %d", expected, len(lock.Images))
		}
	}
	log, err := os.OpenFile(filepath.Join(filepath.Dir(socket), "commands.log"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	imports := 0
	run := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("runtime command has no deadline")
		}
		if name != "ctr" && name != "crictl" {
			t.Fatalf("unexpected command %s", name)
		}
		// Every call must name the dedicated socket, including the direct checks.
		explicitSocket := false
		for _, arg := range args {
			if arg == socket || arg == "unix://"+socket {
				explicitSocket = true
			}
		}
		if !explicitSocket {
			t.Fatalf("runtime command omitted isolated socket: %s %v", name, args)
		}
		if strings.Contains(strings.Join(args, " "), "images import ") {
			imports++
		}
		// Use the same process-group cancellation and stdout-only parsing as the
		// installer; containerd deprecation warnings belong in the evidence log,
		// not in a JSON document or the check for an empty image list.
		cmd := command(ctx, name, args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		output, err := cmd.Output()
		if _, writeErr := fmt.Fprintf(log, "$ %s %q\nstdout:\n%s\nstderr:\n%s\nresult: %v\n", name, args, output, stderr.Bytes(), err); writeErr != nil {
			t.Fatal(writeErr)
		}
		return output, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Minute)
	defer cancel()
	for _, command := range [][]string{
		{"--address", socket, "--namespace", "k8s.io", "images", "list", "--quiet"},
		{"--address", socket, "--namespace", "k8s.io", "content", "list", "--quiet"},
	} {
		output, err := runImageCommand(ctx, run, 30*time.Second, "ctr", command...)
		if err != nil || strings.TrimSpace(string(output)) != "" {
			t.Fatalf("isolated runtime must start with zero images and content: %s / %v", output, err)
		}
	}
	output, err := runImageCommand(ctx, run, 30*time.Second, "crictl", "--runtime-endpoint", "unix://"+socket,
		"--image-endpoint", "unix://"+socket, "images", "--output", "json")
	var initial struct {
		Images []json.RawMessage `json:"images"`
	}
	if err != nil || json.Unmarshal(output, &initial) != nil || len(initial.Images) != 0 {
		t.Fatalf("CRI must start with zero cached images: %s / %v", output, err)
	}
	t.Logf("cold runtime verified: 0 images, 0 content; package has %d images in %d archives", len(lock.Images), len(manifest.Archives))
	opts := ImageOptions{BundleDir: bundle, Runtime: "containerd", ContainerdAddress: socket,
		CRIEndpoint: "unix://" + socket, ImportTimeout: 15 * time.Minute}
	if err := PrepareImages(ctx, run, opts, func(message string) { t.Log(message) }); err != nil {
		t.Fatal(err)
	}
	if imports != len(manifest.Archives) {
		t.Fatalf("expected all %d cold imports, got %d", len(manifest.Archives), imports)
	}
	for _, image := range lock.Images {
		t.Logf("verified %s %s", image.Reference, image.ImageID)
	}
	if err := PrepareImages(ctx, run, opts, func(message string) { t.Log(message) }); err != nil {
		t.Fatal(err)
	}
	if imports != len(manifest.Archives) {
		t.Fatalf("second preparation was not idempotent: %d imports", imports)
	}
	t.Log("PASS: every archive imported, every config ID verified, repeat preparation imported nothing")
}
