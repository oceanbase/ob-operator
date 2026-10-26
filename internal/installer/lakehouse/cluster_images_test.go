package lakehouse

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type clusterImageFake struct {
	t            *testing.T
	bundle       imageTestBundle
	nodes        []PreflightResult
	ids          map[string]map[string]string
	commands     []string
	copied       []string
	imports      []string
	identityBad  string
	runtimeBad   string
	shaBad       bool
	copyBad      bool
	tempResponse string
	block        string
}

func newClusterImageFake(t *testing.T) *clusterImageFake {
	t.Helper()
	f := &clusterImageFake{t: t, bundle: makeImageTestBundle(t), ids: map[string]map[string]string{}}
	for i := 0; i < 3; i++ {
		node := PreflightResult{NodeName: fmt.Sprintf("worker-%d", i+1), NodeInternalIP: fmt.Sprintf("192.0.2.%d", i+11), NodeSystemUUID: fmt.Sprintf("fake-system-uuid-%d", i+1), NodeRuntimeVersion: "containerd://2.2.0-k3s1"}
		f.nodes = append(f.nodes, node)
		f.ids[node.NodeInternalIP] = map[string]string{}
	}
	return f
}

func (f *clusterImageFake) fill(host string) {
	for _, image := range f.bundle.lock.Images {
		f.ids[host][image.Reference] = image.ImageID
	}
}

// Decode only the explicit quoted-argv format produced by clusterShellQuote.
// Reject unquoted shell tokens so tests also catch accidental command expansion.
func splitClusterCommand(t *testing.T, command string) []string {
	t.Helper()
	var result []string
	for len(command) > 0 {
		var word strings.Builder
		for len(command) > 0 && command[0] != ' ' {
			quote := command[0]
			if quote != '\'' && quote != '"' {
				t.Fatalf("unquoted shell token: %q", command)
			}
			command = command[1:]
			end := strings.IndexByte(command, quote)
			if end < 0 {
				t.Fatalf("unterminated quoted shell token: %q", command)
			}
			word.WriteString(command[:end])
			command = command[end+1:]
		}
		result = append(result, word.String())
		command = strings.TrimLeft(command, " ")
	}
	return result
}

func (f *clusterImageFake) run(ctx context.Context, name string, args ...string) ([]byte, error) {
	f.t.Helper()
	if _, ok := ctx.Deadline(); !ok {
		f.t.Fatal("cluster image command has no deadline")
	}
	command := name + " " + strings.Join(args, " ")
	f.commands = append(f.commands, command)
	if f.block != "" && strings.Contains(command, f.block) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	for _, option := range []string{"BatchMode=yes", "StrictHostKeyChecking=yes", "ConnectTimeout=10", "UpdateHostKeys=no", "ControlPath=none", "ForwardAgent=no", "PermitLocalCommand=no"} {
		if !strings.Contains(command, option) {
			f.t.Fatalf("missing connection protection %s: %s", option, command)
		}
	}
	if name == "scp" {
		f.copied = append(f.copied, args[len(args)-2]+" -> "+args[len(args)-1])
		if f.copyBad {
			return nil, errors.New("copy interrupted")
		}
		return nil, nil
	}
	if name != "ssh" || args[len(args)-3] != "--" {
		f.t.Fatalf("unexpected command: %s", command)
	}
	host := args[len(args)-2]
	remote := splitClusterCommand(f.t, args[len(args)-1])
	raw := strings.Join(remote, " ")
	if len(remote) == 3 && remote[0] == "sh" && remote[2] == clusterIdentityScript {
		if host == f.identityBad {
			return []byte("UUID different-machine\n2: eth0 inet 203.0.113.250/24 scope global\n"), nil
		}
		for _, node := range f.nodes {
			if node.NodeInternalIP == host {
				return []byte("UUID " + strings.ToUpper(node.NodeSystemUUID) + "\n"), nil
			}
		}
		f.t.Fatalf("identity queried unknown host %s", host)
	}
	if strings.Contains(raw, "mktemp -d") {
		if f.tempResponse != "" {
			return []byte(f.tempResponse), nil
		}
		return []byte("/tmp/oceanbase-lakehouse-images.ABC123def4\n"), nil
	}
	if remote[0] == "rm" {
		return nil, nil
	}
	if remote[0] == "sha256sum" {
		if f.shaBad {
			return []byte(strings.Repeat("0", 64) + "  " + remote[len(remote)-1]), nil
		}
		return []byte(strings.TrimSuffix(filepath.Base(remote[len(remote)-1]), ".tar") + "  " + remote[len(remote)-1]), nil
	}
	if strings.HasSuffix(raw, " version") || strings.Contains(raw, " info ") {
		if host == f.runtimeBad {
			return nil, errors.New("runtime unavailable")
		}
		return []byte("runtime available"), nil
	}
	if strings.Contains(raw, "inspecti ") || strings.Contains(raw, "image inspect ") {
		id := f.ids[host][remote[len(remote)-1]]
		if id == "" {
			return nil, errors.New("missing image")
		}
		if remote[0] == "docker" {
			return []byte(id), nil
		}
		return []byte(`{"status":{"id":"` + id + `"}}`), nil
	}
	if strings.Contains(raw, " images import ") || strings.Contains(raw, " load --input ") {
		path := remote[len(remote)-1]
		f.imports = append(f.imports, host+":"+path)
		for _, archive := range f.bundle.manifest.Archives {
			if filepath.Base(path) != archive.SHA256+".tar" {
				continue
			}
			for _, ref := range archive.References {
				for _, image := range f.bundle.lock.Images {
					if image.Reference == ref {
						f.ids[host][ref] = image.ImageID
					}
				}
			}
		}
		return nil, nil
	}
	f.t.Fatalf("unexpected remote command: %s", raw)
	return nil, errors.New("unreachable")
}

