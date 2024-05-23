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
	"sort"

	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
)

// log is for logging in this package.
var obclusterlog = logf.Log.WithName("obcluster-resource")

func (r *OBCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-oceanbase-oceanbase-com-v1alpha1-obcluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obclusters,verbs=create;update,versions=v1alpha1,name=mobcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &OBCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *OBCluster) Default() {
	// fill default essential parameters, memory_limit, datafile_maxsize and datafile_next
	logger := obclusterlog.WithValues("namespace", r.Namespace, "name", r.Name)

	parameterMap := make(map[string]apitypes.Parameter, 0)
	memorySize, ok := r.Spec.OBServerTemplate.Resource.Memory.AsInt64()
	if ok {
		memoryLimit := fmt.Sprintf("%dM", memorySize*int64(obcfg.GetConfig().Resource.DefaultMemoryLimitPercent)/100/oceanbaseconst.MegaConverter)
		parameterMap["memory_limit"] = apitypes.Parameter{
			Name:  "memory_limit",
			Value: memoryLimit,
		}
	} else {
		logger.Error(errors.New("Failed to parse memory size"), "parse observer's memory size failed")
	}
	datafileDiskSize, ok := r.Spec.OBServerTemplate.Storage.DataStorage.Size.AsInt64()
	if ok {
		datafileMaxSize := fmt.Sprintf("%dG", datafileDiskSize*int64(obcfg.GetConfig().Resource.DefaultDiskUsePercent)/oceanbaseconst.GigaConverter/100)
		parameterMap["datafile_maxsize"] = apitypes.Parameter{
			Name:  "datafile_maxsize",
			Value: datafileMaxSize,
		}
		datafileNextSize := fmt.Sprintf("%dG", datafileDiskSize*int64(obcfg.GetConfig().Resource.DefaultDiskExpandPercent)/oceanbaseconst.GigaConverter/100)
		parameterMap["datafile_next"] = apitypes.Parameter{
			Name:  "datafile_next",
			Value: datafileNextSize,
		}
	} else {
		logger.Error(errors.New("Failed to parse datafile size"), "parse observer's datafile size failed")
	}
	parameterMap["enable_syslog_recycle"] = apitypes.Parameter{
		Name:  "enable_syslog_recycle",
		Value: "true",
	}
	maxSysLogFileCount := int64(4)
	logSize, ok := r.Spec.OBServerTemplate.Storage.LogStorage.Size.AsInt64()
	if ok {
		// observer has 4 types of log and one logfile limits at 256M considering about wf, maximum of 2G will be occupied for 1 syslog count
		maxSysLogFileCount = logSize * int64(obcfg.GetConfig().Resource.DefaultLogPercent) / oceanbaseconst.GigaConverter / 100 / 2
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

	if r.Spec.UserSecrets.Monitor == "" {
		r.Spec.UserSecrets.Monitor = "monitor-user-" + rand.String(6)
	}
	if r.Spec.UserSecrets.Operator == "" {
		r.Spec.UserSecrets.Operator = "operator-user-" + rand.String(6)
	}
	if r.Spec.UserSecrets.ProxyRO == "" {
		r.Spec.UserSecrets.ProxyRO = "proxyro-user-" + rand.String(6)
	}

	if r.Spec.ServiceAccount == "" {
		r.Spec.ServiceAccount = "default"
	}
	if r.Spec.OBServerTemplate.Storage.DataStorage.StorageClass == "" ||
		r.Spec.OBServerTemplate.Storage.LogStorage.StorageClass == "" ||
		r.Spec.OBServerTemplate.Storage.RedoLogStorage.StorageClass == "" {
		scList := &storagev1.StorageClassList{}
		err := clt.List(context.TODO(), scList)
		var defaults []string
		if err != nil {
			logger.Error(err, "Failed to list storage class")
		} else {
			sort.SliceStable(scList.Items, func(i, j int) bool {
				return scList.Items[i].Name < scList.Items[j].Name
			})
			for _, sc := range scList.Items {
				if sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
					defaults = append(defaults, sc.Name)
				}
			}
			if len(defaults) == 0 {
				logger.Error(nil, "No default storage class found")
			} else {
				if len(defaults) > 1 {
					logger.Info("Multiple default storage class found", "storageClasses", defaults, "selected", defaults[0])
				}
				if r.Spec.OBServerTemplate.Storage.DataStorage.StorageClass == "" {
					r.Spec.OBServerTemplate.Storage.DataStorage.StorageClass = defaults[0]
				}
				if r.Spec.OBServerTemplate.Storage.LogStorage.StorageClass == "" {
					r.Spec.OBServerTemplate.Storage.LogStorage.StorageClass = defaults[0]
				}
				if r.Spec.OBServerTemplate.Storage.RedoLogStorage.StorageClass == "" {
					r.Spec.OBServerTemplate.Storage.RedoLogStorage.StorageClass = defaults[0]
				}
			}
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-oceanbase-oceanbase-com-v1alpha1-obcluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obclusters,verbs=create;update,versions=v1alpha1,name=vobcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OBCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *OBCluster) ValidateCreate() (admission.Warnings, error) {
	return nil, r.validateMutation()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *OBCluster) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	oldCluster, ok := old.(*OBCluster)
	if !ok {
		return nil, errors.New("failed to convert old object to OBCluster")
	}
	oldMode, existOld := oldCluster.GetAnnotations()[oceanbaseconst.AnnotationsMode]
	mode, exist := r.GetAnnotations()[oceanbaseconst.AnnotationsMode]
	oldResource := oldCluster.Spec.OBServerTemplate.Resource
	newResource := r.Spec.OBServerTemplate.Resource
	if existOld && exist && oldMode != mode {
		return nil, errors.New("mode cannot be changed")
	} else if !oldCluster.SupportStaticIP() && (oldResource.Cpu != newResource.Cpu || oldResource.Memory != newResource.Memory) {
		return nil, errors.New("forbid to modify cpu or memory quota of non-static-ip cluster")
	}
	if newResource.Memory.Cmp(oldResource.Memory) < 0 {
		if r.Status.Status != clusterstatus.Running {
			return nil, errors.New("forbid to shrink memory size of non-running cluster")
		}
		conn, err := getSysClient(clt, &obclusterlog, r, oceanbaseconst.OperatorUser, oceanbaseconst.SysTenant, r.Spec.UserSecrets.Operator)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		gvservers, err := conn.ListGVServers(context.Background())
		if err != nil {
			return nil, err
		}
		var maxAssignedMemory int64
		var memoryLimitPercent float64
		for _, gvserver := range gvservers {
			if gvserver.MemAssigned > maxAssignedMemory {
				if oldResource.Memory.Value() < gvserver.MemoryLimit {
					memoryLimitPercent = 0.9
				} else {
					memoryLimitPercent = float64(gvserver.MemoryLimit) / oldResource.Memory.AsApproximateFloat64()
				}
				maxAssignedMemory = gvserver.MemAssigned
			}
		}
		if newResource.Memory.AsApproximateFloat64()*memoryLimitPercent < float64(maxAssignedMemory) {
			return nil, errors.New("Assigned memory is larger than new memory size")
		}
	}
	if r.Spec.BackupVolume == nil && oldCluster.Spec.BackupVolume != nil {
		return nil, errors.New("forbid to remove backup volume")
	}
	var err error
	if r.Spec.BackupVolume != nil && oldCluster.Spec.BackupVolume == nil {
		if !oldCluster.SupportStaticIP() {
			err = errors.New("forbid to add backup volume to non-static-ip cluster")
		}
	}

	newStorage := r.Spec.OBServerTemplate.Storage
	oldStorage := oldCluster.Spec.OBServerTemplate.Storage
	if newStorage.DataStorage.Size.Cmp(oldStorage.DataStorage.Size) > 0 {
		err = errors.Join(err, r.validateStorageClassAllowExpansion(newStorage.DataStorage.StorageClass))
	}
	if newStorage.LogStorage.Size.Cmp(oldStorage.LogStorage.Size) > 0 {
		err = errors.Join(err, r.validateStorageClassAllowExpansion(newStorage.LogStorage.StorageClass))
	}
	if newStorage.RedoLogStorage.Size.Cmp(oldStorage.RedoLogStorage.Size) > 0 {
		err = errors.Join(err, r.validateStorageClassAllowExpansion(newStorage.RedoLogStorage.StorageClass))
	}
	if err != nil {
		return nil, err
	}

	if newStorage.DataStorage.Size.Cmp(oldStorage.DataStorage.Size) < 0 {
		err = errors.Join(err, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("dataStorage").Child("size"), newStorage.DataStorage.Size.String(), "forbid to shrink data storage size"))
	}
	if newStorage.LogStorage.Size.Cmp(oldStorage.LogStorage.Size) < 0 {
		err = errors.Join(err, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("logStorage").Child("size"), newStorage.LogStorage.Size.String(), "forbid to shrink log storage size"))
	}
	if newStorage.RedoLogStorage.Size.Cmp(oldStorage.RedoLogStorage.Size) < 0 {
		err = errors.Join(err, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("redoLogStorage").Child("size"), newStorage.RedoLogStorage.Size.String(), "forbid to shrink redo log storage size"))
	}
	if err != nil {
		return nil, err
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
	mode, modeExist := r.GetAnnotations()[oceanbaseconst.AnnotationsMode]

	// Validate standalone
	if modeExist && mode == oceanbaseconst.ModeStandalone {
		if len(r.Spec.Topology) != 1 {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("topology"), r.Spec.Topology, "standalone mode only support single zone"))
		} else if r.Spec.Topology[0].Replica != 1 {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("topology"), r.Spec.Topology, "standalone mode only support single replica"))
		}
		// validate migration
		migrateAnnoVal, migrateAnnoExist := r.GetAnnotations()[oceanbaseconst.AnnotationsSourceClusterAddress]
		if migrateAnnoExist {
			allErrs = append(allErrs, field.Invalid(field.NewPath("metadata").Child("annotations").Child(oceanbaseconst.AnnotationsSourceClusterAddress), migrateAnnoVal, "migrate obcluster into standalone mode is not supported"))
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

	if r.Spec.OBServerTemplate.Storage.DataStorage.StorageClass == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("dataStorage").Child("storageClass"), "", "storageClass is required, default storage class is not found"))
	}
	if r.Spec.OBServerTemplate.Storage.LogStorage.StorageClass == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("logStorage").Child("storageClass"), "", "storageClass is required, default storage class is not found"))
	}
	if r.Spec.OBServerTemplate.Storage.RedoLogStorage.StorageClass == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("redoLogStorage").Child("storageClass"), "", "storageClass is required, default storage class is not found"))
	}
	if len(allErrs) != 0 {
		return allErrs.ToAggregate()
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
	annos := r.GetAnnotations()
	if annos != nil {
		if _, ok := annos[oceanbaseconst.AnnotationsSinglePVC]; ok {
			if len(storageClassMapping) > 1 {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("*").Child("storageClass"), storageClassMapping, "singlePVC mode only support single storage class"))
			}
		}
	}

	// Validate disk size
	if r.Spec.OBServerTemplate.Storage.DataStorage.Size.Cmp(resource.MustParse(obcfg.GetConfig().Resource.MinDataDiskSize)) < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("dataStorage").Child("size"), r.Spec.OBServerTemplate.Storage.DataStorage.Size.String(), "The minimum data storage size of OBCluster is "+oceanbaseconst.MinDataDiskSize.String()))
	}
	if r.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.Cmp(resource.MustParse(obcfg.GetConfig().Resource.MinRedoLogDiskSize)) < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("redoLogStorage").Child("size"), r.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.String(), "The minimum redo log storage size of OBCluster is "+oceanbaseconst.MinRedoLogDiskSize.String()))
	}
	if r.Spec.OBServerTemplate.Storage.LogStorage.Size.Cmp(resource.MustParse(obcfg.GetConfig().Resource.MinLogDiskSize)) < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("logStorage").Child("size"), r.Spec.OBServerTemplate.Storage.LogStorage.Size.String(), "The minimum log storage size of OBCluster is "+oceanbaseconst.MinLogDiskSize.String()))
	}
	// Validate memory size
	if r.Spec.OBServerTemplate.Resource.Memory.Cmp(resource.MustParse(obcfg.GetConfig().Resource.MinMemorySize)) < 0 {
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
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("parameters"), "memory limit size overflow", "Memory limit exceeds observer's resource"))
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
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("parameters"), "datafile max size overflow", "Datafile maxsize exceeds observer's data storage size"))
		}
	}

	if r.Spec.ServiceAccount != "" {
		sa := v1.ServiceAccount{}
		err := clt.Get(context.Background(), types.NamespacedName{
			Name:      r.Spec.ServiceAccount,
			Namespace: r.Namespace,
		}, &sa)
		if err != nil {
			if apierrors.IsNotFound(err) {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("serviceAccount"), r.Spec.ServiceAccount, "service account not found"))
			} else {
				allErrs = append(allErrs, field.InternalError(field.NewPath("spec").Child("serviceAccount"), err))
			}
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

func (r *OBCluster) createDefaultUserSecret(secretName string) error {
	return clt.Create(context.Background(), &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: r.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: r.APIVersion,
				Kind:       r.Kind,
				Name:       r.GetName(),
				UID:        r.GetUID(),
			}},
		},
		Type: v1.SecretTypeOpaque,
		StringData: map[string]string{
			"password": rand.String(16),
		},
	})
}

func (r *OBCluster) validateStorageClassAllowExpansion(storageClassName string) error {
	sc := storagev1.StorageClass{}
	err := clt.Get(context.Background(), types.NamespacedName{
		Name: storageClassName,
	}, &sc)
	if err != nil {
		return err
	}
	if sc.AllowVolumeExpansion == nil || !*sc.AllowVolumeExpansion {
		return fmt.Errorf("storage class %s does not allow volume expansion", storageClassName)
	}
	return nil
}
