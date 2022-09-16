package core

import (
	v1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/model"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

func (ctrl *OBClusterCtrl) GetAllUnit(clusterIP string) []model.AllUnit {
    res := make([]model.AllUnit, 0)
    sqlOperator, err := ctrl.GetSqlOperator()
    if err == nil {
	    res = sqlOperator.GetAllUnit()
    }
    return res
}

func (ctrl *OBClusterCtrl) StopZone(clusterIP, zoneName string) error {
	klog.Infoln("begin stop OBZone", zoneName)

    sqlOperator, err := ctrl.GetSqlOperator()
    if err == nil {
        return errors.Wrap(err, "get sql operator when stop zone")
    }

	// stop zone
	err = sqlOperator.StopZone(zoneName)
	if err != nil {
		klog.Errorln("stop zone error", zoneName, clusterIP)
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) DeleteZone(clusterIP, zoneName string) error {
	klog.Infoln("begin delete OBZone", zoneName)

    sqlOperator, err := ctrl.GetSqlOperator()
    if err == nil {
        return errors.Wrap(err, "get sql operator when delete zone")
    }

	// delete zone
	err = sqlOperator.DeleteZone(zoneName)
	if err != nil {
		klog.Errorln("Delete zone: error zoneName, clusterIP ", zoneName, clusterIP)
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) DeleteOBZone(clusterIP, zoneName string, statefulApp v1.StatefulApp) error {
	// check unitList is empty
	AllUnitList := ctrl.GetAllUnit(clusterIP)
	for _, unit := range AllUnitList {
		if unit.Zone == zoneName {
			return errors.New("When unitList is not empty, delete OBZone is not supported yet")
		}
	}
	return ctrl.DeleteOBZoneExecuter(clusterIP, zoneName, statefulApp)
}

func (ctrl *OBClusterCtrl) DeleteOBZoneExecuter(clusterIP, zoneName string, statefulApp v1.StatefulApp) error {
	klog.Infoln("Execute Delete OB Zone... ")

	err := ctrl.StopZone(clusterIP, zoneName)
	if err != nil {
		klog.Infoln("StopOBZone err: ", err)
		return err
	}

    sqlOperator, err := ctrl.GetSqlOperator()
    if err == nil {
        return errors.Wrap(err, "get sql operator in delete zone executer")
    }


	// get observerList
	obServerList := sqlOperator.GetOBServer()
	for _, observer := range obServerList {
		if observer.Zone == zoneName && observer.Status != observerconst.OBServerDeleting {
			klog.Infoln("observerList is not nil, begin delete observer...")
			err = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ScaleDown, "", observerconst.OBServerDel)
			if err != nil {
				return err
			}
			return errors.New("observerList is not nil, begin delete observer...")
		}
	}

	//delete ob zone
	err = ctrl.DeleteZone(clusterIP, zoneName)
	if err != nil {
		klog.Infoln("DeleteOBZone err: ", err)
		return err
	}

	return nil

}
