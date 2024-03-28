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

package k8s

import (
	"context"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

type KubeExecRequest struct {
	Namespace string        `json:"namespace"`
	PodName   string        `json:"podName"`
	Container string        `json:"container"`
	Command   []string      `json:"command"`
	Stdin     io.ReadWriter `json:"stdin"`
	Stdout    io.ReadWriter `json:"stdout"`
	Stderr    io.ReadWriter `json:"stderr"`
	TTY       bool          `json:"tty"`

	ResizeQueue remotecommand.TerminalSizeQueue `json:"resizeQueue"`
}

func KubeExec(ctx context.Context, req *KubeExecRequest) error {
	config := client.GetClient().GetConfig()
	execRequest := client.GetClient().ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(req.Namespace).
		Name(req.PodName).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command:   req.Command,
			Container: req.Container,
			Stdin:     req.Stdin != nil,
			Stdout:    req.Stdout != nil,
			Stderr:    req.Stderr != nil,
			TTY:       req.TTY,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", execRequest.URL())
	if err != nil {
		return httpErr.NewInternal(err.Error())
	}
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             req.Stdin,
		Stdout:            req.Stdout,
		Stderr:            req.Stderr,
		Tty:               req.TTY,
		TerminalSizeQueue: req.ResizeQueue,
	})
	if err != nil && strings.Contains(err.Error(), "context canceled") {
		return nil
	}
	return err
}

type ResizeQueue interface {
	remotecommand.TerminalSizeQueue
	SetSize(w, h uint16)
}

func NewResizeQueue() ResizeQueue {
	return &resizeQueue{
		sizes: make(chan *remotecommand.TerminalSize, 1),
	}
}

type resizeQueue struct {
	sizes chan *remotecommand.TerminalSize
}

func (q *resizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-q.sizes
	if !ok {
		return nil
	}
	return size
}

func (q *resizeQueue) SetSize(w, h uint16) {
	size := &remotecommand.TerminalSize{
		Width:  w,
		Height: h,
	}
	select {
	case q.sizes <- size:
	default:
	}
}
