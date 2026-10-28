package lakehouse_test

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func assertUniqueResources(t *testing.T, docs []manifest) {
	t.Helper()
	seen := make(map[string]bool)
	for _, doc := range docs {
		key := fmt.Sprintf("%s/%s/%v/%s", doc["apiVersion"], doc["kind"], field(doc, "metadata", "namespace"), field(doc, "metadata", "name"))
		if seen[key] {
			t.Fatalf("duplicate resource %s", key)
		}
		seen[key] = true
	}
}

func dashboardSecretText(t *testing.T, doc manifest, key string) string {
	t.Helper()
	encoded, ok := field(doc, "data", key).(string)
	if !ok {
		t.Fatalf("Secret/%s has no encoded %s", field(doc, "metadata", "name"), key)
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("Secret/%s key %s is not base64", field(doc, "metadata", "name"), key)
	}
	return string(decoded)
}

func dashboardVolumeClaim(t *testing.T, deployment appsv1.Deployment, name string) string {
	t.Helper()
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == name && volume.PersistentVolumeClaim != nil {
			return volume.PersistentVolumeClaim.ClaimName
		}
	}
	t.Fatalf("missing persistent volume %s", name)
	return ""
}

func TestDashboardIncludedInEveryLakehouseStage(t *testing.T) {
	for _, stage := range []string{"storage", "logservice", "cluster"} {
		t.Run(stage, func(t *testing.T) {
			docs, output, err := render(t, map[string]interface{}{"stage": stage})
			if err != nil {
				t.Fatalf("render: %v\n%s", err, output)
			}
			assertUniqueResources(t, docs)
			var deployment appsv1.Deployment
			decodeScheduledObject(t, find(t, docs, "Deployment", "oceanbase-dashboard-qa"), &deployment)
			if deployment.Namespace != "qa-ns" || deployment.Spec.Template.Spec.ServiceAccountName != "qa-sa" {
				t.Fatal("Dashboard must use its release namespace and ServiceAccount")
			}
			containers := map[string]corev1.Container{}
			for _, container := range deployment.Spec.Template.Spec.Containers {
				containers[container.Name] = container
			}
			for _, name := range []string{"dashboard", "prometheus", "alertmanager"} {
				if containers[name].Image == "" {
					t.Fatalf("missing Dashboard bundle container %s", name)
				}
			}
			dashboard := containers["dashboard"]
			if dashboard.ReadinessProbe == nil || dashboard.ReadinessProbe.HTTPGet == nil ||
				dashboard.ReadinessProbe.HTTPGet.Path != "/api/v1/info" ||
				dashboard.ReadinessProbe.HTTPGet.Port.IntVal != 8080 {
				t.Fatal("Helm readiness must wait for the Dashboard HTTP endpoint")
			}
			env := map[string]string{}
			for _, entry := range dashboard.Env {
				env[entry.Name] = entry.Value
			}
			if env["USER_CREDENTIALS_SECRET"] != "qa-user-credentials" ||
				env["USER_NAMESPACE"] != "qa-ns" || env["RBAC_POLICY_CONFIG_MAP"] != "qa-rbac-policy" {
				t.Fatal("Dashboard is not wired to the parent-managed authentication resources")
			}
			service := find(t, docs, "Service", "oceanbase-dashboard-qa")
			var typedService corev1.Service
			decodeScheduledObject(t, service, &typedService)
			if typedService.Spec.Type != corev1.ServiceTypeClusterIP ||
				len(typedService.Spec.Ports) != 1 || typedService.Spec.Ports[0].Port != 80 ||
				typedService.Spec.Ports[0].TargetPort.IntVal != 8080 || typedService.Spec.Ports[0].NodePort != 0 {
				t.Fatal("Dashboard should have an internal HTTP Service by default")
			}
			for _, resource := range []struct{ kind, name string }{
				{"Secret", "qa-user-credentials"}, {"Secret", "qa-dashboard-bootstrap"},
				{"ConfigMap", "qa-dashboard-identity"}, {"ConfigMap", "qa-rbac-policy"},
				{"PersistentVolumeClaim", "qa-prometheus-pvc"}, {"PersistentVolumeClaim", "qa-shared-volume-pvc"},
			} {
				doc := find(t, docs, resource.kind, resource.name)
				if field(doc, "metadata", "namespace") != "qa-ns" ||
					field(doc, "metadata", "annotations", "helm.sh/resource-policy") != "keep" ||
					field(doc, "metadata", "ownerReferences") != nil {
					t.Fatalf("%s/%s must survive removal or disabling Dashboard", resource.kind, resource.name)
				}
			}
			if dashboardVolumeClaim(t, deployment, "prometheus-storage-volume") != "qa-prometheus-pvc" ||
				dashboardVolumeClaim(t, deployment, "shared-volume") != "qa-shared-volume-pvc" {
				t.Fatal("Dashboard does not mount the parent-managed PVCs")
			}
			roles := map[string]bool{}
			for _, doc := range docs {
				if doc["kind"] == "ClusterRole" {
					name := field(doc, "metadata", "name").(string)
					if !strings.Contains(name, "qa-ns") {
						t.Fatal("cluster-scoped Dashboard role must distinguish equal releases in different namespaces")
					}
					roles[name] = true
				}
			}
			if len(roles) == 0 {
				t.Fatal("Dashboard requires a ClusterRole")
			}
			for _, doc := range docs {
				if doc["kind"] == "ClusterRoleBinding" && !roles[field(doc, "roleRef", "name").(string)] {
					t.Fatal("Dashboard binding refers to a role absent from this release")
				}
			}
			password := dashboardSecretText(t, find(t, docs, "Secret", "qa-dashboard-bootstrap"), "password")
			if len(password) != 32 || password == "admin" {
				t.Fatal("default initial Dashboard password must be a random 32-character value")
			}
			admin := dashboardSecretText(t, find(t, docs, "Secret", "qa-user-credentials"), "admin")
			expected := fmt.Sprintf("%x <|:SEP:|> admin <|:SEP:|> 0 <|:SEP:|> super admin", sha256.Sum256([]byte(password)))
			if admin != expected {
				t.Fatal("initial Dashboard account hash does not match the bootstrap password or first-login metadata")
			}
		})
	}
}

