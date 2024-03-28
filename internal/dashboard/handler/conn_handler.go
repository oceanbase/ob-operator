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
	"net/http"
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

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/clients"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/k8s"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
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
// @Security ApiKeyAuth
func CreateOBClusterConnTerminal(c *gin.Context) (*response.OBConnection, error) {
	nn := &param.K8sObjectIdentity{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
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

	term := &response.OBConnection{}
	term.ClientIP = c.ClientIP()
	term.TerminalID = rand.String(32)
	term.Namespace = nn.Namespace
	term.Cluster = nn.Name
	term.Pod = pods.Items[0].Name
	term.Host = pods.Items[0].Status.PodIP
	term.User = "root"
	term.Password = passwd

	openTermMap.Add(term.TerminalID, term)

	return term, nil
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
// @Security ApiKeyAuth
func CreateOBTenantConnTerminal(c *gin.Context) (*response.OBConnection, error) {
	nn := &param.K8sObjectIdentity{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
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

	obzoneList := &v1alpha1.OBZoneList{}
	err = clients.ZoneClient.List(c, nn.Namespace, obzoneList, metav1.ListOptions{
		LabelSelector: oceanbaseconst.LabelRefOBCluster + "=" + obtenant.Spec.ClusterName,
	})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}

	if len(obzoneList.Items) == 0 {
		return nil, httpErr.NewBadRequest("no obzone found in obcluster " + obtenant.Spec.ClusterName)
	}

	// get full replica observer
	fullReplicaMap := make(map[string]struct{})
	for _, pool := range obtenant.Spec.Pools {
		if pool.Type != nil && strings.EqualFold(pool.Type.Name, "Full") {
			fullReplicaMap[pool.Zone] = struct{}{}
		}
	}
	conn := &response.OBConnection{}
	for _, zone := range obzoneList.Items {
		if _, ok := fullReplicaMap[zone.Spec.Topology.Zone]; ok {
			if len(zone.Status.OBServerStatus) > 0 {
				conn.Host = zone.Status.OBServerStatus[0].Server
				break
			}
		}
	}

	if conn.Host == "" {
		return nil, httpErr.NewBadRequest("no full replica observer found in obtenant")
	}

	conn.ClientIP = c.ClientIP()
	conn.TerminalID = rand.String(32)
	conn.Namespace = nn.Namespace
	conn.Cluster = nn.Name
	conn.User = "root@" + obtenant.Spec.TenantName
	conn.Password = passwd

	openTermMap.Add(conn.TerminalID, conn)

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

	if i, err := strconv.ParseInt(cols, 10, 64); err == nil {
		colsNum = uint16(i)
	}
	if i, err := strconv.ParseInt(rows, 10, 64); err == nil {
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
		cmdArgs[2] = cmdArgs[2] + fmt.Sprintf(" -p'%s'", term.Password)
	}
	sizeQueue := k8s.NewResizeQueue()
	sizeQueue.SetSize(colsNum, rowsNum)

	execReq := &k8s.KubeExecRequest{
		// Namespace: "default",
		// PodName:   "nginx-app-7f6fdf9556-282sw",
		// Container: "nginx",
		// Command:   []string{"/bin/bash"},
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
