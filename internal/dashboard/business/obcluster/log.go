package obcluster

import (
	"context"
	"fmt"
	"os"

	bizconst "github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	jobmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/job"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func DownloadOBClusterLog(ctx context.Context, namespace, name, startTime, endTime string) (*jobmodel.Job, error) {
	jobNamespace := os.Getenv("NAMESPACE")
	sharedPvcName := os.Getenv("SHARED_VOLUME_PVC_NAME")
	sharedMountPath := os.Getenv("SHARED_VOLUME_MOUNT_PATH")
	jobName := fmt.Sprintf("log-%s-%s-%s", namespace, name, rand.String(6))
	attachmentID := jobName
	jobOutputDir := fmt.Sprintf("%s/%s", sharedMountPath, jobName)
	ttlSecondsAfterFinished := int32(24 * 60 * 60)

	labels := map[string]string{
		bizconst.LABEL_MANAGED_BY:        bizconst.DASHBOARD_APP_NAME,
		bizconst.LABEL_JOB_TYPE:          bizconst.JOB_TYPE_LOG,
		bizconst.LABEL_REF_OBCLUSTERNAME: name,
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
				Containers: []corev1.Container{
					{
						Name:    "log",
						Image:   "oceanbase/obdiag:latest",
						Command: []string{"/bin/sh", "-c"},
						Args: []string{
							fmt.Sprintf("mkdir -p %s && obdiag gather log --from %s --to %s --store_dir %s -c /etc/obdiag/config.yaml", jobOutputDir, startTime, endTime, jobOutputDir),
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "shared-volume",
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
						Name: "shared-volume",
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
