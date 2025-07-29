/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package inspection

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/oceanbase/ob-operator/internal/dashboard/config"
	"github.com/pkg/errors"

	logger "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	bizconst "github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	insconst "github.com/oceanbase/ob-operator/internal/dashboard/business/inspection/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/oceanbase"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/inspection"
	insmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/inspection"
	jobmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/job"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func newPolicyFromCronJob(cronJob *batchv1.CronJob, cluster *response.OBClusterOverview, reports []insmodel.ReportBriefInfo) (*insmodel.Policy, error) {
	labels := cronJob.ObjectMeta.GetLabels()
	_, ok := labels[bizconst.LABEL_REF_NAMESPACE]
	if !ok {
		return nil, errors.New("Failed to get object namespace from cronjob labels")
	}
	_, ok = labels[bizconst.LABEL_REF_NAME]
	if !ok {
		return nil, errors.New("Failed to get object name from cronjob labels")
	}
	scenario, ok := labels[bizconst.INSPECTION_SCENARIO]
	if !ok {
		return nil, errors.New("Failed to job scenario from cronjob labels")
	}
	if cluster == nil {
		return nil, errors.New("cluster is nil")
	}
	scheduleStatus := insmodel.ScheduleEnabled
	if cronJob.Spec.Suspend != nil && *cronJob.Spec.Suspend {
		scheduleStatus = insmodel.ScheduleDisabled
	}
	scheduleConfig := insmodel.InspectionScheduleConfig{
		Schedule: cronJob.Spec.Schedule,
		Scenario: insmodel.InspectionScenario(scenario),
	}

	policy := &insmodel.Policy{
		PolicyMeta: insmodel.PolicyMeta{
			OBCluster:       &cluster.OBClusterMeta.OBClusterMetaBasic,
			Status:          scheduleStatus,
			ScheduleConfigs: []insmodel.InspectionScheduleConfig{scheduleConfig},
		},
	}

	if len(reports) > 0 {
		sort.Slice(reports, func(i, j int) bool {
			return reports[i].FinishTime > reports[j].FinishTime
		})
		policy.LatestReports = []insmodel.ReportBriefInfo{reports[0]}
	}

	return policy, nil
}

func listInspectionCronJobs(ctx context.Context, namespace, name, obcluster, scenario string) ([]batchv1.CronJob, error) {
	client := client.GetClient()
	listOptions := metav1.ListOptions{}
	labelSelector := fmt.Sprintf("%s=%s,%s=%s", bizconst.LABEL_MANAGED_BY, bizconst.DASHBOARD_APP_NAME, bizconst.LABEL_JOB_TYPE, bizconst.JOB_TYPE_INSPECTION)
	if namespace != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, bizconst.LABEL_REF_NAMESPACE, namespace)
	}
	if name != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, bizconst.LABEL_REF_NAME, name)
	}
	if obcluster != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, bizconst.LABEL_REF_OBCLUSTERNAME, obcluster)
	}
	if scenario != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, bizconst.INSPECTION_SCENARIO, scenario)
	}
	listOptions.LabelSelector = labelSelector
	cronJobList, err := client.ClientSet.BatchV1().CronJobs("").List(ctx, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list cron jobs")
	}
	return cronJobList.Items, nil
}

