package lakehouse

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ClusterSSHOptions controls access to additional selected Kubernetes workers.
// Hosts maps Kubernetes Node names to SSH hostnames or IPs; the default is each
// Node's InternalIP. Credentials and kubeconfig are never copied to workers.
type ClusterSSHOptions struct {
	User         string
	Port         int
	IdentityFile string
	Hosts        map[string]string
}

var (
	clusterSSHUser   = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]*\$?$`)
	clusterSSHLabel  = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?$`)
	clusterImageTemp = regexp.MustCompile(`^/tmp/oceanbase-lakehouse-images\.[A-Za-z0-9]{10}$`)
)

func validateClusterSSH(ssh ClusterSSHOptions, nodes []PreflightResult) error {
	if ssh.User != "" && !clusterSSHUser.MatchString(ssh.User) {
		return errors.New("SSH user must be a login name without whitespace, shell syntax or options")
	}
	if ssh.Port < 0 || ssh.Port > 65535 {
		return errors.New("SSH port must be between 1 and 65535")
	}
	if ssh.IdentityFile != "" {
		if strings.ContainsAny(ssh.IdentityFile, "\x00\r\n") {
			return errors.New("SSH identity path must not contain control characters")
		}
		info, err := os.Stat(ssh.IdentityFile)
		if err != nil || !info.Mode().IsRegular() {
			return errors.New("SSH identity must name an existing regular file")
		}
	}
	selected := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		if node.NodeName == "" || selected[node.NodeName] {
			return errors.New("cluster image preparation requires distinct named Kubernetes nodes")
		}
		selected[node.NodeName] = true
		host := node.NodeInternalIP
		if override, ok := ssh.Hosts[node.NodeName]; ok {
			host = override
		}
		if !validClusterSSHHost(host) {
			return fmt.Errorf("node %s requires a valid SSH hostname or IP (use --ssh-host for a different SSH address)", node.NodeName)
		}
	}
	for name := range ssh.Hosts {
		if !selected[name] {
			return fmt.Errorf("SSH host override names unselected node %q", name)
		}
	}
	return nil
}

func validClusterSSHHost(host string) bool {
	if ip := net.ParseIP(host); ip != nil {
		return !ip.IsUnspecified() && !ip.IsMulticast() && !ip.IsLoopback()
	}
	if len(host) > 253 || host == "" {
		return false
	}
	for _, label := range strings.Split(strings.TrimSuffix(host, "."), ".") {
		if len(label) > 63 || !clusterSSHLabel.MatchString(label) {
			return false
		}
	}
	return true
}

// PrepareClusterImages checks each selected worker's locked image IDs and only
// transfers archives missing on that worker. CheckOnly performs read-only SSH
// commands: no staging directory, copy, import or cleanup is attempted.
func PrepareClusterImages(ctx context.Context, run RunCommand, opts ImageOptions, nodes []PreflightResult, ssh ClusterSSHOptions, progress Progress) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if run == nil || len(nodes) != 3 {
		return errors.New("cluster image preparation requires a command runner and three selected workers")
	}
	if err := validateClusterSSH(ssh, nodes); err != nil {
		return err
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
	if ssh.User == "" {
		ssh.User = "root"
	}
	if ssh.Port == 0 {
		ssh.Port = 22
	}
	if ssh.IdentityFile != "" {
		var err error
		ssh.IdentityFile, err = filepath.Abs(ssh.IdentityFile)
		if err != nil {
			return fmt.Errorf("resolve SSH identity: %w", err)
		}
	}
	// Verify every SSH destination before the first runtime import. A successful
	// SSH login alone is not evidence that this is the selected Kubernetes node.
	workers := make([]*clusterImageWorker, len(nodes))
	var preflightErrors []error
	for i, node := range nodes {
		if checkLocalWorker(node) == nil {
			continue
		}
		host := node.NodeInternalIP
		if override, ok := ssh.Hosts[node.NodeName]; ok {
			host = override
		}
		worker := &clusterImageWorker{run: run, host: host, ssh: ssh, opts: opts, progress: progress}
		if err := worker.verifyIdentity(ctx, node); err != nil {
			preflightErrors = append(preflightErrors, fmt.Errorf("worker %s SSH preflight: %w", node.NodeName, err))
		}
		workers[i] = worker
	}
	if len(preflightErrors) > 0 {
		return errors.Join(preflightErrors...)
	}
	for i, node := range nodes {
		nodeOpts := opts
		nodeOpts.NodeRuntimeVersion = node.NodeRuntimeVersion
		runtime, err := selectImageRuntime(nodeOpts)
		if err == nil {
			nodeRun := run
			if workers[i] != nil {
				nodeRun = workers[i].remote
			}
			err = runtime.check(ctx, nodeRun, opts.CommandTimeout)
		}
		if err != nil {
			preflightErrors = append(preflightErrors, fmt.Errorf("worker %s runtime preflight: %w", node.NodeName, err))
		}
	}
	if len(preflightErrors) > 0 {
		return errors.Join(preflightErrors...)
	}
	var checkErrors []error
	for i, node := range nodes {
		progress(fmt.Sprintf("  Worker %d/%d: %s", i+1, len(nodes), node.NodeName))
		nodeOpts := opts
		nodeOpts.NodeRuntimeVersion = node.NodeRuntimeVersion
		var err error
		if worker := workers[i]; worker != nil {
			err = worker.prepare(ctx, nodeOpts)
		} else {
			err = PrepareImages(ctx, run, nodeOpts, progress)
		}
		if err != nil {
			err = fmt.Errorf("worker %s image preparation: %w", node.NodeName, err)
			if !opts.CheckOnly {
				return err
			}
			checkErrors = append(checkErrors, err)
		}
	}
	return errors.Join(checkErrors...)
}

