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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	myconfig "github.com/oceanbase/ob-operator/pkg/config"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/cable"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type UpgradeInfo struct {
	ScriptPassedVersion string
	TargetVersion       string
	UpgradeRoute        []string
	ZoneStatus          string
	ClusterStatus       string
	SingleZoneStatus    map[string]string
}

const (
	ExecCheckScriptsCMDTemplate = "python2 ${FILE_NAME} -h${IP} -P${PORT} -uroot"
)

func UpgradeReplacer(filename, clusterIP, port string) *strings.Replacer {
	return strings.NewReplacer("${FILE_NAME}", filename, "${IP}", clusterIP, "${PORT}", port)
}

func GenerateJobName(clusterName, name string) string {
	return fmt.Sprintf("%s-%s", clusterName, name)
}

func (ctrl *OBClusterCtrl) GenerateJobObject(jobName, image string, cmd []string) batchv1.Job {
	var backOffLimit int32
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: ctrl.OBCluster.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    jobName,
							Image:   image,
							Command: cmd,
							Env: []corev1.EnvVar{
								{
									Name:  "LD_LIBRARY_PATH",
									Value: "/home/admin/oceanbase/lib",
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit: &backOffLimit,
		},
	}

	return job
}

func GeneratePodName(clusterName, name string) string {
	return fmt.Sprintf("%s-%s", clusterName, name)
}

