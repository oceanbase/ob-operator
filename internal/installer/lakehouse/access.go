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
			} `json:"status"`
		}
		if err := getJSON(ctx, run, &node, "get", "node", pods.Items[0].Spec.NodeName, "-o", "json"); err != nil {
			return err
		}
		for _, a := range node.Status.Addresses {
			if a.Type == "InternalIP" {
				result.NodeInternalIP = a.Address
			}
			if a.Type == "ExternalIP" {
				result.NodeExternalIP = a.Address
			}
		}
	}
	host := result.NodeExternalIP
	if host == "" {
		host = result.NodeInternalIP
	}
	if o.host != "" {
		host = o.host
	}
	port := svc.Spec.Ports[0].NodePort
	verifyHost := result.NodeInternalIP
	if svc.Spec.Type == "ClusterIP" {
		host = svc.Spec.ClusterIP
		verifyHost = host
		port = svc.Spec.Ports[0].Port
	}
	if host == "" || verifyHost == "" || port == 0 {
		return errors.New("Dashboard address is not yet allocated")
	}
	url := "http://" + net.JoinHostPort(host, strconv.Itoa(int(port)))
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
	summary := fmt.Sprintf("安装完成：release=%s，namespace=%s\nDashboard: %s/#/login\n用户名：admin\n初始密码文件：%s（权限 0600；首次登录后按页面提示修改密码，已改密请使用当前密码）\n已验证节点内 HTTP：%s\n", o.release, o.namespace, url, passwordPath, verifiedURL)
	if svc.Spec.Type == "ClusterIP" {
		summary += "当前选择了 clusterip 访问，仅集群网络内可达。\n"
	} else {
		summary += "浏览器所在网络需能访问上述节点端口；公网 IP 请使用 --dashboard-host 指定，云防火墙/安全组需允许该端口。\n"
	}
	summary += "湖库和 Dashboard 属于同一 Helm release。重复执行相同命令可继续安装，已有数据库和账号保留。\n"
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
		_, err := run(c, tool, args...)
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
		if err := action("安装配套 cert-manager", "kubectl", "--request-timeout=60s", "apply", "-f", manifest); err != nil {
			return err
		}
	}
	for _, name := range []string{"cert-manager", "cert-manager-cainjector", "cert-manager-webhook"} {
		if err := action("等待 "+name, "kubectl", "--request-timeout=30s", "-n", "cert-manager", "rollout", "status", "deployment/"+name, "--timeout=240s"); err != nil {
			return err
		}
	}
	if err := action("安装配套 ob-operator", "helm", "install", "ob-operator", filepath.Join(o.bundle, "charts/ob-operator"), "--namespace", "oceanbase-system", "--create-namespace", "--wait", "--timeout", "240s"); err != nil {
		return err
	}
	return nil
}
