package lakehouse

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ImageOptions describes the local worker selected and checked by the caller.
// Runtime auto uses the selected Kubernetes Node's containerRuntimeVersion,
// never the presence of a docker executable on the installation machine.
type ImageOptions struct {
	BundleDir          string
	Runtime            string
	NodeRuntimeVersion string
	DockerAddress      string
	ContainerdAddress  string
	CRIEndpoint        string
	CheckOnly          bool
	CommandTimeout     time.Duration
	ImportTimeout      time.Duration
}

type lockedImage struct {
	Role      string `json:"role"`
	Reference string `json:"reference"`
	ImageID   string `json:"image_id"`
}

type imageLock struct {
	Images []lockedImage `json:"images"`
}

// Self-contained packages keep every archive relative to the package (bundle),
// so the extracted directory can move without retaining its original parent.
// The delivery base remains supported for older packages with adjacent archives;
// neither form depends on the user's current working directory.
type imageArchive struct {
	Name       string   `json:"name"`
	Base       string   `json:"base"`
	Path       string   `json:"path"`
	SHA256     string   `json:"sha256"`
	References []string `json:"references"`
}

type archiveManifest struct {
	SchemaVersion int            `json:"schemaVersion"`
	Archives      []imageArchive `json:"archives"`
}

type imageRuntime struct {
	name    string
	address string
	cri     string
}

var imageSHA256 = regexp.MustCompile(`^[0-9a-f]{64}$`)

// PrepareImages imports only archives needed by the selected local runtime.
// All required archives pass validation before the first import. A successful
// import is followed by comparison of config image IDs, not merely image tags.
// CheckOnly never imports; it reports whether every locked image is ready.
func PrepareImages(ctx context.Context, run RunCommand, opts ImageOptions, progress Progress) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if run == nil {
		return errors.New("image preparation requires a command runner")
	}
	if progress == nil {
		progress = func(string) {}
	}
	if opts.CommandTimeout <= 0 {
		opts.CommandTimeout = 30 * time.Second
	}
	if opts.ImportTimeout <= 0 {
		opts.ImportTimeout = 10 * time.Minute
	}
	bundle, err := filepath.Abs(opts.BundleDir)
	if err != nil || opts.BundleDir == "" {
		return errors.New("image preparation requires an absolute or relative bundle directory")
	}
	lock, manifest, err := readImageDelivery(bundle)
	if err != nil {
		return err
	}
	runtime, err := selectImageRuntime(opts)
	if err != nil {
		return err
	}
	progress(fmt.Sprintf("检查本机镜像：runtime=%s，目标 Node runtime=%s", runtime.name, opts.NodeRuntimeVersion))
	if err := runtime.check(ctx, run, opts.CommandTimeout); err != nil {
		return err
	}
	missing, err := inspectLockedImages(ctx, run, runtime, opts.CommandTimeout, lock.Images)
	if err != nil {
		return err
	}
	if len(missing) == 0 {
		progress(fmt.Sprintf("本机 %d 个镜像 ID 全部匹配，跳过归档校验和导入", len(lock.Images)))
		return nil
	}
	if opts.CheckOnly {
		return fmt.Errorf("local %s images are not ready: %s; run image preparation on this worker", runtime.name, strings.Join(missing, "; "))
	}
	missingRefs := make(map[string]bool, len(missing))
	for _, detail := range missing {
		missingRefs[strings.SplitN(detail, " ", 2)[0]] = true
	}
	var needed []imageArchive
	for _, archive := range manifest.Archives {
		for _, ref := range archive.References {
			if missingRefs[ref] {
				needed = append(needed, archive)
				break
			}
		}
	}
	progress(fmt.Sprintf("需要准备 %d 个镜像归档；先完整校验，再开始导入", len(needed)))
	paths := make([]string, len(needed))
	var problems []error
	for i, archive := range needed {
		if err := ctx.Err(); err != nil {
			return err
		}
		path, pathErr := resolveImageArchive(bundle, archive)
		if pathErr != nil {
			problems = append(problems, fmt.Errorf("%s: %w", archive.Name, pathErr))
			continue
		}
		paths[i] = path
		info, statErr := os.Stat(path)
		if statErr != nil {
			problems = append(problems, fmt.Errorf("%s: required archive %s: %w", archive.Name, path, statErr))
			continue
		}
		if !info.Mode().IsRegular() {
			problems = append(problems, fmt.Errorf("%s: archive is not a regular file: %s", archive.Name, path))
			continue
		}
		progress(fmt.Sprintf("校验归档 %d/%d：%s（%.1f MiB）", i+1, len(needed), archive.Name, float64(info.Size())/(1024*1024)))
		actual, hashErr := hashImageArchive(ctx, path)
		if hashErr != nil {
			problems = append(problems, fmt.Errorf("%s: %w", archive.Name, hashErr))
		} else if actual != archive.SHA256 {
			problems = append(problems, fmt.Errorf("%s: SHA256 mismatch for %s: expected %s, got %s", archive.Name, path, archive.SHA256, actual))
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("image archives are not ready; nothing was imported:\n%w", errors.Join(problems...))
	}
	for i, archive := range needed {
		progress(fmt.Sprintf("导入归档 %d/%d：%s → %s", i+1, len(needed), archive.Name, runtime.name))
		start := time.Now()
		if err := runtime.importArchive(ctx, run, opts.ImportTimeout, paths[i]); err != nil {
			return fmt.Errorf("import %s failed; previous imports are safe to reuse on retry: %w", archive.Name, err)
		}
		progress(fmt.Sprintf("归档 %s 导入完成（%s）", archive.Name, time.Since(start).Round(time.Second)))
	}
	missing, err = inspectLockedImages(ctx, run, runtime, opts.CommandTimeout, lock.Images)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return fmt.Errorf("image ID verification failed after import: %s", strings.Join(missing, "; "))
	}
	progress(fmt.Sprintf("本机 %d 个镜像 ID 全部匹配；其他目标 Worker 仍需分别准备", len(lock.Images)))
	return nil
}

