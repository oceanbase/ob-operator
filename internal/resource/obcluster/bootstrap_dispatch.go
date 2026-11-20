package obcluster

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	obconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	dbconfig "github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/config"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

// Only the worker that atomically records the claim may invoke bootstrap SQL.
// If it disappears before the claim, a new worker can repeat the preparation.
// A crash between claiming and sending (including during SQL session setup)
// is indistinguishable from a lost response. Subsequent workers only observe.
func (m *OBClusterManager) dispatchClassicBootstrap(prepare func() (func() error, error), observe func() error) error {
	_, claimed, err := m.currentClassicBootstrap(m.Ctx)
	if err != nil {
		return err
	}
	if claimed {
		return observe()
	}
	send, err := prepare()
	if err != nil {
		return err
	}
	var won bool
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cluster, claimed, err := m.currentClassicBootstrap(m.Ctx)
		if err != nil || claimed {
			return err
		}
		before := cluster.DeepCopy()
		if cluster.Annotations == nil {
			cluster.Annotations = make(map[string]string)
		}
		cluster.Annotations[obconst.AnnotationsClassicBootstrapClaim] = cluster.ClassicBootstrapDispatchClaim()
		err = m.Client.Patch(m.Ctx, cluster, client.MergeFromWithOptions(before, client.MergeFromWithOptimisticLock{}))
		won = err == nil
		return err
	})
	if err != nil {
		return err
	}
	if !won {
		return observe()
	}
	if err := send(); err != nil {
		// Preserve the SQL failure if the database does not prove completion.
		if checkErr := observe(); checkErr != nil {
			return errors.Join(err, checkErr)
		}
	}
	return nil
}

func (m *OBClusterManager) currentClassicBootstrap(ctx context.Context) (*v1alpha1.OBCluster, bool, error) {
	reader := m.APIReader
	if reader == nil {
		reader = m.Client
	}
	current := &v1alpha1.OBCluster{}
	if err := reader.Get(ctx, client.ObjectKeyFromObject(m.OBCluster), current); err != nil {
		return nil, false, err
	}
	expected := current.ClassicBootstrapDispatchClaim()
	if expected == "" || current.UID != m.OBCluster.UID {
		return nil, false, errors.New("cluster identity, generation or bootstrap task changed before dispatch")
	}
	value, claimed := current.Annotations[obconst.AnnotationsClassicBootstrapClaim]
	if claimed {
		uid, generation, ok := strings.Cut(value, "/")
		gen, err := strconv.ParseInt(generation, 10, 64)
		if !ok || uid != string(current.UID) || err != nil || gen < 1 || gen > current.Generation || strconv.FormatInt(gen, 10) != generation {
			return nil, false, errors.New("invalid claim for the current cluster identity or generation")
		}
	} else if current.Generation != m.OBCluster.Generation {
		return nil, false, errors.New("cluster generation changed before bootstrap dispatch")
	}
	return current, claimed, nil
}

func (m *OBClusterManager) waitForDispatchedBootstrapResult() error {
	return wait.PollUntilContextTimeout(m.Ctx, 5*time.Second, dbconfig.BootstrapTimeout, true, func(ctx context.Context) (bool, error) {
		live, claimed, err := m.currentClassicBootstrap(ctx)
		if err != nil {
			return false, err
		}
		if !claimed {
			return false, errors.New("bootstrap dispatch claim disappeared during verification")
		}
		// A later generation may change unrelated settings. Verify the physical
		// identity against the live desired topology without issuing SQL again.
		probe := *m
		probe.OBCluster = live
		return probe.verifyBootstrapDatabase(ctx, live)
	})
}

func (m *OBClusterManager) waitForBootstrapResult() error {
	return wait.PollUntilContextTimeout(m.Ctx, 5*time.Second, dbconfig.BootstrapTimeout, true, m.bootstrapResultReady)
}

func (m *OBClusterManager) prepareBootstrap() (*operation.OceanbaseOperationManager, []model.BootstrapServerInfo, error) {
	zones, err := m.readyBootstrapZones(m.Ctx, serverstatus.BootstrapReady)
	if err != nil {
		return nil, nil, fmt.Errorf("read bootstrap zones: %w", err)
	}
	if len(zones) == 0 {
		return nil, nil, errors.New("cluster no longer has a complete set of bootstrap servers")
	}
	var manager *operation.OceanbaseOperationManager
	for attempt := 0; attempt < obcfg.GetConfig().Time.GetConnectionMaxRetries && manager == nil; attempt++ {
		for _, zone := range zones {
			for _, server := range zone.Status.OBServerStatus {
				manager, err = m.bootstrapClient(server.GetConnectAddr())
				if err == nil && manager != nil {
					break
				}
			}
			if manager != nil {
				break
			}
		}
		if manager == nil {
			select {
			case <-m.Ctx.Done():
				return nil, nil, m.Ctx.Err()
			case <-time.After(time.Duration(obcfg.GetConfig().Time.CheckConnectionInterval) * time.Second):
			}
		}
	}
	if manager == nil {
		return nil, nil, errors.New("cannot connect to any of the cluster's bootstrap servers")
	}
	standalone := m.OBCluster.Annotations[obconst.AnnotationsMode] == obconst.ModeStandalone
	return manager, bootstrapSeeds(zones, manager.Connector.DataSource().GetAddress(), standalone), nil
}
