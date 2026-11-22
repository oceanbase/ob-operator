package readiness

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// WriteProxySeeds discovers the per-Observer Service addresses advertised by
// SERVICE-mode clusters. A load-balanced Service IP is not a valid OBProxy
// rootserver identity and causes routing failures after topology refresh.
func WriteProxySeeds(parent context.Context, r Reader, o Options, path string) error {
	o.StorageJob = o.Release + "-storage-init"
	if err := o.Validate(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(parent, o.Timeout)
	defer cancel()
	for {
		seeds, err := proxySeeds(ctx, r, o)
		if err == nil && len(seeds) >= o.ObserverReplicas {
			return os.WriteFile(path, []byte(strings.Join(seeds, ";")+"\n"), 0644)
		}
		var ae *APIError
		if errors.As(err, &ae) && (ae.Status == 401 || ae.Status == 403) {
			return fmt.Errorf("OBProxy discovery authorization failed: %w", err)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("OBProxy discovery timed out or was canceled: %w", ctx.Err())
		case <-time.After(o.Interval):
		}
	}
}

func proxySeeds(ctx context.Context, r Reader, o Options) ([]string, error) {
	var list struct {
		Items []struct {
			Metadata struct {
				Labels            map[string]string
				DeletionTimestamp *string
			}
			Spec struct {
				ClusterIP string
				Ports     []struct {
					Name string
					Port int
				}
			}
		}
	}
	selector := "ref-obcluster=" + o.Release + "-ob,ref-observer"
	if err := r.Get(ctx, "/api/v1/namespaces/"+o.Namespace+"/services?labelSelector="+url.QueryEscape(selector), &list); err != nil {
		return nil, err
	}
	unique := map[string]bool{}
	for _, svc := range list.Items {
		if svc.Metadata.DeletionTimestamp != nil || svc.Metadata.Labels["ref-obcluster"] != o.Release+"-ob" || svc.Metadata.Labels["ref-observer"] == "" || net.ParseIP(svc.Spec.ClusterIP) == nil {
			continue
		}
		for _, port := range svc.Spec.Ports {
			if port.Name == "sql" && port.Port > 0 && port.Port <= 65535 {
				unique[net.JoinHostPort(svc.Spec.ClusterIP, strconv.Itoa(port.Port))] = true
			}
		}
	}
	seeds := make([]string, 0, len(unique))
	for seed := range unique {
		seeds = append(seeds, seed)
	}
	sort.Strings(seeds)
	return seeds, nil
}
