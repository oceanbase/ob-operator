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
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	myconfig "github.com/oceanbase/ob-operator/pkg/config"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/cable"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"
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

func (ctrl *OBClusterCtrl) OBClusterUpgrade(statefulApp cloudv1.StatefulApp) error {
	clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
	upgradeRoute := clusterStatus.UpgradeRoute
	var err error
	err = ctrl.CheckAndSetTargetVersion(clusterStatus.TargetVersion)
	if err != nil {
		return err
	}
	cluster := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
	targetVer := cluster.TargetVersion
	err = ctrl.CheckAndSetUpgradeRoute(statefulApp, upgradeRoute, targetVer)
	if err != nil {
		return err
	}
	return ctrl.UpdateOBClusterAndZoneStatus(observerconst.NeedUpgradeCheck, "", "")
}

func UpgradeReplacer(filename, clusterIP, port string) *strings.Replacer {
	return strings.NewReplacer("${FILE_NAME}", filename, "${IP}", clusterIP, "${PORT}", port)
}

func (ctrl *OBClusterCtrl) GetTargetVersion() (string, error) {
	podIp, err := ctrl.GetHelperPodIP()
	if err != nil {
		return "", err
	}
	return cable.OBServerGetVersion(podIp)
}

func (ctrl *OBClusterCtrl) GetUpgradeRoute(currentVer, targetVer string) ([]string, error) {
	var upgradeRoute []string
	podIp, err := ctrl.GetHelperPodIP()
	if err != nil {
		return upgradeRoute, err
	}
	return cable.OBServerGetUpgradeRoute(podIp, currentVer, targetVer)
}

func (ctrl *OBClusterCtrl) GetCurrentVersion(statefulApp cloudv1.StatefulApp) (string, error) {
	subsets := statefulApp.Status.Subsets
	for subsetsIdx := range subsets {
		for _, pod := range subsets[subsetsIdx].Pods {
			return cable.OBServerGetVersion(pod.PodIP)
		}
	}
	return "", nil
}

func (ctrl *OBClusterCtrl) CheckAndSetTargetVersion(currentTargetVersion string) error {
	targetVersion, err := ctrl.GetTargetVersion()
	if err != nil {
		return err
	}
	if currentTargetVersion == "" {
		klog.Infoln("OBCluster Upgrade Target Verson is ", targetVersion)
		upgradeInfo := UpgradeInfo{
			TargetVersion: targetVersion,
		}
		return ctrl.UpdateOBStatusForUpgrade(upgradeInfo)
	} else if currentTargetVersion != targetVersion {
		klog.Errorln("Can not upgrade OB to another version when current upgrading is not finished")
		return errors.New("Can not upgrade OB to another version when current upgrading is not finished")
	}
	return nil
}

func (ctrl *OBClusterCtrl) CheckAndSetUpgradeRoute(statefulApp cloudv1.StatefulApp, currUpgradeRoute []string, targetVer string) error {
	currentVer, err := ctrl.GetCurrentVersion(statefulApp)
	if err != nil {
		return err
	}
	upgradeRoute, err := ctrl.GetUpgradeRoute(currentVer, targetVer)
	if err != nil {
		return err
	}
	if len(currUpgradeRoute) == 0 {
		upgradeInfo := UpgradeInfo{
			UpgradeRoute: upgradeRoute,
		}
		return ctrl.UpdateOBStatusForUpgrade(upgradeInfo)
	}
	klog.Infoln("CheckAndSetUpgradeRoute: ctrl.OBCluster.Status.Topology[0].UpgradeRoute", ctrl.OBCluster.Status.Topology[0].UpgradeRoute)
	if !reflect.DeepEqual(upgradeRoute, currUpgradeRoute) {
		klog.Errorf("Upgrade Route Does Not Match. Current: %s, Target: %s", currUpgradeRoute, upgradeRoute)
		return errors.New("Upgrade Route Does Not Match")
	}
	return nil
}

func (ctrl *OBClusterCtrl) ExecUpgradePreChecker(statefulApp cloudv1.StatefulApp) error {
	err := ctrl.CreatePreCheckerJob(statefulApp)
	if err != nil {
		return err
	}
	return ctrl.UpdateOBClusterAndZoneStatus(observerconst.UpgradeChecking, "", "")
}

