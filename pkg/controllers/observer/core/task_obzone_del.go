package core

import (
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/model"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

func (ctrl *OBClusterCtrl) GetAllUnit(clusterIP string) []model.AllUnit {
	return sql.GetAllUnit(clusterIP)
}

func (ctrl *OBClusterCtrl) StopZone(clusterIP, zoneName string) error {
	klog.Infoln("begin stop OBZone", zoneName)

	// stop zone
	err := sql.StopZone(clusterIP, zoneName)
	if err != nil {
		klog.Errorln("stop zone error", zoneName, clusterIP)
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) DeleteZone(clusterIP, zoneName string) error {
	klog.Infoln("begin delete OBZone", zoneName)

	// delete zone
	err := sql.DeleteZone(clusterIP, zoneName)
	if err != nil {
		klog.Errorln("Delete zone: error zoneName, clusterIP ", zoneName, clusterIP)
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) DeleteOBZone(clusterIP, zoneName string) error {
	// check unitList is empty
	AllUnitList := ctrl.GetAllUnit(clusterIP)
	for _, unit := range AllUnitList {
		if unit.Zone == zoneName {
			return errors.New("When unitList is not empty, delete OBZone is not supported yet")
		}
	}
	ctrl.DeleteOBZoneExecuter(clusterIP, zoneName)
	return nil
}

func (ctrl *OBClusterCtrl) DeleteOBZoneExecuter(clusterIP, zoneName string) {
	klog.Infoln("Execute Delete OB Zone... ")

	err := ctrl.StopZone(clusterIP, zoneName)
	if err != nil {
		klog.Infoln("StopOBZone err: ", err)
	}

	// get observerList
	obServerList := sql.GetOBServer(clusterIP)
	for _, observer := range obServerList {
		if observer.Zone == zoneName && observer.Status != observerconst.OBServerDeleting {
			klog.Infoln("observerList is not nil, begin delete observer...")
			ctrl.DelOBServerExecuter(clusterIP, zoneName, observer.SvrIP)
		}
	}

	//delete ob zone
	err = ctrl.DeleteZone(clusterIP, zoneName)
	if err != nil {
		klog.Infoln("DeleteOBZone err: ", err)
	}

}
