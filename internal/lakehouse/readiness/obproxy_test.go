package readiness

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProxySeedsUseAdvertisedServicesNotLoadBalancer(t *testing.T) {
	r := &fakeReader{values: map[string]string{
		"/api/v1/namespaces/qa/services?labelSelector=ref-obcluster%3Dob-ob%2Cref-observer": `{"items":[
  {"metadata":{"labels":{"ref-obcluster":"ob-ob","ref-observer":"a"}},"spec":{"clusterIP":"10.0.0.3","ports":[{"name":"sql","port":2881}]}},
  {"metadata":{"labels":{"ref-obcluster":"ob-ob","ref-observer":"b"}},"spec":{"clusterIP":"10.0.0.2","ports":[{"name":"sql","port":2881}]}},
  {"metadata":{"labels":{"ref-obcluster":"ob-ob"}},"spec":{"clusterIP":"10.0.0.1","ports":[{"name":"sql","port":2881}]}},
  {"metadata":{"labels":{"ref-obcluster":"other","ref-observer":"c"}},"spec":{"clusterIP":"10.0.0.4","ports":[{"name":"sql","port":2881}]}}
  ]}`}}
	o := options()
	o.ObserverReplicas = 2
	p := filepath.Join(t.TempDir(), "rslist")
	if err := WriteProxySeeds(context.Background(), r, o, p); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(p)
	if err != nil || string(data) != "10.0.0.2:2881;10.0.0.3:2881\n" {
		t.Fatalf("%q %v", data, err)
	}
}
func TestProxySeedsAuthorizationAndTimeout(t *testing.T) {
	for _, code := range []int{403, 404} {
		r := &fakeReader{err: &APIError{Status: code}}
		o := options()
		o.Timeout = 10 * time.Millisecond
		p := filepath.Join(t.TempDir(), "rslist")
		err := WriteProxySeeds(context.Background(), r, o, p)
		if err == nil {
			t.Fatal("must fail without seeds")
		}
		if code == 403 && !strings.Contains(err.Error(), "authorization") {
			t.Fatal(err)
		}
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatal("must not write invalid seed file")
		}
	}
}