func TestDashboardDisabledOmitsAllItsResources(t *testing.T) {
	docs, output, err := render(t, map[string]interface{}{"stage": "cluster", "dashboard": map[string]interface{}{"enabled": false}})
	if err != nil {
		t.Fatalf("render: %v\n%s", err, output)
	}
	find(t, docs, "OBCluster", "qa-ob")
	find(t, docs, "OBLogServiceCluster", "qa-ls")
	for _, doc := range docs {
		name, _ := field(doc, "metadata", "name").(string)
		if strings.Contains(name, "dashboard") || strings.Contains(name, "prometheus") ||
			strings.Contains(name, "alertmanager") || strings.Contains(name, "shared-volume") ||
			name == "qa-user-credentials" || name == "qa-rbac-policy" {
			t.Fatalf("disabled Dashboard still renders %s/%s", doc["kind"], name)
		}
		switch doc["kind"] {
		case "Deployment", "ClusterRole", "ClusterRoleBinding", "CronJob":
			t.Fatalf("disabled Dashboard still renders %s/%s", doc["kind"], name)
		}
	}
}

func TestDashboardStorageInheritanceAndExistingClaims(t *testing.T) {
	for _, tc := range []struct {
		name, promClass, sharedClass, promClaim, sharedClaim string
		dashboard                                            map[string]interface{}
	}{
		{"inherit", "fast-storage", "fast-storage", "qa-prometheus-pvc", "qa-shared-volume-pvc", nil},
		{"override", "metrics-storage", "reports-storage", "qa-prometheus-pvc", "qa-shared-volume-pvc", map[string]interface{}{
			"prometheus":   map[string]interface{}{"persistence": map[string]interface{}{"storageClassName": "metrics-storage"}},
			"sharedVolume": map[string]interface{}{"persistence": map[string]interface{}{"storageClassName": "reports-storage"}},
		}},
		{"existing", "", "", "my-metrics", "my-reports", map[string]interface{}{
			"prometheus":   map[string]interface{}{"persistence": map[string]interface{}{"existingClaim": "my-metrics"}},
			"sharedVolume": map[string]interface{}{"persistence": map[string]interface{}{"existingClaim": "my-reports"}},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]interface{}{"storageClass": "fast-storage"}
			if tc.dashboard != nil {
				values["dashboard"] = tc.dashboard
			}
			docs, output, err := render(t, values)
			if err != nil {
				t.Fatalf("render: %v\n%s", err, output)
			}
			assertUniqueResources(t, docs)
			var deployment appsv1.Deployment
			decodeScheduledObject(t, find(t, docs, "Deployment", "oceanbase-dashboard-qa"), &deployment)
			if dashboardVolumeClaim(t, deployment, "prometheus-storage-volume") != tc.promClaim ||
				dashboardVolumeClaim(t, deployment, "shared-volume") != tc.sharedClaim {
				t.Fatal("Dashboard ignored the persistent volume selection")
			}
			if tc.name == "existing" {
				for _, doc := range docs {
					if doc["kind"] == "PersistentVolumeClaim" && field(doc, "metadata", "name") != "qa-minio-data" {
						t.Fatal("Dashboard must neither create nor adopt user-supplied PVCs")
					}
				}
				return
			}
			if field(find(t, docs, "PersistentVolumeClaim", "qa-prometheus-pvc"), "spec", "storageClassName") != tc.promClass ||
				field(find(t, docs, "PersistentVolumeClaim", "qa-shared-volume-pvc"), "spec", "storageClassName") != tc.sharedClass {
				t.Fatal("Dashboard PVC storage class inheritance or explicit override was lost")
			}
		})
	}
}

