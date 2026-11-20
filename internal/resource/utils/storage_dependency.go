package utils

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	StorageInitJobAnnotation        = "oceanbase.oceanbase.com/storage-init-job"
	ResolveStorageServiceAnnotation = "oceanbase.oceanbase.com/resolve-storage-service"
)

// WaitForStorageInitialization gates first bootstrap only. It never creates or
// mutates the Job, credentials, or storage; Helm remains their owner.
func WaitForStorageInitialization(ctx context.Context, reader client.Reader, obj metav1.Object, timeout, interval time.Duration) error {
	name := obj.GetAnnotations()[StorageInitJobAnnotation]
	if name == "" {
		return nil
	}
	if timeout <= 0 || interval <= 0 {
		return fmt.Errorf("storage initialization requires positive timeout and interval")
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := wait.PollUntilContextCancel(ctx, interval, true, func(ctx context.Context) (bool, error) {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		job := &batchv1.Job{}
		if err := reader.Get(ctx, types.NamespacedName{Namespace: obj.GetNamespace(), Name: name}, job); err != nil {
			if apierrors.IsNotFound(err) {
				// Helm replaces revision-scoped Jobs on upgrade. Yield to the
				// task framework so its retry obtains the current CR snapshot.
				return false, fmt.Errorf("storage initialization Job %s is absent; retry with the current resource configuration", name)
			}
			return false, err
		}
		if job.DeletionTimestamp != nil {
			return false, fmt.Errorf("storage initialization Job %s is being replaced; retry with the current resource configuration", name)
		}
		for _, condition := range job.Status.Conditions {
			if condition.Status != corev1.ConditionTrue {
				continue
			}
			if condition.Type == batchv1.JobFailed {
				return false, fmt.Errorf("storage initialization Job %s failed (%s); inspect its logs", name, condition.Reason)
			}
			if condition.Type == batchv1.JobComplete {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("waiting for storage initialization Job %s/%s: %w", obj.GetNamespace(), name, err)
	}
	return nil
}

// ResolveStorageServiceURL makes the LS 1.3 path-style endpoint compatible with
// declarative installs. Only explicitly opted-in, same-namespace HTTP Service
// names are resolved. HTTPS must keep its DNS identity. Parent CR specs are
// never changed; callers use the returned URL solely for runtime configuration.
func ResolveStorageServiceURL(ctx context.Context, reader client.Reader, obj metav1.Object, bucketURL string) (string, error) {
	if obj.GetAnnotations()[ResolveStorageServiceAnnotation] != "true" {
		return bucketURL, nil
	}
	bucket, err := url.Parse(bucketURL)
	if err != nil {
		return "", fmt.Errorf("invalid storage bucket URL")
	}
	query, err := url.ParseQuery(bucket.RawQuery)
	if err != nil {
		return "", fmt.Errorf("invalid storage bucket query")
	}
	hosts := query["host"]
	if len(hosts) != 1 {
		return "", fmt.Errorf("storage bucket URL requires exactly one host parameter")
	}
	endpoint, err := url.Parse(hosts[0])
	if err != nil || endpoint.Hostname() == "" || endpoint.User != nil {
		return "", fmt.Errorf("invalid storage endpoint")
	}
	if endpoint.Scheme != "http" {
		return bucketURL, nil
	}
	host := strings.TrimSuffix(strings.ToLower(endpoint.Hostname()), ".")
	parts := strings.Split(host, ".")
	if !(len(parts) == 3 && parts[2] == "svc" || len(parts) == 5 && strings.Join(parts[2:], ".") == "svc.cluster.local") {
		return bucketURL, nil
	}
	if parts[1] != obj.GetNamespace() {
		return "", fmt.Errorf("storage Service must be in the database namespace")
	}
	service := &corev1.Service{}
	if err := reader.Get(ctx, types.NamespacedName{Namespace: obj.GetNamespace(), Name: parts[0]}, service); err != nil {
		return "", fmt.Errorf("get storage Service %s/%s: %w", obj.GetNamespace(), parts[0], err)
	}
	ip := net.ParseIP(service.Spec.ClusterIP)
	if service.DeletionTimestamp != nil || service.Spec.Type == corev1.ServiceTypeExternalName || ip == nil || ip.To4() == nil {
		return "", fmt.Errorf("storage Service %s requires an available IPv4 ClusterIP", parts[0])
	}
	port := endpoint.Port()
	endpoint.Host = ip.String()
	if port != "" {
		endpoint.Host = net.JoinHostPort(ip.String(), port)
	}
	// Keep other parameter spellings/order, including the unescaped endpoint
	// syntax consumed by OceanBase, rather than re-encoding the entire query.
	segments := strings.Split(bucket.RawQuery, "&")
	for i, segment := range segments {
		key, _, _ := strings.Cut(segment, "=")
		decoded, _ := url.QueryUnescape(key)
		if decoded == "host" {
			segments[i] = "host=" + endpoint.String()
		}
	}
	bucket.RawQuery = strings.Join(segments, "&")
	return bucket.String(), nil
}