func (ctrl *OBClusterCtrl) CreatePodForVersion(podName string) error {
	containerImage := fmt.Sprint(ctrl.OBCluster.Spec.ImageRepo, ":", ctrl.OBCluster.Spec.Tag)
	podObject := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ctrl.OBCluster.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  podName,
					Image: containerImage,
					Env: []corev1.EnvVar{
						{
							Name:  "LD_LIBRARY_PATH",
							Value: "/home/admin/oceanbase/lib",
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	// create pod
	podExecuter := resource.NewPodResource(ctrl.Resource)
	err := podExecuter.Create(context.TODO(), podObject)
	if err != nil {
		if kubeerrors.IsAlreadyExists(err) {
			return nil
		}
		klog.Errorln("create pod to get version failed, error: ", err)
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) GetPodIpByName(podName string) (string, error) {
	podExecuter := resource.NewPodResource(ctrl.Resource)
	podObject, err := podExecuter.Get(context.TODO(), ctrl.OBCluster.Namespace, podName)
	if err != nil {
		klog.Errorln("Get PodIp By PodName failed, err: ", err)
		return "", err
	}
	pod := podObject.(corev1.Pod)
	return pod.Status.PodIP, nil
}

func (ctrl *OBClusterCtrl) GetTargetVer() (string, error) {
	podName := GeneratePodName(myconfig.ClusterName, "help")
	err := ctrl.CreatePodForVersion(podName)
	if err != nil {
		return "", err
	}
	podIp, err := ctrl.GetPodIpByName(podName)
	if err != nil {
		return "", err
	}
	time.Sleep(1 * time.Second)
	return cable.OBServerGetVersion(podIp)
}

func (ctrl *OBClusterCtrl) GetUpgradeRoute(currentVer, targetVer string) ([]string, error) {
	var upgradeRoute []string
	podName := GeneratePodName(myconfig.ClusterName, "help")
	podIp, err := ctrl.GetPodIpByName(podName)
	if err != nil {
		return upgradeRoute, err
	}
	return cable.OBServerGetUpgradeRoute(podIp, currentVer, targetVer)
}

func (ctrl *OBClusterCtrl) getCurrentVersion(statefulApp cloudv1.StatefulApp) (string, error) {
	subsets := statefulApp.Status.Subsets
	for subsetsIdx, _ := range subsets {
		for _, pod := range subsets[subsetsIdx].Pods {
			return cable.OBServerGetVersion(pod.PodIP)
		}
	}
	return "", nil
}

func (ctrl *OBClusterCtrl) CheckTargetVersion(currentTargetVersion string) error {
	if currentTargetVersion == "" {
		targetVersion, err := ctrl.GetTargetVer()
		if err != nil {
			return err
		}
		klog.Infoln("OBCluster Upgrade Target Verson is ", targetVersion)
		upgradeInfo := UpgradeInfo{
			TargetVersion: targetVersion,
		}
		return ctrl.UpdateOBStatusForUpgrade(upgradeInfo)
	} // else if currentTargetVersion != targetVersion {
	// 	klog.Errorln("Can not upgrade OB to another version when current upgrading is not finished")
	// 	return errors.New("Can not upgrade OB to another version when current upgrading is not finished")
	// }
	return nil
}

func (ctrl *OBClusterCtrl) CheckUpgradeRoute(statefulApp cloudv1.StatefulApp, upgradeRoute []string, targetVer string) error {
	if upgradeRoute == nil {
		currentVer, err := ctrl.getCurrentVersion(statefulApp)
		if err != nil {
			return err
		}
		upgradeRoute, err = ctrl.GetUpgradeRoute(currentVer, targetVer)
		if err != nil {
			return err
		}
		upgradeInfo := UpgradeInfo{
			UpgradeRoute: upgradeRoute,
		}
		return ctrl.UpdateOBStatusForUpgrade(upgradeInfo)
	}
	// podName := GeneratePodName(myconfig.ClusterName, "help")
	// podExecuter := resource.NewPodResource(ctrl.Resource)
	// podObject, err := podExecuter.Get(context.TODO(), ctrl.OBCluster.Namespace, podName)
	// if err != nil {
	// 	klog.Errorln("Get PodIp By PodName failed, err: ", err)
	// 	return err
	// }
	// return podExecuter.Delete(context.TODO(), podObject)
	return nil
}

func (ctrl *OBClusterCtrl) ExecUpgradePreChecker(statefulApp cloudv1.StatefulApp) error {
	name := "pre-checker"
	jobName := GenerateJobName(myconfig.ClusterName, name)
	containerImage := fmt.Sprint(ctrl.OBCluster.Spec.ImageRepo, ":", ctrl.OBCluster.Spec.Tag)

	rsIP, err := ctrl.GetRsIPFromDB(statefulApp)
	if err != nil {
		return err
	}
	var cmdList []string
	cmd := sql.ReplaceAll(ExecCheckScriptsCMDTemplate, UpgradeReplacer(observerconst.UpgradePreCheckerPath, rsIP, strconv.Itoa(observerconst.MysqlPort)))
	cmdList = append(cmdList, "bash", "-c", cmd)
	jobObject := ctrl.GenerateJobObject(jobName, containerImage, cmdList)
	jobExecuter := resource.NewJobResource(ctrl.Resource)
	err = jobExecuter.Create(context.TODO(), jobObject)
	if err != nil {
		klog.Errorln("Create ", jobName, " job failed, err: ", err)
		return err
	}
	return ctrl.UpdateOBClusterAndZoneStatus(observerconst.UpgradeChecking, "", "")
}

func (ctrl *OBClusterCtrl) ExecUpgradePostChecker(statefulApp cloudv1.StatefulApp) error {
	name := "post-checker"
	jobName := GenerateJobName(myconfig.ClusterName, name)
	containerImage := fmt.Sprint(ctrl.OBCluster.Spec.ImageRepo, ":", ctrl.OBCluster.Spec.Tag)

	jobExecuter := resource.NewJobResource(ctrl.Resource)
	jobObject, err := jobExecuter.Get(context.TODO(), ctrl.OBCluster.Namespace, jobName)

	if err != nil {
		if kubeerrors.IsNotFound(err) {
			rsIP, err := ctrl.GetRsIPFromDB(statefulApp)
			if err != nil {
				return err
			}
			var cmdList []string
			cmd := sql.ReplaceAll(ExecCheckScriptsCMDTemplate, UpgradeReplacer(observerconst.UpgradePostCheckerPath, rsIP, strconv.Itoa(observerconst.MysqlPort)))
			cmdList = append(cmdList, "bash", "-c", cmd)
			jobObject := ctrl.GenerateJobObject(jobName, containerImage, cmdList)
			return jobExecuter.Create(context.TODO(), jobObject)
		} else {
			klog.Errorln("Get ", jobName, " job failed, err: ", err)
			return err
		}
	}
	job := jobObject.(batchv1.Job)
	if job.Status.Succeeded == 0 && job.Status.Failed == 0 {
		return nil
	}
	if job.Status.Succeeded == 1 {
		err = ctrl.UpdateOBClusterAndZoneStatus(observerconst.NeedExecutingPreScripts, "", "")
	}
	if job.Status.Failed == 1 {
		err = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, "", "")
	}
	if err != nil {
		return err
	}
	err = jobExecuter.Delete(context.TODO(), jobObject)
	if err != nil {
		return err
	}
	newStatefulApp := converter.UpdateStatefulAppImage(statefulApp, containerImage)
	statefulAppCtrl := NewStatefulAppCtrl(ctrl, newStatefulApp)
	err = statefulAppCtrl.UpdateStatefulApp()
	if err != nil {
		return err
	}
	return ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, "", "")
}

func (ctrl *OBClusterCtrl) GetPreCheckJobStatus(statefulApp cloudv1.StatefulApp) error {
	name := "pre-checker"
	jobName := GenerateJobName(myconfig.ClusterName, name)
	jobExecuter := resource.NewJobResource(ctrl.Resource)
	jobObject, err := jobExecuter.Get(context.TODO(), ctrl.OBCluster.Namespace, jobName)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			// 是否需要跳转回去新建 job
		} else {
			klog.Errorln("Get ", jobName, " job failed, err: ", err)
			return err
		}
	}
	job := jobObject.(batchv1.Job)
	if job.Status.Succeeded == 0 && job.Status.Failed == 0 {
		return nil
	}
	if job.Status.Succeeded == 1 {
		err = ctrl.UpdateOBClusterAndZoneStatus(observerconst.NeedExecutingPreScripts, "", "")
	}
	if job.Status.Failed == 1 {
		err = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, "", "")
	}
	if err != nil {
		return err
	}
	return jobExecuter.Delete(context.TODO(), jobObject)
}