func ListInspectionPolicies(ctx context.Context, namespace, name, obclusterName string) ([]insmodel.Policy, error) {
	obclusters, err := oceanbase.ListOBClusters(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list obclusters")
	}
	clusterMap := make(map[string]response.OBClusterOverview)
	for i := range obclusters {
		key := fmt.Sprintf("%s/%s", obclusters[i].Namespace, obclusters[i].Name)
		clusterMap[key] = obclusters[i]
	}

	reports, err := ListInspectionReports(ctx, namespace, name, obclusterName, "")
	if err != nil {
		logger.WithError(err).Warn("failed to list inspection reports")
	}
	reportsMap := make(map[string]map[string][]insmodel.ReportBriefInfo)
	for i := range reports {
		report := reports[i]
		key := fmt.Sprintf("%s/%s", report.OBCluster.Namespace, report.OBCluster.Name)
		scenario := string(report.Scenario)
		if _, ok := reportsMap[key]; !ok {
			reportsMap[key] = make(map[string][]insmodel.ReportBriefInfo)
		}
		reportsMap[key][scenario] = append(reportsMap[key][scenario], report)
	}

	cronJobs, err := listInspectionCronJobs(ctx, namespace, name, obclusterName, "")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list cron jobs")
	}
	policyMap := make(map[string]*insmodel.Policy)
	for i := range cronJobs {
		cronJob := cronJobs[i]
		labels := cronJob.ObjectMeta.GetLabels()
		objNamespace, ok := labels[bizconst.LABEL_REF_NAMESPACE]
		if !ok {
			logger.Errorf("Failed to get object namespace from cronjob labels for %s/%s", cronJob.Namespace, cronJob.Name)
			continue
		}
		objName, ok := labels[bizconst.LABEL_REF_NAME]
		if !ok {
			logger.Errorf("Failed to get object name from cronjob labels for %s/%s", cronJob.Namespace, cronJob.Name)
			continue
		}
		scenario, ok := labels[bizconst.INSPECTION_SCENARIO]
		if !ok {
			logger.Errorf("Failed to get scenario from cronjob labels for %s/%s", cronJob.Namespace, cronJob.Name)
			continue
		}

		key := fmt.Sprintf("%s/%s", objNamespace, objName)
		cluster, ok := clusterMap[key]
		if !ok {
			// It's possible that the cluster is not in the list, so we just log it and continue
			logger.Warnf("cluster %s not found, may be deleted", key)
			continue
		}

		var scenarioReports []insmodel.ReportBriefInfo
		if clusterReports, ok := reportsMap[key]; ok {
			scenarioReports = clusterReports[scenario]
		}

		policy, err := newPolicyFromCronJob(&cronJob, &cluster, scenarioReports)
		if err != nil {
			logger.WithError(err).Errorf("Failed to parse inspection policy from cronjob, %s/%s", cronJob.Namespace, cronJob.Name)
			continue
		}

		value, ok := policyMap[key]
		if !ok {
			policyMap[key] = policy
		} else {
			value.ScheduleConfigs = append(value.ScheduleConfigs, policy.ScheduleConfigs...)
			value.LatestReports = append(value.LatestReports, policy.LatestReports...)
		}
	}

	filteredOBClusters := make([]response.OBClusterOverview, 0)
	for _, cluster := range obclusters {
		if namespace != "" && cluster.Namespace != namespace {
			continue
		}
		if name != "" && cluster.Name != name {
			continue
		}
		if obclusterName != "" && cluster.ClusterName != obclusterName {
			continue
		}
		filteredOBClusters = append(filteredOBClusters, cluster)
	}

	policies := make([]insmodel.Policy, 0, len(filteredOBClusters))
	for _, cluster := range filteredOBClusters {
		key := fmt.Sprintf("%s/%s", cluster.Namespace, cluster.Name)
		if policy, ok := policyMap[key]; ok {
			policies = append(policies, *policy)
		} else {
			policies = append(policies, insmodel.Policy{
				PolicyMeta: insmodel.PolicyMeta{
					OBCluster:       &cluster.OBClusterMetaBasic,
					Status:          insmodel.ScheduleDisabled,
					ScheduleConfigs: []insmodel.InspectionScheduleConfig{},
				},
				LatestReports: []insmodel.ReportBriefInfo{},
			})
		}
	}
	return policies, nil
}

func GetInspectionPolicy(ctx context.Context, namespace, name string) (*insmodel.Policy, error) {
	policies, err := ListInspectionPolicies(ctx, namespace, name, "")
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to fetch inspection jobs for %s/%s", namespace, name)
	}
	if len(policies) != 1 {
		return nil, errors.New("Policy not found or found multiple")
	}
	policy := policies[0]
	return &policy, nil
}