func TestDashboardCustomInitialPasswordAndImage(t *testing.T) {
	const password = "test-initial-password"
	const image = "oceanbase-qa/oceanbase-dashboard:offline-test"
	docs, output, err := render(t, map[string]interface{}{"dashboard": map[string]interface{}{"adminPassword": password, "image": image}})
	if err != nil {
		t.Fatalf("render custom Dashboard: %v", err)
	}
	if dashboardSecretText(t, find(t, docs, "Secret", "qa-dashboard-bootstrap"), "password") != password {
		t.Fatal("initial password override was ignored")
	}
	admin := dashboardSecretText(t, find(t, docs, "Secret", "qa-user-credentials"), "admin")
	if !strings.HasPrefix(admin, fmt.Sprintf("%x <|:SEP:|>", sha256.Sum256([]byte(password)))) {
		t.Fatal("supplied initial password was not used for authentication")
	}
	if strings.Contains(output, password) {
		t.Fatal("plaintext initial password leaked outside Secret data")
	}
	var deployment appsv1.Deployment
	decodeScheduledObject(t, find(t, docs, "Deployment", "oceanbase-dashboard-qa"), &deployment)
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == "dashboard" {
			if container.Image != image {
				t.Fatal("offline Dashboard image override was ignored by the dependency")
			}
			return
		}
	}
	t.Fatal("missing Dashboard container")
}

func TestDashboardUsesLakehouseImagePullSecrets(t *testing.T) {
	for _, stage := range []string{"storage", "logservice", "cluster"} {
		t.Run(stage, func(t *testing.T) {
			want := []corev1.LocalObjectReference{{Name: "private-registry"}, {Name: "mirror-registry"}}
			docs, output, err := render(t, map[string]interface{}{"stage": stage, "imagePullSecrets": want})
			if err != nil {
				t.Fatalf("render: %v\n%s", err, output)
			}
			assertUniqueResources(t, docs)
			var deployment appsv1.Deployment
			decodeScheduledObject(t, find(t, docs, "Deployment", "oceanbase-dashboard-qa"), &deployment)
			var account corev1.ServiceAccount
			decodeScheduledObject(t, find(t, docs, "ServiceAccount", deployment.Spec.Template.Spec.ServiceAccountName), &account)
			if account.Name != "qa-sa" || account.Namespace != "qa-ns" || !reflect.DeepEqual(account.ImagePullSecrets, want) {
				t.Fatal("Dashboard ServiceAccount must inherit the lakehouse registry credentials in every stage")
			}
		})
	}
}

