package obzone

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
)

const replacementFor = "oceanbase.oceanbase.com/replaces-observer-uid"

func (m *OBZoneManager) ownsRollingServer(s *v1alpha1.OBServer) bool {
	for _, owner := range s.OwnerReferences {
		if m.OBZone.UID != "" && s.UID != "" && owner.UID == m.OBZone.UID && owner.Name == m.OBZone.Name && owner.Kind == "OBZone" {
			return true
		}
	}
	return false
}

// Finish an interrupted exchange before allocating a spare for another server.
func prioritizePendingReplacements(servers []v1alpha1.OBServer) {
	pending := make(map[string]bool)
	for _, s := range servers {
		if uid := s.Annotations[replacementFor]; uid != "" {
			pending[uid] = true
		}
	}
	sort.SliceStable(servers, func(i, j int) bool { return pending[string(servers[i].UID)] && !pending[string(servers[j].UID)] })
}

func (m *OBZoneManager) getRollingReplacement(old *v1alpha1.OBServer) (*v1alpha1.OBServer, error) {
	servers, err := m.listOBServers()
	if err != nil {
		return nil, err
	}
	var legacy, linked *v1alpha1.OBServer
	var legacyWaiting int
	oldPresent := false
	for i := range servers.Items {
		s := &servers.Items[i]
		if !m.ownsRollingServer(s) || s.DeletionTimestamp != nil {
			return nil, errors.Errorf("observer %s is deleting or has a different owner", s.Name)
		}
		if s.Name == old.Name {
			if s.UID != old.UID {
				return nil, errors.Errorf("observer %s was recreated", old.Name)
			}
			oldPresent = true
		}
		if s.Annotations[replacementFor] == string(old.UID) {
			if linked != nil {
				return nil, errors.Errorf("multiple replacements claim observer %s", old.Name)
			}
			linked = s
		}
		if s.Annotations[replacementFor] != "" || m.checkIfStorageClassChanged(s) {
			continue
		}
		if s.Status.Status != serverstatus.Running {
			legacyWaiting++
		}
		// Legacy releases created an unmarked spare. Prefer any that is still
		// starting, otherwise choose the newest matching, same-zone server.
		if s.Status.Status != serverstatus.Running || legacy == nil {
			legacy = s
		} else if legacy.Status.Status == serverstatus.Running && s.CreationTimestamp.After(legacy.CreationTimestamp.Time) {
			legacy = s
		}
	}
	if !oldPresent {
		return nil, nil
	}
	if linked != nil {
		return linked, nil
	}
	if surplus := len(servers.Items) - m.OBZone.Spec.Topology.Replica; surplus > 0 {
		if surplus != 1 || legacy == nil || legacyWaiting > 1 {
			return nil, errors.Errorf("cannot identify a single replacement for observer %s in zone %s", old.Name, m.OBZone.Name)
		}
		copy := legacy.DeepCopy()
		if legacy.Annotations == nil {
			legacy.Annotations = make(map[string]string)
		}
		legacy.Annotations[replacementFor] = string(old.UID)
		err = m.Client.Patch(m.Ctx, legacy, client.MergeFromWithOptions(copy, client.MergeFromWithOptimisticLock{}))
		return legacy, errors.Wrapf(err, "record legacy replacement %s for %s", legacy.Name, old.Name)
	}
	// Keep the DNS label length within 63 even for a long zone name. The
	// full source UID and owner are also verified before every deletion.
	sum := sha256.Sum256([]byte(old.UID))
	name := fmt.Sprintf("%.50s-%x", m.OBZone.Name, sum[:6])
	s, err := m.createOneOBServer(name, old.UID)
	if kubeerrors.IsAlreadyExists(err) {
		s = &v1alpha1.OBServer{}
		err = m.Client.Get(m.Ctx, m.generateNamespacedName(name), s)
	}
	return s, errors.Wrapf(err, "obtain replacement %s for observer %s", name, old.Name)
}

func (m *OBZoneManager) waitForRollingReplacement(target *v1alpha1.OBServer, oldUID types.UID) error {
	name := target.Name
	timeout := time.Duration(obcfg.GetConfig().Time.DefaultStateWaitTimeout) * time.Second
	err := wait.PollUntilContextTimeout(m.Ctx, time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		s := &v1alpha1.OBServer{}
		if err := m.Client.Get(ctx, m.generateNamespacedName(name), s); err != nil {
			return false, err
		}
		if s.UID != target.UID || !m.ownsRollingServer(s) || s.Annotations[replacementFor] != string(oldUID) || s.DeletionTimestamp != nil || m.checkIfStorageClassChanged(s) {
			return false, errors.Errorf("replacement %s identity, owner or storage class changed", name)
		}
		return s.Status.Status == serverstatus.Running, nil
	})
	return errors.Wrapf(err, "wait for replacement %s of observer UID %s", name, oldUID)
}
