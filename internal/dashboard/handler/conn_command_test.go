/*
Copyright (c) 2026 OceanBase
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
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	kubeclient "github.com/oceanbase/ob-operator/pkg/k8s/client"
)

// The stub consumes the actual generated argv without contacting a database or
// Kubernetes cluster. NUL separators preserve spaces and newlines in values.
func terminalMySQLStub(t *testing.T) {
	t.Helper()
	binDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(binDir, "mysql"), []byte("#!/bin/sh\nprintf '%s\\000' \"$@\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestConnectDatabaseExecutesMySQLDirectly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, user := range []string{"root", "root@tenant"} {
		t.Run(user, func(t *testing.T) {
			term := &response.OBConnection{User: user, Host: "10.0.0.1", Password: "x' & echo injected & echo '"}
			terminalID := "command-test-" + user
			openTermMap.Add(terminalID, term)
			t.Cleanup(func() { openTermMap.Remove(terminalID) })
			t.Setenv("USER_NAMESPACE", "terminal-test")
			t.Setenv("HOSTNAME", "dashboard-test")
			requests := make(chan []string, 1)
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/namespaces/terminal-test/pods/dashboard-test/exec" || r.URL.Query().Get("container") != "dashboard" || r.URL.Query().Get("tty") != "true" {
					t.Error("unexpected terminal exec target or TTY configuration")
				}
				requests <- r.URL.Query()["command"]
				// Stop before SPDY streaming: only the outgoing exec contract is tested.
				http.Error(w, "test exec endpoint", http.StatusBadRequest)
			}))
			defer api.Close()
			clientSet, err := kubernetes.NewForConfig(&rest.Config{Host: api.URL})
			if err != nil {
				t.Fatal(err)
			}
			client := kubeclient.GetClient()
			originalClientSet := client.ClientSet
			client.ClientSet = clientSet
			defer func() { client.ClientSet = originalClientSet }()
			router := gin.New()
			router.GET("/api/v1/terminal/:terminalId", Wrap(ConnectDatabase))
			server := httptest.NewServer(router)
			defer server.Close()
			ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/api/v1/terminal/"+terminalID, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer ws.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			select {
			case command := <-requests:
				want := []string{"mysql", "--user=" + user, "--host=" + term.Host, "--port=2881", "-A", "--password=" + term.Password}
				if !reflect.DeepEqual(command, want) {
					t.Fatal("WebSocket terminal did not send intact mysql argv to pods/exec")
				}
			case <-ctx.Done():
				t.Fatal("terminal did not issue a Kubernetes exec request")
			}
		})
	}
}

func TestDatabaseTerminalCommandDoesNotExecuteCredentials(t *testing.T) {
	terminalMySQLStub(t)
	for _, field := range []string{"password", "password-substitution", "password-backticks", "user", "host"} {
		t.Run(field, func(t *testing.T) {
			marker := filepath.Join(t.TempDir(), "injected")
			term := &response.OBConnection{User: "root", Host: "127.0.0.1", Password: "normal-password"}
			switch field {
			case "password":
				term.Password = fmt.Sprintf("x' & printf injected > '%s' & echo '", marker)
			case "password-substitution":
				term.Password = fmt.Sprintf("x'$(printf injected > '%s')'", marker)
			case "password-backticks":
				term.Password = fmt.Sprintf("x'`printf injected > '%s'`'", marker)
			case "user":
				term.User = fmt.Sprintf("root; printf injected > '%s'; #", marker)
			case "host":
				term.Host = fmt.Sprintf("127.0.0.1; printf injected > '%s'; #", marker)
			}
			command := databaseTerminalCommand(term)
			output, err := exec.Command(command[0], command[1:]...).CombinedOutput()
			if err != nil {
				t.Fatalf("execute mysql stub: %v", err)
			}
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatal("connection credentials executed a shell command")
			}
			want := []string{"--user=" + term.User, "--host=" + term.Host, "--port=2881", "-A", "--password=" + term.Password}
			got := strings.Split(strings.TrimSuffix(string(output), "\x00"), "\x00")
			if !reflect.DeepEqual(got, want) {
				t.Fatal("mysql did not receive connection values as intact arguments")
			}
		})
	}
}

func TestDatabaseTerminalCommandPreservesConnectionValues(t *testing.T) {
	terminalMySQLStub(t)
	cases := []struct {
		name string
		term response.OBConnection
	}{
		{"cluster", response.OBConnection{User: "root", Host: "10.0.0.1", Password: "Normal#123"}},
		{"tenant", response.OBConnection{User: "root@tenant", Host: "10.0.0.2", Password: "Tenant#123"}},
		{"empty-password", response.OBConnection{User: "root@tenant", Host: "10.0.0.2"}},
		{"special-password", response.OBConnection{User: "root", Host: "10.0.0.1", Password: "single' double\" ; & | $() `quote` \\ space\n中文"}},
		{"option-like-values", response.OBConnection{User: "--help", Host: "--init-command=evil", Password: "--execute=evil"}},
		{"ipv6-host", response.OBConnection{User: "root", Host: "2001:db8::1", Password: "password"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			command := databaseTerminalCommand(&tc.term)
			output, err := exec.Command(command[0], command[1:]...).CombinedOutput()
			if err != nil {
				t.Fatalf("execute mysql stub: %v", err)
			}
			want := []string{"--user=" + tc.term.User, "--host=" + tc.term.Host, "--port=2881", "-A"}
			if tc.term.Password != "" {
				want = append(want, "--password="+tc.term.Password)
			}
			got := strings.Split(strings.TrimSuffix(string(output), "\x00"), "\x00")
			if !reflect.DeepEqual(got, want) {
				t.Fatal("mysql did not receive the original user, host and password intact")
			}
		})
	}
}
