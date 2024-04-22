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

package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	secretconst "github.com/oceanbase/ob-operator/internal/const/secret"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	k8sclient "github.com/oceanbase/ob-operator/pkg/k8s/client"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
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

func GetSysOperationClient(c client.Client, logger *logr.Logger, obcluster *v1alpha1.OBCluster) (*operation.OceanbaseOperationManager, error) {
	logger.V(oceanbaseconst.LogLevelTrace).Info("Get cluster sys client", "obCluster", obcluster)
	var manager *operation.OceanbaseOperationManager
	var err error
	_, migrateAnnoExist := GetAnnotationField(obcluster, oceanbaseconst.AnnotationsSourceClusterAddress)
	if migrateAnnoExist && obcluster.Status.Status == clusterstatus.MigrateFromExisting {
		manager, err = getSysClientFromSourceCluster(c, logger, obcluster, oceanbaseconst.RootUser, oceanbaseconst.SysTenant, obcluster.Spec.UserSecrets.Root)
	} else {
		manager, err = getSysClient(c, logger, obcluster, oceanbaseconst.OperatorUser, oceanbaseconst.SysTenant, obcluster.Spec.UserSecrets.Operator)
	}
	return manager, err
}

func GetTenantRootOperationClient(c client.Client, logger *logr.Logger, obcluster *v1alpha1.OBCluster, tenantName, credential string) (*operation.OceanbaseOperationManager, error) {
	logger.V(oceanbaseconst.LogLevelTrace).Info("Get tenant root client", "obCluster", obcluster, "tenantName", tenantName, "credential", credential)
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
	var password string
	if credential != "" {
		password, err = ReadPassword(c, obcluster.Namespace, credential)
		if err != nil {
			return nil, errors.Wrapf(err, "Read password to get oceanbase operation manager of cluster %s", obcluster.Name)
		}
	}

	var s *connector.OceanBaseDataSource
	for _, observer := range observerList.Items {
		address := observer.Status.GetConnectAddr()
		switch obcluster.Status.Status {
		case clusterstatus.New:
			return nil, errors.New("Cluster is not bootstrapped")
		case clusterstatus.Bootstrapped:
			return nil, errors.New("Cluster is not initialized")
		default:
			s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, oceanbaseconst.RootUser, tenantName, password, oceanbaseconst.DefaultDatabase)
		}
		// if err is nil, db connection is already checked available
		rootClient, err := operation.GetOceanbaseOperationManager(s)
		if err == nil && rootClient != nil {
			rootClient.Logger = logger
			return rootClient, nil
		}
		// err is not nil, try to use empty password
		s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, oceanbaseconst.RootUser, tenantName, "", oceanbaseconst.DefaultDatabase)
		rootClient, err = operation.GetOceanbaseOperationManager(s)
		if err == nil && rootClient != nil {
			rootClient.Logger = logger
			return rootClient, nil
		}
	}
	return nil, errors.Errorf("Can not get root operation client of tenant %s in obcluster %s after checked all servers", tenantName, obcluster.Name)
}

func getSysClientFromSourceCluster(c client.Client, logger *logr.Logger, obcluster *v1alpha1.OBCluster, userName, tenantName, secretName string) (*operation.OceanbaseOperationManager, error) {
	sysClient, err := getSysClient(c, logger, obcluster, userName, tenantName, secretName)
	if err == nil {
		return sysClient, nil
	}
	password, err := ReadPassword(c, obcluster.Namespace, secretName)
	if err != nil {
		return nil, errors.Wrapf(err, "Read password to get oceanbase operation manager of cluster %s", obcluster.Name)
	}
	// when obcluster is under migrating, use address from annotation
	migrateAnnoVal, _ := GetAnnotationField(obcluster, oceanbaseconst.AnnotationsSourceClusterAddress)
	servers := strings.Split(migrateAnnoVal, ";")
	for _, server := range servers {
		addressParts := strings.Split(server, ":")
		if len(addressParts) != 2 {
			return nil, errors.New("Parse oceanbase cluster connect address failed")
		}
		sqlPort, err := strconv.ParseInt(addressParts[1], 10, 64)
		if err != nil {
			return nil, errors.New("Parse sql port of obcluster failed")
		}
		s := connector.NewOceanBaseDataSource(addressParts[0], sqlPort, userName, tenantName, password, oceanbaseconst.DefaultDatabase)
		// if err is nil, db connection is already checked available
		sysClient, err := operation.GetOceanbaseOperationManager(s)
		if err == nil && sysClient != nil {
			sysClient.Logger = logger
			return sysClient, nil
		}
		logger.Error(err, "Get operation manager from existing obcluster")
	}
	return nil, errors.Errorf("Failed to get sys client from existing obcluster, address: %s", migrateAnnoVal)
}

