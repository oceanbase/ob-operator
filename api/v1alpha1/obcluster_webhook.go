/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

// log is for logging in this package.
var obclusterlog = logf.Log.WithName("obcluster-resource")

func (r *OBCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-oceanbase-oceanbase-com-v1alpha1-obcluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obclusters,verbs=create;update,versions=v1alpha1,name=mobcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &OBCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *OBCluster) Default() {
	// fill default essential parameters, memory_limit, datafile_maxsize and datafile_next

	obclusterlog.Info("fill in default values of obcluster")
	parameterMap := make(map[string]apitypes.Parameter, 0)
	memorySize, ok := r.Spec.OBServerTemplate.Resource.Memory.AsInt64()
	if ok {
		memoryLimit := fmt.Sprintf("%dG", memorySize*oceanbaseconst.DefaultMemoryLimitPercent/oceanbaseconst.GigaConverter/100)
		parameterMap["memory_limit"] = apitypes.Parameter{
			Name:  "memory_limit",
			Value: memoryLimit,
		}
	} else {
		obclusterlog.Error(errors.New("failed to parse memory size"), "parse observer's memory size failed")
	}
	datafileDiskSize, ok := r.Spec.OBServerTemplate.Storage.DataStorage.Size.AsInt64()
	if ok {
		datafileMaxSize := fmt.Sprintf("%dG", datafileDiskSize*oceanbaseconst.DefaultDiskUsePercent/oceanbaseconst.GigaConverter/100)
		parameterMap["datafile_maxsize"] = apitypes.Parameter{
			Name:  "datafile_maxsize",
			Value: datafileMaxSize,
		}
		datafileNextSize := fmt.Sprintf("%dG", datafileDiskSize*oceanbaseconst.DefaultDiskExpandPercent/oceanbaseconst.GigaConverter/100)
		parameterMap["datafile_next"] = apitypes.Parameter{
			Name:  "datafile_next",
			Value: datafileNextSize,
		}
	} else {
		obclusterlog.Error(errors.New("failed to parse datafile size"), "parse observer's datafile size failed")
	}
	parameterMap["enable_syslog_recycle"] = apitypes.Parameter{
		Name:  "enable_syslog_recycle",
		Value: "true",
	}
	maxSysLogFileCount := int64(4)
	logSize, ok := r.Spec.OBServerTemplate.Storage.LogStorage.Size.AsInt64()
	if ok {
		// observer has 4 types of log and one logfile limits at 256M considering about wf, maximum of 2G will be occupied for 1 syslog count
		maxSysLogFileCount = logSize * oceanbaseconst.DefaultLogPercent / oceanbaseconst.GigaConverter / 100 / 2
	}
	parameterMap["max_syslog_file_count"] = apitypes.Parameter{
		Name:  "max_syslog_file_count",
		Value: fmt.Sprintf("%d", maxSysLogFileCount),
	}

	for _, parameter := range r.Spec.Parameters {
		parameterMap[parameter.Name] = parameter
	}
	parameters := make([]apitypes.Parameter, 0)
	for _, v := range parameterMap {
		parameters = append(parameters, v)
	}
	r.Spec.Parameters = parameters
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-oceanbase-oceanbase-com-v1alpha1-obcluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obclusters,verbs=create;update,versions=v1alpha1,name=vobcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OBCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *OBCluster) ValidateCreate() (admission.Warnings, error) {
	obclusterlog.Info("validate create", "name", r.Name)

	return nil, r.validateMutation()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *OBCluster) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	oldCluster, ok := old.(*OBCluster)
	if !ok {
		return nil, errors.New("failed to convert old object to OBCluster")
	}
	if oldCluster.Spec.Standalone != r.Spec.Standalone {
		return nil, errors.New("standalone mode cannot be changed")
	}
	if !oldCluster.Spec.Standalone && oldCluster.Spec.OBServerTemplate.Resource.Cpu != r.Spec.OBServerTemplate.Resource.Cpu {
		return nil, errors.New("forbid to modify cpu quota of non-standalone cluster")
	}
	if !oldCluster.Spec.Standalone && oldCluster.Spec.OBServerTemplate.Resource.Memory != r.Spec.OBServerTemplate.Resource.Memory {
		return nil, errors.New("forbid to modify memory quota of non-standalone cluster")
	}

	return nil, r.validateMutation()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *OBCluster) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (r *OBCluster) validateMutation() error {
	// Ignore deleting objects
	if r.GetDeletionTimestamp() != nil {
		return nil
	}

	var allErrs field.ErrorList

	// Validate standalone
	if r.Spec.Standalone {
		if len(r.Spec.Topology) != 1 {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("topology"), r.Spec.Topology, "standalone mode only support single zone"))
		} else if r.Spec.Topology[0].Replica != 1 {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("topology"), r.Spec.Topology, "standalone mode only support single replica"))
		}
	}

	// Validate userSecrets
	if r.Spec.UserSecrets != nil {
		if err := r.checkSecretExistence(r.Namespace, r.Spec.UserSecrets.Root, "root"); err != nil {
			allErrs = append(allErrs, err)
		}
		if err := r.checkSecretWithRawError(r.Namespace, r.Spec.UserSecrets.Operator); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("userSecrets").Child("operator"), r.Spec.UserSecrets.Operator, err.Error()))
		}
		if err := r.checkSecretWithRawError(r.Namespace, r.Spec.UserSecrets.Monitor); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("userSecrets").Child("monitor"), r.Spec.UserSecrets.Monitor, err.Error()))
		}
		if err := r.checkSecretWithRawError(r.Namespace, r.Spec.UserSecrets.ProxyRO); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("userSecrets").Child("proxyro"), r.Spec.UserSecrets.ProxyRO, err.Error()))
		}
	}

	// Validate Topology
	if r.Spec.Topology == nil || len(r.Spec.Topology) == 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("topology"), r.Spec.Topology, "empty topology is not permitted"))
	}

	// Validate storageClasses
	storageClassMapping := make(map[string]bool)
	storageClassMapping[r.Spec.OBServerTemplate.Storage.DataStorage.StorageClass] = true
	storageClassMapping[r.Spec.OBServerTemplate.Storage.LogStorage.StorageClass] = true
	storageClassMapping[r.Spec.OBServerTemplate.Storage.RedoLogStorage.StorageClass] = true

	for key := range storageClassMapping {
		err := clt.Get(context.TODO(), types.NamespacedName{
			Name: key,
		}, &storagev1.StorageClass{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("storageClass"), key, fmt.Sprintf("storageClass %s not found", key)))
			} else {
				allErrs = append(allErrs, field.InternalError(field.NewPath("spec").Child("observer").Child("storage").Child("storageClass"), err))
			}
		}
	}

	// Validate disk size
	if r.Spec.OBServerTemplate.Storage.DataStorage.Size.AsApproximateFloat64() < oceanbaseconst.MinDataDiskSize.AsApproximateFloat64() {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("dataStorage").Child("size"), r.Spec.OBServerTemplate.Storage.DataStorage.Size.String(), "The minimum data storage size of OBCluster is "+oceanbaseconst.MinDataDiskSize.String()))
	}
	if r.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.AsApproximateFloat64() < oceanbaseconst.MinRedoLogDiskSize.AsApproximateFloat64() {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("redoLogStorage").Child("size"), r.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.String(), "The minimum redo log storage size of OBCluster is "+oceanbaseconst.MinRedoLogDiskSize.String()))
	}
	if r.Spec.OBServerTemplate.Storage.LogStorage.Size.AsApproximateFloat64() < oceanbaseconst.MinLogDiskSize.AsApproximateFloat64() {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("logStorage").Child("size"), r.Spec.OBServerTemplate.Storage.LogStorage.Size.String(), "The minimum log storage size of OBCluster is "+oceanbaseconst.MinLogDiskSize.String()))
	}

	// Validate memory size
	if r.Spec.OBServerTemplate.Resource.Memory.AsApproximateFloat64() < oceanbaseconst.MinMemorySize.AsApproximateFloat64() {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("resource").Child("memory"), r.Spec.OBServerTemplate.Resource.Memory.String(), "The minimum memory size of OBCluster is "+oceanbaseconst.MinMemorySize.String()))
	}

	// Validate essential parameters
	parameterMap := make(map[string]apitypes.Parameter, 0)
	for _, parameter := range r.Spec.Parameters {
		parameterMap[parameter.Name] = parameter
	}

	// check memory limit
	memoryLimitSize, ok := parameterMap["memory_limit"]
	if ok {
		memoryLimit, err := resource.ParseQuantity(memoryLimitSize.Value)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("parameters"), "memory limit size", "Failed to parse memory limit"))
		} else if memoryLimit.AsApproximateFloat64() > r.Spec.OBServerTemplate.Resource.Memory.AsApproximateFloat64() {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("parameters"), "memory limit size overflow", "memory limit exceeds observer's resource"))
		}

		if r.Spec.OBServerTemplate.Storage.DataStorage.Size.AsApproximateFloat64() < 3*memoryLimit.AsApproximateFloat64() {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("dataStorage").Child("size"), r.Spec.OBServerTemplate.Storage.DataStorage.Size.String(), "The minimum size of data storage should be larger than 3 times of memory limit"))
		}

		if r.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.AsApproximateFloat64() < 3*memoryLimit.AsApproximateFloat64() {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("redoLogStorage").Child("size"), r.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.String(), "The minimum size of redo log storage should be larger than 3 times of memory limit"))
		}
	}

	// check datafile max size
	datafileMaxSize, ok := parameterMap["datafile_maxsize"]
	if ok {
		datafileMax, err := resource.ParseQuantity(datafileMaxSize.Value)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("parameters"), "datafile max size", "Failed to parse datafile max size"))
		} else if datafileMax.AsApproximateFloat64() > r.Spec.OBServerTemplate.Storage.DataStorage.Size.AsApproximateFloat64() {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("parameters"), "datafile max size overflow", "datafile maxsize exceeds observer's data storage size"))
		}
	}

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(GroupVersion.WithKind("OBCluster").GroupKind(), r.Name, allErrs)
}

