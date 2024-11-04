/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package controller

// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=services/finalizers,verbs=update

// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods/log,verbs=get
// +kubebuilder:rbac:groups="",resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=pods/finalizers,verbs=update

// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims/status,verbs=get;update;patch

// +kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumes/status,verbs=get;update;patch

// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs/finalizers,verbs=update

// +kubebuilder:rbac:groups=storage.k8s.io,resources=storageclasses,verbs=get;list;watch

// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obclusters/finalizers,verbs=update

//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obclusteroperations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obclusteroperations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obclusteroperations/finalizers,verbs=update

// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obparameters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obparameters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obparameters/finalizers,verbs=update

// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obresourcerescues,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obresourcerescues/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obresourcerescues/finalizers,verbs=update

// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=observers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=observers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=observers/finalizers,verbs=update

// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenants/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenants/finalizers,verbs=update

//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantrestore,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantrestore/status,verbs=get;update;patch

//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantbackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantbackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantbackups/finalizers,verbs=update

//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantbackuppolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantbackuppolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantbackuppolicies/finalizers,verbs=update

//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantoperations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantoperations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantoperations/finalizers,verbs=update

//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantrestores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantrestores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantrestores/finalizers,verbs=update

// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obzones,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obzones/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obzones/finalizers,verbs=update

/**
**  [GROUP] k8s.oceanbase.com
**/

//+kubebuilder:rbac:groups=k8s.oceanbase.com,resources=k8sclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8s.oceanbase.com,resources=k8sclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8s.oceanbase.com,resources=k8sclusters/finalizers,verbs=update
