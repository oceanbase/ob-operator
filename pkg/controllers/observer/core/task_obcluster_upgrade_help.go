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

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	myconfig "github.com/oceanbase/ob-operator/pkg/config"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/cable"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func GeneratePodName(obclusterName, clusterName, name, version string) string {
	return fmt.Sprintf("%s-%s-%s-%s", obclusterName, clusterName, version, name)
}

func GenerateJobName(obclusterName, clusterName, name, version string) string {
	return fmt.Sprintf("%s-%s-%s-%s", obclusterName, clusterName, version, name)
}

func (ctrl *OBClusterCtrl) GenerateSpecVersion() string {
	specVersion := strings.Split(ctrl.OBCluster.Spec.Tag, "-")
	split := strings.Split(specVersion[0], ".")
	var res string
	for _, x := range split {
		res += x
	}
	return res
}

func (ctrl *OBClusterCtrl) GeneratePodObject(podName string, containerList []corev1.Container) interface{} {
	podObject := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ctrl.OBCluster.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers:    containerList,
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}
	return podObject
}

func (ctrl *OBClusterCtrl) CreatePod(podObject interface{}) error {
	podExecuter := resource.NewPodResource(ctrl.Resource)
	err := podExecuter.Create(context.TODO(), podObject)
	if err != nil {
		if kubeerrors.IsAlreadyExists(err) {
			return nil
		}
		klog.Errorf("Create Pod Failed %s: %s ", podObject, err)
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) DeletePod(podName string) error {
	podExecuter := resource.NewPodResource(ctrl.Resource)
	podObject, err := podExecuter.Get(context.TODO(), ctrl.OBCluster.Namespace, podName)
	if err != nil {
		klog.Errorln("Get PodObject By PodName failed, err: ", err)
		return err
	}
	return podExecuter.Delete(context.TODO(), podObject)
}

func (ctrl *OBClusterCtrl) CreateHelperPod(podName string) error {
	klog.Infoln("Create Helper Pod")
	var envList []corev1.EnvVar
	envList = append(envList, corev1.EnvVar{
		Name:  "LD_LIBRARY_PATH",
		Value: "/home/admin/oceanbase/lib",
	})
	containerImage := fmt.Sprint(ctrl.OBCluster.Spec.ImageRepo, ":", ctrl.OBCluster.Spec.Tag)
	containerList := []corev1.Container{
		{
			Name:  podName,
			Image: containerImage,
			Env:   envList,
		},
	}
	podObject := ctrl.GeneratePodObject(podName, containerList)
	return ctrl.CreatePod(podObject)
}

func (ctrl *OBClusterCtrl) GetHelperPodIP() (string, error) {
	podName := GeneratePodName(ctrl.OBCluster.Name, myconfig.ClusterName, "help", ctrl.GenerateSpecVersion())
	podExecuter := resource.NewPodResource(ctrl.Resource)
	podObject, err := podExecuter.Get(context.TODO(), ctrl.OBCluster.Namespace, podName)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			klog.Errorf("Cannot Find Helper Pod. Creating Helper Pod '%s' Now", podName)
			err = ctrl.CreateHelperPod(podName)
			if err != nil {
				return "", err
			}
			return "", err
		}
		klog.Errorln("Get PodIp By PodName failed, err: ", err)
		return "", err
	}
	pod := podObject.(corev1.Pod)
	if pod.Status.Phase != observerconst.PodRunning {
		err = ctrl.WaitHelperPodReady(podName)
		if err != nil {
			return "", err
		}
	}
	return pod.Status.PodIP, nil
}

func (ctrl *OBClusterCtrl) DeleteHelperPod() error {
	klog.Infoln("Delete Helper Pod")
	podName := GeneratePodName(ctrl.OBCluster.Name, myconfig.ClusterName, "help", ctrl.GenerateSpecVersion())
	return ctrl.DeletePod(podName)
}

func (ctrl *OBClusterCtrl) GenerateJobObject(jobName string, containerList []corev1.Container) batchv1.Job {
	var backOffLimit int32
	jobObject := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: ctrl.OBCluster.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers:    containerList,
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit: &backOffLimit,
		},
	}
	return jobObject
}