func readImageDelivery(bundle string) (imageLock, archiveManifest, error) {
	var lock imageLock
	var manifest archiveManifest
	for _, file := range []struct {
		path string
		to   interface{}
	}{{"image-lock.json", &lock}, {"images/archives.json", &manifest}} {
		data, err := os.ReadFile(filepath.Join(bundle, filepath.FromSlash(file.path)))
		if err != nil {
			return lock, manifest, fmt.Errorf("read %s: %w", file.path, err)
		}
		if err := json.Unmarshal(data, file.to); err != nil {
			return lock, manifest, fmt.Errorf("invalid %s: %w", file.path, err)
		}
	}
	if len(lock.Images) == 0 || manifest.SchemaVersion != 1 || len(manifest.Archives) == 0 {
		return lock, manifest, errors.New("image delivery requires nonempty image-lock images and archives.json schemaVersion 1")
	}
	refs := make(map[string]bool, len(lock.Images))
	for _, image := range lock.Images {
		if image.Reference == "" || strings.HasPrefix(image.Reference, "-") || strings.ContainsAny(image.Reference, " \t\r\n") || !strings.HasPrefix(image.ImageID, "sha256:") || !imageSHA256.MatchString(strings.TrimPrefix(image.ImageID, "sha256:")) {
			return lock, manifest, fmt.Errorf("invalid locked image reference/config ID for role %q", image.Role)
		}
		if refs[image.Reference] {
			return lock, manifest, fmt.Errorf("duplicate locked image %s", image.Reference)
		}
		refs[image.Reference] = true
	}
	covered, names := map[string]bool{}, map[string]bool{}
	for _, archive := range manifest.Archives {
		if archive.Name == "" || names[archive.Name] || !imageSHA256.MatchString(archive.SHA256) || len(archive.References) == 0 {
			return lock, manifest, fmt.Errorf("invalid or duplicate image archive %q", archive.Name)
		}
		names[archive.Name] = true
		if _, err := resolveImageArchive(bundle, archive); err != nil {
			return lock, manifest, fmt.Errorf("invalid archive %s: %w", archive.Name, err)
		}
		for _, ref := range archive.References {
			if !refs[ref] || covered[ref] {
				return lock, manifest, fmt.Errorf("archive %s has untracked or duplicate image %s", archive.Name, ref)
			}
			covered[ref] = true
		}
	}
	var uncovered []string
	for ref := range refs {
		if !covered[ref] {
			uncovered = append(uncovered, ref)
		}
	}
	if len(uncovered) > 0 {
		sort.Strings(uncovered)
		return lock, manifest, fmt.Errorf("images missing from archive manifest: %s", strings.Join(uncovered, ", "))
	}
	return lock, manifest, nil
}

