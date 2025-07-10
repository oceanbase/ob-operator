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

	"github.com/pkg/errors"

	logger "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	bizconst "github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	insconst "github.com/oceanbase/ob-operator/internal/dashboard/business/inspection/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/oceanbase"
	insmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/inspection"
	jobmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/job"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func newPolicyFromCronJob(ctx context.Context, cronJob *batchv1.CronJob) (*insmodel.Policy, error) {
	labels := cronJob.ObjectMeta.GetLabels()
	objNamespace, ok := labels[insconst.INSPECTION_LABEL_REF_NAMESPACE]
	if !ok {
		return nil, errors.New("Failed to get object namespace from cronjob labels")
	}
	objName, ok := labels[insconst.INSPECTION_LABEL_REF_NAME]
	if !ok {
		return nil, errors.New("Failed to get object name from cronjob labels")
	}
	scenario, ok := labels[insconst.INSPECTION_LABEL_SCENARIO]
	if !ok {
		return nil, errors.New("Failed to job scenario from cronjob labels")
	}
	cluster, err := oceanbase.GetOBCluster(ctx, &param.K8sObjectIdentity{
		Namespace: objNamespace,
		Name:      objName,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get corresponding obcluster")
	}
	scheduleStatus := insmodel.ScheduleEnabled
	if cronJob.Spec.Suspend != nil && *cronJob.Spec.Suspend {
		scheduleStatus = insmodel.ScheduleDisabled
	}
	scheduleConfig := insmodel.InspectionScheduleConfig{
		Schedule: cronJob.Spec.Schedule,
		Scenario: insmodel.InspectionScenario(scenario),
	}

	reports, err := ListInspectionReports(ctx, objNamespace, objName, "", scenario)
	if err != nil {
		logger.WithError(err).Warn("failed to list inspection reports")
	}

	policy := &insmodel.Policy{
		OBCluster:       &cluster.OBClusterMeta.OBClusterMetaBasic,
		Status:          scheduleStatus,
		ScheduleConfigs: []insmodel.InspectionScheduleConfig{scheduleConfig},
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
	labelSelector := fmt.Sprintf("%s=%s,%s=%s", insconst.INSPECTION_LABEL_MANAGED_BY, bizconst.DASHBOARD_APP_NAME, insconst.INSPECTION_LABEL_JOB_TYPE, insconst.JOB_TYPE_INSPECTION)
	if namespace != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, insconst.INSPECTION_LABEL_REF_NAMESPACE, namespace)
	}
	if name != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, insconst.INSPECTION_LABEL_REF_NAME, name)
	}
	if obcluster != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, insconst.INSPECTION_LABEL_REF_OBCLUSTERNAME, obcluster)
	}
	if scenario != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, insconst.INSPECTION_LABEL_SCENARIO, scenario)
	}
	listOptions.LabelSelector = labelSelector
	cronJobList, err := client.ClientSet.BatchV1().CronJobs("").List(ctx, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list cron jobs")
	}
	return cronJobList.Items, nil
}

