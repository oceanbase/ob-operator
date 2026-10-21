package lakehouse

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"sigs.k8s.io/yaml"
)

type valueFiles []string

func (v *valueFiles) String() string     { return strings.Join(*v, ",") }
func (v *valueFiles) Set(s string) error { *v = append(*v, s); return nil }

type options struct {
	accessExplicit                                                             bool
	bundle, release, namespace, node, runtime, address, endpoint, host, access string
	check, showAccess                                                          bool
	files                                                                      valueFiles
}

func parseOptions(args []string, stderr io.Writer) (options, error) {
	var o options
	fs := flag.NewFlagSet("lakehouse-installer", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&o.bundle, "bundle", ".", "Extracted bundle directory")
	fs.StringVar(&o.release, "release", "ob-demo", "Helm release (reuse on retry)")
	fs.StringVar(&o.namespace, "namespace", "ob-helm-demo", "Namespace (reuse on retry)")
	fs.StringVar(&o.node, "node", "", "Target worker; automatic when unambiguous")
	fs.StringVar(&o.runtime, "runtime", "auto", "auto, k3s, docker or containerd")
	fs.StringVar(&o.address, "containerd-address", "", "containerd socket for non-K3s installations")
	fs.StringVar(&o.endpoint, "cri-endpoint", "", "CRI Unix socket for non-K3s installations")
	fs.StringVar(&o.host, "dashboard-host", "", "Advertised Dashboard IP or DNS name, e.g. a server public IP")
	fs.StringVar(&o.access, "access", "nodeport", "Dashboard access: nodeport or clusterip")
	fs.BoolVar(&o.check, "check", false, "Read-only preflight; do not import images or install resources")
	fs.BoolVar(&o.showAccess, "show-access", false, "Show URL and save initial password to a private local file")
	fs.Var(&o.files, "f", "Values file; repeatable, later files override earlier ones")
	fs.Var(&o.files, "values", "Values file; repeatable")
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "access" {
			o.accessExplicit = true
		}
	})
	if o.check && o.showAccess {
		return o, errors.New("--check and --show-access are mutually exclusive")
	}
	if fs.NArg() != 0 {
		return o, errors.New("unexpected positional arguments; use --release, --namespace and -f")
	}
	if !regexp.MustCompile(`^[a-z]([a-z0-9-]{0,38}[a-z0-9])?$`).MatchString(o.release) {
		return o, errors.New("release must be 1-40 DNS characters")
	}
	if !regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`).MatchString(o.namespace) {
		return o, errors.New("invalid namespace")
	}
	if o.access != "nodeport" && o.access != "clusterip" {
		return o, errors.New("access must be nodeport or clusterip")
	}
	if o.host != "" && net.ParseIP(o.host) == nil && !regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?$`).MatchString(o.host) {
		return o, errors.New("dashboard-host must be an IP or DNS name without scheme or port")
	}
	var err error
	o.bundle, err = filepath.Abs(o.bundle)
	return o, err
}

func command(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = 5 * time.Second
	return cmd
}
func execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	return command(ctx, name, args...).Output()
}
func bounded(ctx context.Context, run RunCommand, name string, args ...string) ([]byte, error) {
	c, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return run(c, name, args...)
}
func getJSON(ctx context.Context, run RunCommand, dst interface{}, args ...string) error {
	raw, err := bounded(ctx, run, "kubectl", append([]string{"--request-timeout=20s"}, args...)...)
	if err != nil {
		return fmt.Errorf("Kubernetes read failed: %w", err)
	}
	if err = json.Unmarshal(raw, dst); err != nil {
		return errors.New("Kubernetes returned invalid JSON")
	}
	return nil
}

func Run(ctx context.Context, args []string, out, stderr io.Writer) error {
	return runWithCacheDirectory(ctx, args, out, stderr, os.UserCacheDir)
}