func TestDashboardRejectsConflictingOwnershipConfiguration(t *testing.T) {
	cases := map[string]map[string]interface{}{
		"duplicate-credentials":  {"initCredentials": true},
		"external-credentials":   {"userCredentials": "another-secret"},
		"external-users":         {"userNamespace": "another-namespace"},
		"external-config":        {"configNamespace": "another-namespace"},
		"delete-user-data":       {"retainData": false},
		"cluster-role-collision": {"rbac": map[string]interface{}{"includeNamespace": false}},
		"duplicate-account":      {"serviceAccount": map[string]interface{}{"create": true}},
	}
	for _, component := range []string{"prometheus", "sharedVolume"} {
		cases[component+"-duplicate-pvc"] = map[string]interface{}{component: map[string]interface{}{"persistence": map[string]interface{}{"create": true}}}
		cases[component+"-ephemeral-data"] = map[string]interface{}{component: map[string]interface{}{"persistence": map[string]interface{}{"enabled": false}}}
	}
	for name, dashboard := range cases {
		t.Run(name, func(t *testing.T) {
			if _, _, err := render(t, map[string]interface{}{"dashboard": dashboard}); err == nil {
				t.Fatal("conflicting parent/dependency ownership was accepted")
			}
		})
	}
}

func dashboardData(t *testing.T, doc manifest) map[string]string {
	t.Helper()
	var data map[string]interface{}
	switch value := doc["data"].(type) {
	case manifest:
		data = map[string]interface{}(value)
	case map[string]interface{}:
		data = value
	}
	if len(data) == 0 {
		t.Fatalf("%s/%s has no retained data", doc["kind"], field(doc, "metadata", "name"))
	}
	result := map[string]string{}
	for key, value := range data {
		text, ok := value.(string)
		if !ok {
			t.Fatalf("%s/%s data %s must be a string", doc["kind"], field(doc, "metadata", "name"), key)
		}
		result[key] = text
	}
	return result
}

func retainedDashboard(t *testing.T) (*retentionAPI, string) {
	t.Helper()
	api, config := retainedCredentials(t)
	docs, output, err := render(t, map[string]interface{}{})
	if err != nil {
		t.Fatalf("render initial Dashboard fixture: %v\n%s", err, output)
	}
	api.mu.Lock()
	defer api.mu.Unlock()
	for _, name := range []string{"qa-user-credentials", "qa-dashboard-bootstrap"} {
		api.secrets[name] = dashboardData(t, find(t, docs, "Secret", name))
	}
	api.configmaps = map[string]map[string]string{}
	for _, name := range []string{"qa-dashboard-identity", "qa-rbac-policy"} {
		api.configmaps[name] = dashboardData(t, find(t, docs, "ConfigMap", name))
	}
	return api, config
}

func TestDashboardRetainsChangedAccountsAndAuthorization(t *testing.T) {
	api, config := retainedDashboard(t)
	encoded := func(value string) string { return base64.StdEncoding.EncodeToString([]byte(value)) }
	api.mu.Lock()
	api.secrets["qa-user-credentials"] = map[string]string{
		"admin":   encoded(fmt.Sprintf("%x <|:SEP:|> admin <|:SEP:|> 1800000000 <|:SEP:|> changed admin", sha256.Sum256([]byte("changed-through-dashboard")))),
		"auditor": encoded(fmt.Sprintf("%x <|:SEP:|> auditor <|:SEP:|> 1800000001 <|:SEP:|> audit user", sha256.Sum256([]byte("auditor-password")))),
	}
	api.secrets["qa-dashboard-bootstrap"]["extra-retained-key"] = encoded("preserve bootstrap metadata")
	api.configmaps["qa-rbac-policy"] = map[string]string{
		"rbac_policy.csv": "p, admin, *, *, \"Super admin\"\np, reader, /api/v1/obclusters, GET, \"Read clusters\"\ng, admin, admin\ng, auditor, reader",
		"custom-data":     "preserve every ConfigMap data entry",
	}
	api.mu.Unlock()
	for attempt := 0; attempt < 2; attempt++ {
		output, err := renderRetained(t, config, map[string]interface{}{"dashboard": map[string]interface{}{"adminPassword": "must-not-reset-existing-users"}})
		if err != nil {
			t.Fatalf("render retained Dashboard: %v\n%s", err, output)
		}
		docs := releaseDocuments(t, []byte(output))
		for _, name := range []string{"qa-user-credentials", "qa-dashboard-bootstrap"} {
			api.mu.RLock()
			want := api.secrets[name]
			api.mu.RUnlock()
			if !reflect.DeepEqual(dashboardData(t, find(t, docs, "Secret", name)), want) {
				t.Fatalf("Secret/%s was reset or lost accounts/metadata on render %d", name, attempt)
			}
		}
		api.mu.RLock()
		want := api.configmaps["qa-rbac-policy"]
		api.mu.RUnlock()
		if !reflect.DeepEqual(dashboardData(t, find(t, docs, "ConfigMap", "qa-rbac-policy")), want) {
			t.Fatal("Dashboard role changes were overwritten on upgrade/reinstall")
		}
	}
}

