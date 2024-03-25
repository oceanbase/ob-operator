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
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/golang-lru/v2/expirable"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

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
	log.Printf("session expired: %s\n", key)
}

var sessionMap = expirable.NewLRU(100, onEvictFunc, time.Minute*10)

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
// @Router /api/v1/obclusters/namespace/{namespace}/name/{name}/session [PUT]
// @Security ApiKeyAuth
func CreateOBClusterConnSession(c *gin.Context) (*response.OBConnection, error) {
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

	sess := &response.OBConnection{}
	sess.ClientIP = c.ClientIP()
	sess.SessionID = rand.String(32)
	sess.Namespace = nn.Namespace
	sess.Cluster = nn.Name
	sess.Pod = pods.Items[0].Name
	sess.User = "root"
	sess.Password = passwd

	sessionMap.Add(sess.SessionID, sess)

	return sess, nil
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
// @Router /api/v1/obclusters/conn/{sessionId} [GET]
// @Security ApiKeyAuth
func ConnectDatabase(c *gin.Context) (*response.OBConnection, error) {
	type ConnectReq struct {
		SessionId string `json:"sessionId" uri:"sessionId" binding:"required"`
	}
	req := &ConnectReq{}
	err := c.BindUri(req)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	session, ok := sessionMap.Get(req.SessionId)
	if !ok {
		return nil, httpErr.NewBadRequest("session not found")
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

	log.Printf("websocket connected: %s, session ID: %s", conn.RemoteAddr().String(), req.SessionId)

	var cmdPwdPart string
	if session.Password == "" {
		cmdPwdPart = ""
	} else {
		cmdPwdPart = " -p" + session.Password
	}

	ctx, cancel := context.WithCancel(c)
	defer cancel()

	wsOut := newWsWrapper(conn, cancel)
	wsIn := newWsWrapper(conn, cancel)

	sizeQueue := k8s.NewResizeQueue()
	sizeQueue.SetSize(colsNum, rowsNum)

	execReq := &k8s.KubeExecRequest{
		// Namespace: "default",
		// PodName:   "nginx-app-7f6fdf9556-282sw",
		// Container: "nginx",
		// Command:   []string{"/bin/bash"},
		Namespace:   session.Namespace,
		PodName:     session.Pod,
		Container:   "observer",
		Command:     []string{"/bin/bash", "-c", "yum install -y mysql && mysql -uroot -h127.0.0.1 -P2881" + cmdPwdPart},
		Stdin:       wsIn,
		Stdout:      wsOut,
		Stderr:      wsOut,
		TTY:         true,
		ResizeQueue: sizeQueue,
	}

	if err := k8s.KubeExec(ctx, execReq); err != nil {
		log.Printf("kube exec error: %v\n", err)
		return nil, httpErr.NewInternal(err.Error())
	}

	log.Printf("websocket disconnected: %s, session ID: %s\n", conn.RemoteAddr().String(), req.SessionId)
	return nil, nil
}