func (ctrl *OBClusterCtrl) PrepareForPostCheck(statefulApp cloudv1.StatefulApp) error {
	err := ctrl.SetMinVersion()
	if err != nil {
		klog.Errorln(fmt.Sprint("Set Min OB Server Version Error : ", err))
		return err
	}
	err = ctrl.EndUpgrade()
	if err != nil {
		klog.Errorln(fmt.Sprint("End Upgrade Error : ", err))
		return err
	}
	err = ctrl.CheckUpgradeModeAfter()
	if err != nil {
		klog.Errorln(fmt.Sprint("Check Upgrade Mode (End) Error :", err))
		return err
	}

	err = ctrl.RunRootInspection()
	if err != nil {
		klog.Infoln(fmt.Sprint("Run Root Inspection Job Error: ", err))
		return err
	}
	return ctrl.UpdateOBClusterAndZoneStatus(observerconst.UpgradePostChecking, "", "")
}

func (ctrl *OBClusterCtrl) RunRootInspection() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator When Running Root Inspection Job")
	}
	return sqlOperator.RunRootInspection()
}

func (ctrl *OBClusterCtrl) CheckUpgradeModeAfter() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator When End Upgrade")
	}
	zoneUpGradeMode := sqlOperator.GetParameter(observerconst.EnableUpgradeMode)
	isFalse := true
	for _, v := range zoneUpGradeMode {
		if v.Value == "True" {
			isFalse = false
		}
	}
	if !isFalse {
		return errors.New("Upgrade Mode Wrong")
	}
	return nil
}

func (ctrl *OBClusterCtrl) EndUpgrade() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator When End Upgrade")
	}
	return sqlOperator.EndUpgrade()
}

func (ctrl *OBClusterCtrl) SetMinVersion() error {
	klog.Infoln("SetMinVersion: ")
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator When Setting Min OB Server Veriosn")
	}
	clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
	version := clusterStatus.TargetVersion
	return sqlOperator.SetParameter(observerconst.MinObserverVersion, version)
}

