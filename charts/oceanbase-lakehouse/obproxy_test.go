package lakehouse_test

import (
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
			if *deploy.Spec.Replicas != 1 || deploy.Spec.Template.Spec.AutomountServiceAccountToken == nil || *deploy.Spec.Template.Spec.AutomountServiceAccountToken {
				t.Fatal(deploy.Spec)
			}
			c := deploy.Spec.Template.Spec.Containers[0]
			if c.ReadinessProbe == nil || c.StartupProbe == nil || c.LivenessProbe == nil {
				t.Fatal("missing health probes")
			}
			if !strings.Contains(c.Args[0], "exec /home/admin/obproxy/bin/obproxy") || strings.Contains(c.Args[0], "echo") {
				t.Fatal("unexpected startup or credential logging")
			}
			if deploy.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
				t.Fatal("missing node spreading")
			}
			for _, doc := range docs {
				if doc["kind"] == "PodDisruptionBudget" && field(doc, "metadata", "name") == "qa-obproxy" {
					t.Fatal("default single proxy must not block voluntary maintenance")
				}
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
			"spec":     map[string]interface{}{"userSecrets": map[string]string{"proxyro": "operator-proxyro"}, "sharedStorageInfo": map[string]string{"bucketURL": "s3://qa-data?host=http://10.43.0.77:9000&s3_region=us-east-1"}},
		},
	}
	api.secrets["operator-proxyro"] = map[string]string{"password": "cmV0YWluZWQ="}
	api.secrets["qa-proxysys"] = map[string]string{"password": "cmV0YWluZWQ="}
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
