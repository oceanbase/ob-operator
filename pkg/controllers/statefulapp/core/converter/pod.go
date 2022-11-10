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

package converter

import (
	"fmt"
	"sort"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	myconfig "github.com/oceanbase/ob-operator/pkg/config"
	statefulapputil "github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/core/util"
	"github.com/oceanbase/ob-operator/pkg/util"
)

func GeneratePodName(statefulAppName, clusterName, subsetName string, podIndex int) string {
	return fmt.Sprintf("%s-%s-%s-%d", statefulAppName, clusterName, subsetName, podIndex)
}

func GeneratePodsIndexList(expectedReplicas int) []int {
	res := make([]int, 0)
	for i := 0; i < expectedReplicas; i++ {
		res = append(res, i)
	}
	return res
}

func GetPodsIndexList(pods []corev1.Pod) []int {
	res := make([]int, 0)
	for _, pod := range pods {
		index, _ := strconv.Atoi(pod.Labels["index"])
		res = append(res, index)
	}
	return res
}

func FindMissingIndex(subset cloudv1.Subset, pod []corev1.Pod) int {
	podsCurrentIndexList := GetPodsIndexList(pod)
	podsExpectedIndexList := GeneratePodsIndexList(int(subset.Replicas))
	index := util.CompareSlice(podsCurrentIndexList, podsExpectedIndexList)
	return index
}

func GetDeleteIndex(statefulApp cloudv1.StatefulApp, subset cloudv1.Subset) int {
	// TODO: support specify the index
	index := subset.Replicas
	return int(index)
}

func GeneratePodObject(statefulApp cloudv1.StatefulApp, subset cloudv1.Subset, podsCurrent []corev1.Pod) (string, int, corev1.Pod) {
	var podIndex int
	// get pod index
	if len(podsCurrent) == 0 {
		// new subset
		podIndex = 0
	} else {
		// find the missing index
		podIndex = FindMissingIndex(subset, podsCurrent)
	}
	// get pod name
	podName := GeneratePodName(statefulApp.Name, myconfig.ClusterName, subset.Name, podIndex)
	// generate
	podObject := GeneratePodObjectPcress(subset.Name, podName, podIndex, statefulApp, subset.NodeSelector)
	return podName, podIndex, podObject
}

func GeneratePodObjectPcress(subsetName, podName string, podIndex int, statefulApp cloudv1.StatefulApp, nodeSelect map[string]string) corev1.Pod {
	objectMeta := statefulapputil.GenerateObjectMeta(subsetName, podName, podIndex, statefulApp)
	// TODO: support PodSpecial
	// DeepCopy
	podTemplate := statefulApp.DeepCopy().Spec.PodTemplate

	podTemplate.NodeSelector = nodeSelect

	// pvc rewrite
	volumes := podTemplate.Volumes
	if len(volumes) > 0 {
		podTemplate.Volumes = PVCRewrite(podName, volumes)
	}

	pod := corev1.Pod{
		ObjectMeta: objectMeta,
		Spec:       podTemplate,
	}
	return pod
}

func PVCRewrite(podName string, volumes []corev1.Volume) []corev1.Volume {
	newVolumes := make([]corev1.Volume, 0)
	for _, volume := range volumes {
		if volume.Name == "backup" {
			newVolumes = append(newVolumes, volume)
			continue
		}
		klog.Infoln("PVCRewrite volume: ", volume)
		name := volume.PersistentVolumeClaim.ClaimName
		newName := GeneratePVCName(podName, name)
		volume.PersistentVolumeClaim.ClaimName = newName
		newVolumes = append(newVolumes, volume)
	}
	return newVolumes
}

func PodListToPods(podList corev1.PodList) []corev1.Pod {
	res := make([]corev1.Pod, 0)
	if len(podList.Items) > 0 {
		res = podList.Items
	}
	return res
}

func SortPodsDesc(pods []corev1.Pod) []corev1.Pod {
	sort.Slice(pods, func(i, j int) bool {
		indexI, _ := strconv.Atoi(pods[i].Labels["index"])
		indexJ, _ := strconv.Atoi(pods[j].Labels["index"])
		if indexI > indexJ {
			return true
		}
		return false
	})
	return pods
}

func PodCurrentStatusToPodStatus(pod corev1.Pod) cloudv1.PodStatus {
	var podStatus cloudv1.PodStatus
	podStatus.Name = pod.Name
	podStatus.Index, _ = strconv.Atoi(pod.Labels["index"])
	podStatus.PodPhase = pod.Status.Phase
	podStatus.PodIP = pod.Status.PodIP
	podStatus.NodeIP = pod.Status.HostIP
	return podStatus
}

func SortPodsStatus(podsStatus []cloudv1.PodStatus) []cloudv1.PodStatus {
	sort.Slice(podsStatus, func(i, j int) bool {
		if podsStatus[i].Index < podsStatus[j].Index {
			return true
		}
		return false
	})
	return podsStatus
}