func resolveImageArchive(bundle string, archive imageArchive) (string, error) {
	base := bundle
	switch archive.Base {
	case "bundle":
	case "delivery":
		base = filepath.Dir(bundle)
	default:
		return "", fmt.Errorf("base must be bundle or delivery, got %q", archive.Base)
	}
	if archive.Path == "" || filepath.IsAbs(archive.Path) || strings.Contains(archive.Path, "\\") {
		return "", errors.New("archive path must be a nonempty relative path")
	}
	for _, part := range strings.Split(archive.Path, "/") {
		if part == ".." || part == "." || part == "" {
			return "", errors.New("archive path cannot contain empty, dot or parent segments")
		}
	}
	path := filepath.Join(base, filepath.FromSlash(archive.Path))
	// Reject existing symlinks that escape the declared root. Missing archives are
	// aggregated later, and are harmless when the locked images already exist.
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		baseResolved, baseErr := filepath.EvalSymlinks(base)
		if baseErr != nil {
			return "", baseErr
		}
		relative, relErr := filepath.Rel(baseResolved, resolved)
		if relErr != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return "", errors.New("archive symlink escapes its declared base directory")
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}
	return path, nil
}

func selectImageRuntime(opts ImageOptions) (imageRuntime, error) {
	name := opts.Runtime
	if name == "" || name == "auto" {
		switch {
		case strings.HasPrefix(opts.NodeRuntimeVersion, "docker://"):
			name = "docker"
		case strings.HasPrefix(opts.NodeRuntimeVersion, "containerd://"):
			if strings.Contains(strings.ToLower(opts.NodeRuntimeVersion), "k3s") || strings.Contains(opts.ContainerdAddress, "/k3s/") {
				name = "k3s"
			} else {
				name = "containerd"
			}
		default:
			return imageRuntime{}, fmt.Errorf("cannot select runtime from target Node containerRuntimeVersion %q; select a supported worker/runtime", opts.NodeRuntimeVersion)
		}
	}
	if name != "docker" && name != "k3s" && name != "containerd" {
		return imageRuntime{}, fmt.Errorf("unsupported runtime %q; use auto, docker, k3s or containerd", name)
	}
	if opts.NodeRuntimeVersion != "" {
		if (name == "docker" && !strings.HasPrefix(opts.NodeRuntimeVersion, "docker://")) || (name != "docker" && !strings.HasPrefix(opts.NodeRuntimeVersion, "containerd://")) {
			return imageRuntime{}, fmt.Errorf("runtime %s does not match selected Node runtime %s", name, opts.NodeRuntimeVersion)
		}
	}
	runtime := imageRuntime{name: name}
	if name == "docker" {
		address := strings.TrimPrefix(opts.DockerAddress, "unix://")
		if address == "" {
			address = "/var/run/docker.sock"
		}
		if !filepath.IsAbs(address) {
			return runtime, errors.New("Docker image preparation requires an absolute local Unix socket; remote/TCP endpoints are not supported")
		}
		runtime.address = "unix://" + address
	}
	if name == "containerd" {
		if opts.ContainerdAddress == "" || opts.CRIEndpoint == "" {
			return runtime, errors.New("plain containerd requires ctr and crictl plus explicit containerd socket address and CRI endpoint (for example /run/containerd/containerd.sock and unix:///run/containerd/containerd.sock)")
		}
		address := strings.TrimPrefix(opts.ContainerdAddress, "unix://")
		criPath := strings.TrimPrefix(opts.CRIEndpoint, "unix://")
		if !filepath.IsAbs(address) || !strings.HasPrefix(opts.CRIEndpoint, "unix://") || filepath.Clean(address) != filepath.Clean(criPath) {
			return runtime, errors.New("containerd address and CRI endpoint must name the same absolute local Unix socket; remote/TCP endpoints are not supported")
		}
		runtime.address, runtime.cri = address, opts.CRIEndpoint
	}
	return runtime, nil
}

