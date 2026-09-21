package ac

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"testing"
)

// A local, adapter-free enforcer: tests never write the real policy ConfigMap.
func TestObjectStoragePermissionIsolation(t *testing.T) {
	m, err := model.NewModelFromString(modelDefinition)
	if err != nil {
		t.Fatal(err)
	}
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatal(err)
	}
	e.AddNamedMatchingFunc("g", "keyMatch", keyMatch)
	for _, p := range [][]string{{"admin", "*", "*", "admin"}, {"cluster-writer", "obcluster/*", "write", "cluster only"}, {"storage-reader", "objectstorage/test", "read", "read only"}, {"storage-writer", "objectstorage/test", "write", "namespace scoped"}} {
		if _, err := e.AddPolicy(p); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		user, resource, action string
		allowed                bool
	}{
		{"admin", "objectstorage/test", "write", true},
		{"cluster-writer", "objectstorage/test", "read", false},
		{"cluster-writer", "objectstorage/test", "write", false},
		{"storage-reader", "objectstorage/test", "read", true},
		{"storage-reader", "objectstorage/test", "write", false},
		{"storage-writer", "objectstorage/test", "write", true},
		{"storage-writer", "objectstorage/other", "write", false},
	} {
		ok, err := e.Enforce(tc.user, tc.resource, tc.action)
		if err != nil || ok != tc.allowed {
			t.Errorf("%+v: got %v %v", tc, ok, err)
		}
	}
}
