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
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/golang-lru/v2/expirable"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/oceanbase/ob-operator/internal/clients"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/k8s"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	"github.com/oceanbase/ob-operator/internal/dashboard/utils"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func onEvictFunc(key string, _ *response.OBConnection) {
	log.Printf("Terminal connection expired: %s\n", key)
}

var openTermMap = expirable.NewLRU(100, onEvictFunc, time.Minute*10)

// wsWrapper wraps a websocket connection to implement the io.ReadWriteCloser interface.
type wsWrapper struct {
	conn   *websocket.Conn
	cancel context.CancelFunc
}

func newWsWrapper(conn *websocket.Conn, cancel context.CancelFunc) wsWrapper {
	return wsWrapper{
		conn:   conn,
		cancel: cancel,
	}
}

func (w wsWrapper) Read(p []byte) (int, error) {
	_, r, err := w.conn.NextReader()
	if err != nil {
		log.Println("read message error: ", err)
		w.cancel()
		return 0, err
	}
	_ = w.conn.SetReadDeadline(time.Now().Add(10 * time.Minute))
	return r.Read(p)
}

func (w wsWrapper) Write(p []byte) (int, error) {
	err := w.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		log.Println("write message error: ", err)
		w.cancel()
		return 0, err
	}
	_ = w.conn.SetWriteDeadline(time.Now().Add(10 * time.Minute))
	return len(p), nil
}

func (w wsWrapper) Close() error {
	w.cancel()
	return w.conn.Close()
}

// @ID CreateOBClusterConnection
// @Summary Create oceanbase cluster connection
// @Description Create oceanbase cluster connection terminal
// @Tags Terminal
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.OBConnection}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name}/terminal [PUT]
// @Param namespace path string true "namespace"
// @Param name path string true "name"
// @Param channel query string false "channel" Enums(TERMINAL, ODC)
// @Security ApiKeyAuth
func CreateOBClusterConnTerminal(c *gin.Context) (*response.OBConnection, error) {
	nn := &param.K8sObjectIdentity{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	channel := c.Query("channel")
	if channel == "" {
		channel = "TERMINAL"
	}
	if OdcURL == "" && channel == "ODC" {
		return nil, httpErr.NewBadRequest("odc not enabled")
	}
	obcluster, err := clients.GetOBCluster(c, nn.Namespace, nn.Name)
	if err != nil {
		return nil, err
	}

	secret, err := client.GetClient().ClientSet.CoreV1().Secrets(nn.Namespace).Get(c, obcluster.Spec.UserSecrets.Root, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	passwd := string(secret.Data["password"])

	pods, err := client.GetClient().ClientSet.CoreV1().Pods(nn.Namespace).List(c, metav1.ListOptions{
		LabelSelector: oceanbaseconst.LabelRefOBCluster + "=" + obcluster.Name,
		FieldSelector: "status.phase=Running",
	})
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, httpErr.NewBadRequest("no running pods found in obcluster")
	}

	conn := &response.OBConnection{}
	conn.Namespace = nn.Namespace
	conn.Cluster = nn.Name
	conn.Pod = pods.Items[0].Name
	conn.Host = pods.Items[0].Status.PodIP
	conn.ClientIP = c.ClientIP()
	conn.Password = passwd
	conn.User = "root"

	if channel == "TERMINAL" {
		conn.TerminalID = rand.String(32)
		openTermMap.Add(conn.TerminalID, conn)
	} else if channel == "ODC" {
		param, err := generateOdcParam(c, conn.Host, conn.User, passwd)
		if err != nil {
			return nil, err
		}
		visitUrl, err := url.JoinPath(OdcURL, "#", "gateway", param)
		if err != nil {
			return nil, httpErr.NewInternal(err.Error())
		}
		conn.OdcVisitURL = strings.ReplaceAll(visitUrl, "%23", "#")
	}

	return conn, nil
}

// @ID CreateOBTenantConnection
// @Summary Create oceanbase tenant connection
// @Description Create oceanbase tenant connection terminal
// @Tags Terminal
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.OBConnection}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenants/{namespace}/{name}/terminal [PUT]
// @Param namespace path string true "namespace"
// @Param name path string true "name"
// @Param channel query string false "channel" Enums(TERMINAL, ODC)
// @Security ApiKeyAuth
func CreateOBTenantConnTerminal(c *gin.Context) (*response.OBConnection, error) {
	nn := &param.K8sObjectIdentity{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	channel := c.Query("channel")
	if channel == "" {
		channel = "TERMINAL"
	}
	if OdcURL == "" && channel == "ODC" {
		return nil, httpErr.NewBadRequest("odc not enabled")
	}
	obtenant, err := clients.GetOBTenant(c, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewBadRequest("obtenant not found")
		}
		return nil, httpErr.NewInternal(err.Error())
	}

	var passwd string

	if obtenant.Spec.Credentials.Root != "" {
		secret, err := client.GetClient().
			ClientSet.
			CoreV1().
			Secrets(nn.Namespace).
			Get(c, obtenant.Spec.Credentials.Root, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		passwd = string(secret.Data["password"])
	}
	conn := &response.OBConnection{}
	obcluster, err := clients.GetOBCluster(c, nn.Namespace, obtenant.Spec.ClusterName)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}

	// Select unit information from the oceanbase cluster
	db, err := utils.GetOBConnection(c, obcluster, oceanbaseconst.RootUser, oceanbaseconst.SysTenant, obcluster.Spec.UserSecrets.Root)
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	units, err := db.ListUnitsWithTenantId(c, int64(obtenant.Status.TenantRecordInfo.TenantID))
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	if len(units) == 0 {
		return nil, httpErr.NewInternal("no unit found in obtenant " + obtenant.Name)
	}
	conn.Host = units[0].SvrIp

	if conn.Host == "" {
		return nil, httpErr.NewBadRequest("no full replica observer found in obtenant")
	}

	conn.ClientIP = c.ClientIP()
	conn.Namespace = nn.Namespace
	conn.Cluster = nn.Name
	conn.User = "root@" + obtenant.Spec.TenantName
	conn.Password = passwd

	if channel == "TERMINAL" {
		conn.TerminalID = rand.String(32)
		openTermMap.Add(conn.TerminalID, conn)
	} else if channel == "ODC" {
		param, err := generateOdcParam(c, conn.Host, conn.User, passwd)
		if err != nil {
			return nil, err
		}
		visitUrl, err := url.JoinPath(OdcURL, "#", "gateway", param)
		if err != nil {
			return nil, httpErr.NewInternal(err.Error())
		}
		conn.OdcVisitURL = strings.ReplaceAll(visitUrl, "%23", "#")
	}

	return conn, nil
}

