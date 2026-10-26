package lakehouse

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// multiPlacement describes the desired node of each independently persisted
// component. The mappings must survive Pod deletion: PVC owner names encode zones.
type multiPlacement struct {
	nodes       []corev1.Node
	demand      map[string]pfResources
	zoneTargets map[string]string
	first       string
}

func pfMultiTolerations(raw interface{}, field string) ([]corev1.Toleration, error) {
	var out []corev1.Toleration
	if raw == nil {
		return out, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid %s", field)
	}
	if err = json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("invalid %s", field)
	}
	return out, nil
}

func pfMultiPlacement(opts PreflightOptions, nodes []corev1.Node) (multiPlacement, error) {
	p := multiPlacement{demand: map[string]pfResources{}, zoneTargets: map[string]string{}}
	if len(opts.NodeNames) != 3 {
		return p, errors.New("three-node installation requires exactly three node names")
	}
	selected, hostnames := map[string]bool{}, map[string]bool{}
	uuids, internalIPs := map[string]string{}, map[string]string{}
	for _, name := range opts.NodeNames {
		if selected[name] {
			return p, fmt.Errorf("three-node installation requires three distinct nodes; repeated %s", name)
		}
		selected[name] = true
		var found *corev1.Node
		for i := range nodes {
			if nodes[i].Name == name {
				found = &nodes[i]
				break
			}
		}
		if found == nil {
			return p, fmt.Errorf("requested node %s does not exist", name)
		}
		if !pfReady(*found) {
			return p, fmt.Errorf("requested node %s must be Ready and schedulable", name)
		}
		if found.Status.NodeInfo.OperatingSystem != "linux" || found.Status.NodeInfo.Architecture != "amd64" {
			return p, fmt.Errorf("requested node %s must run linux/amd64; found %s/%s", name, found.Status.NodeInfo.OperatingSystem, found.Status.NodeInfo.Architecture)
		}
		if uuid := strings.ToLower(strings.TrimSpace(found.Status.NodeInfo.SystemUUID)); uuid != "" {
			if other := uuids[uuid]; other != "" {
				return p, fmt.Errorf("nodes %s and %s report the same system UUID; three-node installation requires distinct physical workers", other, name)
			}
			uuids[uuid] = name
		}
		for _, address := range found.Status.Addresses {
			if address.Type != corev1.NodeInternalIP {
				continue
			}
			ip := net.ParseIP(address.Address)
			if ip == nil {
				return p, fmt.Errorf("node %s reports an invalid InternalIP", name)
			}
			if other := internalIPs[ip.String()]; other != "" && other != name {
				return p, fmt.Errorf("nodes %s and %s report the same InternalIP; three-node installation requires distinct physical workers", other, name)
			}
			internalIPs[ip.String()] = name
		}
		hostname := found.Labels["kubernetes.io/hostname"]
		if hostname == "" || hostnames[hostname] {
			return p, fmt.Errorf("requested node %s must have a distinct kubernetes.io/hostname label", name)
		}
		for _, other := range nodes {
			if other.Name != name && other.Labels["kubernetes.io/hostname"] == hostname {
				return p, fmt.Errorf("hostname %s is shared by nodes %s and %s; placement is ambiguous", hostname, name, other.Name)
			}
		}
		hostnames[hostname] = true
		p.nodes = append(p.nodes, *found)
	}
	p.first = p.nodes[0].Name
	global, err := pfSelector(pfGet(opts.Values, "scheduling", "nodeSelector"))
	if err != nil {
		return p, err
	}
	tolerations, err := pfMultiTolerations(pfGet(opts.Values, "scheduling", "tolerations"), "scheduling.tolerations")
	if err != nil {
		return p, err
	}
	for _, component := range []string{"logService", "observer"} {
		resources, err := pfQuantities(pfMap(pfGet(opts.Values, component, "resource")), component+".resource")
		if err != nil {
			return p, err
		}
		topology, ok := pfGet(opts.Values, component, "topology").([]interface{})
		if !ok || len(topology) != 3 {
			return p, fmt.Errorf("three-node installation requires three %s zones", component)
		}
		assigned, zones := map[string]bool{}, map[string]bool{}
		for _, raw := range topology {
			zone := pfMap(raw)
			zoneName := pfString(zone["zone"])
			if zoneName == "" || zones[zoneName] || pfString(zone["replica"]) != "1" {
				return p, fmt.Errorf("three-node installation requires distinct %s zones with one replica each", component)
			}
			zones[zoneName] = true
			selector := map[string]string{}
			for k, v := range global {
				selector[k] = v
			}
			overrides, err := pfSelector(zone["nodeSelector"])
			if err != nil {
				return p, err
			}
			for k, v := range overrides {
				selector[k] = v
			}
			var targets []corev1.Node
			for _, node := range nodes {
				if pfMatches(node, []map[string]string{selector}) {
					targets = append(targets, node)
				}
			}
			if len(targets) != 1 || !selected[targets[0].Name] {
				return p, fmt.Errorf("%s zone %s must select exactly one requested node (found %d)", component, zoneName, len(targets))
			}
			target := targets[0]
			if assigned[target.Name] {
				return p, fmt.Errorf("%s zones must use three distinct requested nodes", component)
			}
			if !pfTolerates(target, tolerations) {
				return p, fmt.Errorf("node %s has a taint not tolerated by %s zone %s", target.Name, component, zoneName)
			}
			assigned[target.Name] = true
			demand := p.demand[target.Name]
			demand.add(resources)
			p.demand[target.Name] = demand
			prefix := opts.Release + "-ls-" + zoneName
			if component == "observer" {
				prefix = opts.Release + "-ob-" + pfString(pfGet(opts.Values, "observer", "clusterId")) + "-" + zoneName
			}
			p.zoneTargets[prefix] = target.Name
		}
	}
	mode := pfString(pfGet(opts.Values, "objectStorage", "mode"))
	if mode != "bundled" && mode != "external" {
		return p, errors.New("objectStorage.mode must be bundled or external")
	}
	createBuckets, _ := pfGet(opts.Values, "objectStorage", "createBuckets").(bool)
	if mode == "bundled" || createBuckets {
		if global["kubernetes.io/hostname"] != p.nodes[0].Labels["kubernetes.io/hostname"] || !pfMatches(p.nodes[0], []map[string]string{global}) || !pfTolerates(p.nodes[0], tolerations) {
			return p, errors.New("storage and initialization placement must match the first requested node and its taints")
		}
		resources, err := pfQuantities(pfMap(pfGet(opts.Values, "initialization", "resources", "requests")), "initialization.resources.requests")
		if err != nil {
			return p, err
		}
		d := p.demand[p.first]
		d.add(resources)
		p.demand[p.first] = d
	}
	if mode == "bundled" {
		resources, err := pfQuantities(pfMap(pfGet(opts.Values, "minio", "resources", "requests")), "minio.resources.requests")
		if err != nil {
			return p, err
		}
		d := p.demand[p.first]
		d.add(resources)
		p.demand[p.first] = d
	}
	if enabled, _ := pfGet(opts.Values, "dashboard", "enabled").(bool); enabled {
		selector, err := pfSelector(pfGet(opts.Values, "dashboard", "nodeSelector"))
		if err != nil {
			return p, err
		}
		dashboardTolerations, err := pfMultiTolerations(pfGet(opts.Values, "dashboard", "tolerations"), "dashboard.tolerations")
		if err != nil {
			return p, err
		}
		if selector["kubernetes.io/hostname"] != p.nodes[0].Labels["kubernetes.io/hostname"] || !pfMatches(p.nodes[0], []map[string]string{selector}) || !pfTolerates(p.nodes[0], dashboardTolerations) {
			return p, errors.New("Dashboard placement must match the first requested node and its taints")
		}
		d := p.demand[p.first]
		d.add(pfResources{700, 2*500*1024*1024 + 500000000})
		p.demand[p.first] = d
	}
	return p, nil
}

