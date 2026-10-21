package lakehouse

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// Keep this identical to PROBE_CODE in the pinned obdiag Kubernetes transport.
// Log download always requires Python, even when native tar/gzip can pack logs.
// The archive and transport roundtrip stay in memory and do not write Pod files.
const observerDiagnosticPythonProbe = `import base64,gzip,hashlib,io,json,os,stat,sys,tarfile
assert sys.version_info >= (3,6)
payload=b"lakehouse-diagnostic-probe"
buffer=io.BytesIO()
with tarfile.open(fileobj=buffer,mode="w:gz") as archive:
    member=tarfile.TarInfo("probe")
    member.size=len(payload)
    archive.addfile(member,io.BytesIO(payload))
raw=buffer.getvalue()
assert gzip.decompress(raw)
with tarfile.open(fileobj=io.BytesIO(base64.b64decode(base64.b64encode(raw))),mode="r:gz") as archive:
    actual=archive.extractfile("probe").read()
assert hashlib.sha256(actual).digest()==hashlib.sha256(payload).digest()
`

const observerDiagnosticProbeShell = `export PYTHONDONTWRITEBYTECODE=1
probe=$1; shift; for py do
if command -v "$py" >/dev/null 2>&1 && "$py" -c "$probe" >/dev/null 2>&1; then exit 0; fi
done; exit 1`

// verifyObserverDiagnostics checks every Ready Observer running the configured
// image. It does not collect logs, create jobs, or modify database Pod files.
func verifyObserverDiagnostics(ctx context.Context, run RunCommand, namespace, release string, values map[string]interface{}, interval time.Duration) error {
	image := pfString(pfGet(values, "observer", "image"))
	expected := 0
	zones, _ := pfGet(values, "observer", "topology").([]interface{})
	for _, zone := range zones {
		n, err := strconv.Atoi(pfString(pfMap(zone)["replica"]))
		if err != nil || n < 1 || n > 100 {
			return errors.New("cannot verify diagnostics: invalid Observer topology")
		}
		expected += n
	}
	if expected == 0 || image == "" || interval <= 0 {
		return errors.New("cannot verify diagnostics: missing Observer image/topology or invalid interval")
	}
	for {
		var pods corev1.PodList
		if err := getJSON(ctx, run, &pods, "-n", namespace, "get", "pods", "-l", "ref-obcluster="+release+"-ob", "-o", "json"); err != nil {
			return err
		}
		ready := len(pods.Items) == expected
		for _, pod := range pods.Items {
			configured, running := false, false
			for _, c := range pod.Spec.Containers {
				if c.Name == "observer" && sameImageReference(c.Image, image) {
					configured = true
				}
			}
			for _, c := range pod.Status.ContainerStatuses {
				if c.Name == "observer" && c.Ready && c.State.Running != nil && sameImageReference(c.Image, image) {
					running = true
				}
			}
			ready = ready && configured && running && pod.DeletionTimestamp == nil && pod.Status.Phase == corev1.PodRunning
		}
		if ready {
			for _, pod := range pods.Items {
				_, err := bounded(ctx, run, "kubectl", "--request-timeout=20s", "-n", namespace, "exec", pod.Name, "-c", "observer", "--", "/bin/sh", "-c", observerDiagnosticProbeShell, "observer-diagnostic-probe", observerDiagnosticPythonProbe, "/usr/libexec/platform-python", "python3")
				if err != nil {
					return fmt.Errorf("Observer %s cannot execute the Python >=3.6 in-memory archive/transfer probe required by log download (/usr/libexec/platform-python or python3); retain the installed Observer image and check diagnostic compatibility: %w", pod.Name, err)
				}
			}
			return nil
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("waiting for %d Ready Observer Pods running the configured image: %w", expected, ctx.Err())
		case <-timer.C:
		}
	}
}

func sameImageReference(a, b string) bool {
	return strings.TrimPrefix(a, "docker.io/") == strings.TrimPrefix(b, "docker.io/")
}