func runImageCommand(ctx context.Context, run RunCommand, timeout time.Duration, name string, args ...string) ([]byte, error) {
	limited, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	output, err := run(limited, name, args...)
	if limited.Err() != nil {
		return nil, fmt.Errorf("%s command did not complete within its deadline: %w", name, limited.Err())
	}
	return output, err
}

func (runtime imageRuntime) check(ctx context.Context, run RunCommand, timeout time.Duration) error {
	var err error
	switch runtime.name {
	case "docker":
		_, err = runImageCommand(ctx, run, timeout, "docker", "--host", runtime.address, "info", "--format", "{{.ServerVersion}}")
	case "k3s":
		_, err = runImageCommand(ctx, run, timeout, "k3s", "crictl", "version")
	case "containerd":
		_, err = runImageCommand(ctx, run, timeout, "ctr", "--address", runtime.address, "--namespace", "k8s.io", "version")
		if err == nil {
			_, err = runImageCommand(ctx, run, timeout, "crictl", "--runtime-endpoint", runtime.cri, "--image-endpoint", runtime.cri, "version")
		}
	}
	if err != nil {
		return fmt.Errorf("cannot access selected local %s runtime; check installed tools, daemon/socket and permissions: %w", runtime.name, err)
	}
	return nil
}

func (runtime imageRuntime) inspect(ctx context.Context, run RunCommand, timeout time.Duration, reference string) (string, error) {
	var output []byte
	var err error
	switch runtime.name {
	case "docker":
		output, err = runImageCommand(ctx, run, timeout, "docker", "--host", runtime.address, "image", "inspect", "--format", "{{.Id}}", reference)
		if err == nil {
			return strings.TrimSpace(string(output)), nil
		}
	case "k3s":
		output, err = runImageCommand(ctx, run, timeout, "k3s", "crictl", "inspecti", reference)
	case "containerd":
		output, err = runImageCommand(ctx, run, timeout, "crictl", "--runtime-endpoint", runtime.cri, "--image-endpoint", runtime.cri, "inspecti", reference)
	}
	if err != nil {
		return "", err
	}
	var result struct {
		Status struct {
			ID string `json:"id"`
		} `json:"status"`
	}
	if err := json.Unmarshal(output, &result); err != nil || result.Status.ID == "" {
		return "", errors.New("runtime inspecti did not return status.id")
	}
	return result.Status.ID, nil
}

func inspectLockedImages(ctx context.Context, run RunCommand, runtime imageRuntime, timeout time.Duration, images []lockedImage) ([]string, error) {
	var missing []string
	for _, image := range images {
		actual, err := runtime.inspect(ctx, run, timeout, image.Reference)
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, err
		}
		if err != nil {
			missing = append(missing, image.Reference+" (missing or unreadable)")
		} else if actual != image.ImageID {
			missing = append(missing, fmt.Sprintf("%s (expected %s, actual %s)", image.Reference, image.ImageID, actual))
		}
	}
	return missing, nil
}

func (runtime imageRuntime) importArchive(ctx context.Context, run RunCommand, timeout time.Duration, path string) error {
	var err error
	switch runtime.name {
	case "docker":
		_, err = runImageCommand(ctx, run, timeout, "docker", "--host", runtime.address, "load", "--input", path)
	case "k3s":
		// containerd 2.x's transfer importer accepts gzip Docker-save archives;
		// --local does not, so do not add it to this invocation.
		_, err = runImageCommand(ctx, run, timeout, "k3s", "ctr", "images", "import", path)
	case "containerd":
		_, err = runImageCommand(ctx, run, timeout, "ctr", "--address", runtime.address, "--namespace", "k8s.io", "images", "import", path)
	}
	return err
}

func hashImageArchive(ctx context.Context, path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	buffer := make([]byte, 1024*1024)
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		n, readErr := file.Read(buffer)
		if n > 0 {
			_, _ = hash.Write(buffer[:n])
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return "", readErr
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