func (ctrl *OBClusterCtrl) ExecPostScripts(statefulApp cloudv1.StatefulApp) error {
	clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
	upgradeRoute := clusterStatus.UpgradeRoute
	if upgradeRoute[len(upgradeRoute)-1] == clusterStatus.ScriptPassedVersion {
		return ctrl.UpdateOBClusterAndZoneStatus(observerconst.NeedUpgradePostCheck, "", "")
	}
	containerImage := fmt.Sprint(ctrl.OBCluster.Spec.ImageRepo, ":", ctrl.OBCluster.Spec.Tag)
	rsIP, err := ctrl.GetRsIPFromDB(statefulApp)
	if err != nil {
		return err
	}
	var version string
	var index int
	if clusterStatus.ScriptPassedVersion == "" {
		version = upgradeRoute[1]
		index = 1
	} else {
		for i, ver := range upgradeRoute {
			if ver == clusterStatus.ScriptPassedVersion {
				version = upgradeRoute[i+1]
				index = i + 1
			}
		}
	}
	jobName := GenerateJobName(myconfig.ClusterName, fmt.Sprint("exec-post-scripts-", index))
	jobExecuter := resource.NewJobResource(ctrl.Resource)
	jobObject, err := jobExecuter.Get(context.TODO(), ctrl.OBCluster.Namespace, jobName)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			filename := fmt.Sprint(observerconst.UpgradeScriptsPath, version, observerconst.PostScriptFile)
			cmd := sql.ReplaceAll(ExecCheckScriptsCMDTemplate, UpgradeReplacer(filename, rsIP, strconv.Itoa(observerconst.MysqlPort)))
			var cmdList []string
			cmdList = append(cmdList, "bash", "-c", cmd)
			jobObject = ctrl.GenerateJobObject(jobName, containerImage, cmdList)
			err = jobExecuter.Create(context.TODO(), jobObject)
			if err != nil {
				klog.Errorln("Create ", jobName, " job failed, err: ", err)
				return err
			}
			return nil
		} else {
			klog.Errorln("Get ", jobName, " job failed, err: ", err)
			return err
		}
	}
	job := jobObject.(batchv1.Job)
	if job.Status.Succeeded == 0 && job.Status.Failed == 0 {
		return nil
	}
	if job.Status.Succeeded == 1 {
		upgradeInfo := UpgradeInfo{
			ScriptPassedVersion: version,
		}
		err = ctrl.UpdateOBStatusForUpgrade(upgradeInfo)
		if err != nil {
			return err
		}
	}
	return jobExecuter.Delete(context.TODO(), jobObject)
}

func (ctrl *OBClusterCtrl) ExecPreScripts(statefulApp cloudv1.StatefulApp) error {
	clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
	upgradeRoute := clusterStatus.UpgradeRoute
	if upgradeRoute[len(upgradeRoute)-1] == clusterStatus.ScriptPassedVersion {
		upgradeInfo := UpgradeInfo{
			ScriptPassedVersion: upgradeRoute[0],
			ClusterStatus:       observerconst.NeedUpgrading,
		}

		return ctrl.UpdateOBStatusForUpgrade(upgradeInfo)
	}

	containerImage := fmt.Sprint(ctrl.OBCluster.Spec.ImageRepo, ":", ctrl.OBCluster.Spec.Tag)
	rsIP, err := ctrl.GetRsIPFromDB(statefulApp)
	if err != nil {
		return err
	}

	var version string
	var index int
	if clusterStatus.ScriptPassedVersion == "" {
		version = upgradeRoute[1]
		index = 1
	} else {
		for i, ver := range upgradeRoute {
			if ver == clusterStatus.ScriptPassedVersion {
				version = upgradeRoute[i+1]
				index = i + 1
			}
		}
	}
	jobName := GenerateJobName(myconfig.ClusterName, fmt.Sprint("exec-pre-scripts-", index))
	jobExecuter := resource.NewJobResource(ctrl.Resource)
	jobObject, err := jobExecuter.Get(context.TODO(), ctrl.OBCluster.Namespace, jobName)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			filename := fmt.Sprint(observerconst.UpgradeScriptsPath, version, observerconst.PreScriptFile)
			cmd := sql.ReplaceAll(ExecCheckScriptsCMDTemplate, UpgradeReplacer(filename, rsIP, strconv.Itoa(observerconst.MysqlPort)))
			var cmdList []string
			cmdList = append(cmdList, "bash", "-c", cmd)
			jobObject = ctrl.GenerateJobObject(jobName, containerImage, cmdList)
			err = jobExecuter.Create(context.TODO(), jobObject)
			if err != nil {
				klog.Errorln("Create ", jobName, " job failed, err: ", err)
				return err
			}
			return nil
		} else {
			klog.Errorln("Get ", jobName, " job failed, err: ", err)
			return err
		}
	}
	job := jobObject.(batchv1.Job)
	if job.Status.Succeeded == 0 && job.Status.Failed == 0 {
		return nil
	}
	if job.Status.Succeeded == 1 {
		upgradeInfo := UpgradeInfo{
			ScriptPassedVersion: version,
		}
		err = ctrl.UpdateOBStatusForUpgrade(upgradeInfo)
		if err != nil {
			return err
		}
	}
	return jobExecuter.Delete(context.TODO(), jobObject)
}

