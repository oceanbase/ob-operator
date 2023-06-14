package resource

import (
	"context"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	secretconst "github.com/oceanbase/ob-operator/pkg/const/secret"
	clusterstatus "github.com/oceanbase/ob-operator/pkg/const/status/obcluster"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReadPassword(c client.Client, namespace, secretName string) (string, error) {
	secret := &corev1.Secret{}
	err := c.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      secretName,
	}, secret)
	if err != nil {
		return "", errors.Wrapf(err, "Get password from secret %s failed", secretName)
	}
	return string(secret.Data[secretconst.PasswordKeyName]), err
}

func GetOceanbaseOperationManagerFromOBCluster(c client.Client, obcluster *v1alpha1.OBCluster) (*operation.OceanbaseOperationManager, error) {
	observerList := &v1alpha1.OBServerList{}
	err := c.List(context.Background(), observerList, client.MatchingLabels{
		oceanbaseconst.LabelRefOBCluster: obcluster.Name,
	}, client.InNamespace(obcluster.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "Get observer list")
	}
	if len(observerList.Items) <= 0 {
		return nil, errors.Wrapf(err, "No observer belongs to cluster %s", obcluster.Name)
	}

	var s *connector.OceanBaseDataSource
	for _, observer := range observerList.Items {
		address := observer.Status.PodIp
		switch obcluster.Status.Status {
		case clusterstatus.New:
			s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, oceanbaseconst.RootUser, oceanbaseconst.SysTenant, "", "")
		case clusterstatus.Bootstrapped:
			s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, oceanbaseconst.RootUser, oceanbaseconst.SysTenant, "", oceanbaseconst.DefaultDatabase)
		default:
			// TODO use user operator and read password from secret
			password, err := ReadPassword(c, obcluster.Namespace, obcluster.Spec.UserSecrets.Operator)
			if err != nil {
				return nil, errors.Wrapf(err, "Get oceanbase operation manager of cluster %s", obcluster.Name)
			}
			s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, oceanbaseconst.OperatorUser, oceanbaseconst.SysTenant, password, oceanbaseconst.DefaultDatabase)
		}
		// if err is nil, db connection is already checked available
		oceanbaseOperationManager, err := operation.GetOceanbaseOperationManager(s)
		if err == nil {
			return oceanbaseOperationManager, nil
		}
	}
	return nil, errors.Errorf("Can not get oceanbase operation manager of obcluster %s after checked all server", obcluster.Name)
}
