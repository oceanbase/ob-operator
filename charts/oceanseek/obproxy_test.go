/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package oceanseek_test

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestOBProxyDefaultAndOptOut(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		values := map[string]interface{}{}
		if !enabled {
			values["obproxy"] = map[string]interface{}{"enabled": false}
		}
		docs, output, err := render(t, values)
		if err != nil {
			t.Fatal(err, output)
		}
		var sql corev1.Service
		decodeScheduledObject(t, find(t, docs, "Service", "qa-sql"), &sql)
		if sql.Spec.Ports[0].Port != 2883 {
			t.Fatal("default SQL Service entry must use port 2883", sql.Spec.Ports)
		}
		if enabled {
			if sql.Spec.Ports[0].TargetPort.IntVal != 2883 || sql.Spec.Selector["app.kubernetes.io/component"] != "obproxy" {
				t.Fatal(sql.Spec)
			}
			var direct corev1.Service
			decodeScheduledObject(t, find(t, docs, "Service", "qa-observer"), &direct)
			if direct.Spec.Selector["ref-obcluster"] != "qa-ob" || direct.Spec.Type != "ClusterIP" {
				t.Fatal(direct.Spec)
			}
			var deploy appsv1.Deployment
			decodeScheduledObject(t, find(t, docs, "Deployment", "qa-obproxy"), &deploy)
			if *deploy.Spec.Replicas != 2 || deploy.Spec.Template.Spec.AutomountServiceAccountToken == nil || *deploy.Spec.Template.Spec.AutomountServiceAccountToken {
				t.Fatal(deploy.Spec)
			}
			c := deploy.Spec.Template.Spec.Containers[0]
			init := deploy.Spec.Template.Spec.InitContainers
			if len(init) != 1 || init[0].Name != "discover-observers" || !strings.Contains(strings.Join(init[0].Args, " "), "--obproxy-rslist-file=/bootstrap/rslist") {
				t.Fatal("missing advertised Observer seed discovery")
			}
			if strings.Contains(c.Args[0], "getent") || strings.Contains(c.Args[0], "$address:2881") {
				t.Fatal("load-balanced bootstrap VIP must not be used as a rootserver identity")
			}
			for _, mount := range c.VolumeMounts {
				if mount.Name == "discovery-token" {
					t.Fatal("proxy process must not receive Kubernetes credentials")
				}
			}

			if c.ReadinessProbe == nil || c.StartupProbe == nil || c.LivenessProbe == nil {
				t.Fatal("missing health probes")
			}
			if !strings.Contains(c.Args[0], "exec /home/admin/obproxy/bin/obproxy") || strings.Contains(c.Args[0], "echo") {
				t.Fatal("unexpected startup or credential logging")
			}
			if deploy.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
				t.Fatal("missing node spreading")
			}
			budget := find(t, docs, "PodDisruptionBudget", "qa-obproxy")
			if field(budget, "spec", "maxUnavailable") != 1 {
				t.Fatal("default OBProxy budget must allow one unavailable replica")
			}
			for _, name := range []string{"qa-proxyro", "qa-proxysys"} {
				secret := find(t, docs, "Secret", name)
				if field(secret, "metadata", "annotations", "helm.sh/hook") != "pre-install,pre-upgrade" || field(secret, "metadata", "annotations", "helm.sh/resource-policy") != "keep" {
					t.Fatal(secret)
				}
			}
			if field(find(t, docs, "OBCluster", "qa-ob"), "spec", "userSecrets", "proxyro") != "qa-proxyro" {
				t.Fatal("proxy account is not wired")
			}
		} else {
			if sql.Spec.Ports[0].TargetPort.IntVal != 2881 || sql.Spec.Selector["ref-obcluster"] != "qa-ob" {
				t.Fatal(sql.Spec)
			}
			for _, doc := range docs {
				name, _ := field(doc, "metadata", "name").(string)
				if name == "qa-obproxy" || name == "qa-observer" || name == "qa-proxyro" || name == "qa-proxysys" {
					t.Fatalf("disabled proxy rendered %s", name)
				}
			}
		}
	}
}