func (p multiPlacement) zoneTarget(name string) string {
	// Use the longest zone prefix to keep e.g. zones 'a' and 'a-b' distinct.
	target, longest := "", 0
	for prefix, node := range p.zoneTargets {
		if (name == prefix || strings.HasPrefix(name, prefix+"-")) && len(prefix) > longest {
			target, longest = node, len(prefix)
		}
	}
	return target
}

func (p multiPlacement) podTarget(pod pfPod, opts PreflightOptions) (string, error) {
	for _, key := range []string{"ref-obzone", "ref-oblogservicezone"} {
		if zone := pod.Labels[key]; zone != "" {
			if target := p.zoneTargets[zone]; target != "" {
				return target, nil
			}
			return "", fmt.Errorf("existing Pod %s belongs to an unconfigured zone; automatic relocation is unsupported", pod.Name)
		}
	}
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "OBServer" || owner.Kind == "OBLogServiceNode" {
			if target := p.zoneTarget(owner.Name); target != "" {
				return target, nil
			}
			return "", fmt.Errorf("existing Pod %s belongs to an unconfigured database zone", pod.Name)
		}
	}
	if pod.Labels["ref-obcluster"] == opts.Release+"-ob" || pod.Labels["ref-oblogservicecluster"] == opts.Release+"-ls" {
		if target := p.zoneTarget(pod.Name); target != "" {
			return target, nil
		}
		// Controller maintenance Jobs can carry the cluster label but use an explicit
		// selector for their source node. Keep that placement instead of relocating it.
		if len(pod.Spec.NodeSelector) > 0 {
			var targets []string
			for _, node := range p.nodes {
				if pfMatches(node, []map[string]string{pod.Spec.NodeSelector}) {
					targets = append(targets, node.Name)
				}
			}
			if len(targets) == 1 {
				return targets[0], nil
			}
		}
		return "", fmt.Errorf("cannot safely determine existing Pod %s zone placement", pod.Name)
	}
	return p.first, nil // Chart-owned MinIO, Dashboard and initialization Jobs.
}