func DeleteInspectionPolicy(ctx context.Context, namespace, name, scenario string) error {
	cronJobs, err := listInspectionCronJobs(ctx, namespace, name, "", scenario)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to list all cronjobs for inspection policy of obcluster %s/%s, scenario %s", namespace, name, scenario))
	}
	client := client.GetClient()
	delErrs := make([]error, 0)
	for _, cronJob := range cronJobs {
		err := client.ClientSet.BatchV1().CronJobs(cronJob.Namespace).Delete(ctx, cronJob.Name, metav1.DeleteOptions{})
		if err != nil {
			logger.WithError(err).Errorf("Failed to delete inspection cronjob for object %s/%s, scenario: %s", namespace, name, scenario)
			delErrs = append(delErrs, err)
		}
	}
	if len(delErrs) > 0 {
		return errors.Errorf("Failed to delete inspection policy for object %s/%s, scenario: %s", namespace, name, scenario)
	}
	return nil
}

func createCronJobForInspection(ctx context.Context, obclusterMeta *response.OBClusterMetaBasic, scheduleConfig *insmodel.InspectionScheduleConfig) error {
	cronJobName := fmt.Sprintf("ins-%s-%s", scheduleConfig.Scenario, rand.String(6))
	pvcName := "pvc-" + cronJobName
	configVolumeName := insconst.ConfigVolumeName
	configMountPath := insconst.ConfigMountPath
	configFile := configMountPath + "/config.yaml"
	ttlSecondsAfterFinished := config.GetConfig().Job.Inspection.TTLSecondsAfterFinished
	serviceAccountName := fmt.Sprintf(insconst.ServiceAccountNameFmt, obclusterMeta.Name)
	clusterRoleName := insconst.ClusterRoleName
	clusterRoleBindingName := fmt.Sprintf(insconst.ClusterRoleBindingNameFmt, obclusterMeta.Namespace, obclusterMeta.Name)
	checkPackage := insconst.InspectionPackageBasic
	if scheduleConfig.Scenario == inspection.ScenarioPerformance {
		checkPackage = insconst.InspectionPackagePerformance
	}

	labels := map[string]string{
		bizconst.LABEL_MANAGED_BY:        bizconst.DASHBOARD_APP_NAME,
		bizconst.LABEL_JOB_TYPE:          bizconst.JOB_TYPE_INSPECTION,
		bizconst.LABEL_REF_NAMESPACE:     obclusterMeta.Namespace,
		bizconst.LABEL_REF_NAME:          obclusterMeta.Name,
		bizconst.LABEL_REF_OBCLUSTERNAME: obclusterMeta.ClusterName,
		bizconst.INSPECTION_SCENARIO:     string(scheduleConfig.Scenario),
	}

	ownerRef := metav1.OwnerReference{
		APIVersion: v1alpha1.GroupVersion.String(),
		Kind:       "OBCluster",
		Name:       obclusterMeta.Name,
		UID:        types.UID(obclusterMeta.UID),
	}

	// Create ServiceAccount and ClusterRoleBinding
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:            serviceAccountName,
			Namespace:       obclusterMeta.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerRef},
		},
	}
	client := client.GetClient()
	if _, err := client.ClientSet.CoreV1().ServiceAccounts(obclusterMeta.Namespace).Create(ctx, sa, metav1.CreateOptions{}); err != nil && !k8serrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "Failed to create service account")
	}

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: obclusterMeta.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	if _, err := client.ClientSet.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{}); err != nil && !k8serrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "Failed to create cluster role binding")
	}

	jobSpec := &batchv1.JobSpec{
		TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				ServiceAccountName: serviceAccountName,
				RestartPolicy:      corev1.RestartPolicyNever,
				InitContainers: []corev1.Container{
					{
						Name:            "generate-config",
						Image:           config.GetConfig().Inspection.OBHelper.Image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command:         []string{"bash", "-c", fmt.Sprintf("/home/admin/oceanbase/bin/oceanbase-helper generate obdiag-config -n %s -c %s -o %s", obclusterMeta.Namespace, obclusterMeta.Name, configFile)},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      configVolumeName,
								MountPath: configMountPath,
							},
						},
					},
				},
				Containers: []corev1.Container{
					{
						Name:            "inspection",
						Image:           config.GetConfig().Inspection.OBDiag.Image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command:         []string{"bash", "-c", fmt.Sprintf("obdiag check run --cases %s -c %s --inner_config obdiag.logger.silent=Ture && rm -f %s", checkPackage, configFile, configFile)},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      configVolumeName,
								MountPath: configMountPath,
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: configVolumeName,
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: pvcName,
							},
						},
					},
				},
			},
		},
	}

	spec := &batchv1.CronJobSpec{
		Schedule: scheduleConfig.Schedule,
		JobTemplate: batchv1.JobTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: *jobSpec,
		},
	}

	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cronJobName,
			Namespace:       obclusterMeta.Namespace,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{ownerRef},
		},
		Spec: *spec,
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:            pvcName,
			Namespace:       obclusterMeta.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerRef},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}

	logger.Infof("Create pvc for cronjob %s", cronJobName)
	pvcObject, err := client.ClientSet.CoreV1().PersistentVolumeClaims(obclusterMeta.Namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to create pvc")
	}
	logger.Infof("Successfully created pvc %v", pvcObject)

	logger.Infof("Create cronjob %s", cronJobName)
	cronjobObject, err := client.ClientSet.BatchV1().CronJobs(obclusterMeta.Namespace).Create(ctx, cronJob, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to create cronjob")
	}
	logger.Infof("Successfully created cronjob %v", cronjobObject)
	return nil
}