func (f *clusterImageFake) prepare(check bool) error {
	return PrepareClusterImages(context.Background(), f.run, ImageOptions{BundleDir: f.bundle.dir, Runtime: "auto", CheckOnly: check}, f.nodes, ClusterSSHOptions{}, nil)
}

func assertClusterNoMutation(t *testing.T, commands []string) {
	t.Helper()
	for _, command := range commands {
		if strings.HasPrefix(command, "scp ") || strings.Contains(command, "mktemp ") || strings.Contains(command, "'import'") || strings.Contains(command, "'load'") || strings.Contains(command, "'rm'") {
			t.Fatalf("unexpected mutation: %s", command)
		}
	}
}

func TestClusterImagesReuseCachesWithoutArchivesOrRemoteWrites(t *testing.T) {
	f := newClusterImageFake(t)
	for host := range f.ids {
		f.fill(host)
	}
	for _, path := range f.bundle.paths {
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.prepare(false); err != nil {
		t.Fatal(err)
	}
	assertClusterNoMutation(t, f.commands)
}

func TestClusterImagesCheckAggregatesWorkersWithoutWrites(t *testing.T) {
	f := newClusterImageFake(t)
	err := f.prepare(true)
	if err == nil {
		t.Fatal("missing images were accepted")
	}
	for _, node := range f.nodes {
		if !strings.Contains(err.Error(), node.NodeName) {
			t.Fatalf("missing worker result %s: %v", node.NodeName, err)
		}
	}
	assertClusterNoMutation(t, f.commands)
}

func TestClusterImagesTransfersOnlyMissingArchivesAndVerifiesIDs(t *testing.T) {
	f := newClusterImageFake(t)
	for host := range f.ids {
		f.fill(host)
	}
	missingHost := f.nodes[1].NodeInternalIP
	delete(f.ids[missingHost], "minio/minio:1")
	if err := f.prepare(false); err != nil {
		t.Fatal(err)
	}
	if len(f.copied) != 1 || !strings.Contains(f.copied[0], "minio-cache.tar.gz -> "+missingHost+":/tmp/oceanbase-lakehouse-images.") || len(f.imports) != 1 {
		t.Fatalf("copied/imported cached or unexpected archives: copies=%v imports=%v", f.copied, f.imports)
	}
	all := strings.Join(f.commands, "\n")
	if !strings.Contains(all, "'sha256sum' '--'") || !strings.Contains(all, "'rm' '-rf' '--' '/tmp/oceanbase-lakehouse-images.ABC123def4'") {
		t.Fatalf("missing checksum verification or bounded cleanup: %s", all)
	}
	if f.ids[missingHost]["minio/minio:1"] != f.bundle.lock.Images[2].ImageID {
		t.Fatal("missing image did not become available")
	}
	if err := f.prepare(true); err != nil {
		t.Fatalf("imported worker images did not pass read-only recheck: %v", err)
	}
}

func TestClusterImagesIdentityAndRuntimeAllPassBeforeAnyMutation(t *testing.T) {
	for _, failure := range []string{"identity", "runtime"} {
		t.Run(failure, func(t *testing.T) {
			f := newClusterImageFake(t)
			if failure == "identity" {
				f.identityBad = f.nodes[2].NodeInternalIP
			} else {
				f.runtimeBad = f.nodes[2].NodeInternalIP
			}
			err := f.prepare(false)
			if err == nil || !strings.Contains(err.Error(), f.nodes[2].NodeName) {
				t.Fatalf("expected last-worker preflight failure: %v", err)
			}
			assertClusterNoMutation(t, f.commands)
		})
	}
}

func TestClusterImagesChecksumFailureDoesNotImportAndCleansStaging(t *testing.T) {
	f := newClusterImageFake(t)
	f.shaBad = true
	err := f.prepare(false)
	if err == nil || !strings.Contains(err.Error(), "SHA256 mismatch") || len(f.imports) > 0 {
		t.Fatalf("bad transferred archive imported or accepted: imports=%v error=%v", f.imports, err)
	}
	if !strings.Contains(strings.Join(f.commands, "\n"), "'rm' '-rf' '--'") {
		t.Fatal("failed transfer staging directory not removed")
	}
}

func TestClusterImagesCopyFailureAndUnexpectedTempPath(t *testing.T) {
	for _, test := range []struct{ name, temp, want string }{
		{"copy", "", "copy missing image archive"},
		{"malicious-temp", "/tmp/; rm -rf /\n", "unexpected path"},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := newClusterImageFake(t)
			f.copyBad, f.tempResponse = true, test.temp
			err := f.prepare(false)
			if err == nil || !strings.Contains(err.Error(), test.want) || len(f.imports) > 0 {
				t.Fatalf("unexpected transfer failure: %v", err)
			}
			if test.temp != "" {
				for _, command := range f.commands {
					if strings.HasPrefix(command, "scp ") || strings.Contains(command, "'rm'") {
						t.Fatalf("untrusted temporary path used for copy/cleanup: %s", command)
					}
				}
			}
		})
	}
}