func (ctrl *OBClusterCtrl) GetPreCheckJobStatus(statefulApp cloudv1.StatefulApp) error {
	// Get Job
	name := observerconst.UpgradePreChecker
	jobName := GenerateJobName(ctrl.OBCluster.Name, myconfig.ClusterName, name)
	jobObject, err := ctrl.GetJobObject(jobName)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return ctrl.CreatePreCheckerJob(statefulApp)
		} else {
			klog.Errorln("Get ", jobName, " job failed, err: ", err)
			return err
		}
	}
	// Get Job Status
	jobStatus := ctrl.GetJobStatus(jobObject)
	switch jobStatus {
	case observerconst.JobRunning:
		return nil
	case observerconst.JobSucceeded:
		err = ctrl.UpdateOBClusterAndZoneStatus(observerconst.NeedExecutingPreScripts, "", "")
	case observerconst.JobFailed:
		err = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, "", "")
	}
	if err != nil {
		return err
	}
	// Delete Job
	err = ctrl.DeleteJobObject(jobObject)
	if err != nil {
		klog.Errorln("Delete Job %s Failed, Err: %s", jobName, err)
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) ExecPreScripts(statefulApp cloudv1.StatefulApp) error {
	// All Scripts Finish
	klog.Infoln("ExecPreScripts: ", ctrl.OBCluster.Status.Topology[0].UpgradeRoute)
	finish, err := ctrl.AllScriptsFinish()
	if err != nil {
		return err
	}
	if finish {
		clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
		upgradeRoute := clusterStatus.UpgradeRoute
		upgradeInfo := UpgradeInfo{
			ScriptPassedVersion: upgradeRoute[0],
			ClusterStatus:       observerconst.NeedUpgrading,
		}
		return ctrl.UpdateOBStatusForUpgrade(upgradeInfo)
	}
	// Get Next Version Job
	version, index, err := ctrl.GetNextVersion()
	if err != nil {
		klog.Errorln("ctrl.GetNextVersion(): ", err)
		return ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, "", "")
	}
	jobName := GenerateJobName(ctrl.OBCluster.Name, myconfig.ClusterName, fmt.Sprint(observerconst.UpgradePre, "-", index))
	jobObject, err := ctrl.GetJobObject(jobName)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return ctrl.CreateUpgradePreJob(statefulApp, version, index)
		} else {
			klog.Errorln("Get Job %s Failed, Err: %s", jobName, err)
			return err
		}
	}
	// Get Job status
	jobStatus := ctrl.GetJobStatus(jobObject)
	switch jobStatus {
	case observerconst.JobRunning:
		return nil
	case observerconst.JobSucceeded:
		upgradeInfo := UpgradeInfo{
			ScriptPassedVersion: version,
		}
		err = ctrl.UpdateOBStatusForUpgrade(upgradeInfo)
		if err != nil {
			return err
		}
	}
	return ctrl.DeleteJobObject(jobObject)
}

func (ctrl *OBClusterCtrl) ExecPostScripts(statefulApp cloudv1.StatefulApp) error {
	// All Scripts Finish
	finish, err := ctrl.AllScriptsFinish()
	if err != nil {
		return err
	}
	if finish {
		return ctrl.UpdateOBClusterAndZoneStatus(observerconst.NeedUpgradePostCheck, "", "")
	}
	// Get Next Version Job
	version, index, err := ctrl.GetNextVersion()
	if err != nil {
		return nil
	}
	jobName := GenerateJobName(ctrl.OBCluster.Name, myconfig.ClusterName, fmt.Sprint(observerconst.UpgradePost, "-", index))
	jobObject, err := ctrl.GetJobObject(jobName)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return ctrl.CreateUpgradePostJob(statefulApp, version, index)
		} else {
			klog.Errorln("Get Job %s Failed, Err: %s", jobName, err)
			return err
		}
	}
	// Get Job status
	jobStatus := ctrl.GetJobStatus(jobObject)
	switch jobStatus {
	case observerconst.JobRunning:
		return nil
	case observerconst.JobSucceeded:
		upgradeInfo := UpgradeInfo{
			ScriptPassedVersion: version,
		}
		err = ctrl.UpdateOBStatusForUpgrade(upgradeInfo)
		if err != nil {
			return err
		}
	}
	return ctrl.DeleteJobObject(jobObject)
}