type clusterImageWorker struct {
	run      RunCommand
	host     string
	ssh      ClusterSSHOptions
	opts     ImageOptions
	progress Progress
	tempDir  string
}

func (worker *clusterImageWorker) connectionArgs(scp bool) []string {
	args := []string{"-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=yes", "-o", "ConnectTimeout=10", "-o", "ConnectionAttempts=1", "-o", "ServerAliveInterval=15", "-o", "ServerAliveCountMax=2", "-o", "UpdateHostKeys=no", "-o", "ControlMaster=no", "-o", "ControlPath=none", "-o", "ControlPersist=no", "-o", "ForwardAgent=no", "-o", "ClearAllForwardings=yes", "-o", "PermitLocalCommand=no"}
	if scp {
		args = append(args, "-B", "-P", strconv.Itoa(worker.ssh.Port), "-o", "User="+worker.ssh.User)
	} else {
		args = append(args, "-n", "-T", "-p", strconv.Itoa(worker.ssh.Port), "-l", worker.ssh.User)
	}
	if worker.ssh.IdentityFile != "" {
		args = append(args, "-i", worker.ssh.IdentityFile)
	}
	return args
}

// OpenSSH passes a command string to the login shell, rather than forwarding an
// argv. Quote every argument, including runtime addresses and image references.
func clusterShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func (worker *clusterImageWorker) remote(ctx context.Context, name string, args ...string) ([]byte, error) {
	words := []string{clusterShellQuote(name)}
	for _, arg := range args {
		if strings.ContainsRune(arg, '\x00') {
			return nil, errors.New("remote command argument contains NUL")
		}
		words = append(words, clusterShellQuote(arg))
	}
	sshArgs := worker.connectionArgs(false)
	sshArgs = append(sshArgs, "--", worker.host, strings.Join(words, " "))
	return worker.run(ctx, "ssh", sshArgs...)
}

const clusterIdentityScript = `if [ -r /sys/class/dmi/id/product_uuid ]; then printf 'UUID '; cat /sys/class/dmi/id/product_uuid; printf '\n'; fi
if command -v ip >/dev/null 2>&1; then ip -o address show; fi`

