package lakehouse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Values must contain the complete, merged lakehouse chart values. This check
// never creates resources, chooses different storage for an existing release,
// or silently relocates an existing volume.
type PreflightOptions struct {
	Release, Namespace, NodeName string
	NodeNames                    []string
	Values                       map[string]interface{}
	Existing, PrepareOperator    bool
}

type PreflightResult struct {
	NodeName, StorageClass, KubernetesVersion, HelmVersion             string
	NodeRuntimeVersion, NodeInternalIP, NodeExternalIP, NodeSystemUUID string
	NodeHostname                                                       string
	NodeCount                                                          int
	OperatorMissing                                                    bool
	RequiredCPUMilli, AvailableCPUMilli                                int64
	RequiredMemoryBytes, AvailableMemoryBytes                          int64
	Warnings                                                           []string
	Nodes                                                              []PreflightResult
}

const pfCommandTimeout = 10 * time.Second

// CheckCluster performs bounded, read-only checks for the single-worker or
// explicit three-node experience. Other advanced placement is never guessed.
func CheckCluster(ctx context.Context, run RunCommand, opts PreflightOptions) (PreflightResult, error) {
	var result PreflightResult
	var problems []error
	if run == nil || opts.Release == "" || opts.Namespace == "" || opts.Values == nil {
		return result, errors.New("preflight requires a command runner, release, namespace and complete values")
	}
	read := func(tool string, out interface{}, args ...string) bool {
		child, cancel := context.WithTimeout(ctx, pfCommandTimeout)
		defer cancel()
		if tool == "kubectl" {
			args = append([]string{"--request-timeout=10s"}, args...)
		}
		data, err := run(child, tool, args...)
		if err != nil {
			// Do not echo command output: admission errors may contain user data.
			problems = append(problems, fmt.Errorf("cannot read %s %s (check context, credentials and API access): %w", tool, strings.Join(args, " "), err))
			return false
		}
		if dst, ok := out.(*string); ok {
			*dst = strings.TrimSpace(string(data))
			return true
		}
		if err := json.Unmarshal(data, out); err != nil {
			problems = append(problems, fmt.Errorf("invalid JSON from %s %s", tool, strings.Join(args, " ")))
			return false
		}
		return true
	}
	var version struct {
		ServerVersion struct {
			GitVersion string `json:"gitVersion"`
		} `json:"serverVersion"`
	}
	if read("kubectl", &version, "version", "-o", "json") {
		result.KubernetesVersion = version.ServerVersion.GitVersion
		if !pfVersionAtLeast(result.KubernetesVersion, 1, 27) {
			problems = append(problems, fmt.Errorf("Kubernetes >=1.27 is required; found %q", result.KubernetesVersion))
		}
	}
	if read("helm", &result.HelmVersion, "version", "--template", "{{.Version}}") && !pfVersionAtLeast(result.HelmVersion, 3, 12) {
		problems = append(problems, fmt.Errorf("Helm >=3.12 is required; found %q", result.HelmVersion))
	}
	var classes storagev1.StorageClassList
	if read("kubectl", &classes, "get", "storageclasses", "-o", "json") {
		var err error
		result.StorageClass, err = pfStorageClass(opts.Values, classes.Items)
		if err != nil {
			problems = append(problems, err)
		}
	}
	var nodes corev1.NodeList
	var pods pfPodList
	var claims corev1.PersistentVolumeClaimList
	var volumes corev1.PersistentVolumeList
	nodesOK := read("kubectl", &nodes, "get", "nodes", "-o", "json")
	if nodesOK {
		result.NodeCount = len(nodes.Items)
	}
	podsOK := read("kubectl", &pods, "get", "pods", "--all-namespaces", "-o", "json")
	claimsOK := read("kubectl", &claims, "get", "pvc", "--all-namespaces", "-o", "json")
	volumesOK := read("kubectl", &volumes, "get", "pv", "-o", "json")
	if nodesOK && podsOK && claimsOK && volumesOK {
		checkCapacity := pfCheckCapacity
		if len(opts.NodeNames) == 3 {
			checkCapacity = pfCheckMultiCapacity
		}
		if err := checkCapacity(opts, nodes.Items, pods.Items, claims.Items, volumes.Items, &result); err != nil {
			problems = append(problems, err)
		}
	}
	var crds apiextv1.CustomResourceDefinitionList
	var deployments appsv1.DeploymentList
	var mutating admissionv1.MutatingWebhookConfigurationList
	var validating admissionv1.ValidatingWebhookConfigurationList
	var endpoints corev1.EndpointsList
	crdsOK := read("kubectl", &crds, "get", "crds", "-o", "json")
	if crdsOK {
		for _, crd := range crds.Items {
			if crd.Spec.Group != "oceanbase.oceanbase.com" || crd.Spec.Names.Plural != "obclusters" {
				continue
			}
			var clusters struct {
				Items []struct {
					Metadata struct{ Name string } `json:"metadata"`
					Spec     struct {
						DeploymentMode string                 `json:"deploymentMode"`
						Observer       struct{ Image string } `json:"observer"`
					} `json:"spec"`
				} `json:"items"`
			}
			if read("kubectl", &clusters, "get", "obclusters", "-n", opts.Namespace, "-o", "json") {
				for _, cluster := range clusters.Items {
					if len(opts.NodeNames) == 3 && !opts.Existing && cluster.Metadata.Name == opts.Release+"-ob" {
						problems = append(problems, fmt.Errorf("OBCluster %s/%s already exists without an existing Helm release; recover the original installation instead of adopting orphaned resources", opts.Namespace, cluster.Metadata.Name))
					}
					if cluster.Metadata.Name != opts.Release+"-ob" || cluster.Spec.DeploymentMode != "shared_storage" {
						continue
					}
					// The webhook compares the literal image string. Even a
					// docker.io prefix change is an unsupported version update.
					if cluster.Spec.Observer.Image == "" || cluster.Spec.Observer.Image != pfString(pfGet(opts.Values, "observer", "image")) {
						problems = append(problems, fmt.Errorf("existing shared_storage OBCluster %s/%s does not support changes to observer.image; retain its exact installed image and update only the Dashboard diagnostic images", opts.Namespace, cluster.Metadata.Name))
					}
				}
			}
			break
		}
	}
	if crdsOK && len(opts.NodeNames) == 3 && !opts.Existing {
		for _, crd := range crds.Items {
			if crd.Spec.Group != "oceanbase.oceanbase.com" || crd.Spec.Names.Plural != "oblogserviceclusters" {
				continue
			}
			var clusters struct {
				Items []struct {
					Metadata struct{ Name string } `json:"metadata"`
				} `json:"items"`
			}
			if read("kubectl", &clusters, "get", "oblogserviceclusters", "-n", opts.Namespace, "-o", "json") {
				for _, cluster := range clusters.Items {
					if cluster.Metadata.Name == opts.Release+"-ls" {
						problems = append(problems, fmt.Errorf("OBLogServiceCluster %s/%s already exists without an existing Helm release; recover the original installation instead of adopting orphaned resources", opts.Namespace, cluster.Metadata.Name))
					}
				}
			}
			break
		}
	}
	deploymentsOK := read("kubectl", &deployments, "get", "deployments", "--all-namespaces", "-o", "json")
	mutatingOK := read("kubectl", &mutating, "get", "mutatingwebhookconfigurations", "-o", "json")
	validatingOK := read("kubectl", &validating, "get", "validatingwebhookconfigurations", "-o", "json")
	endpointsOK := read("kubectl", &endpoints, "get", "endpoints", "--all-namespaces", "-o", "json")
	if crdsOK && deploymentsOK && mutatingOK && validatingOK && endpointsOK {
		missing, err := pfCheckOperator(crds.Items, deployments.Items, mutating.Items, validating.Items, endpoints.Items)
		result.OperatorMissing = missing
		if err != nil {
			problems = append(problems, err)
		} else if missing && len(opts.NodeNames) == 3 {
			problems = append(problems, errors.New("three-node installation requires an existing compatible Operator; install a compatible Operator before using this multi-node cluster"))
		} else if missing && !opts.PrepareOperator {
			problems = append(problems, errors.New("SS-capable ob-operator is not installed; enable preparation or install the matching package prerequisites"))
		} else if missing {
			if result.NodeCount != 1 {
				problems = append(problems, errors.New("automatic Operator preparation supports only a single-node cluster; install a compatible Operator before using this multi-node cluster"))
			}
			// Conservative allowance for the bundled controller and cert-manager.
			// Their actual requests are checked again after preparation.
			result.RequiredCPUMilli += 500
			result.RequiredMemoryBytes += 1024 * 1024 * 1024
			if result.NodeName != "" && (result.RequiredCPUMilli > result.AvailableCPUMilli || result.RequiredMemoryBytes > result.AvailableMemoryBytes) {
				problems = append(problems, errors.New("insufficient remaining capacity including the 500m CPU / 1Gi prerequisite preparation allowance"))
			}
			result.Warnings = append(result.Warnings, "Operator preparation is required; rerun preflight after it becomes ready before installing the lakehouse")
		}
	}
	if ctx.Err() != nil {
		problems = append(problems, ctx.Err())
	}
	return result, errors.Join(problems...)
}

