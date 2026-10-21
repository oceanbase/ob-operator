package lakehouse

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type imageTestBundle struct {
	dir      string
	lock     imageLock
	manifest archiveManifest
	paths    []string
}

func makeImageTestBundle(t *testing.T) imageTestBundle {
	t.Helper()
	delivery := t.TempDir()
	b := imageTestBundle{dir: filepath.Join(delivery, "bundle")}
	if err := os.MkdirAll(filepath.Join(b.dir, "images"), 0700); err != nil {
		t.Fatal(err)
	}
	b.lock.Images = []lockedImage{
		{Role: "observer", Reference: "qa/observer:1", ImageID: "sha256:" + strings.Repeat("a", 64)},
		{Role: "dashboard", Reference: "qa/dashboard:1", ImageID: "sha256:" + strings.Repeat("b", 64)},
		{Role: "minio", Reference: "minio/minio:1", ImageID: "sha256:" + strings.Repeat("c", 64)},
	}
	b.manifest.SchemaVersion = 1
	for _, archive := range []imageArchive{
		{Name: "core", Base: "bundle", Path: "images/core.tar.gz", References: []string{"qa/observer:1", "qa/dashboard:1"}},
		{Name: "minio", Base: "delivery", Path: "minio-cache.tar.gz", References: []string{"minio/minio:1"}},
	} {
		payload := []byte("test archive payload " + archive.Name)
		hash := sha256.Sum256(payload)
		archive.SHA256 = hex.EncodeToString(hash[:])
		path, err := resolveImageArchive(b.dir, archive)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, payload, 0600); err != nil {
			t.Fatal(err)
		}
		b.paths = append(b.paths, path)
		b.manifest.Archives = append(b.manifest.Archives, archive)
	}
	b.write(t)
	return b
}