func getSysClient(c client.Client, logger *logr.Logger, obcluster *v1alpha1.OBCluster, userName, tenantName, secretName string) (*operation.OceanbaseOperationManager, error) {
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
	password, err := ReadPassword(c, obcluster.Namespace, secretName)
	if err != nil {
		return nil, errors.Wrapf(err, "Read password to get oceanbase operation manager of cluster %s", obcluster.Name)
	}
	for _, observer := range observerList.Items {
		address := observer.Status.GetConnectAddr()
		switch obcluster.Status.Status {
		case clusterstatus.New:
			s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, oceanbaseconst.RootUser, tenantName, "", "")
		case clusterstatus.Bootstrapped:
			s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, oceanbaseconst.RootUser, tenantName, "", oceanbaseconst.DefaultDatabase)
		default:
			s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, userName, tenantName, password, oceanbaseconst.DefaultDatabase)
		}
		// if err is nil, db connection is already checked available
		sysClient, err := operation.GetOceanbaseOperationManager(s)
		if err == nil && sysClient != nil {
			sysClient.Logger = logger
			return sysClient, nil
		}
	}
	return nil, errors.Errorf("Can not get oceanbase operation manager of obcluster %s after checked all servers", obcluster.Name)
}

func GetJob(ctx context.Context, c client.Client, namespace string, jobName string) (*batchv1.Job, error) {
	job := &batchv1.Job{}
	err := c.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      jobName,
	}, job)
	return job, err
}

func RunJob(ctx context.Context, c client.Client, logger *logr.Logger, namespace string, jobName string, image string, cmd string) (output string, exitCode int32, err error) {
	fullJobName := fmt.Sprintf("%s-%s", jobName, rand.String(6))
	var backoffLimit int32
	var ttl int32 = 300
	container := corev1.Container{
		Name:    "job-runner",
		Image:   image,
		Command: []string{"bash", "-c", cmd},
	}
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fullJobName,
			Namespace: namespace,
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

	err = c.Create(ctx, &job)
	if err != nil {
		return "", exitCode, errors.Wrapf(err, "failed to create job of image: %s", image)
	}

	var jobObject *batchv1.Job
	for i := 0; i < oceanbaseconst.CheckJobMaxRetries; i++ {
		jobObject, err = GetJob(ctx, c, namespace, fullJobName)
		if err != nil {
			logger.Error(err, "Failed to get job")
			// return errors.Wrapf(err, "Failed to get run upgrade script job for obcluster %s", obcluster.Name)
		}
		if jobObject.Status.Succeeded == 0 && jobObject.Status.Failed == 0 {
			logger.V(oceanbaseconst.LogLevelDebug).Info("Job is still running")
		} else {
			logger.V(oceanbaseconst.LogLevelDebug).Info("Job finished")
			break
		}
		time.Sleep(time.Second * oceanbaseconst.CheckJobInterval)
	}
	clientSet := k8sclient.GetClient()
	podList, err := clientSet.ClientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", fullJobName),
	})
	if err != nil || len(podList.Items) == 0 {
		return "", 1, errors.Wrapf(err, "failed to get pods of job %s", jobName)
	}
	var outputBuffer bytes.Buffer
	podLogOpts := corev1.PodLogOptions{}
	pod := podList.Items[0]
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == "job-runner" {
			exitCode = cs.State.Terminated.ExitCode
		}
	}
	if jobObject.Status.Succeeded == 1 {
		logger.V(oceanbaseconst.LogLevelDebug).Info("Job succeeded", "job", fullJobName)
		res := clientSet.ClientSet.CoreV1().Pods(namespace).GetLogs(pod.Name, &podLogOpts)
		logs, err := res.Stream(ctx)
		if err != nil {
			logger.Error(err, "Failed to get job logs")
		} else {
			defer logs.Close()
			_, err = io.Copy(&outputBuffer, logs)
			if err != nil {
				logger.Error(err, "Failed to copy logs")
			}
			output = outputBuffer.String()
		}
	} else {
		logger.V(oceanbaseconst.LogLevelDebug).Info("Job failed", "job", fullJobName)
		return "", exitCode, errors.Wrapf(err, "Failed to run job %s", fullJobName)
	}
	return output, exitCode, nil
}

