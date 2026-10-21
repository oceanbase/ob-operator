package lakehouse

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestReviewCheckCannotExportBootstrapCredential(t *testing.T) {
	if _, err := parseOptions([]string{"--check", "--show-access"}, &bytes.Buffer{}); err == nil {
		t.Fatal("--check and --show-access must be mutually exclusive: access writes the plaintext bootstrap credential")
	}
}

func TestReviewMalformedValuesDoNotExposeCredentialContents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "values.yaml")
	if err := os.WriteFile(path, []byte("dashboard:\n  adminPassword: !!int REVIEW_CREDENTIAL_SENTINEL\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := readValues(path)
	if err == nil {
		t.Fatal("malformed values were accepted")
	}
	if strings.Contains(err.Error(), "REVIEW_CREDENTIAL_SENTINEL") {
		t.Fatal("YAML decoder error exposed secret-bearing file contents")
	}
}

func TestReviewJSONReadsSeparateStderrWarnings(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "kubectl"), []byte("#!/bin/sh\necho 'Warning: v1 Endpoints is deprecated' >&2\nprintf '{\"items\":[]}'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	var got struct {
		Items []interface{} `json:"items"`
	}
	if err := getJSON(context.Background(), execute, &got, "get", "endpoints", "-o", "json"); err != nil {
		t.Fatalf("valid JSON on stdout must remain readable when kubectl writes a warning on stderr: %v", err)
	}
}

func TestReviewAccessVerifiesActualServiceAndProtectsBootstrap(t *testing.T) {
	for _, mode := range []string{"NodePort", "ClusterIP", "http-failure", "redirect"} {
		t.Run(mode, func(t *testing.T) {
			var requests []string
			var requestMu sync.Mutex
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestMu.Lock()
				requests = append(requests, r.URL.Path)
				requestMu.Unlock()
				if mode == "http-failure" {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				if mode == "redirect" {
					w.Header().Set("Location", "http://external.invalid/")
					w.WriteHeader(http.StatusFound)
					return
				}
				if r.URL.Path == "/api/v1/info" {
					fmt.Fprint(w, `{"successful":true,"data":{"publicKey":"example"}}`)
				} else {
					fmt.Fprint(w, "<html></html>")
				}
			}))
			defer server.Close()
			_, portStr, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
			if err != nil {
				t.Fatal(err)
			}
			port, _ := strconv.Atoi(portStr)
			typeName := "NodePort"
			if mode == "ClusterIP" {
				typeName = mode
			}
			secretReads := 0
			run := func(ctx context.Context, name string, args ...string) ([]byte, error) {
				if _, ok := ctx.Deadline(); !ok {
					t.Fatal("unbounded read")
				}
				command := name + " " + strings.Join(args, " ")
				if name != "kubectl" || !strings.Contains(command, " get ") {
					t.Fatalf("unexpected mutation: %s", command)
				}
				if strings.Contains(command, " get service ") {
					return json.Marshal(map[string]interface{}{"spec": map[string]interface{}{"type": typeName, "clusterIP": "127.0.0.1", "ports": []interface{}{map[string]interface{}{"port": port, "nodePort": port}}}})
				}
				if strings.Contains(command, " get secret demo-dashboard-bootstrap ") {
					secretReads++
					return json.Marshal(map[string]interface{}{"data": map[string]string{"password": base64.StdEncoding.EncodeToString([]byte("REVIEW_BOOTSTRAP_SENTINEL"))}})
				}
				t.Fatalf("unexpected read: %s", command)
				return nil, nil
			}
			state := filepath.Join(t.TempDir(), "private-state")
			var out bytes.Buffer
			err = showAccess(context.Background(), run, options{release: "demo", namespace: "demo-ns", host: "192.0.2.99"}, PreflightResult{NodeInternalIP: "127.0.0.1"}, state, &out)
			if mode == "http-failure" || mode == "redirect" {
				if err == nil || secretReads != 0 {
					t.Fatal("failed HTTP readiness must not claim success or export credentials")
				}
				if _, err := os.Stat(state); !os.IsNotExist(err) {
					t.Fatal("failed readiness created local state")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			requestMu.Lock()
			requestPaths := append([]string(nil), requests...)
			requestMu.Unlock()
			if strings.Join(requestPaths, ",") != "/,/api/v1/info" {
				t.Fatalf("missing HTTP gates: %v", requestPaths)
			}
			if strings.Contains(out.String(), "REVIEW_BOOTSTRAP_SENTINEL") {
				t.Fatal("credential leaked to terminal")
			}
			if mode == "NodePort" && !strings.Contains(out.String(), "http://192.0.2.99:"+portStr+"/#/login") {
				t.Fatal("advertised host override was ignored")
			}
			if mode == "ClusterIP" && !strings.Contains(out.String(), "仅集群网络内可达") {
				t.Fatal("ClusterIP reachability not explained")
			}
			password, err := os.ReadFile(filepath.Join(state, "dashboard-initial-password"))
			if err != nil || string(password) != "REVIEW_BOOTSTRAP_SENTINEL" {
				t.Fatal("password file is missing or incorrect")
			}
			for _, name := range []string{"dashboard-initial-password", "ACCESS.txt"} {
				info, err := os.Stat(filepath.Join(state, name))
				if err != nil || info.Mode().Perm() != 0600 {
					t.Fatal("credential/access files must be private")
				}
			}
		})
	}
}
