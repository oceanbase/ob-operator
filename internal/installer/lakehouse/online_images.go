package lakehouse

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func validateOnlineArchive(a imageArchive) error {
	if a.URL == "" {
		return nil
	}
	u, err := url.Parse(a.URL)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("online URL must be HTTP(S) without credentials, query or fragment")
	}
	if a.Base != "bundle" || a.Path != "images/cache/"+a.SHA256+".tar.gz" || a.SizeBytes < 1 || a.SizeBytes > 40<<30 {
		return errors.New("online archives require a digest-named bundle cache path and a bounded size_bytes")
	}
	return nil
}

// Downloads are committed atomically only after byte count and SHA256 match the
// immutable delivery manifest. Completed archives are reused across retries;
// interrupted/invalid partial files are removed. No network is used by --check.
func ensureOnlineArchive(ctx context.Context, bundle string, a imageArchive, progress Progress) error {
	if err := validateOnlineArchive(a); err != nil {
		return err
	}
	path, err := resolveImageArchive(bundle, a)
	if err != nil {
		return err
	}
	if st, err := os.Lstat(path); err == nil {
		if !st.Mode().IsRegular() {
			return errors.New("online cache must contain regular files")
		}
		if st.Size() == a.SizeBytes {
			if sum, err := hashImageArchive(ctx, path); err == nil && sum == a.SHA256 {
				progress("[ready] Reusing verified download: " + a.Name)
				return nil
			}
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	imageDir := filepath.Join(bundle, "images")
	st, err := os.Lstat(imageDir)
	if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return errors.New("online image directory must be a real directory")
	}
	if err := privateDirectory(filepath.Dir(path)); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".download-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	downloadCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	progress(fmt.Sprintf("  Downloading %s (%.1f MiB); verifying before import.", a.Name, float64(a.SizeBytes)/(1<<20)))
	err = withProgress(downloadCtx, progressInterval, progress, "Downloading "+a.Name, func() error {
		req, err := http.NewRequestWithContext(downloadCtx, http.MethodGet, a.URL, nil)
		if err != nil {
			return errors.New("cannot construct image download request")
		}
		client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 || req.URL.User != nil || (via[0].URL.Scheme == "https" && req.URL.Scheme != "https") {
				return errors.New("unsafe download redirect")
			}
			return nil
		}}
		response, err := client.Do(req)
		if err != nil {
			if downloadCtx.Err() != nil {
				return downloadCtx.Err()
			}
			return errors.New("cannot reach image download service; check its network access")
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("image download returned HTTP %d", response.StatusCode)
		}
		if response.ContentLength >= 0 && response.ContentLength != a.SizeBytes {
			return errors.New("image download size differs from manifest")
		}
		hash := sha256.New()
		n, err := io.Copy(io.MultiWriter(file, hash), io.LimitReader(response.Body, a.SizeBytes+1))
		if err != nil {
			return fmt.Errorf("image download interrupted: %w", err)
		}
		if n != a.SizeBytes || hex.EncodeToString(hash.Sum(nil)) != a.SHA256 {
			return errors.New("image download size or SHA256 mismatch")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return err
	}
	progress("[done] Download verified: " + a.Name)
	return nil
}
