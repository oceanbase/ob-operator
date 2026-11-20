package lakehouse

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

func parseThreeNodes(value string) ([]string, error) {
	if value == "" {
		return nil, nil
	}
	names := strings.Split(value, ",")
	if len(names) != 3 {
		return nil, errors.New("--nodes requires exactly three distinct Kubernetes node names, separated by commas")
	}
	seen := map[string]bool{}
	for i := range names {
		names[i] = strings.TrimSpace(names[i])
		if len(validation.IsDNS1123Subdomain(names[i])) != 0 || seen[names[i]] {
			return nil, errors.New("--nodes requires exactly three distinct valid Kubernetes node names (no empty entries)")
		}
		seen[names[i]] = true
	}
	return names, nil
}

func parseSSHHosts(entries []string) (map[string]string, error) {
	hosts := map[string]string{}
	for _, entry := range entries {
		node, host, ok := strings.Cut(entry, "=")
		if !ok || len(validation.IsDNS1123Subdomain(node)) != 0 || host == "" {
			return nil, errors.New("--ssh-host must be NODE=HOST, using the Kubernetes node name and SSH hostname or IP")
		}
		if _, exists := hosts[node]; exists {
			return nil, fmt.Errorf("duplicate --ssh-host for node %s", node)
		}
		hosts[node] = host
	}
	return hosts, nil
}

// threeHostnames recognizes the simple three-worker layout in saved Helm values.
// No local-only marker is required, so a fresh extracted tar can resume it too.
func threeHostnames(values map[string]interface{}) ([]string, bool) {
	var hosts []string
	global := pfString(pfGet(values, "scheduling", "nodeSelector", corev1.LabelHostname))
	for _, component := range []string{"logService", "observer"} {
		entries, _ := pfGet(values, component, "topology").([]interface{})
		if len(entries) != 3 {
			return nil, false
		}
		current := make([]string, 0, 3)
		seen := map[string]bool{}
		for _, entry := range entries {
			zone := pfMap(entry)
			host := pfString(pfGet(zone, "nodeSelector", corev1.LabelHostname))
			if host == "" {
				host = global
			}
			if host == "" || seen[host] || pfString(zone["replica"]) != "1" || pfString(zone["zone"]) == "" {
				return nil, false
			}
			seen[host] = true
			current = append(current, host)
		}
		if hosts != nil && !reflect.DeepEqual(hosts, current) {
			return nil, false
		}
		hosts = current
	}
	return hosts, true
}

func inferThreeNodes(ctx context.Context, run RunCommand, values map[string]interface{}) ([]string, error) {
	hosts, ok := threeHostnames(values)
	if !ok {
		return nil, nil
	}
	var nodes corev1.NodeList
	if err := getJSON(ctx, run, &nodes, "get", "nodes", "-o", "json"); err != nil {
		return nil, err
	}
	names := make([]string, 0, 3)
	for _, host := range hosts {
		var matches []string
		for _, node := range nodes.Items {
			if node.Labels[corev1.LabelHostname] == host {
				matches = append(matches, node.Name)
			}
		}
		if len(matches) != 1 {
			return nil, fmt.Errorf("saved three-worker hostname %q must match exactly one Kubernetes node; automatic data relocation is unsupported", host)
		}
		names = append(names, matches[0])
	}
	return names, nil
}

func placementValues(values map[string]interface{}) map[string]interface{} {
	placement := map[string]interface{}{"scheduling": values["scheduling"]}
	for _, component := range []string{"observer", "logService"} {
		placement[component] = pfGet(values, component, "topology")
	}
	placement["dashboardSelector"] = pfGet(values, "dashboard", "nodeSelector")
	placement["dashboardTolerations"] = pfGet(values, "dashboard", "tolerations")
	return placement
}

func configureThreeNodes(ctx context.Context, run RunCommand, values, defaults, retainedValues map[string]interface{}, names []string, retained bool) (map[string]interface{}, error) {
	if len(names) != 3 {
		return nil, errors.New("three-worker installation requires exactly three nodes")
	}
	complete := mergeValues(defaults, values)
	if retained && !reflect.DeepEqual(placementValues(complete), placementValues(retainedValues)) {
		return nil, errors.New("changing the placement of a saved or installed release is unsupported; reuse the original three-worker configuration")
	}
	var nodes corev1.NodeList
	if err := getJSON(ctx, run, &nodes, "get", "nodes", "-o", "json"); err != nil {
		return nil, err
	}
	hosts := make([]string, 3)
	seen := map[string]bool{}
	for i, name := range names {
		for _, node := range nodes.Items {
			if node.Name == name {
				hosts[i] = node.Labels[corev1.LabelHostname]
				break
			}
		}
		if hosts[i] == "" || seen[hosts[i]] {
			return nil, fmt.Errorf("node %s must exist and have a distinct kubernetes.io/hostname label", name)
		}
		// A hostname selector must identify exactly one physical worker.
		matches := 0
		for _, node := range nodes.Items {
			if node.Labels[corev1.LabelHostname] == hosts[i] {
				matches++
			}
		}
		if matches != 1 {
			return nil, fmt.Errorf("hostname label %s is ambiguous; fix node labels before installation", hosts[i])
		}
		seen[hosts[i]] = true
	}
	if retained {
		original, ok := threeHostnames(retainedValues)
		if !ok || !reflect.DeepEqual(original, hosts) {
			return nil, errors.New("changing an existing or saved topology to different workers is unsupported; single-worker conversion and node-order changes require a separate migration")
		}
		// The root entry owns placement of the auxiliary components too. Refuse
		// adoption of a different Chart layout instead of moving its PVCs.
		if pfString(pfGet(complete, "scheduling", "nodeSelector", corev1.LabelHostname)) != hosts[0] ||
			(pfGet(complete, "dashboard", "enabled") == true && pfString(pfGet(complete, "dashboard", "nodeSelector", corev1.LabelHostname)) != hosts[0]) {
			return nil, errors.New("saved storage and Dashboard placement must remain on the first worker; use the original Chart workflow for other layouts")
		}
		return values, nil
	}
	out := mergeValues(values, nil)
	if pfGet(out, "dashboard", "nodeSelector") == nil {
		// Match the single-worker entry: inherit pool restrictions when no
		// Dashboard-specific placement was supplied, then pin its hostname.
		setValue(out, mergeValues(pfMap(pfGet(complete, "scheduling", "nodeSelector")), nil), "dashboard", "nodeSelector")
	}
	for _, component := range []string{"logService", "observer"} {
		entries, _ := pfGet(complete, component, "topology").([]interface{})
		if len(entries) != 1 && len(entries) != 3 {
			return nil, fmt.Errorf("--nodes supports one or three template zones for %s; use the advanced Chart for other topologies", component)
		}
		topology := make([]interface{}, 3)
		for i, host := range hosts {
			source := entries[0]
			if len(entries) == 3 {
				source = entries[i]
			}
			entry := mergeValues(pfMap(source), nil)
			entry["zone"], entry["replica"] = fmt.Sprintf("zone%d", i+1), 1
			setValue(entry, host, "nodeSelector", corev1.LabelHostname)
			topology[i] = entry
		}
		setValue(out, topology, component, "topology")
	}
	setValue(out, hosts[0], "scheduling", "nodeSelector", corev1.LabelHostname)
	setValue(out, hosts[0], "dashboard", "nodeSelector", corev1.LabelHostname)
	return out, nil
}