func updateCronJobForInspection(ctx context.Context, cronJob *batchv1.CronJob, suspend bool, scheduleConfig *insmodel.InspectionScheduleConfig) error {
	cronJob.Spec.Suspend = &suspend
	cronJob.Spec.Schedule = scheduleConfig.Schedule
	client := client.GetClient()
	_, err := client.ClientSet.BatchV1().CronJobs(cronJob.Namespace).Update(ctx, cronJob, metav1.UpdateOptions{})
	return err
}

func CreateOrUpdateInspectionPolicy(ctx context.Context, policy *insmodel.PolicyMeta) error {
	for _, scheduleConfig := range policy.ScheduleConfigs {
		cronJobs, err := listInspectionCronJobs(ctx, policy.OBCluster.Namespace, policy.OBCluster.Name, policy.OBCluster.ClusterName, string(scheduleConfig.Scenario))
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to list all cronjobs for inspection policy of obcluster %s/%s, scenario %s", policy.OBCluster.Namespace, policy.OBCluster.Name, scheduleConfig.Scenario))
		}
		if len(cronJobs) > 1 {
			return errors.Errorf("Found multiple cronjobs for inspection object %s/%s, scenario: %s", policy.OBCluster.Namespace, policy.OBCluster.Name, scheduleConfig.Scenario)
		} else if len(cronJobs) == 0 {
			err := createCronJobForInspection(ctx, policy.OBCluster, &scheduleConfig)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Failed to create cronjob for inspection object %s/%s, scenario: %s", policy.OBCluster.Namespace, policy.OBCluster.Name, scheduleConfig.Scenario))
			}
		} else {
			err := updateCronJobForInspection(ctx, &cronJobs[0], policy.Status == insmodel.ScheduleDisabled, &scheduleConfig)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Failed to update cronjob for inspection object %s/%s, scenario: %s", policy.OBCluster.Namespace, policy.OBCluster.Name, scheduleConfig.Scenario))
			}
		}
	}
	return nil
}

func TriggerInspection(ctx context.Context, namespace, name, scenario string) (*jobmodel.Job, error) {
	cronJobs, err := listInspectionCronJobs(ctx, namespace, name, "", scenario)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to list all cronjobs for inspection policy of obcluster %s/%s, scenario %s", namespace, name, scenario))
	}
	if len(cronJobs) != 1 {
		return nil, errors.New("No cronjob found or found multiple cronjobs for the same scenario")
	}
	cronJob := cronJobs[0]
	jobName := fmt.Sprintf("ins-%s-%s", scenario, rand.String(6))
	triggeredJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: cronJob.Namespace,
			Labels:    cronJob.Labels,
		},
		Spec: cronJob.Spec.JobTemplate.Spec,
	}
	client := client.GetClient()
	createdJob, err := client.ClientSet.BatchV1().Jobs(cronJob.Namespace).Create(ctx, triggeredJob, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create job")
	}
	return &jobmodel.Job{
		Name:      createdJob.Name,
		Namespace: createdJob.Namespace,
		Status:    jobmodel.JobStatusPending,
	}, nil
}

