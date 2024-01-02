package server

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/router"
	"github.com/oceanbase/oceanbase-dashboard/internal/server/constant"
	"github.com/pkg/errors"
)

type Authorizer interface {
	Authorize(req *http.Request) error
}

type HTTPServer struct {
	Router     *gin.Engine
	Server     *http.Server
	Authorizer *Authorizer
}

func (s *HTTPServer) Run() error {
	s.Server.Handler = s.Router
	// address := fmt.Sprintf("%s:%d", constant.DefaultServerAddress, constant.DefaultServerPort)
	envPort := os.Getenv("LISTEN_PORT")
	var address string
	if envPort == "" {
		address = fmt.Sprintf("%s:%d", constant.DefaultServerHost, constant.DefaultServerPort)
	} else {
		address = fmt.Sprintf("%s:%s", constant.DefaultServerHost, envPort)
	}
	listener, err := net.Listen(constant.DefaultProtocol, address)
	if err != nil {
		return errors.Wrapf(err, "failed to listen address %s", address)
	}
	err = s.Server.Serve(listener)
	if err != nil {
		return errors.Wrap(err, "failed to start server")
	}
	return nil
}

func (s *HTTPServer) RegisterRouter() error {
	router.InitRoutes(s.Router)
	return nil
}

func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		Router: gin.New(),
		Server: &http.Server{},
	}
}
