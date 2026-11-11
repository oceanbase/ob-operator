// Package logservice exposes the supported Operator lifecycle without accepting
// arbitrary Kubernetes metadata or allowing unsupported template mutations.
package logservice

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
	k8s "github.com/oceanbase/ob-operator/pkg/k8s/client"
)

var ClusterGVR = schema.GroupVersionResource{Group: "oceanbase.oceanbase.com", Version: "v1alpha1", Resource: "oblogserviceclusters"}
var NodeGVR = schema.GroupVersionResource{Group: "oceanbase.oceanbase.com", Version: "v1alpha1", Resource: "oblogservicenodes"}
var OBGVR = schema.GroupVersionResource{Group: "oceanbase.oceanbase.com", Version: "v1alpha1", Resource: "obclusters"}

type Service struct {
	Dynamic dynamic.Interface
	Core    kubernetes.Interface
}

func Default() *Service {
	c := k8s.GetClient()
	return &Service{Dynamic: c.DynamicClient, Core: c.ClientSet}
}

type CreateRequest struct {
	Namespace string                           `json:"namespace"`
	Name      string                           `json:"name"`
	Spec      v1alpha1.OBLogServiceClusterSpec `json:"spec" swaggertype:"object"`
}
type ScaleRequest struct {
	ResourceVersion string         `json:"resourceVersion"`
	Replicas        map[string]int `json:"replicas"`
}
type DeleteRequest struct {
	ResourceVersion string `json:"resourceVersion"`
	ConfirmName     string `json:"confirmName"`
}
type Item struct {
	Namespace       string                             `json:"namespace"`
	Name            string                             `json:"name"`
	ResourceVersion string                             `json:"resourceVersion"`
	CreatedAt       metav1.Time                        `json:"createdAt"`
	Deleting        bool                               `json:"deleting"`
	Protected       bool                               `json:"protected"`
	Spec            v1alpha1.OBLogServiceClusterSpec   `json:"spec" swaggertype:"object"`
	Status          v1alpha1.OBLogServiceClusterStatus `json:"status" swaggertype:"object"`
	References      []string                           `json:"references"`
}
type Node struct {
	Name     string                          `json:"name"`
	Zone     string                          `json:"zone"`
	Status   v1alpha1.OBLogServiceNodeStatus `json:"status" swaggertype:"object"`
	Deleting bool                            `json:"deleting"`
}
type Volume struct {
	Name  string                            `json:"name"`
	Phase corev1.PersistentVolumeClaimPhase `json:"phase" swaggertype:"string"`
	Size  string                            `json:"size"`
}
type Event struct {
	Time    metav1.Time `json:"time"`
	Type    string      `json:"type"`
	Reason  string      `json:"reason"`
	Message string      `json:"message"`
	Object  string      `json:"object"`
}
type Detail struct {
	Item
	Nodes   []Node   `json:"nodes"`
	Volumes []Volume `json:"volumes"`
	Events  []Event  `json:"events"`
}