func (b imageTestBundle) write(t *testing.T) {
	t.Helper()
	for path, content := range map[string]interface{}{
		"image-lock.json": b.lock, "images/archives.json": b.manifest,
	} {
		data, err := json.Marshal(content)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(b.dir, path), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

type imageTestRuntime struct {
	t            *testing.T
	bundle       imageTestBundle
	ids          map[string]string
	commands     []string
	imports      []string
	ignoreImport bool
}

func (fake *imageTestRuntime) run(ctx context.Context, name string, args ...string) ([]byte, error) {
	fake.t.Helper()
	if _, ok := ctx.Deadline(); !ok {
		fake.t.Fatal("runtime command has no deadline")
	}
	command := name + " " + strings.Join(args, " ")
	fake.commands = append(fake.commands, command)
	if strings.Contains(command, " info ") || strings.HasSuffix(command, " version") {
		return []byte("runtime is available"), nil
	}
	if strings.Contains(command, "inspecti ") || strings.Contains(command, "image inspect ") {
		id := fake.ids[args[len(args)-1]]
		if id == "" {
			return nil, errors.New("image not found")
		}
		if name == "docker" {
			return []byte(id + "\n"), nil
		}
		return []byte(`{"status":{"id":"` + id + `"}}`), nil
	}
	if strings.Contains(command, " load --input ") || strings.Contains(command, " images import ") {
		path := args[len(args)-1]
		fake.imports = append(fake.imports, path)
		if !fake.ignoreImport {
			for i, archive := range fake.bundle.manifest.Archives {
				if fake.bundle.paths[i] != path {
					continue
				}
				for _, ref := range archive.References {
					for _, image := range fake.bundle.lock.Images {
						if image.Reference == ref {
							fake.ids[ref] = image.ImageID
						}
					}
				}
			}
		}
		return nil, nil
	}
	fake.t.Fatalf("unexpected runtime command: %s", command)
	return nil, errors.New("unreachable")
}

func TestImagesAutoUsesSelectedNodeK3sRuntimeAndVerifiesIDs(t *testing.T) {
	b := makeImageTestBundle(t)
	fake := imageTestRuntime{t: t, bundle: b, ids: map[string]string{}}
	var progress []string
	err := PrepareImages(context.Background(), fake.run, ImageOptions{
		BundleDir: b.dir, Runtime: "auto", NodeRuntimeVersion: "containerd://2.2.0-k3s1",
	}, func(s string) { progress = append(progress, s) })
	if err != nil {
		t.Fatal(err)
	}
	if len(fake.imports) != 2 || fake.imports[1] != filepath.Join(filepath.Dir(b.dir), "minio-cache.tar.gz") {
		t.Fatalf("did not locate delivery-side MinIO archive: %v", fake.imports)
	}
	for _, command := range fake.commands {
		if !strings.HasPrefix(command, "k3s ") || strings.Contains(command, "--local") {
			t.Fatalf("wrong runtime or gzip-incompatible import: %s", command)
		}
	}
	if len(progress) < 5 || !strings.Contains(progress[len(progress)-1], "全部匹配") {
		t.Fatalf("missing useful preparation progress: %v", progress)
	}
}

func TestImagesMatchingCacheSkipsMissingArchivesAndAllImports(t *testing.T) {
	b := makeImageTestBundle(t)
	fake := imageTestRuntime{t: t, bundle: b, ids: map[string]string{}}
	for _, image := range b.lock.Images {
		fake.ids[image.Reference] = image.ImageID
	}
	for _, path := range b.paths {
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}
	if err := PrepareImages(context.Background(), fake.run, ImageOptions{BundleDir: b.dir, Runtime: "docker", NodeRuntimeVersion: "docker://26.1.4"}, nil); err != nil {
		t.Fatal(err)
	}
	if len(fake.imports) != 0 {
		t.Fatalf("cached images should not import: %v", fake.imports)
	}
	for _, command := range fake.commands {
		if !strings.HasPrefix(command, "docker --host unix:///var/run/docker.sock ") {
			t.Fatalf("Docker command may inherit a remote context: %s", command)
		}
	}
}

func TestImagesAggregateArchiveErrorsBeforeAnyImport(t *testing.T) {
	b := makeImageTestBundle(t)
	if err := os.Remove(b.paths[0]); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b.paths[1], []byte("corrupt archive"), 0600); err != nil {
		t.Fatal(err)
	}
	fake := imageTestRuntime{t: t, bundle: b, ids: map[string]string{}}
	err := PrepareImages(context.Background(), fake.run, ImageOptions{BundleDir: b.dir, Runtime: "docker"}, nil)
	if err == nil || !strings.Contains(err.Error(), "required archive") || !strings.Contains(err.Error(), "SHA256 mismatch") || !strings.Contains(err.Error(), "nothing was imported") {
		t.Fatalf("archive failures were not aggregated: %v", err)
	}
	if len(fake.imports) != 0 {
		t.Fatalf("import started before all validation passed: %v", fake.imports)
	}
}

func TestImagesCheckOnlyNeverImports(t *testing.T) {
	b := makeImageTestBundle(t)
	fake := imageTestRuntime{t: t, bundle: b, ids: map[string]string{}}
	err := PrepareImages(context.Background(), fake.run, ImageOptions{BundleDir: b.dir, Runtime: "k3s", CheckOnly: true}, nil)
	if err == nil || !strings.Contains(err.Error(), "images are not ready") || len(fake.imports) != 0 {
		t.Fatalf("check-only must report missing images without importing: %v / %v", err, fake.imports)
	}
}

func TestImagesImportsOnlyArchiveContainingMissingIDs(t *testing.T) {
	b := makeImageTestBundle(t)
	fake := imageTestRuntime{t: t, bundle: b, ids: map[string]string{}}
	for _, image := range b.lock.Images[:2] {
		fake.ids[image.Reference] = image.ImageID
	}
	if err := os.Remove(b.paths[0]); err != nil {
		t.Fatal(err)
	}
	if err := PrepareImages(context.Background(), fake.run, ImageOptions{BundleDir: b.dir, Runtime: "docker"}, nil); err != nil {
		t.Fatal(err)
	}
	if len(fake.imports) != 1 || fake.imports[0] != b.paths[1] {
		t.Fatalf("unnecessary import: %v", fake.imports)
	}
}

func TestImagesRejectSuccessWhenImportedTagHasWrongConfigID(t *testing.T) {
	b := makeImageTestBundle(t)
	fake := imageTestRuntime{t: t, bundle: b, ids: map[string]string{}, ignoreImport: true}
	for _, image := range b.lock.Images {
		fake.ids[image.Reference] = "sha256:" + strings.Repeat("d", 64)
	}
	err := PrepareImages(context.Background(), fake.run, ImageOptions{BundleDir: b.dir, Runtime: "docker"}, nil)
	if err == nil || !strings.Contains(err.Error(), "verification failed after import") {
		t.Fatalf("tag presence was mistaken for config-ID verification: %v", err)
	}
}

func TestImagesPlainContainerdRequiresMatchingExplicitSockets(t *testing.T) {
	for _, opts := range []ImageOptions{
		{Runtime: "auto", NodeRuntimeVersion: "containerd://2.0.1"},
		{Runtime: "containerd", ContainerdAddress: "/run/containerd/containerd.sock"},
		{Runtime: "containerd", ContainerdAddress: "/run/a.sock", CRIEndpoint: "unix:///run/b.sock"},
		{Runtime: "docker", NodeRuntimeVersion: "containerd://2.2.0-k3s1"},
		{Runtime: "auto"},
		{Runtime: "docker", DockerAddress: "tcp://remote:2376"},
	} {
		if _, err := selectImageRuntime(opts); err == nil {
			t.Fatalf("unsafe/ambiguous runtime accepted: %+v", opts)
		}
	}
	b := makeImageTestBundle(t)
	fake := imageTestRuntime{t: t, bundle: b, ids: map[string]string{}}
	err := PrepareImages(context.Background(), fake.run, ImageOptions{
		BundleDir: b.dir, Runtime: "auto", NodeRuntimeVersion: "containerd://2.0.1",
		ContainerdAddress: "/run/containerd/containerd.sock", CRIEndpoint: "unix:///run/containerd/containerd.sock",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range fake.commands {
		if strings.HasPrefix(command, "ctr ") && (!strings.Contains(command, "--address /run/containerd/containerd.sock") || !strings.Contains(command, "--namespace k8s.io")) {
			t.Fatalf("ctr command has wrong socket/namespace: %s", command)
		}
		if strings.HasPrefix(command, "crictl ") && !strings.Contains(command, "--image-endpoint unix:///run/containerd/containerd.sock") {
			t.Fatalf("CRI image checks target wrong socket: %s", command)
		}
	}
}

func TestImagesAllCommandsHaveTimeoutsAndCancellationStopsPreparation(t *testing.T) {
	b := makeImageTestBundle(t)
	run := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	start := time.Now()
	err := PrepareImages(context.Background(), run, ImageOptions{BundleDir: b.dir, Runtime: "docker", CommandTimeout: time.Millisecond}, nil)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > time.Second {
		t.Fatalf("runtime access was not bounded: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := PrepareImages(ctx, run, ImageOptions{BundleDir: b.dir, Runtime: "docker"}, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation was not propagated: %v", err)
	}
}

func TestImagesRejectIncompleteManifestTraversalAndEscapingSymlink(t *testing.T) {
	for _, problem := range []string{"missing-ref", "duplicate-ref", "unknown-ref", "traversal", "absolute", "bad-sha", "symlink"} {
		t.Run(problem, func(t *testing.T) {
			b := makeImageTestBundle(t)
			switch problem {
			case "missing-ref":
				b.manifest.Archives[0].References = []string{"qa/observer:1"}
			case "duplicate-ref":
				b.manifest.Archives[1].References = []string{"qa/observer:1"}
			case "unknown-ref":
				b.manifest.Archives[1].References = []string{"unknown/image:1"}
			case "traversal":
				b.manifest.Archives[0].Path = "../outside.tar.gz"
			case "absolute":
				b.manifest.Archives[0].Path = "/tmp/outside.tar.gz"
			case "bad-sha":
				b.manifest.Archives[0].SHA256 = "not-a-sha"
			case "symlink":
				outside := filepath.Join(t.TempDir(), "outside.tar.gz")
				if err := os.WriteFile(outside, []byte("outside"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(b.paths[0]); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, b.paths[0]); err != nil {
					t.Fatal(err)
				}
			}
			b.write(t)
			called := false
			run := func(context.Context, string, ...string) ([]byte, error) { called = true; return nil, nil }
			if err := PrepareImages(context.Background(), run, ImageOptions{BundleDir: b.dir, Runtime: "docker"}, nil); err == nil || called {
				t.Fatalf("invalid delivery reached runtime commands: %v / called=%v", err, called)
			}
		})
	}
}