func pfMultiStorage(opts PreflightOptions, p multiPlacement, pods []pfPod, claims []corev1.PersistentVolumeClaim, volumes []corev1.PersistentVolume, storageClass string) error {
	targets, supplied, present := map[string]string{}, map[string]bool{}, map[string]bool{}
	classes := map[string]string{}
	if pfString(pfGet(opts.Values, "objectStorage", "mode")) == "bundled" {
		name := pfString(pfGet(opts.Values, "minio", "existingClaim"))
		if name != "" {
			supplied[name] = true
		} else {
			name = opts.Release + "-minio-data"
			classes[name] = storageClass
		}
		targets[name] = p.first
	}
	if enabled, _ := pfGet(opts.Values, "dashboard", "enabled").(bool); enabled {
		for field, suffix := range map[string]string{"prometheus": "prometheus-pvc", "sharedVolume": "shared-volume-pvc"} {
			name := pfString(pfGet(opts.Values, "dashboard", field, "persistence", "existingClaim"))
			if name != "" {
				supplied[name] = true
			} else {
				name = opts.Release + "-" + suffix
				class := pfString(pfGet(opts.Values, "dashboard", field, "persistence", "storageClassName"))
				if class == "" {
					class = storageClass
				}
				classes[name] = class
			}
			targets[name] = p.first
		}
	}
	ownedClaims := map[string]bool{}
	for _, pod := range pods {
		if !pfOwned(pod, opts) {
			continue
		}
		if !opts.Existing {
			return errors.New("resources for this release already exist; resume the existing installation instead of creating a new one")
		}
		target, err := p.podTarget(pod, opts)
		if err != nil {
			return err
		}
		if pfActive(pod) && pod.Spec.NodeName != "" && pod.Spec.NodeName != target {
			return fmt.Errorf("existing Pod %s is on node %s, not requested node %s; automatic migration is unsupported", pod.Name, pod.Spec.NodeName, target)
		}
		for _, v := range pod.Spec.Volumes {
			if v.PersistentVolumeClaim != nil {
				name := v.PersistentVolumeClaim.ClaimName
				if previous := targets[name]; previous != "" && previous != target {
					return fmt.Errorf("PVC %s is used by components with different target nodes", name)
				}
				targets[name] = target
				ownedClaims[name] = true
			}
		}
	}
	for _, claim := range claims {
		if claim.Namespace != opts.Namespace {
			continue
		}
		owned := ownedClaims[claim.Name] || claim.Labels["app.kubernetes.io/instance"] == opts.Release
		if claim.Labels["ref-obcluster"] == opts.Release+"-ob" || claim.Labels["ref-oblogservicecluster"] == opts.Release+"-ls" {
			owned = true
			target := p.zoneTargets[claim.Labels["ref-obzone"]]
			if target == "" {
				target = p.zoneTargets[claim.Labels["ref-oblogservicezone"]]
			}
			if target == "" {
				target = p.zoneTarget(claim.Name)
			}
			if target == "" {
				return fmt.Errorf("PVC %s belongs to an unconfigured database zone; automatic migration is unsupported", claim.Name)
			}
			if previous := targets[claim.Name]; previous != "" && previous != target {
				return fmt.Errorf("PVC %s has conflicting component placement", claim.Name)
			}
			targets[claim.Name], classes[claim.Name] = target, storageClass
		}
		for _, owner := range claim.OwnerReferences {
			if owner.Kind == "OBServer" && strings.HasPrefix(owner.Name, opts.Release+"-ob-") || owner.Kind == "OBLogServiceNode" && strings.HasPrefix(owner.Name, opts.Release+"-ls-") {
				owned = true
				target := p.zoneTarget(owner.Name)
				if target == "" {
					return fmt.Errorf("PVC %s belongs to an unconfigured database zone; automatic migration is unsupported", claim.Name)
				}
				if previous := targets[claim.Name]; previous != "" && previous != target {
					return fmt.Errorf("PVC %s has conflicting component placement", claim.Name)
				}
				targets[claim.Name] = target
				classes[claim.Name] = storageClass
			}
		}
		// Known chart claim names are retained data even if metadata was stripped.
		if targets[claim.Name] != "" && !supplied[claim.Name] {
			owned = true
		}
		if !owned && !supplied[claim.Name] {
			continue
		}
		if !opts.Existing && owned {
			return errors.New("retained PVCs for this release exist; use the existing release configuration")
		}
		if claim.DeletionTimestamp != nil {
			return fmt.Errorf("PVC %s is being deleted", claim.Name)
		}
		target := targets[claim.Name]
		if target == "" {
			return fmt.Errorf("cannot safely determine PVC %s component placement", claim.Name)
		}
		present[claim.Name] = true
		if selected := claim.Annotations["volume.kubernetes.io/selected-node"]; selected != "" && selected != target {
			return fmt.Errorf("PVC %s is bound to node %s, not requested node %s; automatic migration is unsupported", claim.Name, selected, target)
		}
		if expected := classes[claim.Name]; expected != "" && claim.Spec.StorageClassName != nil && *claim.Spec.StorageClassName != expected && !supplied[claim.Name] {
			return fmt.Errorf("PVC %s uses StorageClass %s; changing existing storage is unsupported", claim.Name, *claim.Spec.StorageClassName)
		}
		if claim.Spec.VolumeName != "" {
			found := false
			for _, pv := range volumes {
				if pv.Name == claim.Spec.VolumeName {
					found = true
					if pv.Spec.ClaimRef == nil || pv.Spec.ClaimRef.Namespace != claim.Namespace || pv.Spec.ClaimRef.Name != claim.Name || pv.Spec.ClaimRef.UID != "" && pv.Spec.ClaimRef.UID != claim.UID {
						return fmt.Errorf("PVC %s has inconsistent PV binding", claim.Name)
					}
				}
			}
			if !found {
				return fmt.Errorf("PVC %s references an unavailable PV", claim.Name)
			}
		}
	}
	for name := range supplied {
		if !present[name] {
			return fmt.Errorf("configured existingClaim %s does not exist in namespace %s", name, opts.Namespace)
		}
	}
	for _, pv := range volumes {
		ref := pv.Spec.ClaimRef
		if ref == nil || ref.Namespace != opts.Namespace {
			continue
		}
		target := targets[ref.Name]
		if target == "" {
			target = p.zoneTarget(ref.Name)
		}
		if target == "" {
			continue
		}
		if !present[ref.Name] {
			return fmt.Errorf("retained PV %s references missing PVC %s; recover its original claim before installation", pv.Name, ref.Name)
		}
		var node corev1.Node
		for _, n := range p.nodes {
			if n.Name == target {
				node = n
				break
			}
		}
		if pv.Spec.NodeAffinity != nil && pv.Spec.NodeAffinity.Required != nil && !pfMultiNodeAffinity(node, pv.Spec.NodeAffinity.Required) {
			return fmt.Errorf("PVC %s volume affinity excludes requested node %s; automatic migration is unsupported", ref.Name, target)
		}
	}
	return nil
}