func (ctrl *OBClusterCtrl) PreparingForUpgrade(statefulApp cloudv1.StatefulApp) error {
	upgradeInfo := UpgradeInfo{
		ZoneStatus:    observerconst.NeedUpgrading,
		ClusterStatus: observerconst.Upgrading,
	}
	return ctrl.UpdateOBStatusForUpgrade(upgradeInfo)
}

func (ctrl *OBClusterCtrl) ExecUpgrading(statefulApp cloudv1.StatefulApp) error {
	zoneInfoMap := ctrl.GetInfoForUpgradeByZone()
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return err
	}
	obZoneList := sqlOperator.GetOBZone()
	if zoneInfoMap[observerconst.NeedUpgrading] != nil {
		zoneName := zoneInfoMap[observerconst.NeedUpgrading][0]
		var ip string
		if len(obZoneList) > 2 {
			ip, err = ctrl.GetRsIP(statefulApp, zoneName)
			if err != nil {
				return err
			}
		} else {
			ip, err = ctrl.GetServiceClusterIPByName(ctrl.OBCluster.Namespace, ctrl.OBCluster.Name)
			if err != nil {
				return err
			}
		}
		if len(obZoneList) > 2 {
			isZoneStop, err := ctrl.isOBZoneStop(ip, zoneName)
			if err != nil {
				klog.Errorln("Check OB Zone Status err : ", zoneName, err)
				return err
			}
			if !isZoneStop {
				err = ctrl.StopZone(ip, zoneName)
				if err != nil {
					klog.Errorln("Stop Zone err : ", zoneName, err)
					return err
				}
			}

			err = ctrl.WaitLeaderCountZero(ip, zoneName)
			if err != nil {
				klog.Errorln("Check Zone Leader Count Zero err : ", zoneName, err)
				return err
			}
		}
		err = ctrl.PatchPods(ip, zoneName, statefulApp)
		if err != nil {
			klog.Errorln("Patch Pods err : ", zoneName, err)
			return err
		}

		_, err = ctrl.isOBSeverActive(ip, zoneName, len(obZoneList) > 2)
		if err != nil {
			klog.Errorln("Check OB Sever Status err : ", zoneName, err)
			return err
		}
		if len(obZoneList) > 2 {
			err = ctrl.StartOBZone(ip, zoneName)
			if err != nil {
				klog.Errorln("Start OB Zone err : ", zoneName, err)
				return err
			}
		}
		err = ctrl.waitAllOBSeverAvailable(ip, len(obZoneList) > 2)
		if err != nil {
			klog.Errorln("Check Whether All OB Severs Are Available err : ", err)
			return err
		}
		singleZoneStatus := make(map[string]string)
		singleZoneStatus[zoneName] = observerconst.UpgradingPassed
		upgradeInfo := UpgradeInfo{
			SingleZoneStatus: singleZoneStatus,
		}
		err = ctrl.UpdateOBStatusForUpgrade(upgradeInfo)
		time.Sleep(2 * time.Second)
		return nil
	}

	err = ctrl.UpgradeSchema()
	if err != nil {
		return err
	}
	return ctrl.UpdateOBClusterAndZoneStatus(observerconst.ExecutingPostScripts, "", "")
}