func (ctrl *OBClusterCtrl) CreateJob(jobObject interface{}) error {
	jobExecuter := resource.NewJobResource(ctrl.Resource)
	err := jobExecuter.Create(context.TODO(), jobObject)
	if err != nil {
		klog.Errorln("Create Job Failed, Err: ", err)
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) GetJobObject(jobName string) (interface{}, error) {
	jobExecuter := resource.NewJobResource(ctrl.Resource)
	jobObject, err := jobExecuter.Get(context.TODO(), ctrl.OBCluster.Namespace, jobName)
	if err != nil {
		klog.Errorf("Get JobObject By JobName %s Failed, Err: ", jobName, err)
		return jobObject, err
	}
	return jobObject, nil
}

func (ctrl *OBClusterCtrl) GetJobStatus(jobObject interface{}) string {
	job := jobObject.(batchv1.Job)
	if job.Status.Succeeded == 0 && job.Status.Failed == 0 {
		return observerconst.JobRunning
	}
	if job.Status.Succeeded == 1 {
		return observerconst.JobSucceeded
	}
	if job.Status.Failed == 1 {
		return observerconst.JobFailed
	}
	return ""
}

func (ctrl *OBClusterCtrl) DeleteJobObject(jobObject interface{}) error {
	jobExecuter := resource.NewJobResource(ctrl.Resource)
	err := jobExecuter.Delete(context.TODO(), jobObject)
	if err != nil {
		klog.Errorln("Delete JobObject Failed, err: ", err)
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) CreateExecScriptJob(name, fileName string, statefulApp cloudv1.StatefulApp) error {
	jobName := GenerateJobName(ctrl.OBCluster.Name, myconfig.ClusterName, name, ctrl.GenerateSpecVersion())
	klog.Infoln("Create Job", jobName)
	containerImage := fmt.Sprint(ctrl.OBCluster.Spec.ImageRepo, ":", ctrl.OBCluster.Spec.Tag)
	rsIP, err := ctrl.GetRsIPFromDB(statefulApp)
	if err != nil {
		return err
	}
	var cmdList []string
	cmd := sql.ReplaceAll(ExecCheckScriptsCMDTemplate, UpgradeReplacer(fileName, rsIP, strconv.Itoa(observerconst.MysqlPort)))
	cmdList = append(cmdList, "bash", "-c", cmd)
	var envList []corev1.EnvVar
	envList = append(envList, corev1.EnvVar{
		Name:  "LD_LIBRARY_PATH",
		Value: "/home/admin/oceanbase/lib",
	})
	containerList := []corev1.Container{
		{
			Name:    jobName,
			Image:   containerImage,
			Command: cmdList,
			Env:     envList,
		},
	}
	jobObject := ctrl.GenerateJobObject(jobName, containerList)
	err = ctrl.CreateJob(jobObject)
	if err != nil {
		klog.Errorf("Create Job '%s' Failed, Err: %s", jobName, err)
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) CreatePreCheckerJob(statefulApp cloudv1.StatefulApp) error {
	name := observerconst.UpgradePreChecker
	fileName := observerconst.UpgradePreCheckerPath
	err := ctrl.CreateExecScriptJob(name, fileName, statefulApp)
	if err != nil {
		klog.Errorln("Create Pre Check Job Failed, Err: ", err)
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) CreateHealthCheckerJob(statefulApp cloudv1.StatefulApp, name string) error {
	fileName := observerconst.UpgradeHealthCheckerPath
	err := ctrl.CreateExecScriptJob(name, fileName, statefulApp)
	if err != nil {
		klog.Errorln("create upgrade health check job failed, err: ", err)
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) CreatePostCheckerJob(statefulApp cloudv1.StatefulApp) error {
	name := observerconst.UpgradePostChecker
	fileName := observerconst.UpgradePostCheckerPath
	err := ctrl.CreateExecScriptJob(name, fileName, statefulApp)
	if err != nil {
		klog.Errorln("Create Post Check Job Failed, Err: ", err)
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) CreateUpgradePreJob(statefulApp cloudv1.StatefulApp, version string, index int) error {
	name := fmt.Sprint(observerconst.UpgradePre, "-", index)
	var filename string
	if ctrl.IsUpgradeV3(statefulApp) {
		filename = fmt.Sprint(observerconst.UpgradeScriptsPath, observerconst.PreScriptFile)
	}
	if ctrl.IsUpgradeV4(statefulApp) {
		filename = fmt.Sprint(observerconst.UpgradeScriptsPathV4, observerconst.PreScriptFile)
	}
	err := ctrl.CreateExecScriptJob(name, filename, statefulApp)
	if err != nil {
		klog.Errorln("Create Upgrade Pre Job Failed, Err: ", err)
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) CreateUpgradePostJob(statefulApp cloudv1.StatefulApp, version string, index int) error {
	name := fmt.Sprint(observerconst.UpgradePost, "-", index)
	var filename string
	if ctrl.IsUpgradeV3(statefulApp) {
		filename = fmt.Sprint(observerconst.UpgradeScriptsPath, observerconst.PostScriptFile)
	}
	if ctrl.IsUpgradeV4(statefulApp) {
		filename = fmt.Sprint(observerconst.UpgradeScriptsPathV4, observerconst.PostScriptFile)
	}
	err := ctrl.CreateExecScriptJob(name, filename, statefulApp)
	if err != nil {
		klog.Errorln("Create Upgrade Pre Job Failed, Err: ", err)
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) AllScriptsFinish() (bool, error) {
	clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
	upgradeRoute := clusterStatus.UpgradeRoute
	if len(upgradeRoute) < 2 {
		return false, errors.New("OBCluster Upgrade Route is Wrong When Check All Scripts Finish ")
	}
	return upgradeRoute[len(upgradeRoute)-1] == clusterStatus.ScriptPassedVersion, nil
}

func (ctrl *OBClusterCtrl) GetNextVersion(statefulApp cloudv1.StatefulApp) (string, int, error) {
	var version string
	var index int
	clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
	upgradeRoute := clusterStatus.UpgradeRoute
	if len(upgradeRoute) < 2 {
		return "", 0, errors.New("OBCluster Upgrade Route is Wrong When Get Next Version")
	}
	if ctrl.IsUpgradeV4(statefulApp) {
		return upgradeRoute[len(upgradeRoute)-1], len(upgradeRoute) - 1, nil
	}
	if ctrl.IsUpgradeV3(statefulApp) {
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
		return version, index, nil
	}
	return "", 0, errors.New("not support version")
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

func (ctrl *OBClusterCtrl) WaitLeaderCountZero(rsIP, zoneName string) error {
	klog.Infoln("Wait Leader Count Clear")
	err := ctrl.TickerLeaderCountFromDB(rsIP, zoneName)
	if err != nil {
		return err
	}
	klog.Infoln("Leader Count Is Zero")
	return nil
}

func (ctrl *OBClusterCtrl) WaitAllOBSeverAvailable(rsIP string) error {
	klog.Infoln("Wait All OB Server Available")
	err := ctrl.TickerOBServerAvailableFromDB(rsIP)
	if err != nil {
		return err
	}
	klog.Infoln("All OB Server Available")
	return nil
}

func (ctrl *OBClusterCtrl) TickerOBServerAvailableFromDB(rsIP string) error {
	tick := time.Tick(observerconst.TickPeriodForOBServerStatusCheck)
	var num int
	for {
		select {
		case <-tick:
			if num > observerconst.TickNumForOBServerStatusCheck {
				return errors.New("Wait For OB Server Available Timeout")
			}
			num = num + 1
			res, err := ctrl.isAllOBSeverAvailable(rsIP)
			if res {
				return err
			}
		}
	}
}

func (ctrl *OBClusterCtrl) WaitAndGetVersion(podIP string) (string, error) {
	klog.Infof("Wait and Check Pod '%s' OB Version ", podIP)
	version, err := ctrl.TickerGetPodVersion(podIP)
	if err != nil {
		return "", err
	}
	klog.Infoln("Check OB Servers Version Finish")
	return version, nil
}

func (ctrl *OBClusterCtrl) TickerGetPodVersion(podIP string) (string, error) {
	tick := time.Tick(observerconst.TickPeriodForPodGetObVersion)
	var num int
	for {
		select {
		case <-tick:
			if num > observerconst.TickNumForPodGetObVersion {
				return "", errors.New("Get OB Version Timeout")
			}
			num = num + 1
			res, version, err := ctrl.GetOBVersion(podIP)
			if res {
				return version, err
			}
		}
	}
}

func (ctrl *OBClusterCtrl) GetOBVersion(podIP string) (bool, string, error) {
	currentVersion, err := cable.OBServerGetVersion(podIP)
	if err != nil {
		return false, "", err
	}
	return true, currentVersion, nil
}

func (ctrl *OBClusterCtrl) WaitHelperPodReady(podName string) error {
	klog.Infoln("Wait Helper Pod Running")
	err := ctrl.TickerHelperPodRunning(podName)
	if err != nil {
		return err
	}
	klog.Infoln("Helper Pod Running")
	return nil
}

func (ctrl *OBClusterCtrl) TickerHelperPodRunning(podName string) error {
	tick := time.Tick(observerconst.TickPeriodForPodStatusCheck)
	var num int
	for {
		select {
		case <-tick:
			if num > observerconst.TickNumForPodStatusCheck {
				return errors.New("Wait For Helper Pod Running Timeout")
			}
			num = num + 1
			res, err := ctrl.isHeplerPodRunning(podName)
			if res {
				return err
			}
		}
	}
}

func (ctrl *OBClusterCtrl) isHeplerPodRunning(podName string) (bool, error) {
	podExecuter := resource.NewPodResource(ctrl.Resource)
	podObject, err := podExecuter.Get(context.TODO(), ctrl.OBCluster.Namespace, podName)
	if err != nil {
		return false, err
	}
	pod := podObject.(corev1.Pod)
	if pod.Status.Phase == observerconst.PodRunning {
		return true, nil
	}
	return false, nil

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
			if container.State.Running == nil {
				klog.Errorln(pod.Status.PodIP, " Observer Container Not Running")
				return false, errors.New("Container Not Running")
			}
		}
	}
	return true, nil
}

func (ctrl *OBClusterCtrl) isOBSeverActive(rsIP, zoneName string) (bool, error) {
	sqlOperator, err := ctrl.GetSqlOperator(rsIP)
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

func (ctrl *OBClusterCtrl) isAllOBSeverAvailable(rsIP string) (bool, error) {
	sqlOperator, err := ctrl.GetSqlOperator(rsIP)
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

func (ctrl *OBClusterCtrl) GetConfigAdditionalDir(rsIP, svrIP string) (string, error) {
	sqlOperator, err := ctrl.GetSqlOperator(rsIP)
	if err != nil {
		return "", errors.Wrap(err, "Get Sql Operator When Get Config_Additional_Dir")
	}
	configAdditionalDir := sqlOperator.ShowParameter(observerconst.ConfigAdditionalDir)
	klog.Infoln("configAdditionalDir: ", configAdditionalDir)
	for _, configAdditionalDir := range configAdditionalDir {
		if configAdditionalDir.SvrIP == svrIP {
			return configAdditionalDir.Value, nil
		}
	}
	return "", errors.New(fmt.Sprintf("Cannot Find Server %s ConfigAdditionalDir", svrIP))
}

func (ctrl *OBClusterCtrl) SetMinVersion() error {
	klog.Infoln("Set OB Min Version")
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator When Setting Min OB Server Veriosn")
	}
	clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
	version := clusterStatus.TargetVersion
	return sqlOperator.SetParameter(observerconst.MinObserverVersion, version)
}

func (ctrl *OBClusterCtrl) EndUpgrade() error {
	klog.Infoln("End upgrade")
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator When End Upgrade")
	}
	return sqlOperator.EndUpgrade()
}

func (ctrl *OBClusterCtrl) IsUpgradeModeEnd() (bool, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return false, errors.Wrap(err, "Get Sql Operator When End Upgrade")
	}
	zoneUpGradeMode := sqlOperator.GetParameter(observerconst.EnableUpgradeMode)
	isFalse := true
	for _, v := range zoneUpGradeMode {
		if v.Value == "True" {
			isFalse = false
		}
	}
	if !isFalse {
		return false, errors.New("Upgrade Mode Wrong")
	}
	return true, nil
}

func (ctrl *OBClusterCtrl) TickerUpgradeModeEndFromDB() error {
	tick := time.Tick(observerconst.TickPeriodForOBServerStatusCheck)
	var num int
	for {
		select {
		case <-tick:
			if num > observerconst.TickNumForOBServerStatusCheck {
				return errors.New("Wait For Leader Count Clear Timeout")
			}
			num = num + 1
			res, err := ctrl.IsUpgradeModeEnd()
			if res {
				return err
			}
		}
	}
}

func (ctrl *OBClusterCtrl) CheckAndWaitUpgradeModeEnd() error {
	klog.Infoln("Wait Upgrade Mode End")
	err := ctrl.TickerUpgradeModeEndFromDB()
	if err != nil {
		return err
	}
	klog.Infoln("Check Upgrade Mode OK")
	return nil
}

func (ctrl *OBClusterCtrl) RunRootInspection() error {
	klog.Infoln("Run job 'root_inspection'")
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator When Running Root Inspection Job")
	}
	return sqlOperator.RunRootInspection()
}

func (ctrl *OBClusterCtrl) UpgradeSchema() error {
	klog.Infoln("Upgrade virtual schema")
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when upgrade schema")
	}
	return sqlOperator.UpgradeSchema()
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

func (ctrl *OBClusterCtrl) GetServerParameter() ([]cloudv1.ServerParameter, error) {
	res := make([]cloudv1.ServerParameter, 0)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return res, errors.Wrap(err, "get sql operator error")
	}
	paramsList := []string{observerconst.ServerPermanentOfflineTime, observerconst.EnableRebalance, observerconst.EnableRereplication}
	paramsName := strings.Join(paramsList, ",")
	params := sqlOperator.GetServerParameter(paramsName)
	for _, param := range params {
		serverFromDb := fmt.Sprintf("%s:%d", param.SvrIP, param.SvrPort)
		serverExist := false
		for idx, serverParam := range res {
			if serverFromDb == serverParam.Server {
				serverExist = true
				p := serverParam.Params
				p = append(p, cloudv1.Parameter{Name: param.Name, Value: param.Value})
				serverParam.Params = p
				res[idx] = serverParam
				break
			}
		}
		if !serverExist {
			newParam := []cloudv1.Parameter{{Name: param.Name, Value: param.Value}}
			newServerParam := cloudv1.ServerParameter{Server: serverFromDb, Params: newParam}
			res = append(res, newServerParam)
		}
	}
	return res, nil
}
