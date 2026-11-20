package lakehouse

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func onlineImageFixture(t *testing.T, bad bool) (imageTestBundle, *atomic.Int32) {
	t.Helper()
	b := makeImageTestBundle(t)
	payloads := map[string][]byte{}
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		data := payloads[r.URL.Path]
		if bad {
			data = append(append([]byte(nil), data...), 'x')
		}
		_, _ = w.Write(data)
	}))
	t.Cleanup(server.Close)
	for i := range b.manifest.Archives {
		a := &b.manifest.Archives[i]
		data, err := os.ReadFile(b.paths[i])
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(b.paths[i]); err != nil {
			t.Fatal(err)
		}
		payloads["/"+a.Name] = data
		a.Base, a.Path = "bundle", "images/cache/"+a.SHA256+".tar.gz"
		a.URL, a.SizeBytes = server.URL+"/"+a.Name, int64(len(data))
		b.paths[i] = filepath.Join(b.dir, filepath.FromSlash(a.Path))
	}
	b.write(t)
	return b, &requests
}

func TestOnlineImagesColdDownloadAndCachedRetry(t *testing.T) {
	b, requests := onlineImageFixture(t, false)
	fake := imageTestRuntime{t: t, bundle: b, ids: map[string]string{}}
	opts := ImageOptions{BundleDir: b.dir, Runtime: "k3s"}
	if err := PrepareImages(context.Background(), fake.run, opts, nil); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 2 || len(fake.imports) != 2 {
		t.Fatalf("downloads=%d imports=%v", requests.Load(), fake.imports)
	}
	for _, p := range b.paths {
		if _, err := os.Stat(p); err != nil {
			t.Fatal(err)
		}
	}
	if err := PrepareImages(context.Background(), fake.run, opts, nil); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 2 || len(fake.imports) != 2 {
		t.Fatal("warm retry downloaded or imported images again")
	}
}

func TestOnlineImagesCheckNeverDownloadsOrCreatesCache(t *testing.T) {
	b, requests := onlineImageFixture(t, false)
	fake := imageTestRuntime{t: t, bundle: b, ids: map[string]string{}}
	err := PrepareImages(context.Background(), fake.run, ImageOptions{BundleDir: b.dir, Runtime: "k3s", CheckOnly: true}, nil)
	if err == nil || requests.Load() != 0 || len(fake.imports) != 0 {
		t.Fatalf("err=%v requests=%d imports=%v", err, requests.Load(), fake.imports)
	}
	if _, err := os.Stat(filepath.Join(b.dir, "images/cache")); !os.IsNotExist(err) {
		t.Fatal("read-only preflight created cache")
	}
}

func TestOnlineImagesCorruptDownloadImportsNothing(t *testing.T) {
	b, _ := onlineImageFixture(t, true)
	fake := imageTestRuntime{t: t, bundle: b, ids: map[string]string{}}
	err := PrepareImages(context.Background(), fake.run, ImageOptions{BundleDir: b.dir, Runtime: "k3s"}, nil)
	if err == nil || len(fake.imports) != 0 {
		t.Fatalf("err=%v imports=%v", err, fake.imports)
	}
	for _, p := range b.paths {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatalf("bad download committed to cache %s", p)
		}
	}
	parts, _ := filepath.Glob(filepath.Join(b.dir, "images/cache/.download-*"))
	if len(parts) != 0 {
		t.Fatalf("temporary downloads remain: %v", parts)
	}
}

func TestOnlineImagesRejectUnsafeDownloadManifest(t *testing.T) {
	for _, change := range []string{"credentials", "query", "scheme", "cache-path", "size"} {
		t.Run(change, func(t *testing.T) {
			b, _ := onlineImageFixture(t, false)
			a := &b.manifest.Archives[0]
			switch change {
			case "credentials":
				a.URL = "https://user:pass@example.com/a"
			case "query":
				a.URL += "?token=secret"
			case "scheme":
				a.URL = "file:///tmp/a"
			case "cache-path":
				a.Path = "install.sh"
			case "size":
				a.SizeBytes = 0
			}
			b.write(t)
			_, _, err := readImageDelivery(b.dir)
			if err == nil {
				t.Fatal("unsafe manifest accepted")
			}
		})
	}
}

func TestOnlineImagesCancellationAndCacheSymlink(t *testing.T) {
	b, _ := onlineImageFixture(t, false)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fake := imageTestRuntime{t: t, bundle: b, ids: map[string]string{}}
	if err := PrepareImages(ctx, fake.run, ImageOptions{BundleDir: b.dir, Runtime: "k3s"}, nil); err == nil {
		t.Fatal("cancelled request succeeded")
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(b.dir, "images/cache")); err != nil {
		t.Fatal(err)
	}
	err := PrepareImages(context.Background(), fake.run, ImageOptions{BundleDir: b.dir, Runtime: "k3s"}, nil)
	if err == nil || len(fake.imports) != 0 {
		t.Fatalf("symlink cache accepted: %v", err)
	}
	entries, _ := os.ReadDir(outside)
	if len(entries) != 0 {
		t.Fatal("download escaped bundle")
	}
}

func TestOnlineImagesWrongSHAAndRetainedCacheRecovery(t *testing.T) {
	b, requests := onlineImageFixture(t, false)
	for _, p := range b.paths {
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("corrupt cache"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	fake := imageTestRuntime{t: t, bundle: b, ids: map[string]string{}}
	if err := PrepareImages(context.Background(), fake.run, ImageOptions{BundleDir: b.dir, Runtime: "k3s"}, nil); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 2 {
		t.Fatalf("corrupt cached downloads not replaced: %d", requests.Load())
	}
	// A same-length response with the wrong digest must never be imported.
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, strings.Repeat("x", int(b.manifest.Archives[0].SizeBytes)))
	}))
	defer bad.Close()
	a := &b.manifest.Archives[0]
	a.URL = bad.URL
	_ = os.Remove(b.paths[0])
	b.write(t)
	fake.ids = map[string]string{}
	fake.imports = nil
	if err := PrepareImages(context.Background(), fake.run, ImageOptions{BundleDir: b.dir, Runtime: "k3s"}, nil); err == nil || len(fake.imports) != 0 {
		t.Fatalf("wrong SHA accepted: %v", err)
	}
}

func TestOnlineImagesCancelledDuringDownloadLeavesNoPartialImport(t *testing.T) {
	b, _ := onlineImageFixture(t, false)
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	defer server.Close()
	b.manifest.Archives[0].URL = server.URL
	b.write(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { <-started; cancel() }()
	fake := imageTestRuntime{t: t, bundle: b, ids: map[string]string{}}
	if err := PrepareImages(ctx, fake.run, ImageOptions{BundleDir: b.dir, Runtime: "k3s"}, nil); err == nil || len(fake.imports) != 0 {
		t.Fatalf("cancelled download imported data: %v", err)
	}
	parts, _ := filepath.Glob(filepath.Join(b.dir, "images/cache/.download-*"))
	if len(parts) != 0 {
		t.Fatalf("cancelled downloads remain: %v", parts)
	}
}
