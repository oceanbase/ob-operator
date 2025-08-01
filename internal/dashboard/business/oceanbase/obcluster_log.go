/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package oceanbase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	bizconst "github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/config"
	jobmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/job"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func DownloadOBClusterLog(ctx context.Context, nn *param.K8sObjectIdentity, startTime, endTime string) (*jobmodel.Job, error) {
	jobNamespace := os.Getenv("NAMESPACE")
	sharedPvcName := os.Getenv("SHARED_VOLUME_PVC_NAME")
	sharedMountPath := os.Getenv("SHARED_VOLUME_MOUNT_PATH")
	jobName := fmt.Sprintf("log-%s-%s-%s", nn.Namespace, nn.Name, rand.String(6))
	attachmentID := jobName + ".tar.gz"
	jobOutputDir := filepath.Join(sharedMountPath, jobName)
	configFileName := "config.yaml"
	configFilePath := filepath.Join(jobOutputDir, configFileName)
	k8sConfigPath := filepath.Join(jobOutputDir, "k8s")
	ttlSecondsAfterFinished := config.GetConfig().Job.Normal.TTLSecondsAfterFinished

	labels := map[string]string{
		bizconst.LABEL_MANAGED_BY:        bizconst.DASHBOARD_APP_NAME,
		bizconst.LABEL_JOB_TYPE:          bizconst.JOB_TYPE_LOG,
		bizconst.LABEL_REF_OBCLUSTERNAME: nn.Name,
		bizconst.LABEL_ATTACHMENT_ID:     attachmentID,
	}

	var backoffLimit int32 = 0
	jobSpec := &batchv1.JobSpec{
		BackoffLimit:            &backoffLimit,
		TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				ServiceAccountName: bizconst.SERVICE_ACCOUNT_NAME,
				RestartPolicy:      corev1.RestartPolicyNever,
				InitContainers: []corev1.Container{
					{
						Name:            "generate-config",
						Image:           config.GetConfig().Inspection.OBHelper.Image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command:         []string{"bash", "-c", fmt.Sprintf("mkdir -p %s && /home/admin/oceanbase/bin/oceanbase-helper generate obdiag-config -n %s -c %s -o %s", jobOutputDir, nn.Namespace, nn.Name, jobOutputDir)},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      sharedPvcName,
								MountPath: sharedMountPath,
							},
						},
					},
				},
				Containers: []corev1.Container{
					{
						Name:            "log",
						Image:           config.GetConfig().Inspection.OBDiag.Image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command:         []string{"/bin/sh", "-c"},
						Args: []string{
							fmt.Sprintf("obdiag gather log --from %s --to %s --store_dir %s -c %s && rm -f %s && rm -rf %s && tar -czf %s/%s -C %s . && rm -rf %s", startTime, endTime, jobOutputDir, configFilePath, configFilePath, k8sConfigPath, sharedMountPath, attachmentID, jobOutputDir, jobOutputDir),
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      sharedPvcName,
								MountPath: sharedMountPath,
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: sharedPvcName,
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: sharedPvcName,
							},
						},
					},
				},
			},
		},
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: jobNamespace,
			Labels:    labels,
		},
		Spec: *jobSpec,
	}

	client := client.GetClient()
	createdJob, err := client.ClientSet.BatchV1().Jobs(jobNamespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return &jobmodel.Job{
		Name:      createdJob.Name,
		Namespace: createdJob.Namespace,
		Status:    jobmodel.JobStatusPending,
	}, nil
}
