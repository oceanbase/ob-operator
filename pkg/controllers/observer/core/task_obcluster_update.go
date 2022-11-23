/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package core

import (
	"context"
	"fmt"

	myconfig "github.com/oceanbase/ob-operator/pkg/config"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// func (ctrl *OBClusterCtrl) GetTargetVerJob() (string, error) {
// 	containerImage := ctrl.OBCluster.Spec.ImageRepo + ":" + ctrl.OBCluster.Spec.Tag
// 	klog.Infoln("containerImage: ", containerImage)
// 	var entryCommand []string
// 	entryCommand = append(entryCommand, "bash", "-c", "/home/admin/oceanbase/bin/observer -V")

// 	jobname, jobObject := ctrl.GenerateJobObject(containerImage, entryCommand)
// 	klog.Infoln("jobname: ", jobname)
// 	jobExectuer := resource.NewJobResource(ctrl.Resource)
// 	err := jobExectuer.Create(context.TODO(), jobObject)
// 	if err != nil {
// 		klog.Errorln("jobExectuer.Create: ", err)
// 		return "", err
// 	}
// 	return "", nil
// }

func GenerateJobName(clusterName, name string) string {
	return fmt.Sprintf("%s-%s", clusterName, name)
}

func (ctrl *OBClusterCtrl) GenerateJobObjectPcress(jobName, image string, cmd []string) batchv1.Job {
	var backOffLimit int32
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: ctrl.OBCluster.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    jobName,
							Image:   image,
							Command: cmd,
							Env: []corev1.EnvVar{
								{
									Name:  "LD_LIBRARY_PATH",
									Value: "/home/admin/oceanbase/lib",
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit: &backOffLimit,
		},
	}

	return job
}

func (ctrl *OBClusterCtrl) GenerateJobObject(image string, cmd []string) (string, batchv1.Job) {
	// get Job name
	jobName := GenerateJobName(myconfig.ClusterName, "1")
	// generate
	jobObject := ctrl.GenerateJobObjectPcress(jobName, image, cmd)
	return jobName, jobObject
}

func GeneratePodName(clusterName, name string) string {
	return fmt.Sprintf("%s-%s", clusterName, name)
}

func (ctrl *OBClusterCtrl) CreatePodForVersion() error {
	podName := GeneratePodName(myconfig.ClusterName, "help")
	containerImage := ctrl.OBCluster.Spec.ImageRepo + ":" + ctrl.OBCluster.Spec.Tag
	podObject := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ctrl.OBCluster.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  podName,
					Image: containerImage,
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	// create pod
	podExecuter := resource.NewPodResource(ctrl.Resource)
	err := podExecuter.Create(context.TODO(), podObject)
	if err != nil {
		klog.Errorln("create pod to get version failed, error: ", err)
		if kubeerrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}
	return nil
}