func pfMultiNodeAffinity(node corev1.Node, selector *corev1.NodeSelector) bool {
	for _, term := range selector.NodeSelectorTerms {
		if len(term.MatchExpressions) == 0 && len(term.MatchFields) == 0 {
			continue
		}
		match := true
		for _, req := range term.MatchExpressions {
			if !pfMultiRequirement(node.Labels, req) {
				match = false
			}
		}
		for _, req := range term.MatchFields {
			if !pfMultiRequirement(map[string]string{"metadata.name": node.Name}, req) {
				match = false
			}
		}
		if match {
			return true
		}
	}
	return false
}

func pfMultiRequirement(values map[string]string, req corev1.NodeSelectorRequirement) bool {
	value, exists := values[req.Key]
	contains := false
	for _, v := range req.Values {
		if value == v {
			contains = true
		}
	}
	switch req.Operator {
	case corev1.NodeSelectorOpIn:
		return exists && contains
	case corev1.NodeSelectorOpNotIn:
		return !exists || !contains
	case corev1.NodeSelectorOpExists:
		return exists
	case corev1.NodeSelectorOpDoesNotExist:
		return !exists
	case corev1.NodeSelectorOpGt, corev1.NodeSelectorOpLt:
		if !exists || len(req.Values) != 1 {
			return false
		}
		actual, e1 := strconv.ParseInt(value, 10, 64)
		limit, e2 := strconv.ParseInt(req.Values[0], 10, 64)
		if e1 != nil || e2 != nil {
			return false
		}
		if req.Operator == corev1.NodeSelectorOpGt {
			return actual > limit
		}
		return actual < limit
	default:
		return false
	}
}

