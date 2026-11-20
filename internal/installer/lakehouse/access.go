package lakehouse

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func showAccess(ctx context.Context, run RunCommand, o options, result PreflightResult, stateDir string, out io.Writer) error {
	return showAccessWithDiscovery(ctx, run, o, result, stateDir, out, discoverDashboardHost)
}

type dashboardHostDiscovery func(context.Context, PreflightResult) (string, string)

func dashboardHost(ctx context.Context, o options, result PreflightResult, stateDir string, discover dashboardHostDiscovery) (string, string, error) {
	if o.host != "" {
		return o.host, "--dashboard-host", nil
	}
	saved, err := os.ReadFile(filepath.Join(stateDir, "dashboard-host"))
	if err == nil {
		host := strings.TrimSpace(string(saved))
		if _, err := parseOptions([]string{"--dashboard-host", host}, io.Discard); err != nil || host == "" {
			return "", "", errors.New("saved Dashboard host is invalid; rerun with --dashboard-host <public-IP-or-DNS>")
		}
		return host, "saved --dashboard-host", nil
	}
	if !os.IsNotExist(err) {
		return "", "", fmt.Errorf("cannot read saved Dashboard host: %w", err)
	}
	if isPublicDashboardIP(result.NodeExternalIP) {
		return result.NodeExternalIP, "Kubernetes Node ExternalIP", nil
	}
	if isPublicDashboardIP(result.NodeInternalIP) {
		return result.NodeInternalIP, "public Kubernetes node address", nil
	}
	host, source := discover(ctx, result)
	if isPublicDashboardIP(host) {
		return host, source, nil
	}
	return "", "", nil
}

func showAccessWithDiscovery(ctx context.Context, run RunCommand, o options, result PreflightResult, stateDir string, out io.Writer, discover dashboardHostDiscovery) error {
	var svc struct {
		Spec struct {
			ClusterIP string                           `json:"clusterIP"`
			Type      string                           `json:"type"`
			Ports     []struct{ Port, NodePort int32 } `json:"ports"`
		} `json:"spec"`
	}
	if err := getJSON(ctx, run, &svc, "-n", o.namespace, "get", "service", "oceanbase-dashboard-"+o.release, "-o", "json"); err != nil {
		return fmt.Errorf("Dashboard service unavailable: %w", err)
	}
	if len(svc.Spec.Ports) == 0 {
		return errors.New("Dashboard Service has no ports")
	}
	if result.NodeInternalIP == "" {
		var pods struct {
			Items []struct {
				Spec struct {
					NodeName string `json:"nodeName"`
				} `json:"spec"`
			} `json:"items"`
		}
		if err := getJSON(ctx, run, &pods, "-n", o.namespace, "get", "pods", "-l", "app=oceanbase-dashboard-"+o.release, "-o", "json"); err != nil {
			return err
		}
		if len(pods.Items) == 0 || pods.Items[0].Spec.NodeName == "" {
			return errors.New("Dashboard Pod has no assigned node")
		}
		var node struct {
			Status struct {
				Addresses []struct{ Type, Address string } `json:"addresses"`
				NodeInfo  struct{ SystemUUID string }      `json:"nodeInfo"`
			} `json:"status"`
		}
		if err := getJSON(ctx, run, &node, "get", "node", pods.Items[0].Spec.NodeName, "-o", "json"); err != nil {
			return err
		}
		result.NodeName = pods.Items[0].Spec.NodeName
		result.NodeSystemUUID = node.Status.NodeInfo.SystemUUID
		for _, a := range node.Status.Addresses {
			if a.Type == "InternalIP" {
				result.NodeInternalIP = a.Address
			}
			if a.Type == "ExternalIP" {
				result.NodeExternalIP = a.Address
			}
		}
	}
	port := svc.Spec.Ports[0].NodePort
	verifyHost := result.NodeInternalIP
	if svc.Spec.Type == "ClusterIP" {
		verifyHost = svc.Spec.ClusterIP
		port = svc.Spec.Ports[0].Port
	}
	if verifyHost == "" || port == 0 {
		return errors.New("Dashboard address is not yet allocated")
	}
	verifiedURL := "http://" + net.JoinHostPort(verifyHost, strconv.Itoa(int(port)))
	client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	for _, path := range []string{"/", "/api/v1/info"} {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, verifiedURL+path, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Dashboard is deployed but HTTP check of %s failed; rerun --show-access after checking node networking: %w", verifiedURL, err)
		}
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Dashboard HTTP readiness failed: %s returned %d", path, resp.StatusCode)
		}
	}
	var host, source string
	if svc.Spec.Type != "ClusterIP" {
		var err error
		host, source, err = dashboardHost(ctx, o, result, stateDir, discover)
		if err != nil {
			return err
		}
	}
	var secret struct {
		Data map[string]string `json:"data"`
	}
	if err := getJSON(ctx, run, &secret, "-n", o.namespace, "get", "secret", o.release+"-dashboard-bootstrap", "-o", "json"); err != nil {
		return err
	}
	password, err := base64.StdEncoding.DecodeString(secret.Data["password"])
	if err != nil || len(password) == 0 {
		return errors.New("Dashboard bootstrap password is missing or malformed")
	}
	passwordPath := filepath.Join(stateDir, "dashboard-initial-password")
	if err = writePrivate(passwordPath, password); err != nil {
		return err
	}
	// Only persist an explicit choice. Automatically discovered addresses can
	// change when a node's public IP/EIP is reassigned.
	if o.host != "" && svc.Spec.Type != "ClusterIP" {
		if err = writePrivate(filepath.Join(stateDir, "dashboard-host"), []byte(o.host+"\n")); err != nil {
			return err
		}
	}
	heading := "Installation complete."
	if o.showAccess {
		heading = "Dashboard access"
	}
	summary := fmt.Sprintf("\n%s\nRelease: %s\nNamespace: %s\n", heading, o.release, o.namespace)
	if svc.Spec.Type == "ClusterIP" {
		summary += fmt.Sprintf("Dashboard (cluster network only): %s/#/login\nAccess: ClusterIP; reachable only within the cluster network.\nFor browser access, run on a computer with kubectl access:\n  kubectl -n %s port-forward service/oceanbase-dashboard-%s 18080:%d\nThen open http://127.0.0.1:18080/#/login on that computer.\n", verifiedURL, o.namespace, o.release, port)
	} else {
		if host != "" {
			url := "http://" + net.JoinHostPort(host, strconv.Itoa(int(port)))
			summary += fmt.Sprintf("Dashboard: %s/#/login\nAddress source: %s\n", url, source)
		} else {
			summary += fmt.Sprintf("External Dashboard address: unavailable (could not be detected automatically).\nDashboard (private network only): %s/#/login\nSet your public IP or DNS name without reinstalling:\n  ./install.sh --show-access --release %s --namespace %s --dashboard-host <public-IP-or-DNS>\n", verifiedURL, o.release, o.namespace)
		}
		summary += fmt.Sprintf("Access: NodePort; allow TCP port %d from your browser's network in the firewall/security group.\n", port)
	}
	summary += fmt.Sprintf("Username: admin\nInitial password file: %s (mode 0600)\nHTTP readiness: passed (node network).\n\nChange the initial password when prompted on first login. If already changed, use your current password.\n", passwordPath)
	summary += "Lakehouse and Dashboard share one Helm release. Rerun the same installation command to resume; existing data and accounts are retained.\n"
	if err = writePrivate(filepath.Join(stateDir, "ACCESS.txt"), []byte(summary)); err != nil {
		return err
	}
	fmt.Fprint(out, summary)
	return nil
}