func TestClusterImagesTimeoutBeforeMutation(t *testing.T) {
	f := newClusterImageFake(t)
	f.block = "product_uuid"
	err := PrepareClusterImages(context.Background(), f.run, ImageOptions{BundleDir: f.bundle.dir, Runtime: "auto", CommandTimeout: time.Millisecond}, f.nodes, ClusterSSHOptions{}, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("SSH timeout was not preserved: %v", err)
	}
	assertClusterNoMutation(t, f.commands)
}

func TestClusterSSHIdentityCanMatchInterfaceIP(t *testing.T) {
	for _, test := range []struct {
		name, output string
		ok           bool
	}{
		{"ipv4", "UUID unknown\n2: eth0 inet 192.0.2.11/24 brd 192.0.2.255 scope global\n", true},
		{"ipv6", "UUID unknown\n2: eth0 inet6 2001:db8::11/64 scope global\n", false},
		{"hostname-only", "worker-1\n", false},
		{"ssh-only", "", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			run := func(context.Context, string, ...string) ([]byte, error) { return []byte(test.output), nil }
			worker := clusterImageWorker{run: run, host: "ssh-worker.example", opts: ImageOptions{CommandTimeout: time.Second}}
			err := worker.verifyIdentity(context.Background(), PreflightResult{NodeName: "worker-1", NodeInternalIP: "192.0.2.11", NodeSystemUUID: "expected"})
			if (err == nil) != test.ok {
				t.Fatalf("identity acceptance=%v expected=%v: %v", err == nil, test.ok, err)
			}
		})
	}
}

