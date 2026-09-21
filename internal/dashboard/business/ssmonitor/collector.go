// Package ssmonitor collects SS-specific metrics from the authoritative
// Kubernetes resources and OceanBase system views. Missing samples stay missing.
package ssmonitor

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	ob "github.com/oceanbase/ob-operator/api/v1alpha1"
	obconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	ls "github.com/oceanbase/ob-operator/internal/dashboard/business/logservice"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// IDs verified against OceanBase AI 4.6.2 GV$SYSSTAT. Names, not generic OB
// cache counters, distinguish the SS micro-cache and remote storage pipeline.
var StatIDs = []int{60098, 60099, 60100, 60101, 60102, 60103, 60104, 60105, 60106, 60107, 60108, 60109, 240001, 240002, 240003, 240004, 240009, 240010, 240019, 240022, 240023, 240101, 240102, 240103, 240104}

func statGauge(id int) bool { return id == 240009 || id == 240010 || id == 240019 }

type sample struct {
	ID     int
	Server string
	Tenant int64
	Value  float64
}
type emitter struct {
	bytes.Buffer
	typed map[string]bool
}

func (e *emitter) emit(name, kind string, labels map[string]string, value float64) {
	if !e.typed[name] {
		fmt.Fprintf(&e.Buffer, "# TYPE %s %s\n", name, kind)
		e.typed[name] = true
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+strconv.Quote(labels[k]))
	}
	fmt.Fprintf(&e.Buffer, "%s{%s} %g\n", name, strings.Join(parts, ","), value)
}
func labelsWith(base map[string]string, key, value string) map[string]string {
	m := make(map[string]string, len(base)+1)
	for k, v := range base {
		m[k] = v
	}
	m[key] = value
	return m
}
func boolValue(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
func podReady(p corev1.Pod) bool {
	if p.DeletionTimestamp != nil {
		return false
	}
	for _, c := range p.Status.Conditions {
		if c.Type == corev1.PodReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

// Collect uses request cancellation and bounded SQL/network timeouts. An
// unsuccessful source emits a health metric, never manufactured counter zeros.
func Collect(ctx context.Context, s *ls.Service) (string, error) {
	e := &emitter{typed: map[string]bool{}}
	clusters, err := s.List(ctx, "")
	if err != nil {
		return "", err
	}
	for _, c := range clusters {
		if c.Deleting {
			continue
		}
		labels := map[string]string{"namespace": c.Namespace, "logservice": c.Name}
		pods, err := s.Core.CoreV1().Pods(c.Namespace).List(ctx, metav1.ListOptions{LabelSelector: obconst.LabelRefOBLogServiceCluster + "=" + c.Name})
		e.emit("ss_ls_collection_success", "gauge", labels, boolValue(err == nil))
		if err != nil {
			continue
		}
		desired := 0
		ready := 0
		for _, z := range c.Spec.Topology {
			desired += z.Replica
			count := 0
			for _, p := range pods.Items {
				if p.Labels[obconst.LabelRefOBLogServiceZone] == c.Name+"-"+z.Zone {
					ok := podReady(p)
					if ok {
						count++
					}
					nodeLabels := labelsWith(labelsWith(labels, "zone", z.Zone), "node", p.Name)
					e.emit("ss_ls_node_ready", "gauge", nodeLabels, boolValue(ok))
					for _, cs := range p.Status.ContainerStatuses {
						e.emit("ss_ls_container_restarts_total", "counter", labelsWith(nodeLabels, "container", cs.Name), float64(cs.RestartCount))
					}
				}
			}
			zl := labelsWith(labels, "zone", z.Zone)
			e.emit("ss_ls_desired_replicas", "gauge", zl, float64(z.Replica))
			e.emit("ss_ls_ready_replicas", "gauge", zl, float64(count))
			ready += count
		}
		e.emit("ss_ls_running", "gauge", labels, boolValue(c.Status.Status == "running"))
		e.emit("ss_ls_ready_ratio", "gauge", labels, float64(ready)/float64(desired))
		active, err := queryRegistry(ctx, s, c.Namespace, c.Name)
		e.emit("ss_ls_registry_success", "gauge", labels, boolValue(err == nil))
		if err == nil {
			for _, z := range c.Spec.Topology {
				e.emit("ss_ls_active_nodes", "gauge", labelsWith(labels, "zone", z.Zone), float64(active[z.Zone]))
			}
		}
	}
	obs, err := s.Dynamic.Resource(ls.OBGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, item := range obs.Items {
		c := &ob.OBCluster{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, c); err != nil {
			return "", err
		}
		if string(c.Spec.DeploymentMode) != "shared_storage" || c.DeletionTimestamp != nil {
			continue
		}
		labels := map[string]string{"namespace": c.Namespace, "obcluster": c.Name, "ob_cluster_name": c.Spec.ClusterName}
		samples, err := queryStats(ctx, s, c)
		e.emit("ss_ob_collection_success", "gauge", labels, boolValue(err == nil))
		if err != nil {
			continue
		}
		e.emit("ss_ob_last_success_timestamp_seconds", "gauge", labels, float64(time.Now().Unix()))
		for _, v := range samples {
			sl := labelsWith(labelsWith(labelsWith(labels, "stat_id", strconv.Itoa(v.ID)), "svr_ip", v.Server), "tenant_id", strconv.FormatInt(v.Tenant, 10))
			if statGauge(v.ID) {
				e.emit("ss_ob_stat", "gauge", sl, v.Value)
			} else {
				e.emit("ss_ob_stat_total", "counter", sl, v.Value)
			}
		}
	}
	e.emit("ss_collector_success", "gauge", nil, 1)
	return e.String(), nil
}

func queryRegistry(ctx context.Context, s *ls.Service, ns, name string) (map[string]int, error) {
	nodes, err := s.Dynamic.Resource(ls.NodeGVR).Namespace(ns).List(ctx, metav1.ListOptions{LabelSelector: obconst.LabelRefOBLogServiceCluster + "=" + name})
	if err != nil {
		return nil, err
	}
	for _, item := range nodes.Items {
		n := &ob.OBLogServiceNode{}
		if runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, n) != nil || n.DeletionTimestamp != nil || n.Status.GetConnectAddr() == "" {
			continue
		}
		// Destination comes only from Operator node status, not a user-supplied URL.
		if net.ParseIP(n.Status.GetConnectAddr()) == nil {
			continue
		}
		httpPort := n.Spec.HttpPort
		if httpPort == 0 {
			httpPort = obconst.LogServiceHttpPort
		}
		endpoint := "http://" + net.JoinHostPort(n.Status.GetConnectAddr(), strconv.Itoa(int(httpPort))) + "/manager/ln/show"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			continue
		}
		cl := &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		resp, err := cl.Do(req)
		if err != nil {
			continue
		}
		var result struct {
			Message string `json:"msg"`
			Data    struct {
				Nodes []struct {
					Zone   string `json:"az"`
					Status string `json:"status"`
				} `json:"ln_list"`
			} `json:"data"`
		}
		err = json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&result)
		resp.Body.Close()
		if err != nil || resp.StatusCode != 200 || result.Message != "" || result.Data.Nodes == nil {
			continue
		}
		active := map[string]int{}
		for _, node := range result.Data.Nodes {
			if node.Status == "Active" {
				active[node.Zone]++
			}
		}
		return active, nil
	}
	return nil, fmt.Errorf("LS registry unavailable")
}

func queryStats(ctx context.Context, s *ls.Service, c *ob.OBCluster) ([]sample, error) {
	secret, err := s.Core.CoreV1().Secrets(c.Namespace).Get(ctx, c.Spec.UserSecrets.Root, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("database credentials unavailable")
	}
	password, ok := secret.Data["password"]
	if !ok {
		return nil, fmt.Errorf("database credential key unavailable")
	}
	servers, err := s.Dynamic.Resource(schema.GroupVersionResource{Group: ls.OBGVR.Group, Version: ls.OBGVR.Version, Resource: "observers"}).Namespace(c.Namespace).List(ctx, metav1.ListOptions{LabelSelector: obconst.LabelRefOBCluster + "=" + c.Name})
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(StatIDs))
	for i, id := range StatIDs {
		ids[i] = strconv.Itoa(id)
	}
	query := "SELECT STAT_ID, SVR_IP, CON_ID, VALUE FROM oceanbase.GV$SYSSTAT WHERE STAT_ID IN (" + strings.Join(ids, ",") + ")"
	for _, item := range servers.Items {
		server := &ob.OBServer{}
		if runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, server) != nil || server.Status.Status != "running" {
			continue
		}
		conf := mysql.NewConfig()
		conf.User = "root"
		conf.Passwd = string(password)
		conf.DBName = "oceanbase"
		conf.Net = "tcp"
		conf.Addr = net.JoinHostPort(server.Status.GetConnectAddr(), "2881")
		conf.Timeout = 3 * time.Second
		conf.ReadTimeout = 5 * time.Second
		conf.WriteTimeout = 3 * time.Second
		connector, err := mysql.NewConnector(conf)
		if err != nil {
			continue
		}
		db := sql.OpenDB(connector)
		db.SetMaxOpenConns(1)
		samples, err := readStats(ctx, db, query)
		db.Close()
		if err == nil {
			return samples, nil
		}
		if ctx.Err() != nil {
			break
		}
	}
	return nil, fmt.Errorf("SS system statistics are unavailable")
}
func readStats(ctx context.Context, db *sql.DB, query string) ([]sample, error) {
	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []sample{}
	for rows.Next() {
		v := sample{}
		if err := rows.Scan(&v.ID, &v.Server, &v.Tenant, &v.Value); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("SS statistic IDs not supported")
	}
	return out, nil
}