func prepareOperator(ctx context.Context, run RunCommand, o options, result PreflightResult, progress Progress) error {
	if result.NodeCount != 1 {
		return errors.New("automatic prerequisite installation supports a single-node cluster; prepare an SS-capable Operator on this multi-node cluster first")
	}
	var certCRDs struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Status struct {
				Conditions []struct{ Type, Status string } `json:"conditions"`
			} `json:"status"`
		} `json:"items"`
	}
	if err := getJSON(ctx, run, &certCRDs, "get", "crds", "-o", "json"); err != nil {
		return err
	}
	certCount := 0
	established := map[string]bool{}
	for _, crd := range certCRDs.Items {
		if strings.HasSuffix(crd.Metadata.Name, ".cert-manager.io") || strings.HasSuffix(crd.Metadata.Name, ".acme.cert-manager.io") {
			certCount++
			for _, condition := range crd.Status.Conditions {
				if condition.Type == "Established" && condition.Status == "True" {
					established[crd.Metadata.Name] = true
				}
			}
		}
	}
	requiredCRDs := []string{"certificates.cert-manager.io", "certificaterequests.cert-manager.io", "issuers.cert-manager.io", "clusterissuers.cert-manager.io", "challenges.acme.cert-manager.io", "orders.acme.cert-manager.io"}
	if certCount > 0 {
		for _, name := range requiredCRDs {
			if !established[name] {
				return fmt.Errorf("existing cert-manager is incomplete: %s is missing or not Established; repair it before continuing", name)
			}
		}
	}
	action := func(label, tool string, args ...string) error {
		progress(label)
		c, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		err := withProgress(c, progressInterval, progress, label, func() error {
			_, err := run(c, tool, args...)
			return err
		})
		if err != nil {
			return fmt.Errorf("%s failed: %w; repair the partially initialized prerequisite before retrying; existing shared components will not be overwritten", label, err)
		}
		return nil
	}
	if certCount == 0 {
		var deployments struct {
			Items []interface{} `json:"items"`
		}
		if err := getJSON(ctx, run, &deployments, "-n", "cert-manager", "get", "deployments", "-o", "json"); err != nil {
			return err
		}
		if len(deployments.Items) > 0 {
			return errors.New("partial cert-manager installation detected; repair existing components before continuing")
		}
		manifest := filepath.Join(o.bundle, "prerequisites/cert-manager-v1.12.4.yaml")
		if _, err := os.Stat(manifest); err != nil {
			return err
		}
		if err := action("Installing bundled cert-manager", "kubectl", "--request-timeout=60s", "apply", "-f", manifest); err != nil {
			return err
		}
	}
	for _, name := range []string{"cert-manager", "cert-manager-cainjector", "cert-manager-webhook"} {
		if err := action("Waiting for "+name+" readiness", "kubectl", "--request-timeout=30s", "-n", "cert-manager", "rollout", "status", "deployment/"+name, "--timeout=240s"); err != nil {
			return err
		}
	}
	if err := action("Installing bundled ob-operator", "helm", "install", "ob-operator", filepath.Join(o.bundle, "charts/ob-operator"), "--namespace", "oceanbase-system", "--create-namespace", "--wait", "--timeout", "240s"); err != nil {
		return err
	}
	return nil
}