func apiError(err error) error {
	if err == nil {
		return nil
	}
	if apierrors.IsNotFound(err) {
		return oberr.NewNotFound(err.Error())
	}
	if apierrors.IsConflict(err) || apierrors.IsAlreadyExists(err) {
		return oberr.New(oberr.ErrConflict, "Resource changed or already exists; refresh and retry")
	}
	if apierrors.IsInvalid(err) || apierrors.IsBadRequest(err) || apierrors.IsForbidden(err) {
		return oberr.NewBadRequest(err.Error())
	}
	return oberr.NewInternal("Kubernetes request failed")
}
func decode(u *unstructured.Unstructured) (*v1alpha1.OBLogServiceCluster, error) {
	r := &v1alpha1.OBLogServiceCluster{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, r)
	return r, err
}
func publicItem(r *v1alpha1.OBLogServiceCluster) Item {
	// Never expose embedded URL credentials from objects created outside Dashboard.
	spec := r.DeepCopy().Spec
	if u, err := url.Parse(spec.ObjectStoreURL.BucketURL); err == nil {
		u.User = nil
		// OceanBase's storage URL grammar expects raw http:// in host. Do not
		// round-trip safe URLs through url.Values.Encode (which changes them).
		safe := []string{}
		for _, part := range strings.Split(u.RawQuery, "&") {
			key, value, _ := strings.Cut(part, "=")
			decoded, _ := url.QueryUnescape(value)
			if strings.Contains(decoded, "@") {
				continue
			} // endpoint userinfo is a credential, too
			if key == "host" || key == "s3_region" || key == "endpoint" {
				safe = append(safe, part)
			}
		}
		u.RawQuery = strings.Join(safe, "&")
		spec.ObjectStoreURL.BucketURL = u.String()
	} else {
		spec.ObjectStoreURL.BucketURL = "<invalid URL>"
	}
	for i, p := range spec.Parameters {
		n := strings.ToLower(p.Name)
		if strings.Contains(n, "password") || strings.Contains(n, "secret") || strings.Contains(n, "access_key") || strings.Contains(n, "access_id") {
			spec.Parameters[i].Value = "<redacted>"
		}
	}
	return Item{Namespace: r.Namespace, Name: r.Name, ResourceVersion: r.ResourceVersion, CreatedAt: r.CreationTimestamp, Deleting: r.DeletionTimestamp != nil, Protected: r.Annotations[obconst.AnnotationsIgnoreDeletion] == "true", Spec: spec, Status: r.Status, References: []string{}}
}
func (s *Service) references(ctx context.Context, ns, name string) ([]string, error) {
	list, err := s.Dynamic.Resource(OBGVR).Namespace(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, apiError(err)
	}
	refs := []string{}
	for _, r := range list.Items {
		ref, _, _ := unstructured.NestedString(r.Object, "spec", "logServiceRef", "name")
		if ref == name {
			refs = append(refs, r.GetName())
		}
	}
	sort.Strings(refs)
	return refs, nil
}
func (s *Service) List(ctx context.Context, ns string) ([]Item, error) {
	list, err := s.Dynamic.Resource(ClusterGVR).Namespace(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, apiError(err)
	}
	out := make([]Item, 0, len(list.Items))
	for i := range list.Items {
		r, e := decode(&list.Items[i])
		if e != nil {
			return nil, e
		}
		out = append(out, publicItem(r))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Namespace+"/"+out[i].Name < out[j].Namespace+"/"+out[j].Name })
	return out, nil
}
func (s *Service) Get(ctx context.Context, ns, name string) (*Detail, error) {
	u, err := s.Dynamic.Resource(ClusterGVR).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, apiError(err)
	}
	r, err := decode(u)
	if err != nil {
		return nil, err
	}
	d := &Detail{Item: publicItem(r), Nodes: []Node{}, Volumes: []Volume{}, Events: []Event{}}
	d.References, err = s.references(ctx, ns, name)
	if err != nil {
		return nil, err
	}
	nodes, err := s.Dynamic.Resource(NodeGVR).Namespace(ns).List(ctx, metav1.ListOptions{LabelSelector: obconst.LabelRefOBLogServiceCluster + "=" + name})
	if err != nil {
		return nil, apiError(err)
	}
	names := map[string]bool{name: true}
	for _, u := range nodes.Items {
		n := v1alpha1.OBLogServiceNode{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &n); err != nil {
			return nil, err
		}
		d.Nodes = append(d.Nodes, Node{Name: n.Name, Zone: n.Spec.Zone, Status: n.Status, Deleting: n.DeletionTimestamp != nil})
		names[n.Name] = true
		if n.Status.PodName != "" {
			names[n.Status.PodName] = true
		}
	}
	pvcs, err := s.Core.CoreV1().PersistentVolumeClaims(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, apiError(err)
	}
	for _, v := range pvcs.Items {
		owned := false
		for _, owner := range v.OwnerReferences {
			if names[owner.Name] {
				owned = true
			}
		}
		if owned {
			size := v.Spec.Resources.Requests[corev1.ResourceStorage]
			d.Volumes = append(d.Volumes, Volume{Name: v.Name, Phase: v.Status.Phase, Size: size.String()})
			names[v.Name] = true
		}
	}
	events, err := s.Core.CoreV1().Events(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, apiError(err)
	}
	for _, e := range events.Items {
		if names[e.InvolvedObject.Name] {
			_, last, _ := common.K8sEventTiming(e)
			d.Events = append(d.Events, Event{Time: last, Type: e.Type, Reason: e.Reason, Message: e.Message, Object: e.InvolvedObject.Name})
		}
	}
	sort.Slice(d.Events, func(i, j int) bool { return d.Events[j].Time.Before(&d.Events[i].Time) })
	if len(d.Events) > 50 {
		d.Events = d.Events[:50]
	}
	return d, nil
}