func (ctrl *OBClusterCtrl) WaitLeaderCountZero(rsIP, zoneName string) error {
	klog.Infoln("Wait Leader Count Clear")
	err := ctrl.TickerLeaderCountFromDB(rsIP, zoneName)
	if err != nil {
		return err
	}
	klog.Infoln("Leader Count Is Zero")
	return nil
}

func (ctrl *OBClusterCtrl) TickerLeaderCountFromDB(rsIP, zoneName string) error {
	tick := time.Tick(observerconst.TickPeriodForOBServerStatusCheck)
	var num int
	for {
		select {
		case <-tick:
			if num > observerconst.TickNumForOBServerStatusCheck {
				return errors.New("Wait For Leader Count Clear Timeout")
			}
			num = num + 1
			res, err := ctrl.isLeaderCountZero(rsIP, zoneName)
			if res {
				return err
			}
		}
	}
}

func (ctrl *OBClusterCtrl) waitAllOBSeverAvailable(rsIP string, rolling_upgrade bool) error {
	klog.Infoln("Wait All OB Server Available")
	err := ctrl.TickerOBServerAvailableFromDB(rsIP, rolling_upgrade)
	if err != nil {
		return err
	}
	klog.Infoln("All OB Server Available")
	return nil
}

func (ctrl *OBClusterCtrl) TickerOBServerAvailableFromDB(rsIP string, rolling_upgrade bool) error {
	tick := time.Tick(observerconst.TickPeriodForOBServerStatusCheck)
	var num int
	for {
		select {
		case <-tick:
			if num > observerconst.TickNumForOBServerStatusCheck {
				return errors.New("Wait For OB Server Available Timeout")
			}
			num = num + 1
			res, err := ctrl.isAllOBSeverAvailable(rsIP, rolling_upgrade)
			if res {
				return err
			}
		}
	}
}

func (ctrl *OBClusterCtrl) WaitAllContainerRunning(pod corev1.Pod) error {
	klog.Infoln("Wait All OB Server Container Running")
	err := ctrl.TickerOBServerContainerRunning(pod)
	if err != nil {
		return err
	}
	klog.Infoln("All OB Server Container Running")
	return nil
}

func (ctrl *OBClusterCtrl) TickerOBServerContainerRunning(pod corev1.Pod) error {
	tick := time.Tick(observerconst.TickPeriodForOBServerStatusCheck)
	var num int
	for {
		select {
		case <-tick:
			if num > observerconst.TickNumForOBServerStatusCheck {
				return errors.New("Wait For OB Server Container Running Timeout")
			}
			num = num + 1
			res, err := ctrl.isAllOBSeverContainerRunning(pod)
			if res {
				return err
			}
		}
	}
}

func (ctrl *OBClusterCtrl) isAllOBSeverContainerRunning(pod corev1.Pod) (bool, error) {
	for _, container := range pod.Status.ContainerStatuses {
		if container.Name == observerconst.ImgOb {
			klog.Infoln("container.State.Running == nil ", container.State.Running == nil)
			if container.State.Running == nil {
				klog.Errorln(pod.Status.PodIP, " Observer Container Not Running")
				return false, errors.New("Container Not Running")
			}
		}
	}
	return true, nil
}

