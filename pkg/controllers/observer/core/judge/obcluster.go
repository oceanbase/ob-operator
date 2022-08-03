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

package judge

import (
	"reflect"
	"strings"

	kubeerrors "k8s.io/apimachinery/pkg/api/errors"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
	"github.com/pkg/errors"
)

func VersionIsModified(version string, statefulApp cloudv1.StatefulApp) (bool, error) {
	var versionCurrent string
	for _, container := range statefulApp.Spec.PodTemplate.Containers {
		if container.Name == observerconst.ImgOb {
			versionCurrent = strings.Split(container.Image, ":")[len(strings.Split(container.Image, ":"))-1]
			break
		}
	}
	if version == versionCurrent {
		return false, nil
	}
	return true, nil
}

func ResourcesIsModified(clusterList []cloudv1.Cluster, obCluster cloudv1.OBCluster, statefulApp cloudv1.StatefulApp) (bool, error) {
	cluster := converter.GetClusterSpecFromOBTopology(clusterList)
	statefulAppNew := converter.GenerateStatefulAppObject(cluster, obCluster)
	podTemplateCompareStatus := reflect.DeepEqual(statefulApp.Spec.PodTemplate, statefulAppNew.Spec.PodTemplate)
	storageTemplatesCompareStatus := reflect.DeepEqual(statefulApp.Spec.StorageTemplates, statefulAppNew.Spec.StorageTemplates)
	if podTemplateCompareStatus && storageTemplatesCompareStatus {
		return false, nil
	}
	return true, nil
}

func ZoneNumberIsModified(clusterList []cloudv1.Cluster, obCluster cloudv1.OBCluster, statefulApp cloudv1.StatefulApp) (string, error) {
	cluster := converter.GetClusterSpecFromOBTopology(clusterList)
	zoneNumberNew := len(cluster.Zone)
	if zoneNumberNew == 0 {
		return observerconst.Maintain, kubeerrors.NewServiceUnavailable("can't scale Zone to zero")
	}

	// zoneNumberCurrent := len(statefulApp.Spec.Subsets)
	podIP := statefulApp.Status.Subsets[0].Pods[0].PodIP
	obZoneList := sql.GetOBZone(podIP)
	zoneNumberCurrent := len(obZoneList)
	if zoneNumberCurrent == 0 {
		return "", errors.New(observerconst.DataBaseError)
	}
	if zoneNumberNew > zoneNumberCurrent {
		return observerconst.ScaleUP, nil
	} else if zoneNumberNew < zoneNumberCurrent {
		return observerconst.ZoneScaleDown, nil
	} else {
		return observerconst.Maintain, nil
	}
}