func (worker *clusterImageWorker) verifyIdentity(ctx context.Context, node PreflightResult) error {
	output, err := runImageCommand(ctx, worker.remote, worker.opts.CommandTimeout, "sh", "-c", clusterIdentityScript)
	if err != nil {
		return fmt.Errorf("cannot read worker identity; verify SSH tools, key authentication and trusted known_hosts entry: %w", err)
	}
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "UUID" && node.NodeSystemUUID != "" && strings.EqualFold(fields[1], node.NodeSystemUUID) {
			return nil
		}
		for i, field := range fields {
			if (field == "inet" || field == "inet6") && i+1 < len(fields) {
				ip, _, err := net.ParseCIDR(fields[i+1])
				if err == nil && !ip.IsLoopback() && node.NodeInternalIP != "" && ip.Equal(net.ParseIP(node.NodeInternalIP)) {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("SSH host %s UUID/interface IP does not match Kubernetes node %s; verify --ssh-host before importing images", worker.host, node.NodeName)
}

func (worker *clusterImageWorker) prepare(ctx context.Context, opts ImageOptions) (result error) {
	bundle, err := filepath.Abs(opts.BundleDir)
	if err != nil {
		return err
	}
	_, manifest, err := readImageDelivery(bundle)
	if err != nil {
		return err
	}
	archiveSHA := make(map[string]string, len(manifest.Archives))
	for _, archive := range manifest.Archives {
		path, err := resolveImageArchive(bundle, archive)
		if err != nil {
			return err
		}
		archiveSHA[path] = archive.SHA256
	}
	defer func() {
		if worker.tempDir == "" {
			return
		}
		// A cancelled import must not leave multi-gigabyte temporary archives.
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		if _, err := worker.remote(cleanupCtx, "rm", "-rf", "--", worker.tempDir); err != nil {
			result = errors.Join(result, fmt.Errorf("could not remove temporary image directory %s on %s: %w", worker.tempDir, worker.host, err))
		}
	}()
	proxy := func(commandCtx context.Context, name string, args ...string) ([]byte, error) {
		isImport := (name == "docker" && len(args) >= 3 && args[len(args)-2] == "--input") ||
			((name == "k3s" || name == "ctr") && len(args) >= 3 && args[len(args)-2] == "import")
		if !isImport {
			return worker.remote(commandCtx, name, args...)
		}
		if opts.CheckOnly {
			return nil, errors.New("read-only image check cannot transfer or import archives")
		}
		localPath := args[len(args)-1]
		remotePath, err := worker.transfer(commandCtx, localPath, archiveSHA[localPath])
		if err != nil {
			return nil, err
		}
		remoteArgs := append([]string(nil), args...)
		remoteArgs[len(remoteArgs)-1] = remotePath
		output, err := worker.remote(commandCtx, name, remoteArgs...)
		if err == nil {
			// Keep peak staging space to one archive. Deferred cleanup also handles
			// failed transfers, verification, imports or cancellation.
			_, err = worker.remote(commandCtx, "rm", "-f", "--", remotePath)
		}
		return output, err
	}
	return PrepareImages(ctx, proxy, opts, worker.progress)
}

func (worker *clusterImageWorker) transfer(ctx context.Context, localPath, expected string) (string, error) {
	// Compare the remote bytes with the delivery manifest, not a new hash of a
	// potentially changed local file after PrepareImages verified the archives.
	if !imageSHA256.MatchString(expected) {
		return "", errors.New("cannot transfer an archive absent from the verified delivery manifest")
	}
	if worker.tempDir == "" {
		output, err := runImageCommand(ctx, worker.remote, worker.opts.CommandTimeout, "sh", "-c", "umask 077; mktemp -d /tmp/oceanbase-lakehouse-images.XXXXXXXXXX")
		if err != nil {
			return "", fmt.Errorf("create private remote image directory: %w", err)
		}
		path := strings.TrimSpace(string(output))
		if !clusterImageTemp.MatchString(path) {
			return "", errors.New("remote mktemp returned an unexpected path; refusing to copy or remove it")
		}
		worker.tempDir = path
	}
	remotePath := worker.tempDir + "/" + expected + ".tar"
	host := worker.host
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	args := worker.connectionArgs(true)
	args = append(args, "--", localPath, host+":"+remotePath)
	worker.progress("  Transferring missing image archive to worker " + worker.host)
	if _, err := worker.run(ctx, "scp", args...); err != nil {
		return "", fmt.Errorf("copy missing image archive (requires scp and key authentication): %w", err)
	}
	output, err := runImageCommand(ctx, worker.remote, worker.opts.CommandTimeout, "sha256sum", "--", remotePath)
	if err != nil {
		return "", fmt.Errorf("verify transferred image archive: %w", err)
	}
	fields := strings.Fields(string(output))
	if len(fields) < 1 || fields[0] != expected {
		return "", errors.New("transferred image archive SHA256 mismatch; archive was not imported")
	}
	return remotePath, nil
}
