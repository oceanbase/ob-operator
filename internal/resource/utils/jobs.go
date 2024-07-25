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
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	cmdconst "github.com/oceanbase/ob-operator/internal/const/cmd"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	k8sclient "github.com/oceanbase/ob-operator/pkg/k8s/client"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

func GetJob(ctx context.Context, c client.Client, namespace string, jobName string) (*batchv1.Job, error) {
	job := &batchv1.Job{}
	err := c.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      jobName,
	}, job)
	return job, err
}

type JobContainerVolumes struct {
	VolumeMounts []corev1.VolumeMount
	Volumes      []corev1.Volume
}

func RunJob(ctx context.Context, c client.Client, logger *logr.Logger, namespace string, jobName string, image string, cmd string, volumeConfigs ...JobContainerVolumes) (output string, exitCode int32, err error) {
	fullJobName := fmt.Sprintf("%s-%s", jobName, rand.String(6))
	var backoffLimit int32
	var ttl int32 = 300
	var mounts []corev1.VolumeMount
	var volumes []corev1.Volume
	for _, vc := range volumeConfigs {
		mounts = append(mounts, vc.VolumeMounts...)
		volumes = append(volumes, vc.Volumes...)
	}

	container := corev1.Container{
		Name:         "job-runner",
		Image:        image,
		Command:      []string{"bash", "-c", cmd},
		VolumeMounts: mounts,
	}
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fullJobName,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers:      []corev1.Container{container},
					RestartPolicy:   corev1.RestartPolicyNever,
					Volumes:         volumes,
					SecurityContext: GetDefaultSecurityContext(),
				},
			},
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttl,
		},
	}

	err = c.Create(ctx, &job)
	if err != nil {
		return "", int32(cmdconst.ExitCodeNotExecuted), errors.Wrapf(err, "failed to create job of image: %s", image)
	}

	// Wait for the job to be created before fetching it
	time.Sleep(time.Second)

	var jobObject *batchv1.Job
	finished := false
	for i := 0; i < obcfg.GetConfig().Time.CheckJobMaxRetries; i++ {
		jobObject, err = GetJob(ctx, c, namespace, fullJobName)
		if err != nil {
			logger.Error(err, "Failed to get job")
			// return errors.Wrapf(err, "Failed to get run upgrade script job for obcluster %s", obcluster.Name)
		}
		if jobObject.Status.Succeeded == 0 && jobObject.Status.Failed == 0 {
			logger.V(oceanbaseconst.LogLevelDebug).Info("Job is still running")
		} else {
			logger.V(oceanbaseconst.LogLevelDebug).Info("Job finished")
			finished = true
			break
		}
		time.Sleep(time.Second * time.Duration(obcfg.GetConfig().Time.CheckJobInterval))
	}
	if !finished {
		return "", int32(cmdconst.ExitCodeNotExecuted), errors.Wrapf(err, "Run job %s timeout", fullJobName)
	}
	clientSet := k8sclient.GetClient()
	podList, err := clientSet.ClientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", fullJobName),
	})
	if err != nil || len(podList.Items) == 0 {
		return "", int32(cmdconst.ExitCodeNotExecuted), errors.Wrapf(err, "failed to get pods of job %s", jobName)
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
		return "", exitCode, errors.Errorf("Failed to run job %s", fullJobName)
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
	observers, err := oceanbaseOperationManager.ListServers(ctx)
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
		Command: []string{"bash", "-c", fmt.Sprintf("if [[ `command -v python2` ]]; then ln -sf /usr/bin/python2 /usr/bin/python; fi && python %s -h%s -P%d -uroot -p'%s' %s", filepath, rootserver.Ip, rootserver.SqlPort, password, extraOpt)},
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
					Containers:      []corev1.Container{container},
					RestartPolicy:   corev1.RestartPolicyNever,
					SecurityContext: GetDefaultSecurityContext(),
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
	err = CheckJobWithTimeout(check, time.Second*time.Duration(obcfg.GetConfig().Time.WaitForJobTimeoutSeconds))
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
	timeout := time.Second * time.Duration(obcfg.GetConfig().Time.DefaultStateWaitTimeout)
	interval := time.Second * time.Duration(obcfg.GetConfig().Time.CheckJobInterval)
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
