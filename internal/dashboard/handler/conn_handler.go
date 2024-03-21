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
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/oceanbase/ob-operator/internal/clients"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
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

func onEvictfunc(key string, value *Session) {
	return
}

var sessionMap = expirable.NewLRU(100, onEvictfunc, time.Minute*10)

type Session struct {
	Stdin     *os.File
	Stdout    *os.File
	Stderr    *os.File
	Namespace string `json:"namespace"`
	Cluster   string `json:"cluster,omitempty"`
	Tenant    string `json:"tenant,omitempty"`
	Pod       string `json:"pod"`
	ClientIP  string `json:"clientIp"`
	SessionID string `json:"sessionId"`
	User      string `json:"user"`
	Password  string `json:"password"`
}

type TerminalSizeQueue struct {
	sizes chan *remotecommand.TerminalSize
}

func (q *TerminalSizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-q.sizes
	if !ok {
		return nil
	}
	return size
}

func (q *TerminalSizeQueue) SetSize(size *remotecommand.TerminalSize) {
	select {
	case q.sizes <- size:
	default:
	}
}

func CreateOBClusterConnSession(c *gin.Context) (*Session, error) {
	nn := &param.K8sObjectIdentity{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	obcluster, err := clients.GetOBCluster(c, nn.Namespace, nn.Name)
	if err != nil {
		return nil, err
	}
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

	sess := &Session{}
	sess.ClientIP = c.ClientIP()
	sess.SessionID = rand.String(32)
	sess.Namespace = nn.Namespace
	sess.Cluster = nn.Name
	sess.Pod = pods.Items[0].Name

	sessionMap.Add(sess.SessionID, sess)

	return sess, nil
}

func ConnectDatabase(c *gin.Context) (*Session, error) {
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

	logrus.Infof("websocket connected: %s, session ID: %s", conn.RemoteAddr().String(), req.SessionId)

	config := client.GetClient().GetConfig()

	execRequest := client.GetClient().ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(session.Namespace).
		Name(session.Pod).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command:   []string{"/bin/bash"},
			Container: "observer",
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	sizeQueue := &TerminalSizeQueue{sizes: make(chan *remotecommand.TerminalSize, 1)}
	sizeQueue.SetSize(&remotecommand.TerminalSize{Width: colsNum, Height: rowsNum})

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", execRequest.URL())
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}

	err = exec.StreamWithContext(c, remotecommand.StreamOptions{
		Stdout:            websocketWrapper{conn},
		Stderr:            websocketWrapper{conn},
		Stdin:             websocketWrapper{conn},
		Tty:               true,
		TerminalSizeQueue: sizeQueue,
	})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}

	return session, nil
}

type websocketWrapper struct {
	conn *websocket.Conn
}

var _ io.Reader = websocketWrapper{}
var _ io.Writer = websocketWrapper{}

func (w websocketWrapper) Read(p []byte) (int, error) {
	_, r, err := w.conn.NextReader()
	if err != nil {
		return 0, err
	}
	return r.Read(p)
}

func (w websocketWrapper) Write(p []byte) (int, error) {
	err := w.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
