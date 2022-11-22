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
	"os"
	"path/filepath"
	"strings"

	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	"k8s.io/klog/v2"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	clientcmd "k8s.io/client-go/tools/clientcmd"
)

func (ctrl *OBClusterCtrl) connectToK8s() (*kubernetes.Clientset, error) {
	home, exists := os.LookupEnv("HOME")
	klog.Infoln("home, exists : ", home, exists)
	if !exists {
		home = "/root"
	}

	configPath := filepath.Join(home, ".kube", "config")
	klog.Infoln("configPath: ", configPath)
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		klog.Errorln("failed to create K8s config, ", err)
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorln("failed to create K8s config, ", err)
		return nil, err
	}
	return clientset, nil
}

func (ctrl *OBClusterCtrl) launchK8sJob(clientset *kubernetes.Clientset, jobName *string, image *string, cmd *string) error {
	jobs := clientset.BatchV1().Jobs("oceanbase-system")
	var backOffLimit int32 = 0

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      *jobName,
			Namespace: "oceanbase-system",
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    *jobName,
							Image:   *image,
							Command: strings.Split(*cmd, " "),
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
			BackoffLimit: &backOffLimit,
		},
	}

	_, err := jobs.Create(context.TODO(), jobSpec, metav1.CreateOptions{})
	if err != nil {
		klog.Errorln("Failed to create K8s job, ", err)
		return err
	}
	klog.Infoln("Created job successfully")
	return nil
}

func (ctrl *OBClusterCtrl) GetTargetVerJob() (string, error) {
	// jobName := flag.String("jobname", "test-job-createOB", "name of job")
	// containerImage := flag.String("image", ctrl.OBCluster.Spec.ImageRepo+":"+ctrl.OBCluster.Spec.Tag, "容器镜像的名字")
	// entryCommand := flag.String("command", "pwd", "command to execute")

	// flag.Parse()

	// clientset, err := ctrl.connectToK8s()
	// if err != nil {
	// 	return "", err
	// }
	// err = ctrl.launchK8sJob(clientset, jobName, containerImage, entryCommand)
	// if err != nil {
	// 	return "", err
	// }
	// return "", nil

	containerImage := ctrl.OBCluster.Spec.ImageRepo + ":" + ctrl.OBCluster.Spec.Tag
	klog.Infoln("containerImage: ", containerImage)
	entryCommand := "pwd"

	jobname, jobObject := converter.GenerateJobObject(containerImage, entryCommand)
	klog.Infoln("jobname: ", jobname)
	klog.Infoln("jobname: ", jobObject)
	jobExectuer := resource.NewJobResource(ctrl.Resource)
	err := jobExectuer.Create(context.TODO(), jobObject)
	if err != nil {
		klog.Errorln("jobExectuer.Create: ", err)
		return "", err
	}
	return "", nil
}