func listInspectionJobs(ctx context.Context, namespace, name, obcluster, scenario string) ([]batchv1.Job, error) {

	client := client.GetClient()
	listOptions := metav1.ListOptions{}
	labelSelector := fmt.Sprintf("%s=%s,%s=%s", bizconst.LABEL_MANAGED_BY, bizconst.DASHBOARD_APP_NAME, bizconst.LABEL_JOB_TYPE, bizconst.JOB_TYPE_INSPECTION)
	if namespace != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, bizconst.LABEL_REF_NAMESPACE, namespace)
	}
	if name != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, bizconst.LABEL_REF_NAME, name)
	}
	if obcluster != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, bizconst.LABEL_REF_OBCLUSTERNAME, obcluster)
	}
	if scenario != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, bizconst.INSPECTION_SCENARIO, scenario)
	}
	listOptions.LabelSelector = labelSelector
	jobList, err := client.ClientSet.BatchV1().Jobs("").List(ctx, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list jobs")
	}
	return jobList.Items, nil
}

func ListInspectionReports(ctx context.Context, namespace, name, obcluster, scenario string) ([]insmodel.ReportBriefInfo, error) {
	obclusters, err := oceanbase.ListOBClusters(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list obclusters")
	}
	clusterMap := make(map[string]response.OBClusterOverview)
	for i := range obclusters {
		key := fmt.Sprintf("%s/%s", obclusters[i].Namespace, obclusters[i].Name)
		clusterMap[key] = obclusters[i]
	}

	jobs, err := listInspectionJobs(ctx, namespace, name, obcluster, scenario)
	logger.Infof("Found %d corresponding jobs", len(jobs))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list jobs")
	}
	reports := make([]insmodel.ReportBriefInfo, 0, len(jobs))
	for idx, job := range jobs {
		labels := job.ObjectMeta.GetLabels()
		objNamespace, ok := labels[bizconst.LABEL_REF_NAMESPACE]
		if !ok {
			logger.Errorf("Failed to get object namespace from job labels for %s/%s", job.Namespace, job.Name)
			continue
		}
		objName, ok := labels[bizconst.LABEL_REF_NAME]
		if !ok {
			logger.Errorf("Failed to get object name from job labels for %s/%s", job.Namespace, job.Name)
			continue
		}
		key := fmt.Sprintf("%s/%s", objNamespace, objName)
		cluster, ok := clusterMap[key]
		if !ok {
			logger.Warnf("cluster %s not found, may be deleted", key)
			continue
		}
		report, err := newReportFromJob(ctx, &jobs[idx], &cluster)
		if err != nil {
			logger.WithError(err).Errorf("Failed to parse report from job, %s/%s", job.Namespace, job.Name)
			continue
		}
		reports = append(reports, report.ReportBriefInfo)
	}
	return reports, nil
}

func GetInspectionReport(ctx context.Context, namespace, name string) (*insmodel.Report, error) {
	client := client.GetClient()
	job, err := client.ClientSet.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get job")
	}
	labels := job.ObjectMeta.GetLabels()
	objNamespace, ok := labels[bizconst.LABEL_REF_NAMESPACE]
	if !ok {
		return nil, errors.New("Failed to get object namespace from job labels")
	}
	objName, ok := labels[bizconst.LABEL_REF_NAME]
	if !ok {
		return nil, errors.New("Failed to get object name from job labels")
	}
	cluster, err := oceanbase.GetOBCluster(ctx, &param.K8sObjectIdentity{
		Namespace: objNamespace,
		Name:      objName,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get corresponding obcluster")
	}
	return newReportFromJob(ctx, job, &cluster.OBClusterOverview)
}