func ValidateCreate(p *CreateRequest) error {
	bad := func(s string) error { return oberr.NewBadRequest(s) }
	if len(validation.IsDNS1123Label(p.Namespace)) > 0 || len(validation.IsDNS1123Subdomain(p.Name)) > 0 || len(p.Name) > 63 {
		return bad("Valid namespace and name (maximum 63 characters) are required")
	}
	v := p.Spec
	if v.ClusterId <= 0 || v.ClusterId > 9007199254740991 {
		return bad("clusterId must be a positive safe integer")
	}
	if v.LogService == nil || v.LogService.Image == "" || v.LogService.Resource.Cpu.Sign() <= 0 || v.LogService.Resource.Memory.Sign() <= 0 {
		return bad("Image, positive CPU and memory are required")
	}
	st := v.LogService.Storage
	if st == nil || st.StoreStorage == nil || st.LogStorage == nil || st.StoreStorage.Size.Sign() <= 0 || st.LogStorage.Size.Sign() <= 0 {
		return bad("Positive storeStorage and logStorage sizes are required")
	}
	if len(v.Topology) == 0 {
		return bad("At least one zone is required")
	}
	seen := map[string]bool{}
	bootstrapReplicas := int64(0)
	for _, z := range v.Topology {
		if len(validation.IsDNS1123Label(z.Zone)) > 0 || seen[z.Zone] || z.Replica < 1 {
			return bad("Zones must be unique DNS labels with at least one replica")
		}
		seen[z.Zone] = true
		// Dashboard currently supports the three-node bootstrap contract, verified
		// with oblogservice:1.3.0. This is creation-only, not a scaling constraint.
		if z.Replica > 3 {
			return bad("Dashboard requires exactly 3 initial LogService nodes for bootstrap; scale replicas after initialization")
		}
		bootstrapReplicas += int64(z.Replica)
		if z.RpcPort < 1 || z.RpcPort > 65535 || z.HttpPort < 1 || z.HttpPort > 65535 || z.RpcPort == z.HttpPort {
			return bad("Distinct RPC and HTTP ports between 1 and 65535 are required")
		}
		if z.Resource != nil && (z.Resource.Cpu.Sign() <= 0 || z.Resource.Memory.Sign() <= 0) {
			return bad("Zone resource overrides must be positive")
		}
	}
	if bootstrapReplicas != 3 {
		return bad("Dashboard requires exactly 3 initial LogService nodes for bootstrap; scale replicas after initialization")
	}
	u, err := url.Parse(v.ObjectStoreURL.BucketURL)
	if err != nil || u.Scheme != "s3" || u.Host == "" || u.User != nil || u.Fragment != "" {
		return bad("Object store URL must be an s3:// bucket URL without inline credentials")
	}
	for k := range u.Query() {
		if k != "host" && k != "s3_region" && k != "endpoint" {
			return bad("Object store URL supports only host, endpoint and s3_region; use a Secret for credentials")
		}
		if strings.Contains(u.Query().Get(k), "@") || strings.ContainsAny(u.Query().Get(k), "\r\n") {
			return bad("Object store endpoints must not contain inline credentials or line breaks")
		}
	}
	if len(validation.IsDNS1123Subdomain(v.ObjectStoreURL.SecretRef.Name)) > 0 {
		return bad("Object store Secret name is required")
	}
	for _, p := range v.Parameters {
		if p.Name == "" || obconst.ContainsManagedParameter(p.Name, p.Value, obconst.LogServiceManagedParameters[:]) {
			return bad("Operator-managed parameters cannot be overridden")
		}
	}
	return nil
}
func (s *Service) Create(ctx context.Context, p *CreateRequest) (*Item, error) {
	if err := ValidateCreate(p); err != nil {
		return nil, err
	}
	if _, err := s.Core.CoreV1().Namespaces().Get(ctx, p.Namespace, metav1.GetOptions{}); err != nil {
		return nil, apiError(err)
	}
	secret, err := s.Core.CoreV1().Secrets(p.Namespace).Get(ctx, p.Spec.ObjectStoreURL.SecretRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, oberr.NewBadRequest("Object storage Secret is unavailable in this namespace")
	}
	if len(secret.Data["access_id"]) == 0 || len(secret.Data["access_key"]) == 0 {
		return nil, oberr.NewBadRequest("Secret must contain non-empty access_id and access_key")
	}
	// Storage classes are checked before any resource is created.
	for _, sc := range []string{p.Spec.LogService.Storage.StoreStorage.StorageClass, p.Spec.LogService.Storage.LogStorage.StorageClass} {
		if sc != "" {
			if _, err := s.Core.StorageV1().StorageClasses().Get(ctx, sc, metav1.GetOptions{}); err != nil {
				return nil, oberr.NewBadRequest("StorageClass not found: " + sc)
			}
		}
	}
	existing, err := s.List(ctx, "")
	if err != nil {
		return nil, err
	}
	for _, r := range existing {
		if r.Spec.ClusterId == p.Spec.ClusterId {
			return nil, oberr.NewBadRequest("clusterId is already in use")
		}
	}
	r := &v1alpha1.OBLogServiceCluster{TypeMeta: metav1.TypeMeta{APIVersion: "oceanbase.oceanbase.com/v1alpha1", Kind: "OBLogServiceCluster"}, ObjectMeta: metav1.ObjectMeta{Name: p.Name, Namespace: p.Namespace, Annotations: map[string]string{obconst.AnnotationsMode: obconst.ModeService}}, Spec: p.Spec}
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(r)
	if err != nil {
		return nil, err
	}
	u, err := s.Dynamic.Resource(ClusterGVR).Namespace(p.Namespace).Create(ctx, &unstructured.Unstructured{Object: obj}, metav1.CreateOptions{})
	if err != nil {
		return nil, apiError(err)
	}
	r, err = decode(u)
	if err != nil {
		return nil, err
	}
	result := publicItem(r)
	return &result, nil
}
func (s *Service) Scale(ctx context.Context, ns, name string, p *ScaleRequest) (*Item, error) {
	u, err := s.Dynamic.Resource(ClusterGVR).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, apiError(err)
	}
	r, err := decode(u)
	if err != nil {
		return nil, err
	}
	if p.ResourceVersion == "" || r.ResourceVersion != p.ResourceVersion {
		return nil, oberr.New(oberr.ErrConflict, "Resource changed; refresh before scaling")
	}
	if r.DeletionTimestamp != nil || r.Status.Status != "running" {
		return nil, oberr.NewBadRequest("Wait for the LogService to be running before scaling")
	}
	if len(p.Replicas) != len(r.Spec.Topology) {
		return nil, oberr.NewBadRequest("Supply every existing zone; adding or removing zones is unsupported")
	}
	for i, z := range r.Spec.Topology {
		n, ok := p.Replicas[z.Zone]
		if !ok || n < 1 {
			return nil, oberr.NewBadRequest("Each existing zone must retain at least one replica")
		}
		r.Spec.Topology[i].Replica = n
	}
	// Preserve all unrelated metadata/spec/status and use resourceVersion for an
	// atomic optimistic update; never retry a stale user's scale intention.
	topology, err := runtime.DefaultUnstructuredConverter.ToUnstructured(r)
	if err != nil {
		return nil, err
	}
	u.Object["spec"] = topology["spec"]
	result, err := s.Dynamic.Resource(ClusterGVR).Namespace(ns).Update(ctx, u, metav1.UpdateOptions{})
	if err != nil {
		return nil, apiError(err)
	}
	r, err = decode(result)
	if err != nil {
		return nil, err
	}
	item := publicItem(r)
	return &item, nil
}
func (s *Service) Delete(ctx context.Context, ns, name string, p *DeleteRequest) (bool, error) {
	u, err := s.Dynamic.Resource(ClusterGVR).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, apiError(err)
	}
	if p.ConfirmName != name || p.ResourceVersion == "" {
		return false, oberr.NewBadRequest("Confirm the exact name and resourceVersion before deleting")
	}
	if p.ResourceVersion != u.GetResourceVersion() {
		return false, oberr.New(oberr.ErrConflict, "Resource changed; refresh before deleting")
	}
	if u.GetAnnotations()[obconst.AnnotationsIgnoreDeletion] == "true" {
		return false, oberr.NewBadRequest("LogService is protected from deletion")
	}
	refs, err := s.references(ctx, ns, name)
	if err != nil {
		return false, err
	}
	if len(refs) > 0 {
		return false, oberr.NewBadRequest(fmt.Sprintf("LogService is referenced by OBClusters: %s", strings.Join(refs, ", ")))
	}
	uid := u.GetUID()
	policy := metav1.DeletePropagationForeground
	err = s.Dynamic.Resource(ClusterGVR).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &uid, ResourceVersion: &p.ResourceVersion}, PropagationPolicy: &policy})
	return err == nil, apiError(err)
}