func TestDashboardPartialRetainedAuthenticationFailsClosed(t *testing.T) {
	for _, missing := range []string{"qa-user-credentials", "qa-dashboard-bootstrap", "qa-rbac-policy"} {
		t.Run(missing, func(t *testing.T) {
			api, config := retainedDashboard(t)
			api.mu.Lock()
			delete(api.secrets, missing)
			delete(api.configmaps, missing)
			api.mu.Unlock()
			if _, err := renderRetained(t, config, map[string]interface{}{}); err == nil {
				t.Fatalf("missing retained authentication resource %s caused a silent reset", missing)
			}
		})
	}
	// Even if the marker is lost, one remaining Secret/ConfigMap is evidence that
	// this release has user data. It must not be treated as a clean installation.
	for _, remaining := range []string{"qa-user-credentials", "qa-dashboard-bootstrap", "qa-rbac-policy", "qa-dashboard-identity"} {
		t.Run("only-"+remaining, func(t *testing.T) {
			api, config := retainedDashboard(t)
			api.mu.Lock()
			for _, name := range []string{"qa-user-credentials", "qa-dashboard-bootstrap"} {
				if name != remaining {
					delete(api.secrets, name)
				}
			}
			for _, name := range []string{"qa-rbac-policy", "qa-dashboard-identity"} {
				if name != remaining {
					delete(api.configmaps, name)
				}
			}
			api.mu.Unlock()
			if _, err := renderRetained(t, config, map[string]interface{}{}); err == nil {
				t.Fatal("partial prior authentication state was mistaken for a fresh installation")
			}
		})
	}
}

func TestExistingLakehouseCanAddDashboardWithoutChangingDatabaseCredentials(t *testing.T) {
	api, config := retainedCredentials(t)
	output, err := renderRetained(t, config, map[string]interface{}{})
	if err != nil {
		t.Fatalf("existing lakehouse should permit first Dashboard installation: %v\n%s", err, output)
	}
	docs := releaseDocuments(t, []byte(output))
	find(t, docs, "Deployment", "oceanbase-dashboard-qa")
	find(t, docs, "Secret", "qa-user-credentials")
	find(t, docs, "Secret", "qa-dashboard-bootstrap")
	for _, name := range []string{"qa-db-root", "qa-minio-root", "qa-storage"} {
		api.mu.RLock()
		want := api.secrets[name]
		api.mu.RUnlock()
		if !reflect.DeepEqual(dashboardData(t, find(t, docs, "Secret", name)), want) {
			t.Fatalf("adding Dashboard changed database/storage credential %s", name)
		}
	}
	if field(find(t, docs, "ConfigMap", "qa-storage-identity"), "data", "identity") != api.identity {
		t.Fatal("adding Dashboard must not change the retained lakehouse storage identity")
	}
}