func (r *OBCluster) checkSecretExistence(ns, secretName, fieldName string) *field.Error {
	if secretName == "" {
		return field.Invalid(field.NewPath("spec").Child("userSecrets").Child(fieldName), secretName, fmt.Sprintf("Empty credential %s is not permitted", fieldName))
	}
	secret := &v1.Secret{}
	err := tenantClt.Get(context.Background(), types.NamespacedName{
		Namespace: ns,
		Name:      secretName,
	}, secret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return field.Invalid(field.NewPath("spec").Child("userSecrets").Child(fieldName), secretName, fmt.Sprintf("Given %s credential %s not found", fieldName, secretName))
		}
		return field.InternalError(field.NewPath("spec").Child("userSecrets").Child(fieldName), err)
	}
	if _, ok := secret.Data["password"]; !ok {
		return field.Invalid(field.NewPath("spec").Child("userSecrets").Child(fieldName), secretName, fmt.Sprintf("password field not found in given credential %s ", secretName))
	}
	return nil
}

func (r *OBCluster) checkSecretWithRawError(ns, secretName string) error {
	if secretName == "" {
		return nil
	}
	secret := &v1.Secret{}
	err := tenantClt.Get(context.Background(), types.NamespacedName{
		Namespace: ns,
		Name:      secretName,
	}, secret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if _, ok := secret.Data["password"]; !ok {
		return errors.New("password field not found in given credential")
	}
	return nil
}
