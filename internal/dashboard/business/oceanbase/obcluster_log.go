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
	attachmentID := jobName
	jobOutputDir := filepath.Join(sharedMountPath, jobName)
	configFileName := "config.yaml"
	configFilePath := filepath.Join(jobOutputDir, configFileName)
	ttlSecondsAfterFinished := int32(24 * 60 * 60)

	labels := map[string]string{
		bizconst.LABEL_MANAGED_BY:        bizconst.DASHBOARD_APP_NAME,
		bizconst.LABEL_JOB_TYPE:          bizconst.JOB_TYPE_LOG,
		bizconst.LABEL_REF_OBCLUSTERNAME: nn.Name,
		bizconst.LABEL_ATTACHMENT_ID:     attachmentID,
	}

	jobSpec := &batchv1.JobSpec{
		TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyNever,
				InitContainers: []corev1.Container{
					{
						Name:            "generate-config",
						Image:           "oceanbase/oceanbase-helper:latest",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command:         []string{"bash", "-c", fmt.Sprintf("/home/admin/oceanbase/bin/oceanbase-helper generate obdiag-config -n %s -c %s -o %s", nn.Namespace, nn.Name, configFilePath)},
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
						Name:    "log",
						Image:   "oceanbase/obdiag:latest",
						Command: []string{"/bin/sh", "-c"},
						Args: []string{
							fmt.Sprintf("mkdir -p %s && obdiag gather log --from %s --to %s --store_dir %s -c %s", jobOutputDir, startTime, endTime, jobOutputDir, configFilePath),
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      sharedPvcName,
								MountPath: sharedMountPath,
							},
						},
						Lifecycle: &corev1.Lifecycle{
							PreStop: &corev1.LifecycleHandler{
								Exec: &corev1.ExecAction{
									Command: []string{"/bin/sh", "-c", fmt.Sprintf("rm -rf %s", jobOutputDir)},
								},
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