func runWithCacheDirectory(ctx context.Context, args []string, out, stderr io.Writer, userCacheDir func() (string, error)) error {
	o, err := parseOptions(args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, tool := range []string{"helm", "kubectl", "jq", "bash"} {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("install prerequisite %s before running this command", tool)
		}
	}
	progress := func(s string) { fmt.Fprintln(out, s) }
	progress("[1/5] 检查集群与已安装配置")
	var identity struct {
		Metadata struct {
			UID string `json:"uid"`
		} `json:"metadata"`
	}
	if err = getJSON(ctx, execute, &identity, "get", "namespace", "kube-system", "-o", "json"); err != nil {
		return err
	}
	if identity.Metadata.UID == "" {
		return errors.New("cluster identity is missing")
	}
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(identity.Metadata.UID+"/"+o.namespace+"/"+o.release)))[:16]
	stateDir := filepath.Join(o.bundle, ".lakehouse-install", key)
	if o.showAccess {
		return showAccess(ctx, execute, o, PreflightResult{}, stateDir, out)
	}
	if !o.check {
		cacheDir, err := userCacheDir()
		if err != nil {
			return fmt.Errorf("cannot locate installation lock directory: %w", err)
		}
		cacheDir = filepath.Join(cacheDir, "oceanbase-lakehouse")
		if err = privateDirectory(cacheDir); err != nil {
			return err
		}
		lockDir := filepath.Join(cacheDir, "locks")
		if err = privateDirectory(lockDir); err != nil {
			return err
		}
		// Share the lock across extracted bundles for this user and release. Helm
		// additionally serializes server mutations. Never remove a held lock file.
		lock, err := os.OpenFile(filepath.Join(lockDir, key), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
		if err != nil {
			return err
		}
		defer lock.Close()
		if err = lock.Chmod(0600); err != nil {
			return err
		}
		if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
			return errors.New("another installation of this release is running")
		}
		defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	}
	var releases []struct {
		Name string `json:"name"`
	}
	raw, err := bounded(ctx, execute, "helm", "list", "-n", o.namespace, "--all", "--filter", "^"+o.release+"$", "-o", "json")
	if err != nil {
		return fmt.Errorf("cannot read installed Helm release: %w", err)
	}
	if err = json.Unmarshal(raw, &releases); err != nil {
		return errors.New("invalid Helm release list")
	}
	existing := len(releases) > 0
	var values map[string]interface{}
	if existing {
		raw, err = bounded(ctx, execute, "helm", "get", "values", o.release, "-n", o.namespace, "-o", "json")
		if err != nil {
			return fmt.Errorf("cannot read existing user values: %w", err)
		}
		if err = json.Unmarshal(raw, &values); err != nil {
			return errors.New("invalid existing user values")
		}
		effectiveRaw, readErr := bounded(ctx, execute, "helm", "get", "values", o.release, "-n", o.namespace, "--all", "-o", "json")
		if readErr != nil {
			return fmt.Errorf("cannot read effective installed values: %w", readErr)
		}
		var effective map[string]interface{}
		if err = json.Unmarshal(effectiveRaw, &effective); err != nil {
			return errors.New("invalid effective installed values")
		}
		values = retainConfiguration(effective, values)
		progress("已读取当前 release 的用户参数；新传入的 values 优先。")
	} else {
		values, err = readValues(filepath.Join(stateDir, "values.yaml"))
		if os.IsNotExist(err) {
			values, err = initialValues(o.bundle)
		}
		if err != nil {
			return err
		}
	}
	if values == nil {
		values = map[string]interface{}{}
	}
	for _, path := range o.files {
		overlay, err := readValues(path)
		if err != nil {
			return err
		}
		values = mergeValues(values, overlay)
	}
	if o.accessExplicit || (!existing && pfGet(values, "dashboard", "service", "type") == nil) {
		if o.access == "nodeport" {
			setValue(values, "NodePort", "dashboard", "service", "type")
		} else {
			setValue(values, "ClusterIP", "dashboard", "service", "type")
		}
	}
	defaults, err := defaultValues(o.bundle)
	if err != nil {
		return err
	}
	complete := mergeValues(defaults, values)
	// Dashboard shares the placement of the lakehouse. Explicit user overrides
	// remain subject to the same single-worker scheduling validation.
	if pfGet(values, "dashboard", "nodeSelector") == nil {
		setValue(values, pfGet(complete, "scheduling", "nodeSelector"), "dashboard", "nodeSelector")
	}
	if pfGet(values, "dashboard", "tolerations") == nil {
		setValue(values, pfGet(complete, "scheduling", "tolerations"), "dashboard", "tolerations")
	}
	complete = mergeValues(defaults, values)
	result, err := CheckCluster(ctx, execute, PreflightOptions{Release: o.release, Namespace: o.namespace, NodeName: o.node, Values: complete, Existing: existing, PrepareOperator: true})
	if err != nil {
		return err
	}
	if err = checkLocalWorker(result); err != nil {
		return err
	}
	progress(fmt.Sprintf("目标节点 %s；CPU 需要 %.2f / 可用 %.2f 核，内存需要 %.1f / 可用 %.1f GiB", result.NodeName, float64(result.RequiredCPUMilli)/1000, float64(result.AvailableCPUMilli)/1000, float64(result.RequiredMemoryBytes)/(1<<30), float64(result.AvailableMemoryBytes)/(1<<30)))
	for _, warning := range result.Warnings {
		progress(warning)
	}
	if err = validateConfiguredImages(o.bundle, complete); err != nil {
		return err
	}
	progress("[2/5] 校验镜像及缓存（已匹配的镜像自动跳过）")
	if err = PrepareImages(ctx, execute, ImageOptions{BundleDir: o.bundle, Runtime: o.runtime, NodeRuntimeVersion: result.NodeRuntimeVersion, ContainerdAddress: o.address, CRIEndpoint: o.endpoint, CheckOnly: o.check}, progress); err != nil {
		return err
	}
	if o.check {
		progress("预检查通过；没有导入镜像或修改 Kubernetes 资源。")
		return nil
	}
	// Use the node's actual hostname label, not its metadata.name.
	var node struct {
		Metadata struct {
			Labels map[string]string `json:"labels"`
		} `json:"metadata"`
	}
	if err = getJSON(ctx, execute, &node, "get", "node", result.NodeName, "-o", "json"); err != nil {
		return err
	}
	hostname := node.Metadata.Labels["kubernetes.io/hostname"]
	if hostname == "" {
		return errors.New("target node has no hostname label")
	}
	if !existing {
		setValue(values, hostname, "scheduling", "nodeSelector", "kubernetes.io/hostname")
	}
	setValue(values, hostname, "dashboard", "nodeSelector", "kubernetes.io/hostname")
	setValue(values, result.StorageClass, "storageClass")
	raw, err = yaml.Marshal(values)
	if err != nil {
		return err
	}

	if err = writePrivate(filepath.Join(stateDir, "values.yaml"), raw); err != nil {
		return err
	}
	progress("[3/5] 检查并准备 Operator（已有组件保留）")
	if result.OperatorMissing {
		if err = prepareOperator(ctx, execute, o, result, progress); err != nil {
			return err
		}
		_, err = CheckCluster(ctx, execute, PreflightOptions{Release: o.release, Namespace: o.namespace, NodeName: result.NodeName, Values: mergeValues(defaults, values), Existing: existing})
		if err != nil {
			return err
		}
	}
	progress("[4/5] 安装湖库和 Dashboard，按阶段等待就绪")
	installCtx, cancel := context.WithTimeout(ctx, 2*time.Hour)
	defer cancel()
	cmd := command(installCtx, "bash", filepath.Join(o.bundle, "charts/oceanbase-lakehouse/install.sh"), o.release, o.namespace, "-f", filepath.Join(stateDir, "values.yaml"))
	cmd.Env = append(os.Environ(), "LAKEHOUSE_MANAGED_ACCESS=1")
	cmd.Stdout = out
	cmd.Stderr = stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("安装未完成，处理提示后重跑同一命令（不自动删除数据）: %w", err)
	}
	if enabled, _ := pfGet(complete, "dashboard", "enabled").(bool); !enabled {
		progress("安装完成；已按显式配置关闭内置 Dashboard。")
		return nil
	}
	progress("[5/5] 验证日志采集依赖、Dashboard 并生成访问说明")
	diagnosticCtx, diagnosticCancel := context.WithTimeout(ctx, 15*time.Minute)
	defer diagnosticCancel()
	if err = verifyObserverDiagnostics(diagnosticCtx, execute, o.namespace, o.release, complete, 3*time.Second); err != nil {
		return err
	}
	progress("Observer 日志采集所需的 tar/gzip 已验证；诊断镜像身份已校验。")
	return showAccess(ctx, execute, o, result, stateDir, out)
}

func checkLocalWorker(result PreflightResult) error {
	uuid, err := os.ReadFile("/sys/class/dmi/id/product_uuid")
	if err == nil && result.NodeSystemUUID != "" && strings.EqualFold(strings.TrimSpace(string(uuid)), result.NodeSystemUUID) {
		return nil
	}
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	for _, a := range addresses {
		ip, _, err := net.ParseCIDR(a.String())
		if err == nil && !ip.IsLoopback() && result.NodeInternalIP != "" && ip.Equal(net.ParseIP(result.NodeInternalIP)) {
			return nil
		}
	}
	return fmt.Errorf("run this bundle on target worker %s: this host's UUID/IP does not match; importing into a different host would not make images available to Pods", result.NodeName)
}
