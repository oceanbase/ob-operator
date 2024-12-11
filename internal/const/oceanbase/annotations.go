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

package oceanbase

const (
	AnnotationCalicoValidate  = "cni.projectcalico.org/podIP"
	AnnotationCalicoIpAddrs   = "cni.projectcalico.org/ipAddrs"
	AnnotationKubeOvnValidate = "ovn.kubernetes.io/ip_address"
	AnnotationKubeOvnIpAddrs  = "ovn.kubernetes.io/ip_address"
)

const (
	AnnotationsIndependentPVCLifecycle = "oceanbase.oceanbase.com/independent-pvc-lifecycle"
	AnnotationsSinglePVC               = "oceanbase.oceanbase.com/single-pvc"
	AnnotationsMode                    = "oceanbase.oceanbase.com/mode"
	AnnotationsSourceClusterAddress    = "oceanbase.oceanbase.com/source-cluster-address"
	AnnotationsIgnoreDeletion          = "oceanbase.oceanbase.com/ignore-deletion"
	AnnotationsPauseReconciling        = "oceanbase.oceanbase.com/pause-reconciling"
	AnnotationsSupportStaticIP         = "oceanbase.oceanbase.com/support-static-ip"
	AnnotationsDeletionPriority        = "oceanbase.oceanbase.com/deletion-priority"
)

const (
	ModeStandalone = "standalone"
	ModeService    = "service"
)

const (
	CNICalico  = "calico"
	CNIKubeOvn = "kube-ovn"
	CNIUnknown = "unknown"
)