func (ctrl *OBClusterCtrl) ExecUpgradePostChecker(statefulApp cloudv1.StatefulApp) error {
	// Get Job
	name := observerconst.UpgradePostChecker
	jobName := GenerateJobName(ctrl.OBCluster.Name, myconfig.ClusterName, name)
	jobObject, err := ctrl.GetJobObject(jobName)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return ctrl.CreatePostCheckerJob(statefulApp)
		} else {
			klog.Errorln("Get Job %s Failed, Err: %s", jobName, err)
			return err
		}
	}
	// Get Job Status
	jobStatus := ctrl.GetJobStatus(jobObject)
	switch jobStatus {
	case observerconst.JobRunning:
		return nil
	case observerconst.JobSucceeded:
		err = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, "", "")
		if err != nil {
			return err
		}
	case observerconst.JobFailed:
		return ctrl.DeleteJobObject(jobObject)
	}
	// Delete Job
	err = ctrl.DeleteJobObject(jobObject)
	if err != nil {
		klog.Errorln("Delete Job %s Failed, Err: %s", jobName, err)
		return err
	}

	err = ctrl.UpdateStatefulAppImage(statefulApp)
	if err != nil {
		klog.Errorln("Update StatefulApp Failed, Err: ", err)
		return err
	}
	return ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, "", "")
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
	err = ctrl.CheckUpgradeModeEnd()
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
		rollingUpgrade := len(obZoneList) > 2
		if rollingUpgrade {
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
		if rollingUpgrade {
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
		err = ctrl.PatchAndStartContainer(ip, zoneName, statefulApp)
		if err != nil {
			klog.Errorln("Patch Pods err : ", zoneName, err)
			return err
		}
		_, err = ctrl.isOBSeverActive(ip, zoneName)
		if err != nil {
			klog.Errorln("Check OB Sever Status err : ", zoneName, err)
			return err
		}
		if rollingUpgrade {
			err = ctrl.StartOBZone(ip, zoneName)
			if err != nil {
				klog.Errorln("Start OB Zone err : ", zoneName, err)
				return err
			}
		}
		err = ctrl.WaitAllOBSeverAvailable(ip)
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

func (ctrl *OBClusterCtrl) PatchAndStartContainer(rsIP, zoneName string, statefulApp cloudv1.StatefulApp) error {
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
				if currentVersion != clusterStatus.TargetVersion {
					return errors.New(fmt.Sprint("OB Server version Is Not Target Version : ", zoneName))
				}
			}
		}
	}
	klog.Infoln("Check OB Servers Version Finish")
	rsName := converter.GenerateRootServiceName(ctrl.OBCluster.Name)
	rsCtrl := NewRootServiceCtrl(ctrl)
	rsCurrent, err := rsCtrl.GetRootServiceByName(ctrl.OBCluster.Namespace, rsName)
	if err != nil {
		return err
	}

	// Recovery Etc From Additional dir
	for _, subset := range subsets {
		podList := subset.Pods
		if subset.Name == zoneName {
			for _, pod := range podList {
				err = cable.OBRecoverConfig(pod.PodIP)
				if err != nil {
					klog.Errorln("Recover OBServer Config Failed, Err: ", err)
					return err
				}
			}
		}
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

func (ctrl *OBClusterCtrl) CheckUpgradeModeBegin(statefulApp cloudv1.StatefulApp) error {
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

func (ctrl *OBClusterCtrl) UpdateStatefulAppImage(statefulApp cloudv1.StatefulApp) error {
	image := fmt.Sprint(ctrl.OBCluster.Spec.ImageRepo, ":", ctrl.OBCluster.Spec.Tag)
	newStatefulApp := converter.UpdateStatefulAppImage(statefulApp, image)
	statefulAppCtrl := NewStatefulAppCtrl(ctrl, newStatefulApp)
	return statefulAppCtrl.UpdateStatefulApp()
}