func ListInspectionPolicies(ctx context.Context, namespace, name, obclusterName string) ([]insmodel.Policy, error) {
	cronJobs, err := listInspectionCronJobs(ctx, namespace, name, obclusterName, "")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list cron jobs")
	}
	policyMap := make(map[string]*insmodel.Policy)
	for i := range cronJobs {
		cronJob := cronJobs[i]
		policy, err := newPolicyFromCronJob(ctx, &cronJob)
		if err != nil {
			logger.WithError(err).Errorf("Failed to parse inspection policy from cronjob, %s/%s", cronJob.Namespace, cronJob.Name)
			continue
		}
		key := fmt.Sprintf("%s/%s", policy.OBCluster.Namespace, policy.OBCluster.Name)
		value, ok := policyMap[key]
		if !ok {
			policyMap[key] = policy
		} else {
			value.ScheduleConfigs = append(value.ScheduleConfigs, policy.ScheduleConfigs...)
			value.LatestReports = append(value.LatestReports, policy.LatestReports...)
		}
	}

	obclusters, err := oceanbase.ListOBClusters(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list obclusters")
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
				OBCluster:       &cluster.OBClusterMetaBasic,
				Status:          insmodel.ScheduleDisabled,
				ScheduleConfigs: []insmodel.InspectionScheduleConfig{},
				LatestReports:   []insmodel.ReportBriefInfo{},
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
	configVolumeName := "config"
	configMountPath := "/etc/config"
	configFile := configMountPath + "/config.yaml"
	ttlSecondsAfterFinished := int32(7 * 24 * 60 * 60)

	labels := map[string]string{
		insconst.INSPECTION_LABEL_MANAGED_BY:        bizconst.DASHBOARD_APP_NAME,
		insconst.INSPECTION_LABEL_JOB_TYPE:          insconst.JOB_TYPE_INSPECTION,
		insconst.INSPECTION_LABEL_REF_NAMESPACE:     obclusterMeta.Namespace,
		insconst.INSPECTION_LABEL_REF_NAME:          obclusterMeta.Name,
		insconst.INSPECTION_LABEL_REF_OBCLUSTERNAME: obclusterMeta.ClusterName,
		insconst.INSPECTION_LABEL_SCENARIO:          string(scheduleConfig.Scenario),
	}

	jobSpec := &batchv1.JobSpec{
		TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				ServiceAccountName: "inspection-job-sa",
				RestartPolicy:      corev1.RestartPolicyNever,
				InitContainers: []corev1.Container{
					{
						Name:            "generate-config",
						Image:           "oceanbase/oceanbase-helper:latest",
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
						Image:           "oceanbase/obdiag:latest",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command:         []string{"obdiag", "check", "run", "-c", configFile, "--inner_config", "obdiag.logger.silent=Ture"},
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

	ownerRef := metav1.OwnerReference{
		APIVersion: v1alpha1.GroupVersion.String(),
		Kind:       "OBCluster",
		Name:       obclusterMeta.Name,
		UID:        types.UID(obclusterMeta.UID),
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
	client := client.GetClient()
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
	_, err := client.ClientSet.BatchV1().CronJobs("").Update(ctx, cronJob, metav1.UpdateOptions{})
	return err
}

func CreateOrUpdateInspectionPolicy(ctx context.Context, policy *insmodel.Policy) error {
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
	labelSelector := fmt.Sprintf("%s=%s,%s=%s", insconst.INSPECTION_LABEL_MANAGED_BY, bizconst.DASHBOARD_APP_NAME, insconst.INSPECTION_LABEL_JOB_TYPE, insconst.JOB_TYPE_INSPECTION)
	if namespace != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, insconst.INSPECTION_LABEL_REF_NAMESPACE, namespace)
	}
	if name != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, insconst.INSPECTION_LABEL_REF_NAME, name)
	}
	if obcluster != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, insconst.INSPECTION_LABEL_REF_OBCLUSTERNAME, obcluster)
	}
	if scenario != "" {
		labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, insconst.INSPECTION_LABEL_SCENARIO, scenario)
	}
	listOptions.LabelSelector = labelSelector
	jobList, err := client.ClientSet.BatchV1().Jobs("").List(ctx, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list jobs")
	}
	return jobList.Items, nil
}

func ListInspectionReports(ctx context.Context, namespace, name, obcluster, scenario string) ([]insmodel.ReportBriefInfo, error) {
	jobs, err := listInspectionJobs(ctx, namespace, name, obcluster, scenario)
	logger.Infof("Found %d corresponding jobs", len(jobs))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list jobs")
	}
	reports := make([]insmodel.ReportBriefInfo, 0, len(jobs))
	for idx, job := range jobs {
		report, err := newReportFromJob(ctx, &jobs[idx])
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
	return newReportFromJob(ctx, job)
}

func newReportFromJob(ctx context.Context, job *batchv1.Job) (*insmodel.Report, error) {
	labels := job.ObjectMeta.GetLabels()
	objNamespace, ok := labels[insconst.INSPECTION_LABEL_REF_NAMESPACE]
	if !ok {
		return nil, errors.New("Failed to get object namespace from job labels")
	}
	objName, ok := labels[insconst.INSPECTION_LABEL_REF_NAME]
	if !ok {
		return nil, errors.New("Failed to get object name from job labels")
	}
	scenario, ok := labels[insconst.INSPECTION_LABEL_SCENARIO]
	if !ok {
		return nil, errors.New("Failed to job scenario from job labels")
	}
	cluster, err := oceanbase.GetOBCluster(ctx, &param.K8sObjectIdentity{
		Namespace: objNamespace,
		Name:      objName,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get corresponding obcluster")
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
