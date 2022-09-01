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
	"bytes"
	"encoding/json"
	"fmt"
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
)

type ConfigsJson struct {
	Configs []Configs `json:"configs"`
}
type Configs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (ctrl *OBClusterCtrl) CreateUserForObagent(statefulApp cloudv1.StatefulApp) error {
	subsets := statefulApp.Status.Subsets
	podIp := subsets[0].Pods[0].PodIP
	err := sql.CreateUser(podIp, "ocp_monitor", "root")
	klog.Infoln("CreateUser podIP is :", podIp)
	if err != nil {
		return err
	}
	err = sql.GrantPrivilege(podIp, "select", "*", "ocp_monitor")
	if err != nil {
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) ReviseConfig(podIP string, zoneName string) error {
	obCluster := ctrl.OBCluster
	clusterName := obCluster.Name
	clusterID := fmt.Sprintf("%d", obCluster.Spec.ClusterID)
	config := ConfigsJson{
		[]Configs{
			{Key: "monagent.ob.monitor.user", Value: "ocp_monitor"},
			{Key: "monagent.ob.monitor.password", Value: "root"},
			{Key: "monagent.host.ip", Value: podIP},
			{Key: "monagent.ob.cluster.name", Value: clusterName},
			{Key: "monagent.ob.cluster.id", Value: clusterID},
			{Key: "monagent.ob.zone.name", Value: zoneName},
			{Key: "monagent.pipeline.node.status", Value: "active"}}}
	updateUrl := fmt.Sprintf("http://%s:%d%s", podIP, observerconst.MonagentPort, observerconst.MonagentUpdateUrl)
	body, _ := json.Marshal(config)
	resp, err := http.Post(updateUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		klog.Errorln("update obagent config failed,", podIP, err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			klog.Errorln("Fail to read response:", err)
			return err
		}
		jsonStr := string(body)
		klog.Infoln("Update config Response: ", jsonStr)
	} else {
		//The status is not Created. print the error.
		klog.Errorln("Get failed with error: ", resp.Status)
	}
	return nil
}

func (ctrl *OBClusterCtrl) ReviseAllOBAgentConfig(statefulApp cloudv1.StatefulApp) error {
	subsets := statefulApp.Status.Subsets
	// 获得所有的 obagent
	for subsetsIdx, _ := range subsets {
		zoneName := subsets[subsetsIdx].Name
		for _, pod := range subsets[subsetsIdx].Pods {
			err := ctrl.ReviseConfig(pod.PodIP, zoneName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
