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

package types

import corev1 "k8s.io/api/core/v1"

type LogServiceZoneTopology struct {
	Zone         string              `json:"zone"`
	Region       string              `json:"region"`
	Replica      int                 `json:"replica"`
	Resource     *ResourceSpec       `json:"resource,omitempty"`
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Affinity     *corev1.Affinity    `json:"affinity,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
	RpcPort      int32               `json:"rpcPort,omitempty"`
	HttpPort     int32               `json:"httpPort,omitempty"`
}

type ObjectStoreConfig struct {
	BucketURL string                      `json:"bucketURL"`
	SecretRef corev1.LocalObjectReference `json:"secretRef"`
}

type LogServiceStorageSpec struct {
	StoreStorage *StorageSpec `json:"storeStorage"`
	LogStorage   *StorageSpec `json:"logStorage"`
}

type LogServiceZoneReplicaStatus struct {
	Zone   string `json:"zone"`
	Status string `json:"status"`
}

type LogServiceNodeReplicaStatus struct {
	NodeName string `json:"nodeName"`
	Status   string `json:"status"`
}
