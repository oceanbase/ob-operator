/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package handler

import (
	"encoding/base64"
	"encoding/json"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"

	"github.com/oceanbase/ob-operator/internal/dashboard/business/auth"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	crypto "github.com/oceanbase/ob-operator/pkg/crypto"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
)

func loggingCreateOBClusterParam(param *param.CreateOBClusterParam) {
	logger.
		WithField("Name", param.Name).
		WithField("Namespace", param.Namespace).
		WithField("ClusterName", param.ClusterName).
		WithField("ClusterId", param.ClusterId).
		WithField("Mode", param.Mode).
		WithField("Topology", param.Topology).
		WithField("OBServer", param.OBServer).
		WithField("Monitor", param.Monitor).
		WithField("Parameters", param.Parameters).
		Infof("Create OBCluster param")
}

func loggingCreateOBTenantParam(param *param.CreateOBTenantParam) {
	logger.
		WithField("Name", param.Name).
		WithField("Namespace", param.Namespace).
		WithField("ClusterName", param.ClusterName).
		WithField("TenantName", param.TenantName).
		WithField("UnitNumber", param.UnitNumber).
		WithField("ConnectWhiteList", param.ConnectWhiteList).
		WithField("Charset", param.Charset).
		WithField("UnitConfig", param.UnitConfig).
		WithField("Pools", param.Pools).
		WithField("TenantRole", param.TenantRole).
		WithField("Source", param.Source).
		Infof("Create OBTenant param")
}

type OdcBastionData struct {
	AccountVerifyToken string `json:"accountVerifyToken"`
	Type               string `json:"type"`
	UnionDbUser        string `json:"unionDbUser"`
	Host               string `json:"host"`
	Password           string `json:"password"`
	Port               int    `json:"port"`
}

type OdcBastion struct {
	Action  string         `json:"action"`
	Data    OdcBastionData `json:"data"`
	Encrypt bool           `json:"encrypt"`
}

func generateOdcParam(c *gin.Context, host, user, passwd string) (string, error) {
	session := sessions.Default(c)
	var username string
	if session.Get("username") == nil {
		username = "user"
	} else {
		username = session.Get("username").(string)
	}

	token := auth.GenerateAuthToken(&auth.AuthUser{
		Username: username,
		Nickname: username,
	})

	odcBastionData := OdcBastionData{
		AccountVerifyToken: token.String(),
		Type:               "OB_MYSQL",
		UnionDbUser:        user,
		Host:               host,
		Password:           passwd,
		Port:               2881,
	}

	odcBastion := OdcBastion{
		Action: "newTempSession",
		Data:   odcBastionData,
	}
	jsonData, err := json.Marshal(odcBastion)
	if err != nil {
		return "", httpErr.NewInternal(err.Error())
	}
	paramStr := base64.StdEncoding.EncodeToString(jsonData)
	return paramStr, nil
}

func extractPassword(param *param.CreateOBClusterParam) error {
	var err error
	param.RootPassword, err = crypto.DecryptWithPrivateKey(param.RootPassword)
	if err != nil {
		return err
	}
	param.ProxyroPassword, err = crypto.DecryptWithPrivateKey(param.ProxyroPassword)
	return err
}
