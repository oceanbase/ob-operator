package lakehouse

import (
	"context"
	"reflect"
	"testing"
)

func TestThreeNodesInheritPoolWithoutMutatingInput(t *testing.T) {
	fixture := newThreeNodeRunFixture(t)
	values := mergeValues(fixture.defaults, nil)
	setValue(values, "lakehouse", "scheduling", "nodeSelector", "example.com/pool")
	before := mergeValues(values, nil)
	got, err := configureThreeNodes(context.Background(), execute, values, fixture.defaults, nil, []string{"worker-1", "worker-2", "worker-3"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(values, before) {
		t.Fatal("topology generation mutated its input (retained values must be immutable)")
	}
	if pfGet(got, "dashboard", "nodeSelector", "example.com/pool") != "lakehouse" {
		t.Fatal("Dashboard lost the inherited pool restriction")
	}
	if pfGet(got, "dashboard", "nodeSelector", "kubernetes.io/hostname") != "host-1" {
		t.Fatal("Dashboard did not use the first node's actual hostname")
	}
	setValue(values, "ui", "dashboard", "nodeSelector", "example.com/pool")
	got, err = configureThreeNodes(context.Background(), execute, values, fixture.defaults, nil, []string{"worker-1", "worker-2", "worker-3"}, false)
	if err != nil || pfGet(got, "dashboard", "nodeSelector", "example.com/pool") != "ui" {
		t.Fatalf("explicit Dashboard pool was not retained for preflight validation: %v", err)
	}
}