func newReportFromJob(ctx context.Context, job *batchv1.Job, cluster *response.OBClusterOverview) (*insmodel.Report, error) {
	labels := job.ObjectMeta.GetLabels()
	_, ok := labels[bizconst.LABEL_REF_NAMESPACE]
	if !ok {
		return nil, errors.New("Failed to get object namespace from job labels")
	}
	_, ok = labels[bizconst.LABEL_REF_NAME]
	if !ok {
		return nil, errors.New("Failed to get object name from job labels")
	}
	scenario, ok := labels[bizconst.INSPECTION_SCENARIO]
	if !ok {
		return nil, errors.New("Failed to job scenario from job labels")
	}
	if cluster == nil {
		return nil, errors.New("cluster is nil")
	}

	status := jobmodel.JobStatusPending
	if job.Status.Succeeded > 0 {
		status = jobmodel.JobStatusSuccessful
	} else if job.Status.Failed > 0 {
		status = jobmodel.JobStatusFailed
	} else if job.Status.Active > 0 {
		status = jobmodel.JobStatusRunning
	}

	report := &insmodel.Report{
		ReportBriefInfo: insmodel.ReportBriefInfo{
			Namespace:        job.Namespace,
			Name:             job.Name,
			OBCluster:        cluster.OBClusterMeta,
			Scenario:         insmodel.InspectionScenario(scenario),
			Status:           status,
			ResultStatistics: insmodel.ResultStatistics{},
		},
		ResultDetail: insmodel.ResultDetail{},
	}
	if job.Status.StartTime != nil {
		report.StartTime = job.Status.StartTime.Unix()
	}
	if job.Status.CompletionTime != nil {
		report.FinishTime = job.Status.CompletionTime.Unix()
	}

	if status == jobmodel.JobStatusSuccessful {
		pod, err := getPodFromJob(ctx, job)
		if err != nil {
			return nil, err
		}
		logs, err := getPodLogs(ctx, pod)
		if err != nil {
			return nil, err
		}

		type ObdiagResult struct {
			Data struct {
				Observer struct {
					Fail     map[string][]string `json:"fail"`
					Critical map[string][]string `json:"critical"`
					Warning  map[string][]string `json:"warning"`
					All      map[string][]string `json:"all"`
				} `json:"observer"`
			} `json:"data"`
		}

		var result ObdiagResult
		if err := json.Unmarshal([]byte(logs), &result); err != nil {
			logger.WithError(err).Error("Failed to unmarshal obdiag result")
		} else {
			report.ResultDetail.FailedItems = newInspectionItemsFromMap(result.Data.Observer.Fail)
			report.ResultDetail.CriticalItems = newInspectionItemsFromMap(result.Data.Observer.Critical)
			report.ResultDetail.ModerateItems = newInspectionItemsFromMap(result.Data.Observer.Warning)
			report.ResultStatistics.FailedCount = len(report.ResultDetail.FailedItems)
			report.ResultStatistics.CriticalCount = len(report.ResultDetail.CriticalItems)
			report.ResultStatistics.ModerateCount = len(report.ResultDetail.ModerateItems)

			// Populate NegligibleItems with "all pass" results
			negligibleItems := make([]insmodel.InspectionItem, 0)
			for name, results := range result.Data.Observer.All {
				if len(results) == 1 && results[0] == "all pass" {
					negligibleItems = append(negligibleItems, insmodel.InspectionItem{
						Name:    name,
						Results: results,
					})
				}
			}
			report.ResultDetail.NegligibleItems = negligibleItems
		}
	}

	return report, nil
}

func getPodFromJob(ctx context.Context, job *batchv1.Job) (*corev1.Pod, error) {
	client := client.GetClient()
	selector := labels.Set(job.Spec.Selector.MatchLabels).String()
	pods, err := client.ClientSet.CoreV1().Pods(job.Namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list pods")
	}
	if len(pods.Items) == 0 {
		return nil, errors.New("No pod found for job")
	}
	return &pods.Items[0], nil
}

func getPodLogs(ctx context.Context, pod *corev1.Pod) (string, error) {
	client := client.GetClient()
	podLogOpts := corev1.PodLogOptions{}
	req := client.ClientSet.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", errors.Wrap(err, "error in opening stream")
	}
	defer podLogs.Close()

	logs, err := io.ReadAll(podLogs)
	if err != nil {
		return "", errors.Wrap(err, "error in read logs")
	}
	return string(logs), nil
}

func newInspectionItemsFromMap(m map[string][]string) []insmodel.InspectionItem {
	items := make([]insmodel.InspectionItem, 0, len(m))
	for name, results := range m {
		items = append(items, insmodel.InspectionItem{
			Name:    name,
			Results: results,
		})
	}
	return items
}