func (ctrl *OBClusterCtrl) PatchPods(rsIP, zoneName string, statefulApp cloudv1.StatefulApp) error {
	// patch all pods in a zone
	subsets := statefulApp.Status.Subsets
	podExecuter := resource.NewPodResource(ctrl.Resource)
	var startSubset []cloudv1.SubsetStatus
	for _, subset := range subsets {
		podList := subset.Pods
		if subset.Name == zoneName {
			startSubset = append(startSubset, subset)
			for _, pod := range podList {
				podName := pod.Name
				podObject, err := podExecuter.Get(context.TODO(), ctrl.OBCluster.Namespace, podName)
				if err != nil {
					klog.Errorln("Get PodObject By PodName failed, err: ", err)
					return err
				}
				podObjectReal := podObject.(corev1.Pod)
				newPodObject := podObjectReal.DeepCopy()
				for idx, container := range newPodObject.Spec.Containers {
					if container.Name == observerconst.ImgOb {
						newPodObject.Spec.Containers[idx].Image = fmt.Sprint(ctrl.OBCluster.Spec.ImageRepo, ":", ctrl.OBCluster.Spec.Tag)
						err = podExecuter.Patch(context.TODO(), *newPodObject, client.MergeFrom(podObjectReal.DeepCopyObject().(client.Object)))
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
	// wait observer container running
	for _, subset := range subsets {
		podList := subset.Pods
		if subset.Name == zoneName {
			for _, pod := range podList {
				podName := pod.Name
				podObject, err := podExecuter.Get(context.TODO(), ctrl.OBCluster.Namespace, podName)
				if err != nil {
					klog.Errorln("Get PodObject By PodName failed, err: ", err)
					return err
				}
				podObjectReal := podObject.(corev1.Pod)
				err = ctrl.WaitAllContainerRunning(podObjectReal)
				if err != nil {
					return err
				}
			}
		}
	}
	// wait observer available
	clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
	for _, subset := range subsets {
		podList := subset.Pods
		if subset.Name == zoneName {
			for _, pod := range podList {
				currentVersion, err := cable.OBServerGetVersion(pod.PodIP)
				if err != nil {
					klog.Errorln(pod.PodIP, "OB Server Get Version Failed, Error: ", err)
					return err
				}
				klog.Infoln("currentVersion: ", currentVersion)
				if currentVersion != clusterStatus.TargetVersion {
					return errors.New(fmt.Sprint("Ob Server version Is Not Target Version : ", zoneName))
				}
			}
		}
	}
	klog.Infoln("get versiob finish")

	rsName := converter.GenerateRootServiceName(ctrl.OBCluster.Name)
	rsCtrl := NewRootServiceCtrl(ctrl)
	rsCurrent, err := rsCtrl.GetRootServiceByName(ctrl.OBCluster.Namespace, rsName)
	if err != nil {
		return err
	}
	rsList := cable.GenerateRSListFromRootServiceStatus(rsCurrent.Status.Topology)
	cable.OBServerStart(ctrl.OBCluster, startSubset, rsList)
	for _, subset := range subsets {
		podList := subset.Pods
		if subset.Name == zoneName {
			for _, pod := range podList {
				err = ctrl.WaitOBServerActive(rsIP, zoneName, pod.PodIP, statefulApp)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (ctrl *OBClusterCtrl) isAllOBSeverAvailable(rsIP string, rolling_upgrade bool) (bool, error) {
	var sqlOperator *sql.SqlOperator
	var err error
	if rolling_upgrade {
		sqlOperator, err = ctrl.GetSqlOperator(rsIP)
	} else {
		sqlOperator, err = ctrl.GetSqlOperator()

	}
	if err != nil {
		return false, errors.Wrap(err, "get sql operator when recover server")
	}
	clogStatList := sqlOperator.GetClogStat()
	if len(clogStatList) == 0 {
		return true, nil
	} else {
		return false, errors.New("Not All Server Available")
	}
}

func (ctrl *OBClusterCtrl) isOBSeverActive(rsIP, zoneName string, rolling_upgrade bool) (bool, error) {
	var sqlOperator *sql.SqlOperator
	var err error
	if rolling_upgrade {
		sqlOperator, err = ctrl.GetSqlOperator(rsIP)
	} else {
		sqlOperator, err = ctrl.GetSqlOperator()
	}
	if err != nil {
		return false, errors.New("Get Sql Operator When Check OBSever Status")
	}
	obServerList := sqlOperator.GetOBServer()
	if len(obServerList) == 0 {
		return false, errors.New(observerconst.DataBaseError)
	}
	for _, obServer := range obServerList {
		if obServer.Zone == zoneName && obServer.Status != observerconst.OBServerActive {
			return false, errors.New(fmt.Sprint("OBServers Are Not Active In Zone: ", zoneName))
		}
	}
	return true, nil
}

func (ctrl *OBClusterCtrl) UpgradeSchema() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when upgrade schema")
	}
	return sqlOperator.UpgradeSchema()
}

func (ctrl *OBClusterCtrl) isLeaderCountZero(rsIP, zoneName string) (bool, error) {
	sqlOperator, err := ctrl.GetSqlOperator(rsIP)
	if err != nil {
		return false, errors.Wrap(err, "get sql operator when check info leader count")
	}
	zoneLeaderCount := sqlOperator.GetLeaderCount()
	for _, zone := range zoneLeaderCount {
		if zone.Zone == zoneName {
			if zone.LeaderCount == 0 {
				return true, nil
			} else {
				return false, errors.New(fmt.Sprint("Leader Count Is Not Zero: ", zoneName))
			}
		}
	}
	return false, errors.New(fmt.Sprint("Can Not Get Zone Leader Count : ", zoneName))
}

func (ctrl *OBClusterCtrl) isOBZoneStop(rsIP, zoneName string) (bool, error) {
	sqlOperator, err := ctrl.GetSqlOperator(rsIP)
	if err != nil {
		return false, errors.Wrap(err, "Get Sql Operator When Check OB Zone Stop")
	}
	obZoneList := sqlOperator.GetOBZone()
	if len(obZoneList) == 0 {
		return false, errors.New(observerconst.DataBaseError)
	}

	for _, zone := range obZoneList {
		if zone.Zone == zoneName {
			if zone.Info == observerconst.OBZoneInactive {
				return true, nil
			}
			return false, nil
		}
	}
	return false, errors.New("Can Not Get Zone ")
}

func (ctrl *OBClusterCtrl) GetInfoForUpgradeByZone() map[string][]string {
	infoMap := make(map[string][]string)
	clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
	for _, zone := range clusterStatus.Zone {
		if zone.ZoneStatus == observerconst.NeedUpgrading {
			infoMap[observerconst.NeedUpgrading] = append(infoMap[observerconst.NeedUpgrading], zone.Name)
		} else if zone.ZoneStatus == observerconst.Upgrading {
			infoMap[observerconst.Upgrading] = append(infoMap[observerconst.Upgrading], zone.Name)
		}

	}
	return infoMap
}

func (ctrl *OBClusterCtrl) CheckUpgradeMode(statefulApp cloudv1.StatefulApp) error {
	sqlOperator, err := ctrl.GetSqlOperatorFromStatefulApp(statefulApp)
	if err != nil {
		klog.Errorln("Get Sql Operator From StatefulApp Failed, Err: ", err)
		return err
	}
	isOK := true
	zoneUpGradeMode := sqlOperator.GetParameter(observerconst.EnableUpgradeMode)
	for _, v := range zoneUpGradeMode {
		if v.Value == "False" {
			isOK = false
		}
	}
	if isOK {
		return ctrl.UpdateOBClusterAndZoneStatus(observerconst.ExecutingPreScripts, "", "")
	} else {
		return sqlOperator.BeginUpgrade()
	}
}

func (ctrl *OBClusterCtrl) GetRsIPFromDB(statefulApp cloudv1.StatefulApp) (string, error) {
	var rsIP string
	sqlOperator, err := ctrl.GetSqlOperatorFromStatefulApp(statefulApp)
	if err != nil {
		klog.Errorln("Get Sql Operator From StatefulApp Failed, Err: ", err)
		return rsIP, err
	}
	rsList := sqlOperator.GetRootService()
	for _, zone := range rsList {
		if zone.Role == 1 {
			rsIP = zone.SvrIP
			return rsIP, nil
		}
	}
	return rsIP, errors.New("Get RS IP Failed. Cannot Find RS")
}

func (ctrl *OBClusterCtrl) GetRsIP(statefulApp cloudv1.StatefulApp, zoneName string) (string, error) {
	rsName := converter.GenerateRootServiceName(ctrl.OBCluster.Name)
	rsCtrl := NewRootServiceCtrl(ctrl)
	rsCurrent, err := rsCtrl.GetRootServiceByName(ctrl.OBCluster.Namespace, rsName)
	if err != nil {
		return "", err
	}
	for _, cluster := range rsCurrent.Status.Topology {
		if cluster.Cluster == myconfig.ClusterName {
			for _, zone := range cluster.Zone {
				if zone.ServerIP != "" && zone.Name != zoneName {
					return zone.ServerIP, nil
				}
			}
		}
	}
	return "", errors.New("Get RS IP Failed. Cannot Find RS")
}
