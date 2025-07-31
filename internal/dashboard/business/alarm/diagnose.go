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

package alarm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bizconst "github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/oceanbase"
	"github.com/oceanbase/ob-operator/internal/dashboard/config"
	"github.com/oceanbase/ob-operator/internal/dashboard/generated/bindata"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/alert"
	jobmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/job"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func DiagnoseAlert(ctx context.Context, param *alert.AnalyzeParam) (*jobmodel.Job, error) {
	jobNamespace := os.Getenv("NAMESPACE")
	sharedPvcName := os.Getenv("SHARED_VOLUME_PVC_NAME")
	sharedMountPath := os.Getenv("SHARED_VOLUME_MOUNT_PATH")
	jobName := fmt.Sprintf("diagnose-%s-%s", param.Instance.OBCluster, rand.String(6))
	attachmentID := jobName + ".tar.gz"
	jobOutputDir := filepath.Join(sharedMountPath, jobName)
	configFileName := "config.yaml"
	configFilePath := filepath.Join(jobOutputDir, configFileName)
	k8sConfigPath := filepath.Join(jobOutputDir, "k8s")
	ttlSecondsAfterFinished := config.GetConfig().Job.Normal.TTLSecondsAfterFinished

	labels := map[string]string{
		bizconst.LABEL_MANAGED_BY:        bizconst.DASHBOARD_APP_NAME,
		bizconst.LABEL_JOB_TYPE:          bizconst.JOB_TYPE_DIAGNOSE,
		bizconst.LABEL_REF_OBCLUSTERNAME: param.Instance.OBCluster,
		bizconst.LABEL_ATTACHMENT_ID:     attachmentID,
	}

	var obclusterObj *response.OBClusterOverview = nil
	obclusters, err := oceanbase.ListOBClusters(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list obclusters")
	}
	for idx, obcluster := range obclusters {
		if obcluster.ClusterName == param.Instance.OBCluster {
			obclusterObj = &obclusters[idx]
			break
		}
	}

	if obclusterObj == nil {
		return nil, errors.Errorf("Can not found obcluster object with obcluster name: %s", param.Instance.OBCluster)
	}

	scene := "observer.unknown"
	ruleSceneMapBytes, err := bindata.Asset("internal/assets/dashboard/rule-scene-map.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read rule-scene-map.yaml")
	}
	var ruleSceneMap map[string]string
	if err := yaml.Unmarshal(ruleSceneMapBytes, &ruleSceneMap); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal rule-scene-map.yaml")
	}
	if s, ok := ruleSceneMap[param.Rule]; ok {
		scene = s
	}

	startTime := time.Unix(param.StartsAt-60, 0).In(time.Local)
	var endTime time.Time
	if param.EndsAt < param.StartsAt {
		endTime = time.Unix(param.StartsAt, 0).In(time.Local)
	} else {
		endTime = time.Unix(param.EndsAt, 0).In(time.Local)
	}
	from := startTime.Format("2006-01-02 15:04:05")
	to := endTime.Format("2006-01-02 15:04:05")

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
						Command:         []string{"bash", "-c", fmt.Sprintf("mkdir -p %s && /home/admin/oceanbase/bin/oceanbase-helper generate obdiag-config -n %s -c %s -o %s", jobOutputDir, obclusterObj.Namespace, obclusterObj.Name, jobOutputDir)},
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
						Name:            "diagnose",
						Image:           config.GetConfig().Inspection.OBDiag.Image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command:         []string{"/bin/sh", "-c"},
						Args: []string{
							fmt.Sprintf("obdiag gather scene run --scene=%s --from '%s' --to '%s' --store_dir %s -c %s && rm -f %s && rm -rf %s && tar -czf %s/%s -C %s . && rm -rf %s", scene, from, to, jobOutputDir, configFilePath, configFilePath, k8sConfigPath, sharedMountPath, attachmentID, jobOutputDir, jobOutputDir),
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
