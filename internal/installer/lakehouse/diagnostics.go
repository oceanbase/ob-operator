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

// verifyObserverDiagnostics waits for the configured image, not just an older
// Ready Pod. obdiag's Kubernetes transport executes tar inside each Observer.
// It does not collect logs, create jobs, or modify files in the database Pod.
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
				_, err := bounded(ctx, run, "kubectl", "--request-timeout=20s", "-n", namespace, "exec", pod.Name, "-c", "observer", "--", "/bin/sh", "-c", "command -v tar >/dev/null && command -v gzip >/dev/null && tar --version >/dev/null && gzip --version >/dev/null")
				if err != nil {
					return fmt.Errorf("Observer %s cannot execute tar/gzip required by log download; use a verified diagnostics-capable Observer image: %w", pod.Name, err)
				}
			}
			return nil
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("waiting for %d Observer Pods running the configured diagnostics-capable image: %w", expected, ctx.Err())
		case <-timer.C:
		}
	}
}

func sameImageReference(a, b string) bool {
	return strings.TrimPrefix(a, "docker.io/") == strings.TrimPrefix(b, "docker.io/")
}
