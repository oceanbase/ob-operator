package lakehouse

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestContainerdImportsPlainTarAndCleansTemporaryFile(t *testing.T) {
	for _, runtime := range []string{"containerd", "k3s", "docker"} {
		for _, fail := range []bool{false, true} {
			t.Run(runtime+map[bool]string{false: "/success", true: "/failure"}[fail], func(t *testing.T) {
				dir := t.TempDir()
				path := filepath.Join(dir, "archive.tar.gz")
				payload := []byte("verified tar stream for old and new containerd")
				var compressed bytes.Buffer
				writer := gzip.NewWriter(&compressed)
				_, _ = writer.Write(payload)
				_ = writer.Close()
				if err := os.WriteFile(path, compressed.Bytes(), 0600); err != nil {
					t.Fatal(err)
				}
				imported := ""
				run := func(ctx context.Context, name string, args ...string) ([]byte, error) {
					imported = args[len(args)-1]
					data, err := os.ReadFile(imported)
					if err != nil {
						t.Fatal(err)
					}
					want := payload
					if runtime == "docker" {
						want = compressed.Bytes()
					}
					if !bytes.Equal(data, want) {
						t.Fatalf("%s did not receive a compatible archive", runtime)
					}
					if runtime != "docker" {
						if imported == path || filepath.Dir(imported) != dir {
							t.Fatal("expected private temporary tar beside the archive", imported)
						}
						st, _ := os.Stat(imported)
						if st.Mode().Perm() != 0600 {
							t.Fatal("temporary tar permissions", st.Mode())
						}
					}
					if fail {
						return nil, errors.New("runtime import failed")
					}
					return nil, nil
				}
				err := (imageRuntime{name: runtime, address: "/isolated.sock"}).importArchive(context.Background(), run, time.Minute, path)
				if (err != nil) != fail {
					t.Fatal(err)
				}
				if runtime != "docker" {
					if _, err := os.Stat(imported); !os.IsNotExist(err) {
						t.Fatal("temporary tar was not removed", err)
					}
				}
				if got, _ := os.ReadFile(path); !bytes.Equal(got, compressed.Bytes()) {
					t.Fatal("verified cache changed")
				}
			})
		}
	}
}

func TestContainerdRejectsBrokenGzipBeforeImport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.tar.gz")
	_ = os.WriteFile(path, []byte{0x1f, 0x8b, 0, 0}, 0600)
	called := false
	err := (imageRuntime{name: "containerd"}).importArchive(context.Background(), func(context.Context, string, ...string) ([]byte, error) { called = true; return nil, nil }, time.Minute, path)
	if err == nil || called {
		t.Fatal("broken gzip reached runtime", err)
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatal("temporary files remain", entries)
	}
}

func TestContainerdCanceledDecompressionDoesNotImport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "archive.tar.gz")
	var data bytes.Buffer
	writer := gzip.NewWriter(&data)
	_, _ = writer.Write([]byte(strings.Repeat("x", 4096)))
	_ = writer.Close()
	_ = os.WriteFile(path, data.Bytes(), 0600)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	err := (imageRuntime{name: "containerd"}).importArchive(ctx, func(context.Context, string, ...string) ([]byte, error) { called = true; return nil, nil }, time.Minute, path)
	if !errors.Is(err, context.Canceled) || called {
		t.Fatal("canceled decompression reached runtime", err)
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatal("temporary files remain", entries)
	}
}

func TestClusterImagesTransfersDecompressedTarWithDerivedDigest(t *testing.T) {
	f := newClusterImageFake(t)
	originals := make([][]byte, len(f.bundle.paths))
	for i, path := range f.bundle.paths {
		originals[i], _ = os.ReadFile(path)
		var zipped bytes.Buffer
		writer := gzip.NewWriter(&zipped)
		_, _ = writer.Write(originals[i])
		_ = writer.Close()
		if err := os.WriteFile(path, zipped.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(zipped.Bytes())
		f.bundle.manifest.Archives[i].SHA256 = hex.EncodeToString(sum[:])
	}
	f.bundle.write(t)
	transferred := map[string]int{}
	run := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "scp" {
			path := args[len(args)-2]
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			found := -1
			for i, want := range originals {
				if bytes.Equal(data, want) {
					found = i
					break
				}
			}
			if found < 0 {
				t.Fatal("worker received gzip or unverified content")
			}
			digest := sha256.Sum256(data)
			expected := hex.EncodeToString(digest[:]) + ".tar"
			if filepath.Base(args[len(args)-1]) != expected {
				t.Fatal("worker did not verify the derived tar digest")
			}
			transferred[expected] = found
		}
		output, err := f.run(ctx, name, args...)
		if name == "ssh" && err == nil {
			remote := splitClusterCommand(t, args[len(args)-1])
			if strings.Contains(strings.Join(remote, " "), " images import ") {
				index, ok := transferred[filepath.Base(remote[len(remote)-1])]
				if !ok {
					t.Fatal("imported an unverified transfer")
				}
				host := args[len(args)-2]
				for _, ref := range f.bundle.manifest.Archives[index].References {
					for _, image := range f.bundle.lock.Images {
						if image.Reference == ref {
							f.ids[host][ref] = image.ImageID
						}
					}
				}
			}
		}
		return output, err
	}
	if err := PrepareClusterImages(context.Background(), run, ImageOptions{BundleDir: f.bundle.dir, Runtime: "auto"}, f.nodes, ClusterSSHOptions{}, nil); err != nil {
		t.Fatal(err)
	}
	if len(f.imports) != 6 {
		t.Fatal("expected two archives per worker", f.imports)
	}
	for _, path := range f.bundle.paths {
		entries, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".import-*"))
		if len(entries) != 0 {
			t.Fatal("temporary decompressed archives remain", entries)
		}
	}
}
