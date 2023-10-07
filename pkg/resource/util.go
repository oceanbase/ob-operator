/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package resource

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	secretconst "github.com/oceanbase/ob-operator/pkg/const/secret"
	clusterstatus "github.com/oceanbase/ob-operator/pkg/const/status/obcluster"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
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

func GetOceanbaseOperationManagerFromOBCluster(c client.Client, logger *logr.Logger, obcluster *v1alpha1.OBCluster) (*operation.OceanbaseOperationManager, error) {
	logger.Info("Get cluster root client", "obCluster", obcluster)
	return getOperationClient(c, logger, obcluster, oceanbaseconst.OperatorUser, oceanbaseconst.SysTenant, obcluster.Spec.UserSecrets.Operator)
}

func GetTenantOperationClient(c client.Client, logger *logr.Logger, obcluster *v1alpha1.OBCluster, tenantName, credential string) (*operation.OceanbaseOperationManager, error) {
	logger.Info("Get tenant root client", "obCluster", obcluster, "tenantName", tenantName, "credential", credential)
	return getOperationClient(c, logger, obcluster, oceanbaseconst.RootUser, tenantName, credential)
}

func getOperationClient(c client.Client, logger *logr.Logger, obcluster *v1alpha1.OBCluster, userName, tenantName, secretName string) (*operation.OceanbaseOperationManager, error) {
	observerList := &v1alpha1.OBServerList{}
	err := c.List(context.Background(), observerList, client.MatchingLabels{
		oceanbaseconst.LabelRefOBCluster: obcluster.Name,
	}, client.InNamespace(obcluster.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "Get observer list")
	}
	if len(observerList.Items) == 0 {
		return nil, errors.Errorf("No observer belongs to cluster %s", obcluster.Name)
	}

	var s *connector.OceanBaseDataSource
	for _, observer := range observerList.Items {
		address := observer.Status.PodIp
		switch obcluster.Status.Status {
		case clusterstatus.New:
			s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, oceanbaseconst.RootUser, tenantName, "", "")
		case clusterstatus.Bootstrapped:
			s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, oceanbaseconst.RootUser, tenantName, "", oceanbaseconst.DefaultDatabase)
		default:
			// TODO use user operator and read password from secret
			password, err := ReadPassword(c, obcluster.Namespace, secretName)
			if err != nil {
				return nil, errors.Wrapf(err, "Read password to get oceanbase operation manager of cluster %s", obcluster.Name)
			}
			s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, userName, tenantName, password, oceanbaseconst.DefaultDatabase)
		}
		// if err is nil, db connection is already checked available
		oceanbaseOperationManager, err := operation.GetOceanbaseOperationManager(s)
		if err == nil && oceanbaseOperationManager != nil {
			oceanbaseOperationManager.Logger = logger
			return oceanbaseOperationManager, nil
		}
	}
	return nil, errors.Errorf("Can not get oceanbase operation manager of obcluster %s after checked all server", obcluster.Name)
}

func GetJob(c client.Client, namespace string, jobName string) (*batchv1.Job, error) {
	job := &batchv1.Job{}
	err := c.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      jobName,
	}, job)
	return job, err
}