// @ID ConnectDatabase
// @Summary Connect to oceanbase database
// @Description Connect to oceanbase database in websocket
// @Tags Terminal
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.OBConnection}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/conn/{terminalId} [GET]
// @Param terminalId path string true "terminalId"
// @Security ApiKeyAuth
func ConnectDatabase(c *gin.Context) (*response.OBConnection, error) {
	type ConnectReq struct {
		TerminalId string `json:"terminalId" uri:"terminalId" binding:"required"`
	}
	req := &ConnectReq{}
	err := c.BindUri(req)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	term, ok := openTermMap.Get(req.TerminalId)
	if !ok {
		return nil, httpErr.NewBadRequest("terminal not found")
	}
	var colsNum uint16 = 160
	var rowsNum uint16 = 60

	cols := c.Query("cols")
	rows := c.Query("rows")

	if i, err := strconv.ParseInt(cols, 10, 16); err == nil && i >= 0 && i <= math.MaxUint16 {
		colsNum = uint16(i)
	}
	if i, err := strconv.ParseInt(rows, 10, 16); err == nil && i >= 0 && i <= math.MaxUint16 {
		rowsNum = uint16(i)
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	defer conn.Close()

	log.Printf("websocket connected: %s, terminal ID: %s", conn.RemoteAddr().String(), req.TerminalId)

	ctx, cancel := context.WithCancel(c)
	defer cancel()

	wsOut := newWsWrapper(conn, cancel)
	wsIn := newWsWrapper(conn, cancel)

	cmdArgs := []string{"/bin/bash", "-c", fmt.Sprintf("mysql -u%s -h%s -P2881 -A", term.User, term.Host)}
	if term.Password != "" {
		cmdArgs[2] += fmt.Sprintf(" -p'%s'", term.Password)
	}
	sizeQueue := k8s.NewResizeQueue()
	sizeQueue.SetSize(colsNum, rowsNum)

	execReq := &k8s.KubeExecRequest{
		Namespace:   os.Getenv("USER_NAMESPACE"),
		PodName:     os.Getenv("HOSTNAME"),
		Container:   "dashboard",
		Command:     cmdArgs,
		Stdin:       wsIn,
		Stdout:      wsOut,
		Stderr:      wsOut,
		TTY:         true,
		ResizeQueue: sizeQueue,
	}

	if err := k8s.KubeExec(ctx, execReq); err != nil {
		log.Printf("kube exec error: %v\n", err)
		_, _ = wsOut.Write([]byte(err.Error()))
		return nil, httpErr.NewInternal(err.Error())
	}

	log.Printf("websocket disconnected: %s, terminal ID: %s\n", conn.RemoteAddr().String(), req.TerminalId)
	return nil, nil
}