func TestClusterSSHValidation(t *testing.T) {
	nodes := []PreflightResult{{NodeName: "one", NodeInternalIP: "192.0.2.1"}}
	for _, test := range []struct {
		name string
		opts ClusterSSHOptions
	}{
		{"user option", ClusterSSHOptions{User: "-oProxyCommand=bad"}},
		{"user shell", ClusterSSHOptions{User: "root;echo"}},
		{"negative port", ClusterSSHOptions{Port: -1}},
		{"large port", ClusterSSHOptions{Port: 65536}},
		{"unknown node", ClusterSSHOptions{Hosts: map[string]string{"typo": "192.0.2.2"}}},
		{"host option", ClusterSSHOptions{Hosts: map[string]string{"one": "-oProxyCommand=bad"}}},
		{"host shell", ClusterSSHOptions{Hosts: map[string]string{"one": "example;echo"}}},
		{"host login", ClusterSSHOptions{Hosts: map[string]string{"one": "user@example.com"}}},
		{"empty host", ClusterSSHOptions{Hosts: map[string]string{"one": ""}}},
		{"identity missing", ClusterSSHOptions{IdentityFile: "/nonexistent/identity"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := validateClusterSSH(test.opts, nodes); err == nil {
				t.Fatal("unsafe SSH option accepted")
			}
		})
	}
	for _, host := range []string{"192.0.2.20", "2001:db8::1", "worker-2.example.com"} {
		if err := validateClusterSSH(ClusterSSHOptions{Hosts: map[string]string{"one": host}}, nodes); err != nil {
			t.Fatalf("valid host %q rejected: %v", host, err)
		}
	}
}

func TestClusterSSHArgumentsKeepShellSyntaxLiteral(t *testing.T) {
	sentinel := filepath.Join(t.TempDir(), "executed")
	value := "x'; touch " + sentinel + "; $(echo unsafe) & `echo unsafe`"
	output, err := exec.Command("sh", "-c", "printf '%s' "+clusterShellQuote(value)).Output()
	if err != nil || string(output) != value {
		t.Fatalf("shell quoting changed argv: %q %v", output, err)
	}
	if _, err := os.Stat(sentinel); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("shell syntax was executed")
	}
}

func TestClusterImagesImportDeadlineStillCleansRemoteStaging(t *testing.T) {
	f := newClusterImageFake(t)
	f.block = "scp "
	err := PrepareClusterImages(context.Background(), f.run, ImageOptions{BundleDir: f.bundle.dir, Runtime: "auto", ImportTimeout: time.Millisecond}, f.nodes, ClusterSSHOptions{}, nil)
	if !errors.Is(err, context.DeadlineExceeded) || len(f.imports) != 0 {
		t.Fatalf("copy deadline failed or imported an incomplete archive: %v", err)
	}
	if !strings.Contains(strings.Join(f.commands, "\n"), "'rm' '-rf' '--'") {
		t.Fatal("canceled transfer left its staging directory")
	}
}

func TestClusterImagesLocalCorruptionFailsBeforeCopy(t *testing.T) {
	f := newClusterImageFake(t)
	if err := os.WriteFile(f.bundle.paths[1], []byte("corrupt archive"), 0600); err != nil {
		t.Fatal(err)
	}
	err := f.prepare(false)
	if err == nil || !strings.Contains(err.Error(), "SHA256 mismatch") {
		t.Fatalf("corrupt local archive accepted: %v", err)
	}
	assertClusterNoMutation(t, f.commands)
}

func TestClusterSSHTransferUsesIPv6ScpDestinationAndExplicitIdentity(t *testing.T) {
	f := newClusterImageFake(t)
	identity := filepath.Join(t.TempDir(), "id ed25519")
	if err := os.WriteFile(identity, []byte("unused fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	worker := clusterImageWorker{run: f.run, host: "2001:db8::2", ssh: ClusterSSHOptions{User: "deploy", Port: 2222, IdentityFile: identity}, opts: ImageOptions{CommandTimeout: time.Second}, progress: func(string) {}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := worker.transfer(ctx, f.bundle.paths[0], f.bundle.manifest.Archives[0].SHA256); err != nil {
		t.Fatal(err)
	}
	if len(f.copied) != 1 || !strings.Contains(f.copied[0], " -> [2001:db8::2]:/tmp/") {
		t.Fatalf("invalid IPv6 scp destination: %v", f.copied)
	}
	all := strings.Join(f.commands, "\n")
	if !strings.Contains(all, "-B -P 2222 -o User=deploy -i "+identity) {
		t.Fatalf("scp does not preserve port/user/identity options: %s", all)
	}
}