func TestOBProxyAbsentBeforeClusterAndSingleReplica(t *testing.T) {
	for _, stage := range []string{"storage", "logservice"} {
		docs, output, err := render(t, map[string]interface{}{"stage": stage})
		if err != nil {
			t.Fatal(err, output)
		}
		for _, doc := range docs {
			if field(doc, "metadata", "name") == "qa-obproxy" {
				t.Fatal("proxy rendered before cluster stage")
			}
		}
	}
	docs, output, err := render(t, map[string]interface{}{"obproxy": map[string]interface{}{"replicas": 1}})
	if err != nil {
		t.Fatal(err, output)
	}
	for _, doc := range docs {
		if doc["kind"] == "PodDisruptionBudget" && field(doc, "metadata", "name") == "qa-obproxy" {
			t.Fatal("single instance PDB blocks maintenance")
		}
	}
}

func TestOBProxyReusesExistingAccountAndDoesNotReseed(t *testing.T) {
	api, config := retainedCredentials(t)
	api.crs = map[string]interface{}{
		"/apis/oceanbase.oceanbase.com/v1alpha1/namespaces/qa-ns/obclusters/qa-ob": map[string]interface{}{
			"apiVersion": "oceanbase.oceanbase.com/v1alpha1", "kind": "OBCluster",
			"metadata": map[string]string{"name": "qa-ob", "namespace": "qa-ns"},
			"spec":     map[string]interface{}{"userSecrets": map[string]string{"proxyro": "operator-proxyro"}, "sharedStorageInfo": map[string]string{"bucketURL": "s3://qa-data?host=http://192.0.2.77:9000&s3_region=us-east-1"}},
		},
	}
	encodedPassword := base64.StdEncoding.EncodeToString([]byte(rand.Text()))
	api.secrets["operator-proxyro"] = map[string]string{"password": encodedPassword}
	api.secrets["qa-proxysys"] = map[string]string{"password": encodedPassword}
	for _, enabled := range []bool{true, false, true} {
		output, err := renderRetained(t, config, map[string]interface{}{"obproxy": map[string]interface{}{"enabled": enabled}, "dashboard": map[string]interface{}{"enabled": false}})
		if err != nil {
			t.Fatal(err, output)
		}
		if !strings.Contains(output, "proxyro: operator-proxyro") {
			t.Fatal("existing account identity changed")
		}
		if strings.Contains(output, "name: qa-proxyro\n") || strings.Contains(output, "password: cmV0YWluZWQ=") {
			t.Fatal("existing credentials adopted or replayed")
		}
	}
	api.mu.Lock()
	delete(api.secrets, "operator-proxyro")
	api.mu.Unlock()
	output, err := renderRetained(t, config, map[string]interface{}{"dashboard": map[string]interface{}{"enabled": false}})
	if err == nil || !strings.Contains(output, "proxyro Secret is missing") {
		t.Fatal("missing existing credentials must fail without reset", err, output)
	}
}

func TestOBProxyRejectsInvalidConfiguration(t *testing.T) {
	for _, values := range []map[string]interface{}{
		{"enabled": "yes"}, {"replicas": 0}, {"image": ""}, {"memoryLimit": "invalid"},
	} {
		_, _, err := render(t, map[string]interface{}{"obproxy": values})
		if err == nil {
			t.Fatalf("invalid OBProxy configuration accepted: %v", values)
		}
	}
}

func TestOBProxyMultipleReplicasKeepDisruptionBudget(t *testing.T) {
	docs, output, err := render(t, map[string]interface{}{"obproxy": map[string]interface{}{"replicas": 2}})
	if err != nil {
		t.Fatal(err, output)
	}
	find(t, docs, "PodDisruptionBudget", "qa-obproxy")
}