func TestDashboardCanRestoreOnlyItsMissingMarker(t *testing.T) {
	api, config := retainedDashboard(t)
	api.mu.Lock()
	delete(api.configmaps, "qa-dashboard-identity")
	api.mu.Unlock()
	output, err := renderRetained(t, config, map[string]interface{}{})
	if err != nil {
		t.Fatalf("complete authentication data should permit restoring a missing marker: %v\n%s", err, output)
	}
	docs := releaseDocuments(t, []byte(output))
	find(t, docs, "ConfigMap", "qa-dashboard-identity")
	for _, name := range []string{"qa-user-credentials", "qa-dashboard-bootstrap"} {
		api.mu.RLock()
		want := api.secrets[name]
		api.mu.RUnlock()
		if !reflect.DeepEqual(dashboardData(t, find(t, docs, "Secret", name)), want) {
			t.Fatalf("restoring a marker changed Secret/%s", name)
		}
	}
}

func TestLakehousePackageContainsDashboardDependency(t *testing.T) {
	helm := os.Getenv("HELM_BINARY")
	if helm == "" {
		helm = "helm"
	}
	destination := t.TempDir()
	output, err := exec.Command(helm, "package", ".", "--destination", destination).CombinedOutput()
	if err != nil {
		t.Fatalf("package must work with the bundled dependency: %v\n%s", err, output)
	}
	archives, err := filepath.Glob(filepath.Join(destination, "oceanbase-lakehouse-*.tgz"))
	if err != nil || len(archives) != 1 {
		t.Fatal("expected one packaged lakehouse chart")
	}
	file, err := os.Open(archives[0])
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gzipReader.Close()
	reader := tar.NewReader(gzipReader)
	hasDashboard := false
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasSuffix(header.Name, "_test.go") {
			t.Fatalf("repository-only test leaked into chart archive: %s", header.Name)
		}
		if strings.Contains(header.Name, "/charts/") &&
			(strings.HasSuffix(header.Name, "/templates/bundle.yaml") || strings.HasSuffix(header.Name, "/oceanbase-dashboard-0.6.2.tgz")) {
			hasDashboard = true
		}
	}
	if !hasDashboard {
		t.Fatal("single-chart package omits the Dashboard dependency")
	}
	// Render from an empty working directory so sibling source charts, lock
	// files, cached repositories, or local values cannot mask an incomplete tgz.
	command := exec.Command(helm, "template", "qa", archives[0], "-n", "qa-ns", "--set", "stage=cluster")
	command.Dir = t.TempDir()
	output, err = command.CombinedOutput()
	if err != nil {
		t.Fatalf("isolated chart archive cannot render: %v\n%s", err, output)
	}
	docs := releaseDocuments(t, output)
	assertUniqueResources(t, docs)
	find(t, docs, "Deployment", "oceanbase-dashboard-qa")
	find(t, docs, "OBCluster", "qa-ob")
	find(t, docs, "OBLogServiceCluster", "qa-ls")
	find(t, docs, "Secret", "qa-dashboard-bootstrap")
}

func TestPackagedDashboardDependencyPropagatesAllRuntimeImageOverrides(t *testing.T) {
	const utility = "registry.example/utility@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const prometheus = "registry.example/prometheus:v2.51.1"
	const alertmanager = "registry.example/alertmanager:v0.27.0"
	docs, output, err := render(t, map[string]interface{}{"dashboard": map[string]interface{}{
		"utilityImage": utility,
		"prometheus":   map[string]interface{}{"image": prometheus},
		"alertmanager": map[string]interface{}{"image": alertmanager},
	}})
	if err != nil {
		t.Fatal(err, output)
	}
	var deployment appsv1.Deployment
	decodeScheduledObject(t, find(t, docs, "Deployment", "oceanbase-dashboard-qa"), &deployment)
	if deployment.Spec.Template.Spec.InitContainers[0].Image != utility {
		t.Fatal("packaged dependency lost utility image override")
	}
	for _, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == "prometheus" && c.Image != prometheus {
			t.Fatal("packaged dependency lost Prometheus override")
		}
		if c.Name == "alertmanager" && c.Image != alertmanager {
			t.Fatal("packaged dependency lost Alertmanager override")
		}
	}
	if !strings.Contains(output, "image: \""+utility+"\"") || strings.Count(output, utility) != 2 {
		t.Fatal("cleanup and volume initialization must share the overridden utility image")
	}
	if strings.Contains(output, "busybox:latest") {
		t.Fatal("mutable utility default remains in the packaged dependency")
	}
}