func pfCheckMultiCapacity(opts PreflightOptions, nodes []corev1.Node, pods []pfPod, claims []corev1.PersistentVolumeClaim, volumes []corev1.PersistentVolume, result *PreflightResult) error {
	placement, err := pfMultiPlacement(opts, nodes)
	if err != nil {
		return err
	}
	if err = pfMultiStorage(opts, placement, pods, claims, volumes, result.StorageClass); err != nil {
		return err
	}
	var problems []error
	for _, node := range placement.nodes {
		demand := placement.demand[node.Name]
		detail := PreflightResult{NodeName: node.Name, NodeHostname: node.Labels["kubernetes.io/hostname"], NodeRuntimeVersion: node.Status.NodeInfo.ContainerRuntimeVersion, NodeSystemUUID: node.Status.NodeInfo.SystemUUID, StorageClass: result.StorageClass, KubernetesVersion: result.KubernetesVersion, HelmVersion: result.HelmVersion, NodeCount: len(nodes), RequiredCPUMilli: demand.cpu, RequiredMemoryBytes: demand.memory}
		for _, address := range node.Status.Addresses {
			if address.Type == corev1.NodeInternalIP && detail.NodeInternalIP == "" {
				detail.NodeInternalIP = address.Address
			}
			if address.Type == corev1.NodeExternalIP && detail.NodeExternalIP == "" {
				detail.NodeExternalIP = address.Address
			}
		}
		available := pfResourceList(node.Status.Allocatable)
		for _, pod := range pods {
			if !pfActive(pod) || opts.Existing && pfOwned(pod, opts) {
				continue
			}
			if pod.Spec.NodeName != node.Name {
				if pod.Spec.NodeName != "" || !pfMatches(node, []map[string]string{pod.Spec.NodeSelector}) || !pfTolerates(node, pod.Spec.Tolerations) {
					continue
				}
				if pod.Spec.Affinity != nil && pod.Spec.Affinity.NodeAffinity != nil && pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil && !pfMultiNodeAffinity(node, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution) {
					continue
				}
			}
			used := pfPodRequests(pod)
			available.cpu -= used.cpu
			available.memory -= used.memory
		}
		detail.AvailableCPUMilli, detail.AvailableMemoryBytes = available.cpu, available.memory
		result.Nodes = append(result.Nodes, detail)
		if demand.cpu > available.cpu || demand.memory > available.memory {
			problems = append(problems, fmt.Errorf("node %s lacks resources: installation needs %dm CPU / %dMi memory, available after other workloads %dm / %dMi", node.Name, demand.cpu, demand.memory/(1024*1024), available.cpu, available.memory/(1024*1024)))
		}
	}
	details := result.Nodes
	*result = details[0]
	result.Nodes = details
	return errors.Join(problems...)
}