func ExecuteUpgradeScript(c client.Client, logger *logr.Logger, obcluster *v1alpha1.OBCluster, filepath string, extraOpt string) error {
	password, err := ReadPassword(c, obcluster.Namespace, obcluster.Spec.UserSecrets.Root)
	if err != nil {
		return errors.Wrapf(err, "Failed to get root password")
	}
	oceanbaseOperationManager, err := GetOceanbaseOperationManagerFromOBCluster(c, logger, obcluster)
	if err != nil {
		return errors.Wrapf(err, "Get operation manager failed for obcluster %s", obcluster.Name)
	}
	observers, err := oceanbaseOperationManager.ListServers()
	if err != nil {
		return errors.Wrapf(err, "Failed to list all servers for obcluster %s", obcluster.Name)
	}
	var rootserver model.OBServer
	for _, observer := range observers {
		rootserver = observer
		if observer.WithRootserver > 0 {
			logger.Info(fmt.Sprintf("Found rootserver, %s:%d", observer.Ip, observer.Port))
			break
		}
	}

	parts := strings.Split(uuid.New().String(), "-")
	suffix := parts[len(parts)-1]
	jobName := fmt.Sprintf("%s-%s", "script-runner", suffix)
	var backoffLimit int32
	var ttl int32 = 300
	container := corev1.Container{
		Name:    "script-runner",
		Image:   obcluster.Spec.OBServerTemplate.Image,
		Command: []string{"bash", "-c", fmt.Sprintf("python2 %s -h%s -P%d -uroot -p'%s' %s", filepath, rootserver.Ip, rootserver.SqlPort, password, extraOpt)},
	}
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: obcluster.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers:    []corev1.Container{container},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttl,
		},
	}
	logger.Info("Create run upgrade script job", "script", filepath)
	err = c.Create(context.Background(), &job)
	if err != nil {
		return errors.Wrapf(err, "Failed to create run upgrade script job for obcluster %s", obcluster.Name)
	}

	var jobObject *batchv1.Job
	for {
		jobObject, err = GetJob(c, obcluster.Namespace, jobName)
		if err != nil {
			logger.Error(err, "Failed to get job")
			// return errors.Wrapf(err, "Failed to get run upgrade script job for obcluster %s", obcluster.Name)
		}
		if jobObject.Status.Succeeded == 0 && jobObject.Status.Failed == 0 {
			logger.Info("job is still running")
		} else {
			logger.Info("job finished")
			break
		}
		time.Sleep(time.Second * oceanbaseconst.CheckJobInterval)
	}
	if jobObject.Status.Succeeded == 1 {
		logger.Info("job succeeded")
	} else {
		logger.Info("job failed", "job", jobName)
		return errors.Wrap(err, "Failed to run upgrade script job")
	}
	return nil
}

func GetCNIFromAnnotation(pod *corev1.Pod) string {
	_, found := pod.Annotations[oceanbaseconst.AnnotationCalicoValidate]
	if found {
		return oceanbaseconst.CNICalico
	}
	return oceanbaseconst.CNIUnknown
}

func NeedAnnotation(pod *corev1.Pod, cni string) bool {
	switch cni {
	case oceanbaseconst.CNICalico:
		_, found := pod.Annotations[oceanbaseconst.AnnotationCalicoIpAddrs]
		return !found
	default:
		return false
	}
}

func getTenantRestoreSource(ctx context.Context, clt client.Client, logger *logr.Logger, con *operation.OceanbaseOperationManager, ns, tenantCR string) (string, error) {
	var restoreSource string
	var err error

	primary := &v1alpha1.OBTenant{}
	err = clt.Get(ctx, types.NamespacedName{
		Namespace: ns,
		Name:      tenantCR,
	}, primary)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return "", err
		}
	} else {
		// Get ip_list from primary tenant
		aps, err := con.ListTenantAccessPoints(tenantCR)
		if err != nil {
			return "", err
		}
		ipList := make([]string, 0)
		for _, ap := range aps {
			ipList = append(ipList, fmt.Sprintf("%s:%d", ap.SvrIP, ap.SqlPort))
		}
		standbyRoPwd, err := ReadPassword(clt, ns, primary.Status.Credentials.StandbyRO)
		if err != nil {
			logger.Error(err, "Failed to read standby ro password")
			return "", err
		}
		// Set restore source
		restoreSource = fmt.Sprintf("SERVICE=%s USER=%s@%s PASSWORD=%s", strings.Join(ipList, ";"), "standby_ro", primary.Spec.TenantName, standbyRoPwd)
	}

	return restoreSource, nil
}
