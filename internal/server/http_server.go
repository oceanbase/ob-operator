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

package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	serverconst "github.com/oceanbase/ob-operator/internal/const/server"
)

type Authorizer interface {
	Authorize(req *http.Request) error
}

type HTTPServer struct {
	Router     *gin.Engine
	Server     *http.Server
	Authorizer *Authorizer
	ctx        context.Context
}

func (s *HTTPServer) Run() error {
	s.Server.Handler = s.Router
	envPort := os.Getenv("LISTEN_PORT")
	var address string
	if envPort == "" {
		address = fmt.Sprintf("%s:%d", serverconst.DefaultServerHost, serverconst.DefaultServerPort)
	} else {
		address = fmt.Sprintf("%s:%s", serverconst.DefaultServerHost, envPort)
	}
	listener, err := net.Listen(serverconst.DefaultProtocol, address)
	if err != nil {
		return errors.Wrapf(err, "failed to listen address %s", address)
	}
	err = s.Server.Serve(listener)
	if err != nil {
		return errors.Wrap(err, "failed to start server")
	}
	return nil
}

func (s *HTTPServer) Stop() error {
	return s.Server.Shutdown(s.ctx)
}

func NewHTTPServer(ctx context.Context) *HTTPServer {
	if os.Getenv("DEBUG_DASHBOARD") != "true" {
		gin.SetMode(gin.ReleaseMode)
	}
	httpServer := &HTTPServer{
		Router: gin.New(),
		Server: &http.Server{},
		ctx:    ctx,
	}
	return httpServer
}
