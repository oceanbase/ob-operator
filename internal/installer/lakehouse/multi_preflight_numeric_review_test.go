package lakehouse

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// Exercise the two persisted-value paths used by a retry: Helm JSON and the
// installer's saved YAML. Both decode otherwise exact integer IDs as float64.
func numericReviewValues(t *testing.T, values map[string]interface{}, savedYAML bool) map[string]interface{} {
	t.Helper()
	raw, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if !savedYAML {
		return decoded
	}
	raw, err = yaml.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "values.yaml")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	decoded, err = readValues(path)
	if err != nil {
		t.Fatal(err)
	}
	return decoded
}

func numericReviewClaim(id int64) (corev1.PersistentVolumeClaim, corev1.PersistentVolume) {
	claim, pv := pfMultiClaimAndPV()
	name := fmt.Sprintf("demo-ob-%d-z2-abcdef", id)
	claim.Name = name + "-data"
	claim.OwnerReferences[0].Name = name
	pv.Spec.ClaimRef.Name = claim.Name
	return claim, pv
}

func TestMultiPreflightNumericReviewRoundTripsRetainedState(t *testing.T) {
	// The schema permits 1..2147483647. Include the decimal/scientific-format
	// transition and the cluster ID from the live three-worker acceptance run.
	for _, id := range []int64{1, 999999, 1000000, 821692080, 2147483647} {
		for _, savedYAML := range []bool{false, true} {
			format := "helm-json"
			if savedYAML {
				format = "saved-yaml"
			}
			t.Run(fmt.Sprintf("%d/%s", id, format), func(t *testing.T) {
				opts, nodes := pfMultiFixture()
				opts.Existing = true
				setValue(opts.Values, id, "observer", "clusterId")
				opts.Values = numericReviewValues(t, opts.Values, savedYAML)
				pod := pfTestPod("2", true)
				pod.Name = fmt.Sprintf("demo-ob-%d-z2-abcdef", id)
				pod.Labels["ref-obzone"] = fmt.Sprintf("demo-ob-%d-z2", id)
				pod.Spec.NodeName = "worker-2"
				if err := pfCheckMultiCapacity(opts, nodes, []pfPod{pod}, nil, nil, &PreflightResult{StorageClass: "local-path"}); err != nil {
					t.Fatalf("same cluster and Pod placement rejected after %s: %v", format, err)
				}
				claim, pv := numericReviewClaim(id)
				if err := pfCheckMultiCapacity(opts, nodes, nil, []corev1.PersistentVolumeClaim{claim}, []corev1.PersistentVolume{pv}, &PreflightResult{StorageClass: "local-path"}); err != nil {
					t.Fatalf("same retained PVC placement rejected without Pods after %s: %v", format, err)
				}
				claim.OwnerReferences = nil
				claim.Labels = map[string]string{"ref-obcluster": "demo-ob", "ref-obzone": fmt.Sprintf("demo-ob-%d-z2", id)}
				if err := pfCheckMultiCapacity(opts, nodes, nil, []corev1.PersistentVolumeClaim{claim}, []corev1.PersistentVolume{pv}, &PreflightResult{StorageClass: "local-path"}); err != nil {
					t.Fatalf("independent-lifecycle PVC rejected after %s: %v", format, err)
				}
			})
		}
	}
}

func TestMultiPreflightNumericReviewKeepsIdentityAndMigrationGuards(t *testing.T) {
	const id int64 = 821692080
	tests := []struct {
		name   string
		want   string
		change func(*PreflightOptions, *corev1.PersistentVolumeClaim, *corev1.PersistentVolume)
	}{
		{"different cluster id", "unconfigured database zone", func(o *PreflightOptions, c *corev1.PersistentVolumeClaim, p *corev1.PersistentVolume) {
			setValue(o.Values, float64(id+1), "observer", "clusterId")
		}},
		{"unknown zone", "unconfigured database zone", func(o *PreflightOptions, c *corev1.PersistentVolumeClaim, p *corev1.PersistentVolume) {
			c.OwnerReferences[0].Name = fmt.Sprintf("demo-ob-%d-oldzone-abcdef", id)
		}},
		{"selected node changed", "automatic migration", func(o *PreflightOptions, c *corev1.PersistentVolumeClaim, p *corev1.PersistentVolume) {
			c.Annotations["volume.kubernetes.io/selected-node"] = "worker-3"
		}},
		{"pv affinity changed", "affinity excludes", func(o *PreflightOptions, c *corev1.PersistentVolumeClaim, p *corev1.PersistentVolume) {
			p.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Values = []string{"host-3"}
		}},
		{"new install with retained data", "retained PVC", func(o *PreflightOptions, c *corev1.PersistentVolumeClaim, p *corev1.PersistentVolume) {
			o.Existing = false
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, nodes := pfMultiFixture()
			opts.Existing = true
			setValue(opts.Values, id, "observer", "clusterId")
			opts.Values = numericReviewValues(t, opts.Values, true)
			claim, pv := numericReviewClaim(id)
			tt.change(&opts, &claim, &pv)
			err := pfCheckMultiCapacity(opts, nodes, nil, []corev1.PersistentVolumeClaim{claim}, []corev1.PersistentVolume{pv}, &PreflightResult{StorageClass: "local-path"})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("retained data guard weakened: want %q, got %v", tt.want, err)
			}
		})
	}

	for _, zone := range []string{"demo-ob-821692081-z2", "demo-ob-821692080-oldzone", "demo-ob-8.2169208e+08-z2"} {
		t.Run("pod-"+zone, func(t *testing.T) {
			opts, nodes := pfMultiFixture()
			opts.Existing = true
			setValue(opts.Values, id, "observer", "clusterId")
			opts.Values = numericReviewValues(t, opts.Values, false)
			pod := pfTestPod("2", true)
			pod.Name = zone + "-abcdef"
			pod.Labels["ref-obzone"] = zone
			pod.Spec.NodeName = "worker-2"
			err := pfCheckMultiCapacity(opts, nodes, []pfPod{pod}, nil, nil, &PreflightResult{StorageClass: "local-path"})
			if err == nil || !strings.Contains(err.Error(), "unconfigured zone") {
				t.Fatalf("unknown Pod identity accepted: %v", err)
			}
		})
	}
}
