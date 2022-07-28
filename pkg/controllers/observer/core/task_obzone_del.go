package core

import (
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/model"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
)

func (ctrl *OBClusterCtrl) GetAllUnit(clusterIP string) []model.AllUnit {
	return sql.GetAllUnit(clusterIP)
}
