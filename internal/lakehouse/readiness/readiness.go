// Package readiness implements the read-only native Helm readiness hook.
// It has no kubectl, shell, Helm, cluster-write, or secret-reading dependency.
package readiness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Reader interface {
	Get(context.Context, string, any) error
}

type Options struct {
	Namespace, Release, StorageJob, Dashboard string
	ObserverReplicas, LogServiceReplicas      int
	Timeout, Interval                         time.Duration
}

var dnsName = regexp.MustCompile(`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)

func (o Options) Validate() error {
	for _, name := range []string{o.Namespace, o.Release, o.StorageJob} {
		if !dnsName.MatchString(name) || len(name) > 253 {
			return errors.New("namespace, release and storage-job must be valid Kubernetes names")
		}
	}
	if o.Dashboard != "" && (!dnsName.MatchString(o.Dashboard) || len(o.Dashboard) > 253) {
		return errors.New("dashboard must be a valid Deployment name")
	}
	if o.ObserverReplicas < 1 || o.LogServiceReplicas < 1 {
		return errors.New("expected replica counts must be positive")
	}
	if o.Timeout <= 0 || o.Interval <= 0 {
		return errors.New("timeout and interval must be positive")
	}
	return nil
}

type condition struct{ Type, Status, Reason string }
type metadata struct {
	Name              string
	Generation        int64
	DeletionTimestamp *string
}
type resource struct {
	Metadata metadata
	Status   struct {
		Status                                            string
		Conditions                                        []condition
		ObservedGeneration                                int64
		ReadyReplicas, UpdatedReplicas, AvailableReplicas int
	}
	Spec struct{ Replicas *int }
}
type pod struct {
	Metadata metadata
	Status   struct {
		Phase                                    string
		Conditions                               []condition
		InitContainerStatuses, ContainerStatuses []struct {
			State struct {
				Waiting    *struct{ Reason string }
				Terminated *struct {
					Reason   string
					ExitCode int
				}
			}
		}
	}
}

// Wait reports changes and a bounded heartbeat, and returns only after all
// components are ready. API replies (which may contain credentials) are never
// logged. State names and a small set of Pod failure reasons are safe to print.
func Wait(parent context.Context, reader Reader, o Options, output io.Writer) error {
	if err := o.Validate(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(parent, o.Timeout)
	defer cancel()
	started := time.Now()
	last, lastPrinted := "", time.Time{}
	for {
		detail, ready, err := snapshot(ctx, reader, o)
		if err != nil {
			if ctx.Err() != nil {
				return fmt.Errorf("readiness timed out or was canceled after %s; last state: %s", time.Since(started).Round(time.Second), last)
			}
			var ae *APIError
			if errors.As(err, &ae) && (ae.Status == 401 || ae.Status == 403) {
				return fmt.Errorf("readiness API authorization failed (HTTP %d); check the release readiness Role and RoleBinding", ae.Status)
			}
			detail = "Kubernetes API temporarily unavailable; retrying"
		}
		if ready {
			fmt.Fprintf(output, "Lakehouse ready: LogService and OceanBase running; all database Pods ready%s (%s)\n", dashboardSummary(o), time.Since(started).Round(time.Second))
			return nil
		}
		if strings.HasPrefix(detail, "Storage initialization failed") {
			return errors.New(detail)
		}
		if detail != last || time.Since(lastPrinted) >= 30*time.Second {
			fmt.Fprintf(output, "Waiting: %s (%s elapsed)\n", detail, time.Since(started).Round(time.Second))
			lastPrinted = time.Now()
		}
		last = detail
		timer := time.NewTimer(o.Interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("readiness timed out or was canceled after %s; last state: %s; inspect Pods and storage Job in namespace %s", time.Since(started).Round(time.Second), last, o.Namespace)
		case <-timer.C:
		}
	}
}
func dashboardSummary(o Options) string {
	if o.Dashboard != "" {
		return "; Dashboard ready"
	}
	return ""
}
func getResource(ctx context.Context, r Reader, path string) (resource, bool, error) {
	var v resource
	err := r.Get(ctx, path, &v)
	var ae *APIError
	if errors.As(err, &ae) && ae.Status == 404 {
		return v, false, nil
	}
	return v, err == nil, err
}
func snapshot(ctx context.Context, r Reader, o Options) (string, bool, error) {
	ns := "/namespaces/" + url.PathEscape(o.Namespace) + "/"
	job, exists, err := getResource(ctx, r, "/apis/batch/v1"+ns+"jobs/"+url.PathEscape(o.StorageJob))
	if err != nil {
		return "", false, err
	}
	if !exists {
		return "Storage initialization Job has not been created", false, nil
	}
	if job.Metadata.DeletionTimestamp != nil {
		return "Storage initialization Job is being replaced", false, nil
	}
	complete := false
	for _, c := range job.Status.Conditions {
		if c.Status == "True" && c.Type == "Failed" {
			return "Storage initialization failed; inspect Job " + o.StorageJob + " logs and retry helm upgrade", false, nil
		}
		if c.Status == "True" && c.Type == "Complete" {
			complete = true
		}
	}
	if !complete {
		return "Storage initialization", false, nil
	}
	for _, item := range []struct {
		kind, name, label, title, podLabel string
		replicas                           int
	}{
		{"oblogserviceclusters", o.Release + "-ls", "ref-oblogservicecluster", "LogService", "oblogservice-node", o.LogServiceReplicas},
		{"obclusters", o.Release + "-ob", "ref-obcluster", "OceanBase", "ref-observer", o.ObserverReplicas},
	} {
		cr, exists, err := getResource(ctx, r, "/apis/oceanbase.oceanbase.com/v1alpha1"+ns+item.kind+"/"+url.PathEscape(item.name))
		if err != nil {
			return "", false, err
		}
		if !exists {
			return item.title + " resource has not been created", false, nil
		}
		if cr.Metadata.DeletionTimestamp != nil {
			return item.title + " is being deleted", false, nil
		}
		if cr.Status.Status != "running" {
			return item.title + ": " + safeStatus(cr.Status.Status), false, nil
		}
		var pods struct{ Items []pod }
		if err := r.Get(ctx, "/api/v1"+ns+"pods?labelSelector="+url.QueryEscape(item.label+"="+item.name+","+item.podLabel), &pods); err != nil {
			return "", false, err
		}
		ready, active, reason := 0, 0, ""
		for _, p := range pods.Items {
			// Completed bootstrap Jobs use the same cluster label; they are not
			// database Pods and must neither satisfy nor prevent readiness.
			if p.Metadata.DeletionTimestamp != nil || p.Status.Phase == "Succeeded" || p.Status.Phase == "Failed" {
				continue
			}
			active++
			for _, c := range p.Status.Conditions {
				if c.Type == "Ready" && c.Status == "True" {
					ready++
					break
				}
			}
			for _, c := range append(p.Status.InitContainerStatuses, p.Status.ContainerStatuses...) {
				if c.State.Waiting != nil {
					switch c.State.Waiting.Reason {
					case "ImagePullBackOff", "ErrImagePull", "CrashLoopBackOff", "CreateContainerConfigError":
						reason = c.State.Waiting.Reason
					}
				}
			}
		}
		if ready < item.replicas || ready != active {
			if reason != "" {
				reason = "; " + reason
			}
			return fmt.Sprintf("%s Pods ready %d/%d%s", item.title, ready, item.replicas, reason), false, nil
		}
	}
	if o.Dashboard != "" {
		d, exists, err := getResource(ctx, r, "/apis/apps/v1"+ns+"deployments/"+url.PathEscape(o.Dashboard))
		if err != nil {
			return "", false, err
		}
		desired := 1
		if d.Spec.Replicas != nil {
			desired = *d.Spec.Replicas
		}
		if !exists || d.Metadata.DeletionTimestamp != nil || desired < 1 || d.Status.ObservedGeneration < d.Metadata.Generation || d.Status.UpdatedReplicas < desired || d.Status.ReadyReplicas < desired || d.Status.AvailableReplicas < desired {
			return "Dashboard Deployment and HTTP readiness probes", false, nil
		}
	}
	return "", true, nil
}
func safeStatus(value string) string {
	switch value {
	case "new", "bootstrap", "pending", "running", "failed", "initializing", "creating", "starting":
		return value
	case "":
		return "waiting for Operator"
	default:
		return "initializing"
	}
}

// DecodeJSON enforces a bounded API response to keep the helper small even
// when a namespace contains a large number of Pods.
func DecodeJSON(body io.Reader, out any) error {
	data, err := io.ReadAll(io.LimitReader(body, (8<<20)+1))
	if err != nil {
		return err
	}
	if len(data) > 8<<20 {
		return errors.New("Kubernetes API reply exceeds 8 MiB")
	}
	return json.Unmarshal(data, out)
}
