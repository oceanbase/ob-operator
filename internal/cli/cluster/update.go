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
package cluster

import (
	"errors"

	"github.com/spf13/cobra"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/cli/generic"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	param "github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

type UpdateOptions struct {
	generic.ResourceOption
	Resource     common.ResourceSpec             `json:"resource"`
	Storage      *param.OBServerStorageSpec      `json:"storage"`
	UpdateType   string                          `json:"updateType"`
	ModifyConfig *v1alpha1.ModifyOBServersConfig `json:"modifyConfig"`
}

func NewUpdateOptions() *UpdateOptions {
	return &UpdateOptions{
		Storage: &param.OBServerStorageSpec{},
	}
}

// GetUpdateOperation creates update opertaions
func GetUpdateOperation(o *UpdateOptions) *v1alpha1.OBClusterOperation {
	updateOp := &v1alpha1.OBClusterOperation{
		ObjectMeta: v1.ObjectMeta{
			Name:      o.Name + "-update-" + rand.String(6),
			Namespace: o.Namespace,
			Labels:    map[string]string{oceanbaseconst.LabelRefOBClusterOp: o.Name},
		},
		Spec: v1alpha1.OBClusterOperationSpec{
			OBCluster:       o.Name,
			Type:            apiconst.ClusterOpTypeModifyOBServers,
			ModifyOBServers: o.ModifyConfig,
		},
	}
	return updateOp
}

func (o *UpdateOptions) Validate() error {
	updateTypeCount := 0
	if o.CheckIfFlagChanged(FLAG_OBSERVER_CPU, FLAG_OBSERVER_MEMORY) {
		updateTypeCount++
		o.UpdateType = "resource"
	}
	if o.CheckIfFlagChanged(FLAG_DATA_STORAGE_CLASS, FLAG_LOG_STORAGE_CLASS, FLAG_REDO_LOG_STORAGE_CLASS) {
		updateTypeCount++
		o.UpdateType = "modifyStorageClass"
	}
	if o.CheckIfFlagChanged(FLAG_DATA_STORAGE_SIZE, FLAG_LOG_STORAGE_SIZE, FLAG_REDO_LOG_STORAGE_SIZE) {
		updateTypeCount++
		o.UpdateType = "expandStorageSize"
	}
	if updateTypeCount > 1 {
		return errors.New("Only one type of update is allowed at a time")
	}
	if updateTypeCount == 0 {
		return errors.New("No update type specified, support cpu/memory/storage")
	}
	return nil
}

func (o *UpdateOptions) Complete() error {
	switch o.UpdateType {
	case "resource":
		resource := &types.ResourceSpec{}
		if o.Resource.Cpu != 0 {
			resource.Cpu = *apiresource.NewQuantity(o.Resource.Cpu, apiresource.DecimalSI)
		}
		if o.Resource.MemoryGB != 0 {
			resource.Memory = *apiresource.NewQuantity(o.Resource.MemoryGB*constant.GB, apiresource.BinarySI)
		}
		o.ModifyConfig = &v1alpha1.ModifyOBServersConfig{Resource: resource}
	case "modifyStorageClass":
		modifyStorageClass := &v1alpha1.ModifyStorageClassConfig{}
		if o.Storage.Data.StorageClass != "" {
			modifyStorageClass.DataStorage = o.Storage.Data.StorageClass
		}
		if o.Storage.Log.StorageClass != "" {
			modifyStorageClass.LogStorage = o.Storage.Log.StorageClass
		}
		if o.Storage.RedoLog.StorageClass != "" {
			modifyStorageClass.RedoLogStorage = o.Storage.RedoLog.StorageClass
		}
		o.ModifyConfig = &v1alpha1.ModifyOBServersConfig{ModifyStorageClass: modifyStorageClass}
	case "expandStorageSize":
		expandStorageSize := &v1alpha1.ExpandStorageSizeConfig{}
		if o.Storage.Data.SizeGB != 0 {
			expandStorageSize.DataStorage = apiresource.NewQuantity(o.Storage.Data.SizeGB*constant.GB, apiresource.BinarySI)
		}
		if o.Storage.RedoLog.SizeGB != 0 {
			expandStorageSize.RedoLogStorage = apiresource.NewQuantity(o.Storage.RedoLog.SizeGB*constant.GB, apiresource.BinarySI)
		}
		if o.Storage.Log.SizeGB != 0 {
			expandStorageSize.LogStorage = apiresource.NewQuantity(o.Storage.Log.SizeGB*constant.GB, apiresource.BinarySI)
		}
		o.ModifyConfig = &v1alpha1.ModifyOBServersConfig{ExpandStorageSize: expandStorageSize}
	default:
		return errors.New("UpdateType Error")
	}
	return nil
}

// AddFlags for update options
func (o *UpdateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Namespace, FLAG_NAMESPACE, DEFAULT_NAMESPACE, "namespace of ob cluster")
	cmd.Flags().Int64Var(&o.Resource.Cpu, FLAG_OBSERVER_CPU, DEFAULT_OBSERVER_CPU, "The cpu of the observer")
	cmd.Flags().Int64Var(&o.Resource.MemoryGB, FLAG_OBSERVER_MEMORY, DEFAULT_OBSERVER_MEMORY, "The memory of the observer")
	cmd.Flags().StringVar(&o.Storage.Data.StorageClass, FLAG_DATA_STORAGE_CLASS, "", "The storage class of the data storage")
	cmd.Flags().StringVar(&o.Storage.RedoLog.StorageClass, FLAG_REDO_LOG_STORAGE_CLASS, "", "The storage class of the redo log storage")
	cmd.Flags().StringVar(&o.Storage.Log.StorageClass, FLAG_LOG_STORAGE_CLASS, "", "The storage class of the log storage")
	cmd.Flags().Int64Var(&o.Storage.Data.SizeGB, FLAG_DATA_STORAGE_SIZE, DEFAULT_DATA_STORAGE_SIZE, "The size of the data storage")
	cmd.Flags().Int64Var(&o.Storage.RedoLog.SizeGB, FLAG_REDO_LOG_STORAGE_SIZE, DEFAULT_REDO_LOG_STORAGE_SIZE, "The size of the redo log storage")
	cmd.Flags().Int64Var(&o.Storage.Log.SizeGB, FLAG_LOG_STORAGE_SIZE, DEFAULT_LOG_STORAGE_SIZE, "The size of the log storage")
}