func ExecuteUpgradeScript(ctx context.Context, c client.Client, logger *logr.Logger, obcluster *v1alpha1.OBCluster, filepath string, extraOpt string) error {
	password, err := ReadPassword(c, obcluster.Namespace, obcluster.Spec.UserSecrets.Root)
	if err != nil {
		return errors.Wrapf(err, "Failed to get root password")
	}
	oceanbaseOperationManager, err := GetSysOperationClient(c, logger, obcluster)
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

	jobName := fmt.Sprintf("%s-%s", "script-runner", rand.String(6))
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
			OwnerReferences: []metav1.OwnerReference{{
				Kind:       obcluster.Kind,
				APIVersion: obcluster.APIVersion,
				Name:       obcluster.Name,
				UID:        obcluster.UID,
			}},
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
	err = c.Create(ctx, &job)
	if err != nil {
		return errors.Wrapf(err, "Failed to create run upgrade script job for obcluster %s", obcluster.Name)
	}

	var jobObject *batchv1.Job
	check := func() (bool, error) {
		jobObject, err = GetJob(ctx, c, obcluster.Namespace, jobName)
		if err != nil {
			return false, errors.Wrapf(err, "Failed to get run upgrade script job for obcluster %s", obcluster.Name)
		}
		if jobObject.Status.Succeeded == 0 && jobObject.Status.Failed == 0 {
			logger.V(oceanbaseconst.LogLevelDebug).Info("Job is still running")
			return false, nil
		} else if jobObject.Status.Succeeded == 1 {
			logger.V(oceanbaseconst.LogLevelDebug).Info("Job succeeded")
			return true, nil
		} else {
			logger.V(oceanbaseconst.LogLevelDebug).Info("Job failed", "job", jobName)
			return false, errors.Wrap(err, "Failed to run upgrade script job")
		}
	}
	err = CheckJobWithTimeout(check, time.Second*oceanbaseconst.WaitForJobTimeoutSeconds)
	if err != nil {
		return errors.Wrap(err, "Failed to wait for job to finish")
	}
	return nil
}

type CheckJobFunc func() (bool, error)

// CheckJobWithTimeout checks job with timeout, return error if timeout or job failed.
// First parameter is the function to check job status, return true if job finished, false if not.
// Second parameter is the timeout duration, default to 1800s.
// Third parameter is the interval to check job status, default to 3s.
func CheckJobWithTimeout(f CheckJobFunc, ds ...time.Duration) error {
	timeout := time.Second * oceanbaseconst.DefaultStateWaitTimeout
	interval := time.Second * oceanbaseconst.CheckJobInterval
	if len(ds) > 0 {
		timeout = ds[0]
	}
	if len(ds) > 1 {
		interval = ds[1]
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			return errors.New("Timeout to wait for job")
		default:
			time.Sleep(interval)
			finished, err := f()
			if err != nil {
				return err
			}
			if finished {
				return nil
			}
		}
	}
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

// GetTenantRestoreSource gets restore source from tenant CR. If tenantCR is in form of ns/name, the parameter ns is ignored.
func GetTenantRestoreSource(ctx context.Context, clt client.Client, logger *logr.Logger, ns, tenantCR string) (string, error) {
	finalNs := ns
	finalTenantCR := tenantCR
	splits := strings.Split(tenantCR, "/")
	if len(splits) == 2 {
		finalNs = splits[0]
		finalTenantCR = splits[1]
	}
	var restoreSource string
	var err error

	primary := &v1alpha1.OBTenant{}
	err = clt.Get(ctx, types.NamespacedName{
		Namespace: finalNs,
		Name:      finalTenantCR,
	}, primary)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return "", err
		}
	} else {
		obcluster := &v1alpha1.OBCluster{}
		err := clt.Get(ctx, types.NamespacedName{
			Namespace: finalNs,
			Name:      primary.Spec.ClusterName,
		}, obcluster)
		if err != nil {
			return "", errors.Wrap(err, "get obcluster")
		}
		con, err := GetSysOperationClient(clt, logger, obcluster)
		if err != nil {
			return "", errors.Wrap(err, "get oceanbase operation manager")
		}
		// Get ip_list from primary tenant
		aps, err := con.ListTenantAccessPoints(primary.Spec.TenantName)
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
		restoreSource = fmt.Sprintf("SERVICE=%s USER=%s@%s PASSWORD=%s", strings.Join(ipList, ";"), oceanbaseconst.StandbyROUser, primary.Spec.TenantName, standbyRoPwd)
	}

	return restoreSource, nil
}