var pfVersionRE = regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)(?:\.[0-9]+)?(?:[-+].*)?$`)

func pfVersionAtLeast(v string, major, minor int) bool {
	m := pfVersionRE.FindStringSubmatch(strings.TrimSpace(v))
	if len(m) != 3 {
		return false
	}
	a, _ := strconv.Atoi(m[1])
	b, _ := strconv.Atoi(m[2])
	return a > major || a == major && b >= minor
}

func pfMap(v interface{}) map[string]interface{} {
	m, _ := v.(map[string]interface{})
	return m
}

func pfGet(v map[string]interface{}, path ...string) interface{} {
	var current interface{} = v
	for _, key := range path {
		current = pfMap(current)[key]
	}
	return current
}

func pfString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

func pfStorageClass(values map[string]interface{}, classes []storagev1.StorageClass) (string, error) {
	name := pfString(values["storageClass"])
	known := map[string]bool{}
	var defaults []string
	for _, sc := range classes {
		known[sc.Name] = true
		if sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" || sc.Annotations["storageclass.beta.kubernetes.io/is-default-class"] == "true" {
			defaults = append(defaults, sc.Name)
		}
	}
	if name == "" {
		if len(defaults) != 1 {
			return "", errors.New("select storageClass explicitly: exactly one default StorageClass was not found")
		}
		name = defaults[0]
	}
	if !known[name] {
		return "", fmt.Errorf("StorageClass %q does not exist", name)
	}
	if enabled, _ := pfGet(values, "dashboard", "enabled").(bool); !enabled {
		return name, nil
	}
	for _, field := range []string{"prometheus", "sharedVolume"} {
		other := pfString(pfGet(values, "dashboard", field, "persistence", "storageClassName"))
		if other != "" && !known[other] {
			return "", fmt.Errorf("Dashboard %s StorageClass %q does not exist", field, other)
		}
	}
	return name, nil
}

type pfResources struct{ cpu, memory int64 }

func (r *pfResources) add(other pfResources) {
	if other.cpu > math.MaxInt64-r.cpu {
		r.cpu = math.MaxInt64
	} else {
		r.cpu += other.cpu
	}
	if other.memory > math.MaxInt64-r.memory {
		r.memory = math.MaxInt64
	} else {
		r.memory += other.memory
	}
}

func pfQuantities(v map[string]interface{}, field string) (pfResources, error) {
	var result pfResources
	for _, key := range []string{"cpu", "memory"} {
		q, err := resource.ParseQuantity(pfString(v[key]))
		if err != nil || q.Sign() <= 0 {
			return result, fmt.Errorf("%s.%s must be a positive Kubernetes quantity in complete values", field, key)
		}
		if key == "cpu" {
			if q.Cmp(resource.MustParse("1000000000")) > 0 {
				return result, fmt.Errorf("%s.cpu exceeds supported capacity", field)
			}
			result.cpu = q.MilliValue()
		} else {
			if q.Cmp(resource.MustParse("9223372036854775807")) > 0 {
				return result, fmt.Errorf("%s.memory exceeds supported capacity", field)
			}
			result.memory = q.Value()
		}
	}
	return result, nil
}

func pfSelector(v interface{}) (map[string]string, error) {
	out := map[string]string{}
	if v == nil {
		return out, nil
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil, errors.New("nodeSelector must be a string map")
	}
	for key, value := range m {
		s, ok := value.(string)
		if !ok {
			return nil, errors.New("nodeSelector values must be strings")
		}
		out[key] = s
	}
	return out, nil
}

func pfDemand(values map[string]interface{}) (pfResources, []map[string]string, []corev1.Toleration, error) {
	var demand pfResources
	global, err := pfSelector(pfGet(values, "scheduling", "nodeSelector"))
	if err != nil {
		return demand, nil, nil, err
	}
	selectors := []map[string]string{global}
	var tolerations []corev1.Toleration
	if raw := pfGet(values, "scheduling", "tolerations"); raw != nil {
		data, _ := json.Marshal(raw)
		if err := json.Unmarshal(data, &tolerations); err != nil {
			return demand, nil, nil, errors.New("invalid scheduling.tolerations")
		}
	}
	for _, component := range []string{"logService", "observer"} {
		r, err := pfQuantities(pfMap(pfGet(values, component, "resource")), component+".resource")
		if err != nil {
			return demand, nil, nil, err
		}
		topology, ok := pfGet(values, component, "topology").([]interface{})
		if !ok || len(topology) == 0 {
			return demand, nil, nil, fmt.Errorf("%s.topology must be present in complete values", component)
		}
		for _, entry := range topology {
			zone := pfMap(entry)
			replicas, err := strconv.Atoi(pfString(zone["replica"]))
			if err != nil || replicas < 1 || replicas > 100 {
				return demand, nil, nil, fmt.Errorf("invalid %s.topology replica count", component)
			}
			if r.cpu > math.MaxInt64/int64(replicas) || r.memory > math.MaxInt64/int64(replicas) {
				return demand, nil, nil, fmt.Errorf("%s resource total exceeds supported capacity", component)
			}
			selector := map[string]string{}
			for k, v := range global {
				selector[k] = v
			}
			overrides, err := pfSelector(zone["nodeSelector"])
			if err != nil {
				return demand, nil, nil, err
			}
			for k, v := range overrides {
				selector[k] = v
			}
			selectors = append(selectors, selector)
			demand.add(pfResources{r.cpu * int64(replicas), r.memory * int64(replicas)})
		}
	}
	if pfString(pfGet(values, "objectStorage", "mode")) == "bundled" {
		r, err := pfQuantities(pfMap(pfGet(values, "minio", "resources", "requests")), "minio.resources.requests")
		if err != nil {
			return demand, nil, nil, err
		}
		demand.add(r)
	}
	mode := pfString(pfGet(values, "objectStorage", "mode"))
	if mode != "bundled" && mode != "external" {
		return demand, nil, nil, errors.New("objectStorage.mode must be bundled or external")
	}
	createBuckets, _ := pfGet(values, "objectStorage", "createBuckets").(bool)
	if mode == "bundled" || createBuckets {
		r, err := pfQuantities(pfMap(pfGet(values, "initialization", "resources", "requests")), "initialization.resources.requests")
		if err != nil {
			return demand, nil, nil, err
		}
		demand.add(r) // A retry/upgrade also creates this storage-init Job.
	}
	if enabled, _ := pfGet(values, "dashboard", "enabled").(bool); enabled {
		selector, err := pfSelector(pfGet(values, "dashboard", "nodeSelector"))
		if err != nil {
			return demand, nil, nil, err
		}
		selectors = append(selectors, selector)
		// These are the actual three container requests in the bundled child chart.
		demand.add(pfResources{700, 2*500*1024*1024 + 500000000})
	}
	return demand, selectors, tolerations, nil
}

type pfPodList struct {
	Items []pfPod `json:"items"`
}
type pfPod struct {
	corev1.Pod
	PodResources *corev1.ResourceRequirements
}

func (p *pfPod) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &p.Pod); err != nil {
		return err
	}
	var newer struct {
		Spec struct {
			Resources *corev1.ResourceRequirements `json:"resources"`
		} `json:"spec"`
	}
	if err := json.Unmarshal(data, &newer); err != nil {
		return err
	}
	p.PodResources = newer.Spec.Resources
	return nil
}

func pfResourceList(r corev1.ResourceList) pfResources {
	return pfResources{r.Cpu().MilliValue(), r.Memory().Value()}
}

func pfPodRequests(p pfPod) pfResources {
	var total, sidecars, initMax pfResources
	for _, c := range p.Spec.Containers {
		total.add(pfResourceList(c.Resources.Requests))
	}
	for _, c := range p.Spec.InitContainers {
		r := pfResourceList(c.Resources.Requests)
		if c.RestartPolicy != nil && *c.RestartPolicy == corev1.ContainerRestartPolicyAlways {
			sidecars.add(r)
			total.add(r)
			r = sidecars
		} else {
			r.add(sidecars)
		}
		initMax.cpu = max(initMax.cpu, r.cpu)
		initMax.memory = max(initMax.memory, r.memory)
	}
	total.cpu = max(total.cpu, initMax.cpu)
	total.memory = max(total.memory, initMax.memory)
	if p.PodResources != nil {
		r := pfResourceList(p.PodResources.Requests)
		total.cpu = max(total.cpu, r.cpu)
		total.memory = max(total.memory, r.memory)
	}
	total.add(pfResourceList(p.Spec.Overhead))
	return total
}

func pfOwned(p pfPod, opts PreflightOptions) bool {
	if p.Namespace != opts.Namespace {
		return false
	}
	l := p.Labels
	return l["app.kubernetes.io/instance"] == opts.Release || l["ref-obcluster"] == opts.Release+"-ob" || l["ref-oblogservicecluster"] == opts.Release+"-ls" || l["app"] == "oceanbase-dashboard-"+opts.Release
}

func pfActive(p pfPod) bool {
	return p.Status.Phase != corev1.PodSucceeded && p.Status.Phase != corev1.PodFailed
}

func pfMatches(node corev1.Node, selectors []map[string]string) bool {
	for _, selector := range selectors {
		for k, v := range selector {
			if node.Labels[k] != v {
				return false
			}
		}
	}
	return true
}

func pfTolerates(node corev1.Node, tolerations []corev1.Toleration) bool {
	for _, taint := range node.Spec.Taints {
		if taint.Effect != corev1.TaintEffectNoSchedule && taint.Effect != corev1.TaintEffectNoExecute {
			continue
		}
		found := false
		for _, t := range tolerations {
			if t.ToleratesTaint(&taint) && (taint.Effect != corev1.TaintEffectNoExecute || t.TolerationSeconds == nil) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func pfReady(node corev1.Node) bool {
	if node.Spec.Unschedulable {
		return false
	}
	for _, c := range node.Status.Conditions {
		if c.Type == corev1.NodeReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

func pfCheckCapacity(opts PreflightOptions, nodes []corev1.Node, pods []pfPod, claims []corev1.PersistentVolumeClaim, volumes []corev1.PersistentVolume, result *PreflightResult) error {
	demand, selectors, tolerations, err := pfDemand(opts.Values)
	if err != nil {
		return err
	}
	result.RequiredCPUMilli, result.RequiredMemoryBytes = demand.cpu, demand.memory
	pinned := map[string]bool{}
	ownedClaims := map[string]bool{}
	suppliedClaims := map[string]bool{}
	expectedClasses := map[string]string{opts.Release + "-minio-data": result.StorageClass}
	if pfString(pfGet(opts.Values, "objectStorage", "mode")) == "bundled" {
		if name := pfString(pfGet(opts.Values, "minio", "existingClaim")); name != "" {
			suppliedClaims[name] = false
		}
	}
	if enabled, _ := pfGet(opts.Values, "dashboard", "enabled").(bool); enabled {
		for field, suffix := range map[string]string{"prometheus": "prometheus-pvc", "sharedVolume": "shared-volume-pvc"} {
			if name := pfString(pfGet(opts.Values, "dashboard", field, "persistence", "existingClaim")); name != "" {
				suppliedClaims[name] = false
			}
			class := pfString(pfGet(opts.Values, "dashboard", field, "persistence", "storageClassName"))
			if class == "" {
				class = result.StorageClass
			}
			expectedClasses[opts.Release+"-"+suffix] = class
		}
	}
	for _, p := range pods {
		if !pfOwned(p, opts) {
			continue
		}
		if !opts.Existing {
			return errors.New("resources for this release already exist; resume the existing installation instead of creating a new one")
		}
		if pfActive(p) && p.Spec.NodeName != "" {
			pinned[p.Spec.NodeName] = true
		}
		for _, v := range p.Spec.Volumes {
			if v.PersistentVolumeClaim != nil {
				ownedClaims[v.PersistentVolumeClaim.ClaimName] = true
			}
		}
	}
	for _, c := range claims {
		if c.Namespace != opts.Namespace {
			continue
		}
		owned := ownedClaims[c.Name] || c.Labels["app.kubernetes.io/instance"] == opts.Release
		_, supplied := suppliedClaims[c.Name]
		for _, owner := range c.OwnerReferences {
			if owner.Kind == "OBLogServiceNode" && strings.HasPrefix(owner.Name, opts.Release+"-ls-") || owner.Kind == "OBServer" && strings.HasPrefix(owner.Name, opts.Release+"-ob-"+pfString(pfGet(opts.Values, "observer", "clusterId"))+"-") {
				owned = true
				expectedClasses[c.Name] = result.StorageClass
			}
		}
		if !owned && !supplied {
			continue
		}
		if !opts.Existing && owned {
			return errors.New("retained PVCs for this release exist; use the existing release configuration")
		}
		if c.DeletionTimestamp != nil {
			return fmt.Errorf("PVC %s is being deleted", c.Name)
		}
		if supplied {
			suppliedClaims[c.Name] = true
		}
		ownedClaims[c.Name] = true
		if node := c.Annotations["volume.kubernetes.io/selected-node"]; node != "" {
			pinned[node] = true
		}
		if expected := expectedClasses[c.Name]; expected != "" && c.Spec.StorageClassName != nil && *c.Spec.StorageClassName != expected && !supplied {
			return fmt.Errorf("PVC %s uses StorageClass %s; changing existing storage is unsupported", c.Name, *c.Spec.StorageClassName)
		}
	}
	for name, exists := range suppliedClaims {
		if !exists {
			return fmt.Errorf("configured existingClaim %s does not exist in namespace %s", name, opts.Namespace)
		}
	}
	for _, pv := range volumes {
		if pv.Spec.ClaimRef == nil || pv.Spec.ClaimRef.Namespace != opts.Namespace || !ownedClaims[pv.Spec.ClaimRef.Name] || pv.Spec.NodeAffinity == nil || pv.Spec.NodeAffinity.Required == nil {
			continue
		}
		terms := pv.Spec.NodeAffinity.Required.NodeSelectorTerms
		if len(terms) != 1 {
			return fmt.Errorf("PVC %s has advanced volume placement; use the existing Chart deployment workflow", pv.Spec.ClaimRef.Name)
		}
		for _, e := range terms[0].MatchExpressions {
			if e.Key == "kubernetes.io/hostname" && e.Operator == corev1.NodeSelectorOpIn && len(e.Values) == 1 {
				// PV affinity names are hostname label values, not necessarily node names.
				matched := false
				for _, n := range nodes {
					if n.Labels[e.Key] == e.Values[0] {
						pinned[n.Name] = true
						matched = true
					}
				}
				if !matched {
					return fmt.Errorf("PVC %s is bound to an unavailable node", pv.Spec.ClaimRef.Name)
				}
			} else {
				return fmt.Errorf("PVC %s has advanced volume placement; use the existing Chart deployment workflow", pv.Spec.ClaimRef.Name)
			}
		}
		if len(terms[0].MatchFields) > 0 {
			return errors.New("advanced PV node matchFields are unsupported by the single-worker installer")
		}
	}
	if len(pinned) > 1 {
		return errors.New("existing release spans multiple nodes; the single-worker installer will not relocate it")
	}
	target := opts.NodeName
	for node := range pinned {
		if target != "" && target != node {
			return fmt.Errorf("existing data is on node %s, not requested node %s; automatic migration is unsupported", node, target)
		}
		target = node
	}
	var eligible []corev1.Node
	for _, node := range nodes {
		if target != "" && node.Name != target {
			continue
		}
		if pfReady(node) && pfMatches(node, selectors) && pfTolerates(node, tolerations) {
			eligible = append(eligible, node)
		}
	}
	if len(eligible) != 1 {
		return fmt.Errorf("single-worker installation requires one Ready, schedulable node matching all zone selectors/tolerations (found %d); select a node or use the advanced Chart workflow", len(eligible))
	}
	node := eligible[0]
	if enabled, _ := pfGet(opts.Values, "dashboard", "enabled").(bool); enabled {
		var dashboardTolerations []corev1.Toleration
		if raw := pfGet(opts.Values, "dashboard", "tolerations"); raw != nil {
			data, _ := json.Marshal(raw)
			if err := json.Unmarshal(data, &dashboardTolerations); err != nil {
				return errors.New("invalid dashboard.tolerations")
			}
		}
		if !pfTolerates(node, dashboardTolerations) {
			return errors.New("target node has a taint not tolerated by Dashboard; database tolerations alone do not place Dashboard")
		}
	}
	result.NodeName = node.Name
	result.NodeHostname = node.Labels["kubernetes.io/hostname"]
	if result.NodeHostname == "" {
		return errors.New("target node is missing kubernetes.io/hostname; automatic placement is unsupported")
	}
	result.NodeRuntimeVersion = node.Status.NodeInfo.ContainerRuntimeVersion
	result.NodeSystemUUID = node.Status.NodeInfo.SystemUUID
	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeInternalIP && result.NodeInternalIP == "" {
			result.NodeInternalIP = address.Address
		}
		if address.Type == corev1.NodeExternalIP && result.NodeExternalIP == "" {
			result.NodeExternalIP = address.Address
		}
	}
	available := pfResourceList(node.Status.Allocatable)
	for _, p := range pods {
		if !pfActive(p) {
			continue
		}
		if p.Spec.NodeName != node.Name {
			if p.Spec.NodeName != "" || !pfMatches(node, []map[string]string{p.Spec.NodeSelector}) || !pfTolerates(node, p.Spec.Tolerations) {
				continue
			}
		}
		// Replacing the release's existing request footprint avoids double charging
		// upgrades. New desired requests above already include its complete footprint.
		if opts.Existing && pfOwned(p, opts) {
			continue
		}
		used := pfPodRequests(p)
		available.cpu -= used.cpu
		available.memory -= used.memory
	}
	result.AvailableCPUMilli, result.AvailableMemoryBytes = available.cpu, available.memory
	if demand.cpu > available.cpu || demand.memory > available.memory {
		return fmt.Errorf("node %s lacks resources: installation needs %dm CPU / %dMi memory, available after other workloads %dm / %dMi", node.Name, demand.cpu, demand.memory/(1024*1024), available.cpu, available.memory/(1024*1024))
	}
	return nil
}

func pfCheckOperator(crds []apiextv1.CustomResourceDefinition, deployments []appsv1.Deployment, mutating []admissionv1.MutatingWebhookConfiguration, validating []admissionv1.ValidatingWebhookConfiguration, endpoints []corev1.Endpoints) (bool, error) {
	required := map[string]bool{"obclusters": false, "oblogserviceclusters": false, "oblogservicezones": false, "oblogservicenodes": false}
	anyOperator := false
	var problems []error
	for _, crd := range crds {
		if crd.Spec.Group != "oceanbase.oceanbase.com" {
			continue
		}
		anyOperator = true
		if _, ok := required[crd.Spec.Names.Plural]; !ok {
			continue
		}
		required[crd.Spec.Names.Plural] = true
		established := false
		for _, c := range crd.Status.Conditions {
			if c.Type == apiextv1.Established && c.Status == apiextv1.ConditionTrue {
				established = true
			}
		}
		if !established {
			problems = append(problems, fmt.Errorf("CRD %s is not Established", crd.Name))
		}
		served := false
		for _, version := range crd.Spec.Versions {
			if version.Name != "v1alpha1" || !version.Served {
				continue
			}
			served = true
			if crd.Spec.Names.Plural == "obclusters" {
				compatible := false
				if version.Schema != nil && version.Schema.OpenAPIV3Schema != nil {
					mode := version.Schema.OpenAPIV3Schema.Properties["spec"].Properties["deploymentMode"]
					for _, value := range mode.Enum {
						if string(value.Raw) == `"shared_storage"` {
							compatible = true
						}
					}
				}
				if !compatible {
					problems = append(problems, errors.New("installed OBCluster CRD does not advertise shared_storage; do not overwrite the existing Operator automatically"))
				}
			}
		}
		if !served {
			problems = append(problems, fmt.Errorf("CRD %s does not serve v1alpha1", crd.Name))
		}
	}
	readyNamespaces := map[string]bool{}
	for _, d := range deployments {
		isOperator := d.Name == "oceanbase-controller-manager" || strings.Contains(d.Labels["app.kubernetes.io/part-of"], "ob-operator")
		for _, c := range d.Spec.Template.Spec.Containers {
			if strings.Contains(c.Image, "/ob-operator:") || strings.Contains(c.Image, "/ob-operator@") {
				isOperator = true
			}
		}
		if !isOperator {
			continue
		}
		anyOperator = true
		if d.Status.ObservedGeneration >= d.Generation && d.Status.AvailableReplicas > 0 && d.Status.ReadyReplicas > 0 {
			readyNamespaces[d.Namespace] = true
		}
	}
	readyEndpoints := map[string]bool{}
	for _, ep := range endpoints {
		for _, subset := range ep.Subsets {
			if len(subset.Addresses) > 0 && len(subset.Ports) > 0 {
				readyEndpoints[ep.Namespace+"/"+ep.Name] = true
			}
		}
	}
	mutates, validates := map[string]bool{}, map[string]bool{}
	checkWebhook := func(config admissionv1.WebhookClientConfig, rules []admissionv1.RuleWithOperations, seen map[string]bool) {
		for _, rule := range rules {
			if !pfContains(rule.APIGroups, "oceanbase.oceanbase.com") {
				continue
			}
			anyOperator = true
			for _, kind := range []string{"obclusters", "oblogserviceclusters"} {
				if !pfContains(rule.Resources, kind) || !pfContains(rule.APIVersions, "v1alpha1") {
					continue
				}
				create, update := false, false
				for _, operation := range rule.Operations {
					create = create || operation == admissionv1.Create || operation == admissionv1.OperationAll
					update = update || operation == admissionv1.Update || operation == admissionv1.OperationAll
				}
				if !create || !update {
					continue
				}
				if config.Service != nil && len(config.CABundle) > 0 && readyEndpoints[config.Service.Namespace+"/"+config.Service.Name] && readyNamespaces[config.Service.Namespace] {
					seen[kind] = true
				}
			}
		}
	}
	for _, cfg := range mutating {
		for _, w := range cfg.Webhooks {
			checkWebhook(w.ClientConfig, w.Rules, mutates)
		}
	}
	for _, cfg := range validating {
		for _, w := range cfg.Webhooks {
			checkWebhook(w.ClientConfig, w.Rules, validates)
		}
	}
	if !anyOperator {
		return true, nil
	}
	var missing []string
	for name, exists := range required {
		if !exists {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		problems = append(problems, fmt.Errorf("partial/incompatible Operator installation: missing SS CRDs %s; automatic replacement is blocked", strings.Join(missing, ", ")))
	}
	if len(readyNamespaces) == 0 {
		problems = append(problems, errors.New("ob-operator Deployment is not ready"))
	}
	for _, kind := range []string{"obclusters", "oblogserviceclusters"} {
		if !mutates[kind] || !validates[kind] {
			problems = append(problems, fmt.Errorf("%s admission webhooks lack a ready Operator backend or CA bundle", kind))
		}
	}
	return false, errors.Join(problems...)
}

func pfContains(values []string, target string) bool {
	for _, v := range values {
		if v == target || v == "*" {
			return true
		}
	}
	return false
}
